package kline

import (
	"context"
	"errors"
	"sort"
	"strings"
	"sync"
	"time"
)

// ---------------- Configuration and Interface ----------------

const (
	DefaultSymbolRefreshInterval = 5 * time.Minute
	// Failure backoff bounds
	symbolBackoffBase = 800 * time.Millisecond
	symbolBackoffMax  = 15 * time.Second
	// Timing jitter half-amplitude (±)
	symbolJitterHalf = 0.15
)

// SymbolFetcher abstracts "getting currently supported symbols (uppercase)"
type SymbolFetcher interface {
	FetchSymbols(ctx context.Context) ([]string, error)
}

// FilteredBinanceFuturesFetcher: default USDM perpetual/USDT/TRADING filter
type FilteredBinanceFuturesFetcher struct {
	cli *KlineHTTPClient
	// Optional custom filter; returns true to keep
	Predicate func(s ExchangeSymbol) bool
}

func NewFilteredBinanceFuturesFetcher(cli *KlineHTTPClient) *FilteredBinanceFuturesFetcher {
	return &FilteredBinanceFuturesFetcher{
		cli: cli,
		Predicate: func(s ExchangeSymbol) bool {
			return strings.EqualFold(s.Status, SymbolStatusTrading) &&
				strings.EqualFold(s.ContractType, ContractTypePerpetual) &&
				strings.EqualFold(s.QuoteAsset, QuoteAssetUSDT)
		},
	}
}

// Your existing types: only showing field signatures we use for clarity
type ExchangeSymbol struct {
	Symbol       string
	Status       string
	ContractType string
	QuoteAsset   string
}

type ExchangeInfo struct {
	Symbols []ExchangeSymbol
}

// Requires KlineHTTPClient to have this method; if method name differs, update the call here.
func (f *FilteredBinanceFuturesFetcher) FetchSymbols(ctx context.Context) ([]string, error) {
	info, err := f.cli.GetFuturesExchangeInfo(ctx)
	if err != nil {
		return nil, err
	}
	out := make([]string, 0, len(info.Symbols))
	for _, s := range info.Symbols {
		if f.Predicate == nil || f.Predicate(s) {
			out = append(out, strings.ToUpper(strings.TrimSpace(s.Symbol)))
		}
	}
	// Deduplicate
	m := make(map[string]struct{}, len(out))
	uniq := uniqKeepUpper(out, m)
	// Sort for auditing/comparison testing
	sort.Strings(uniq)
	return uniq, nil
}

// uniqKeepUpper: deduplicates slice (elements already uppercase), reuses passed map to reduce allocation
func uniqKeepUpper(in []string, seen map[string]struct{}) []string {
	for k := range seen {
		delete(seen, k)
	}
	out := make([]string, 0, len(in))
	for _, s := range in {
		if s == "" || strings.TrimSpace(s) == "" {
			continue
		}
		if _, ok := seen[s]; ok {
			continue
		}
		seen[s] = struct{}{}
		out = append(out, s)
	}
	return out
}

// ---------------- SymbolSync Implementation ----------------

type SymbolSync struct {
	mu       sync.Mutex
	interval time.Duration
	fetcher  SymbolFetcher

	// Last synced set (uppercase)
	lastSet map[string]struct{}
	// Optional: preset initial set, if empty then first sync uses remote response as baseline
	initialSeed []string
}

// NewSymbolSync: if interval<=0 uses default 5m; if fetcher is nil uses default Binance USDM filter implementation.
func NewSymbolSync(interval time.Duration, fetcher SymbolFetcher) *SymbolSync {
	if interval <= 0 {
		interval = DefaultSymbolRefreshInterval
	}
	if fetcher == nil {
		fetcher = NewFilteredBinanceFuturesFetcher(NewKlineHTTPClient())
	}
	return &SymbolSync{
		interval: interval,
		fetcher:  fetcher,
		lastSet:  make(map[string]struct{}),
	}
}

// WithInitialSeed: sets baseline set at startup (uppercase), avoids treating full network as "added" in first round
func (ss *SymbolSync) WithInitialSeed(seed []string) *SymbolSync {
	ss.mu.Lock()
	defer ss.mu.Unlock()
	ss.initialSeed = normalizeUpper(seed)
	ss.lastSet = make(map[string]struct{}, len(ss.initialSeed))
	for _, s := range ss.initialSeed {
		ss.lastSet[s] = struct{}{}
	}
	return ss
}

// Start: periodic background sync; attempts one fetch on startup (triggers callback on success), failures use exponential backoff retry.
// onChange is called when additions/removals are detected (executed after unlock to avoid deadlock).
func (ss *SymbolSync) Start(ctx context.Context, onChange func(added, removed []string)) {
	// Run once first (with backoff)
	ss.loopOnceWithRetry(ctx, onChange)

	// Periodic ticker (with jitter)
	t := time.NewTicker(jitterDuration(ss.interval, symbolJitterHalf))
	defer t.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-t.C:
			// Randomize next interval on each tick to avoid cluster synchronization
			t.Reset(jitterDuration(ss.interval, symbolJitterHalf))
			ss.loopOnceWithRetry(ctx, onChange)
		}
	}
}

// ComputeDiffOnce: only computes diff, does not update internal lastSet.
func (ss *SymbolSync) ComputeDiffOnce(ctx context.Context) (added, removed []string, err error) {
	curr, err := ss.fetcher.FetchSymbols(ctx)
	if err != nil {
		return nil, nil, err
	}
	currSet := toSet(curr)
	ss.mu.Lock()
	defer ss.mu.Unlock()
	added, removed = diffSets(ss.lastSet, currSet)
	sort.Strings(added)
	sort.Strings(removed)
	return added, removed, nil
}

// ---------------- Internal Logic ----------------

func (ss *SymbolSync) loopOnceWithRetry(ctx context.Context, onChange func(added, removed []string)) {
	backoff := symbolBackoffBase
	for {
		// Check if context is cancelled
		select {
		case <-ctx.Done():
			return
		default:
		}

		added, removed, err := ss.computeAndUpdate(ctx, onChange)
		if err == nil {
			_ = added
			_ = removed
			return
		}
		// Exit if ctx is cancelled
		if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
			return
		}
		// Exponential backoff + jitter
		sleep := jitterDuration(backoff, 0.25)
		if err := sleepCtx(ctx, sleep); err != nil {
			return
		}
		backoff = nextBackoff(backoff, symbolBackoffMax)
	}
}

func (ss *SymbolSync) computeAndUpdate(ctx context.Context, onChange func(added, removed []string)) ([]string, []string, error) {
	curr, err := ss.fetcher.FetchSymbols(ctx)
	if err != nil {
		return nil, nil, err
	}
	currSet := toSet(curr)

	// Compute diff and update lastSet in one critical section, but callback outside lock
	ss.mu.Lock()
	added, removed := diffSets(ss.lastSet, currSet)
	// If first time and lastSet is empty and initialSeed provided, use initialSeed as baseline to avoid "added storm"
	if len(ss.lastSet) == 0 && len(ss.initialSeed) > 0 {
		added, removed = nil, nil
	}
	// Update lastSet
	ss.lastSet = currSet
	ss.mu.Unlock()

	// Trigger callback (outside lock)
	if onChange != nil && (len(added) > 0 || len(removed) > 0) {
		// Sort only for readability/testability
		sort.Strings(added)
		sort.Strings(removed)
		onChange(added, removed)
	}
	return added, removed, nil
}

// ---------------- Utilities ----------------

func toSet(list []string) map[string]struct{} {
	m := make(map[string]struct{}, len(list))
	for _, s := range list {
		su := strings.ToUpper(strings.TrimSpace(s))
		if su != "" {
			m[su] = struct{}{}
		}
	}
	return m
}

// diffSets: O(n) computes additions/removals of curr relative to prev
func diffSets(prev, curr map[string]struct{}) (added, removed []string) {
	for s := range curr {
		if _, ok := prev[s]; !ok {
			added = append(added, s)
		}
	}
	for s := range prev {
		if _, ok := curr[s]; !ok {
			removed = append(removed, s)
		}
	}
	return
}

func normalizeUpper(in []string) []string {
	out := make([]string, 0, len(in))
	for _, s := range in {
		if su := strings.ToUpper(strings.TrimSpace(s)); su != "" {
			out = append(out, su)
		}
	}
	sort.Strings(out)
	return out
}

// jitter: returns uniform random in [d*(1-half), d*(1+half)]
func jitterDuration(d time.Duration, half float64) time.Duration {
	if d <= 0 || half <= 0 {
		return d
	}
	span := float64(d) * half
	j := (randFloat64()*2 - 1) * span
	out := float64(d) + j
	if out < float64(time.Millisecond) {
		out = float64(time.Millisecond)
	}
	return time.Duration(out)
}
