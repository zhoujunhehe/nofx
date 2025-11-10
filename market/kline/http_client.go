// file: kline/kline_http_client.go
package kline

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"
)

// ====== Constants and Default Configuration ======

const (
	// Header names (Binance returns one of these)
	HTTPHeaderRetryAfter         = "Retry-After"
	HTTPHeaderUsedWeight         = "X-MBX-USED-WEIGHT"
	HTTPHeaderUsedWeightPrefix   = "X-MBX-USED-WEIGHT-"
	HTTPHeaderUsedWeight1mUpper  = "X-MBX-USED-WEIGHT-1M"
	HTTPHeaderUsedWeight1mLower  = "X-MBX-USED-WEIGHT-1m"
	defaultWeight1mLimit         = 6000 // Adjustable based on account/IP current limit
	defaultBackoffBase           = 150 * time.Millisecond
	defaultBackoffMax            = 5 * time.Second
	defaultHeaderBackoffRatio    = 0.75 // Start gentle backoff when usage ≥75%
	defaultHTTPTimeout           = 30 * time.Second
	defaultMaxAttemptsPerRequest = 6
	errorBodyLimitBytes          = 2 << 10 // 2KB
)

// Default USDT-margined futures base URLs (can be overridden via injection)
var FuturesBaseURLs = []string{
	"https://fapi.binance.com",
	"https://fapi1.binance.com",
	"https://fapi2.binance.com",
	"https://fapi3.binance.com",
	"https://fapi4.binance.com",
}

// ====== Optional Token Bucket Governance (integrates with Governor) ======

type TokenGovernor interface {
	// Acquire blocks until tokens are available based on cost (weight) and optional bucket order, ctx can be cancelled
	Acquire(ctx context.Context, cost float64, buckets ...string) error
	// Optional: temporarily reduce rate when receiving 429
	Feedback429(bucket string, factor float64, decay time.Duration)
}

// ====== K-line Structure (can be replaced when integrating with nofx/market package) ======

type marketKline struct {
	OpenTime            int64
	Open                float64
	High                float64
	Low                 float64
	Close               float64
	Volume              float64
	CloseTime           int64
	QuoteVolume         float64
	Trades              int
	TakerBuyBaseVolume  float64
	TakerBuyQuoteVolume float64
	Symbol              string // Optional backfill
	Interval            string // Optional backfill
}

// ====== Client Configuration and Construction ======

type BackoffPolicy struct {
	Base time.Duration
	Max  time.Duration
}

type HeaderBackoff struct {
	Enabled       bool
	Weight1mLimit int     // Default 6000
	TriggerRatio  float64 // Start backoff when ratio is reached, default 0.75
	Base          time.Duration
	Max           time.Duration
}

type HttpOptions struct {
	HTTPClient    *http.Client
	BaseURLs      []string
	MaxAttempts   int
	Backoff       BackoffPolicy
	HeaderBackoff HeaderBackoff
	Governor      TokenGovernor // Can be nil
	// Bucket name order for Acquire (e.g., global->reconcile)
	GovernorBuckets []string
}

func defaultOptions() HttpOptions {
	return HttpOptions{
		HTTPClient:  &http.Client{Timeout: defaultHTTPTimeout},
		BaseURLs:    append([]string(nil), FuturesBaseURLs...),
		MaxAttempts: defaultMaxAttemptsPerRequest,
		Backoff: BackoffPolicy{
			Base: defaultBackoffBase,
			Max:  defaultBackoffMax,
		},
		HeaderBackoff: HeaderBackoff{
			Enabled:       true,
			Weight1mLimit: defaultWeight1mLimit,
			TriggerRatio:  defaultHeaderBackoffRatio,
			Base:          120 * time.Millisecond,
			Max:           1500 * time.Millisecond,
		},
	}
}

type KlineHTTPClient struct {
	client  *http.Client
	bases   []string
	idx     int
	maxTry  int
	backoff BackoffPolicy

	hb       HeaderBackoff
	gov      TokenGovernor
	govOrder []string
}

func NewKlineHTTPClient(opts ...func(*HttpOptions)) *KlineHTTPClient {
	o := defaultOptions()
	for _, fn := range opts {
		fn(&o)
	}
	if o.HTTPClient == nil {
		o.HTTPClient = &http.Client{Timeout: defaultHTTPTimeout}
	}
	if len(o.BaseURLs) == 0 {
		o.BaseURLs = append([]string(nil), FuturesBaseURLs...)
	}
	if o.MaxAttempts <= 0 {
		o.MaxAttempts = defaultMaxAttemptsPerRequest
	}
	return &KlineHTTPClient{
		client:  o.HTTPClient,
		bases:   o.BaseURLs,
		idx:     0,
		maxTry:  o.MaxAttempts,
		backoff: o.Backoff,

		hb:       o.HeaderBackoff,
		gov:      o.Governor,
		govOrder: o.GovernorBuckets,
	}
}

// Rotate to next base URL
func (c *KlineHTTPClient) rotate() {
	if len(c.bases) == 0 {
		return
	}
	c.idx = (c.idx + 1) % len(c.bases)
}

// ====== Public API ======

// GetKlines fetches the most recent limit klines (without start/end, suitable for "recent window")
func (c *KlineHTTPClient) GetKlines(ctx context.Context, symbol, interval string, limit int) ([]marketKline, error) {
	return c.getKlinesInternal(ctx, symbol, interval, limit, 0, 0)
}

// GetKlinesWindow fetches klines with startTime/endTime (in milliseconds)
func (c *KlineHTTPClient) GetKlinesWindow(ctx context.Context, symbol, interval string, startMs, endMs int64, limit int) ([]marketKline, error) {
	return c.getKlinesInternal(ctx, symbol, interval, limit, startMs, endMs)
}

func (c *KlineHTTPClient) getKlinesInternal(ctx context.Context, symbol, interval string, limit int, startMs, endMs int64) ([]marketKline, error) {
	if limit <= 0 {
		limit = 60
	}
	cost := FuturesKlinesCost(limit)

	// Optional: perform token bucket governance before sending request
	if c.gov != nil {
		if err := c.gov.Acquire(ctx, cost, c.govOrder...); err != nil {
			return nil, err
		}
	}

	baseAttempts := len(c.bases)
	if baseAttempts == 0 {
		return nil, errors.New("no base URLs configured")
	}

	attempt := 0
	// Exponential backoff multiplier
	var backoff time.Duration = c.backoff.Base

	for attempt < c.maxTry {
		base := c.bases[c.idx]
		u := base + "/fapi/v1/klines"

		req, err := http.NewRequestWithContext(ctx, http.MethodGet, u, nil)
		if err != nil {
			return nil, err
		}
		q := req.URL.Query()
		q.Set("symbol", strings.ToUpper(symbol))
		q.Set("interval", interval)
		q.Set("limit", strconv.Itoa(limit))
		if startMs > 0 {
			q.Set("startTime", strconv.FormatInt(startMs, 10))
		}
		if endMs > 0 {
			q.Set("endTime", strconv.FormatInt(endMs, 10))
		}
		req.URL.RawQuery = q.Encode()

		resp, err := c.client.Do(req)
		if err != nil {
			// Network error: rotate domain + exponential backoff
			log.Printf("[REST] network error on %s: %v (rotate+backoff)", base, err)
			c.rotate()
			if err := sleepCtx(ctx, withJitter(backoff)); err != nil {
				return nil, err
			}
			backoff = nextBackoff(backoff, c.backoff.Max)
			attempt++
			continue
		}

		// Use local variables to store results, avoiding global variable concurrency issues
		var result []marketKline
		var success bool

		// Ensure response body is closed
		func() {
			defer resp.Body.Close()

			// 429/rate limit: respect Retry-After; attempt domain rotation
			if resp.StatusCode == http.StatusTooManyRequests {
				// Dynamic rate reduction (optional)
				if c.gov != nil {
					// Assume backfill bucket name is "reconcile"; ignore if not available
					c.gov.Feedback429("reconcile", 0.7, 30*time.Second)
				}
				delay := retryAfterDelay(resp.Header, c.backoff.Base)
				log.Printf("[REST] 429 received on %s, sleep %v then rotate", base, delay)
				_ = sleepCtx(ctx, delay)
				c.rotate()
				attempt++
				return
			}

			// 418 (temporarily blacklisted) or 5xx: exponential backoff + rotation
			if resp.StatusCode == http.StatusTeapot || resp.StatusCode/100 == 5 {
				body := readAtMost(resp.Body, errorBodyLimitBytes)
				log.Printf("[REST] %d on %s: %s (rotate+backoff)", resp.StatusCode, base, body)
				c.rotate()
				_ = sleepCtx(ctx, withJitter(backoff))
				backoff = nextBackoff(backoff, c.backoff.Max)
				attempt++
				return
			}

			// Non-2xx is also treated as failure (but not forced rotation, can backoff)
			if resp.StatusCode/100 != 2 {
				body := readAtMost(resp.Body, errorBodyLimitBytes)
				log.Printf("[REST] non-2xx(%d) on %s: %s (backoff)", resp.StatusCode, base, body)
				_ = sleepCtx(ctx, withJitter(backoff))
				backoff = nextBackoff(backoff, c.backoff.Max)
				attempt++
				return
			}

			// Adaptive header backoff (header-aware backoff, ctx-aware)
			if c.hb.Enabled {
				if d := c.headerBackoffDelay(resp.Header); d > 0 {
					_ = sleepCtx(ctx, d)
				}
			}

			// Parse JSON: use Decoder + UseNumber for precision
			dec := json.NewDecoder(resp.Body)
			dec.UseNumber()
			var raw [][]interface{}
			if err := dec.Decode(&raw); err != nil {
				log.Printf("[REST] json decode error on %s: %v", base, err)
				_ = sleepCtx(ctx, withJitter(backoff))
				backoff = nextBackoff(backoff, c.backoff.Max)
				attempt++
				return
			}

			out := make([]marketKline, 0, len(raw))
			for _, row := range raw {
				k, perr := parseKlineArray(row)
				if perr != nil {
					continue
				}
				k.Symbol = strings.ToUpper(symbol)
				k.Interval = interval
				out = append(out, k)
			}

			// Success: set local variables
			result = out
			success = true
		}()

		if success {
			return result, nil
		}
		// Next attempt (rotation/backoff already decided above)
	}

	return nil, fmt.Errorf("all attempts failed for %s %s (limit=%d)", symbol, interval, limit)
}

// ====== Internal Utilities ======

func parseKlineArray(arr []interface{}) (marketKline, error) {
	// Expected at least 11 items
	if len(arr) < 11 {
		return marketKline{}, fmt.Errorf("invalid kline length=%d", len(arr))
	}
	num := func(v interface{}) float64 {
		switch t := v.(type) {
		case json.Number:
			f, _ := t.Float64()
			return f
		case string:
			f, _ := strconv.ParseFloat(t, 64)
			return f
		case float64:
			return t
		case int64:
			return float64(t)
		case int:
			return float64(t)
		default:
			return 0
		}
	}
	i64 := func(v interface{}) int64 {
		switch t := v.(type) {
		case json.Number:
			i, _ := strconv.ParseInt(string(t), 10, 64)
			return i
		case string:
			f, _ := strconv.ParseFloat(t, 64)
			return int64(f)
		case float64:
			return int64(t)
		case int64:
			return t
		case int:
			return int64(t)
		default:
			return 0
		}
	}

	k := marketKline{
		OpenTime:            i64(arr[0]),
		Open:                num(arr[1]),
		High:                num(arr[2]),
		Low:                 num(arr[3]),
		Close:               num(arr[4]),
		Volume:              num(arr[5]),
		CloseTime:           i64(arr[6]),
		QuoteVolume:         num(arr[7]),
		Trades:              int(i64(arr[8])),
		TakerBuyBaseVolume:  num(arr[9]),
		TakerBuyQuoteVolume: num(arr[10]),
	}
	return k, nil
}

// Adaptive header backoff: performs ctx-aware small delay based on 1-minute weight usage ratio
func (c *KlineHTTPClient) headerBackoffDelay(h http.Header) time.Duration {
	limit := c.hb.Weight1mLimit
	if limit <= 0 {
		limit = defaultWeight1mLimit
	}
	used := getUsedWeight1m(h)
	if used <= 0 {
		return 0
	}
	ratio := float64(used) / float64(limit)
	if ratio < c.hb.TriggerRatio {
		return 0
	}
	// Linearly increasing delay, capped at hb.Max
	alpha := (ratio - c.hb.TriggerRatio) / (1 - c.hb.TriggerRatio)
	if alpha < 0 {
		alpha = 0
	}
	if alpha > 1 {
		alpha = 1
	}
	delay := time.Duration(float64(c.hb.Base) * (0.5 + alpha)) // 0.5~1.5x base
	if delay > c.hb.Max {
		delay = c.hb.Max
	}
	return withJitter(delay)
}

func getUsedWeight1m(h http.Header) int {
	// Prefer 1M/1m, then fallback to X-MBX-USED-WEIGHT
	for _, k := range []string{HTTPHeaderUsedWeight1mUpper, HTTPHeaderUsedWeight1mLower, HTTPHeaderUsedWeight} {
		v := h.Get(k)
		if v == "" {
			continue
		}
		n, err := strconv.Atoi(v)
		if err == nil {
			return n
		}
	}
	// Compatible with variants like X-MBX-USED-WEIGHT-1m, and X-MBX-USED-WEIGHT itself
	for k, vals := range h {
		kUpper := strings.ToUpper(k)
		// Check if it starts with "X-MBX-USED-WEIGHT-" or is "X-MBX-USED-WEIGHT" itself
		if strings.HasPrefix(kUpper, HTTPHeaderUsedWeightPrefix) || kUpper == HTTPHeaderUsedWeight {
			for _, v := range vals {
				if n, err := strconv.Atoi(v); err == nil {
					return n
				}
			}
		}
	}
	return 0
}

func retryAfterDelay(h http.Header, fallback time.Duration) time.Duration {
	if ra := h.Get(HTTPHeaderRetryAfter); ra != "" {
		// seconds
		if secs, err := strconv.Atoi(ra); err == nil && secs >= 0 {
			return time.Duration(secs) * time.Second
		}
		// http-date
		if t, err := http.ParseTime(ra); err == nil {
			d := time.Until(t)
			if d > 0 {
				return d
			}
			return 0
		}
	}
	return withJitter(fallback)
}

func readAtMost(r io.Reader, n int64) string {
	lr := &io.LimitedReader{R: r, N: n}
	b, _ := io.ReadAll(lr)
	return string(b)
}

func nextBackoff(cur, max time.Duration) time.Duration {
	if cur <= 0 {
		return defaultBackoffBase
	}
	next := time.Duration(float64(cur) * 1.6)
	if next > max {
		next = max
	}
	return next
}

func withJitter(d time.Duration) time.Duration {
	if d <= 0 {
		return 0
	}
	// Jitter range ±25%
	j := float64(d) * 0.25
	return time.Duration(float64(d) - j + (randFloat64() * 2 * j))
}

// Simple reentrant pseudo-random
var rndSeed = uint64(time.Now().UnixNano())

func randFloat64() float64 {
	// xorshift64*
	x := rndSeed
	x ^= x << 13
	x ^= x >> 7
	x ^= x << 17
	rndSeed = x
	// 0..1
	return float64(x&0x3FFFFFFFFFFFFFFF) / float64(0x3FFFFFFFFFFFFFFF)
}

func sleepCtx(ctx context.Context, d time.Duration) error {
	if d <= 0 {
		return nil
	}
	t := time.NewTimer(d)
	defer t.Stop()
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-t.C:
		return nil
	}
}

// GetFuturesExchangeInfo fetches futures exchange metadata (domain rotation + 429/weight adaptive)
func (c *KlineHTTPClient) GetFuturesExchangeInfo(ctx context.Context) (*ExchangeInfo, error) {
	baseAttempts := len(c.bases)
	if baseAttempts == 0 {
		return nil, errors.New("no base URLs configured")
	}
	attempt := 0
	var backoff time.Duration = c.backoff.Base
	for attempt < c.maxTry {
		// Check if context is cancelled
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		default:
		}

		base := c.bases[c.idx]
		u := base + "/fapi/v1/exchangeInfo"

		req, err := http.NewRequestWithContext(ctx, http.MethodGet, u, nil)
		if err != nil {
			return nil, err
		}
		resp, err := c.client.Do(req)
		if err != nil {
			// Check if it's a context cancellation error
			if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
				return nil, err
			}
			log.Printf("exchangeInfo request error on %s: %v", base, err)
			c.rotate()
			attempt++
			if err := sleepCtx(ctx, backoff); err != nil {
				return nil, err
			}
			backoff = nextBackoff(backoff, c.backoff.Max)
			continue
		}
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
		resp.Body.Close()

		// Handle 429
		if resp.StatusCode == http.StatusTooManyRequests {
			// Respect Retry-After, otherwise use default backoff
			if d := retryAfterDelay(resp.Header, c.backoff.Base); d > 0 {
				if err := sleepCtx(ctx, d); err != nil {
					return nil, err
				}
			}
			c.rotate()
			attempt++
			continue
		}

		// Header adaptive backoff
		if c.hb.Enabled {
			if d := c.headerBackoffDelay(resp.Header); d > 0 {
				if err := sleepCtx(ctx, d); err != nil {
					return nil, err
				}
			}
		}

		if resp.StatusCode/100 != 2 {
			log.Printf("exchangeInfo non-2xx on %s: %s", base, string(body))
			c.rotate()
			attempt++
			if err := sleepCtx(ctx, backoff); err != nil {
				return nil, err
			}
			backoff = nextBackoff(backoff, c.backoff.Max)
			continue
		}

		var info ExchangeInfo
		if err := json.Unmarshal(body, &info); err != nil {
			return nil, fmt.Errorf("decode exchangeInfo: %w", err)
		}
		return &info, nil
	}
	return nil, errors.New("exchangeInfo request failed after retries")
}
