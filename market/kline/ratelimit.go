package kline

import (
	"context"
	"errors"
	"sync"
	"time"
)

var (
	ErrCostTooLarge   = errors.New("cost is larger than bucket capacity")
	ErrContextTimeout = errors.New("context deadline exceeded before tokens available")
)

// Bucket is a weighted token bucket with time-precise refill (tokens can be fractional).
// - capacity: maximum number of tokens (recommended to model as "per-minute limit" capacity, e.g., 6000 weight/min)
// - ratePerSec: refill rate per second (capacity/60)
// - tokens: current remaining tokens (float64, supports fractional refill)
// - lastRefill: last refill time (uses monotonic clock)
type Bucket struct {
	mu         sync.Mutex
	capacity   float64
	tokens     float64
	ratePerSec float64
	lastRefill time.Time

	// For smooth dynamic rate adjustment: records a multiplier (e.g., 0.5 means temporarily reduce 50% rate)
	scale float64
}

func NewBucket(capacity float64, ratePerSec float64) *Bucket {
	if capacity <= 0 {
		capacity = 1000
	}
	if ratePerSec <= 0 {
		ratePerSec = 100
	}
	now := time.Now()
	return &Bucket{
		capacity:   capacity,
		tokens:     capacity, // Full bucket at startup
		ratePerSec: ratePerSec,
		lastRefill: now,
		scale:      1.0,
	}
}

// internal: refill based on time (considering dynamic scaling scale)
func (b *Bucket) refillLocked(now time.Time) {
	elapsed := now.Sub(b.lastRefill).Seconds()
	if elapsed <= 0 {
		return
	}
	refill := elapsed * b.ratePerSec * b.scale
	b.tokens += refill
	if b.tokens > b.capacity {
		b.tokens = b.capacity
	}
	b.lastRefill = now
}

// TryTake is non-blocking: if tokens are sufficient, deduct and return true, otherwise return false.
func (b *Bucket) TryTake(cost float64) bool {
	if cost <= 0 {
		return true
	}
	b.mu.Lock()
	defer b.mu.Unlock()
	b.refillLocked(time.Now())
	if cost > b.capacity {
		// Cost can never be satisfied, directly reject (or let upper layer split request)
		return false
	}
	if b.tokens >= cost {
		b.tokens -= cost
		return true
	}
	return false
}

// Reserve calculates wait duration (without deduction). Returns 0 if can execute immediately; otherwise returns precise wait duration needed.
func (b *Bucket) Reserve(cost float64) time.Duration {
	if cost <= 0 {
		return 0
	}
	b.mu.Lock()
	defer b.mu.Unlock()
	now := time.Now()
	b.refillLocked(now)
	if cost > b.capacity {
		// Usually should be split by upper layer; here returns theoretical wait (may be very long)
		def := cost - b.tokens
		waitSec := def / (b.ratePerSec * b.scale)
		return time.Duration(waitSec * float64(time.Second))
	}
	if b.tokens >= cost {
		return 0
	}
	def := cost - b.tokens
	waitSec := def / (b.ratePerSec * b.scale)
	return time.Duration(waitSec * float64(time.Second))
}

// WaitN is blocking: waits for sufficient tokens before ctx deadline, deducts on success; returns ErrContextTimeout or ErrCostTooLarge on failure.
func (b *Bucket) WaitN(ctx context.Context, cost float64) error {
	if cost <= 0 {
		return nil
	}
	// Fast path
	if b.TryTake(cost) {
		return nil
	}
	// Calculate wait
	for {
		wait := b.Reserve(cost)
		if wait <= 0 {
			// Retry deduction
			if b.TryTake(cost) {
				return nil
			}
			// Extreme race condition: loop again
			continue
		}
		// Wait or cancel
		timer := time.NewTimer(wait)
		select {
		case <-ctx.Done():
			timer.Stop()
			return ErrContextTimeout
		case <-timer.C:
			// Loop and retry
		}
	}
}

// Feedback429 is used for temporary rate reduction after receiving 429 / Retry-After (e.g., when Binance weight is nearly exhausted).
// factor ∈ (0,1], e.g., 0.5 means reduce to 50% rate; recovers to 1.0 after decay time.
func (b *Bucket) Feedback429(factor float64, decay time.Duration) {
	if factor <= 0 || factor > 1 {
		factor = 0.8
	}
	b.mu.Lock()
	b.scale = factor
	b.mu.Unlock()

	// Scheduled recovery
	go func() {
		timer := time.NewTimer(decay)
		<-timer.C
		b.mu.Lock()
		b.scale = 1.0
		b.mu.Unlock()
	}()
}

func (b *Bucket) SetRatePerSec(newRate float64) {
	if newRate <= 0 {
		return
	}
	b.mu.Lock()
	b.refillLocked(time.Now())
	b.ratePerSec = newRate
	b.mu.Unlock()
}

func (b *Bucket) SetCapacity(newCap float64) {
	if newCap <= 0 {
		return
	}
	b.mu.Lock()
	b.refillLocked(time.Now())
	b.capacity = newCap
	if b.tokens > b.capacity {
		b.tokens = b.capacity
	}
	b.mu.Unlock()
}

// Snapshot for observation
type BucketSnapshot struct {
	Capacity   float64
	Tokens     float64
	RatePerSec float64
	Scale      float64
}

func (b *Bucket) Snapshot() BucketSnapshot {
	b.mu.Lock()
	defer b.mu.Unlock()
	// Refresh view
	b.refillLocked(time.Now())
	return BucketSnapshot{
		Capacity:   b.capacity,
		Tokens:     b.tokens,
		RatePerSec: b.ratePerSec,
		Scale:      b.scale,
	}
}

// ---------------------------- Multi-Bucket Governance ----------------------------

// Governor manages multiple buckets (e.g., global bucket + backfill bucket + tail validation bucket + emergency bucket).
// Acquire waits and deducts on all buckets in order—only succeeds if all pass.
type Governor struct {
	mu     sync.RWMutex
	bucket map[string]*Bucket
	order  []string // Default order for Acquire (e.g., global -> class -> team -> lane)
}

func NewGovernor() *Governor {
	return &Governor{
		bucket: make(map[string]*Bucket),
	}
}

func (g *Governor) AddBucket(name string, b *Bucket) {
	g.mu.Lock()
	defer g.mu.Unlock()
	g.bucket[name] = b
	g.rebuildOrderLocked()
}

func (g *Governor) SetOrder(names []string) {
	g.mu.Lock()
	defer g.mu.Unlock()
	g.order = append([]string(nil), names...)
}

func (g *Governor) rebuildOrderLocked() {
	// If not explicitly set, insertion order is not guaranteed, here only preserve existing order
	if len(g.order) == 0 {
		// Default map key order is unstable, recommend external SetOrder.
	}
}

// Acquire deducts in order on given bucket list (blocking), returns nil only if all succeed.
// If buckets is empty, use g.order; cost is the "weight" of this request.
func (g *Governor) Acquire(ctx context.Context, cost float64, buckets ...string) error {
	if cost <= 0 {
		return nil
	}
	g.mu.RLock()
	names := buckets
	if len(names) == 0 {
		names = g.order
	}
	// Snapshot references
	list := make([]*Bucket, 0, len(names))
	for _, n := range names {
		if b, ok := g.bucket[n]; ok {
			list = append(list, b)
		}
	}
	g.mu.RUnlock()

	if len(list) == 0 {
		return nil // No buckets means no limit
	}

	// Wait bucket by bucket; here uses "serial wait" strategy, simple and effective
	for _, b := range list {
		// Fast path first
		if b.TryTake(cost) {
			continue
		}
		// Blocking wait
		if err := b.WaitN(ctx, cost); err != nil {
			return err
		}
	}
	return nil
}

// Feedback429 performs temporary rate reduction on a bucket.
func (g *Governor) Feedback429(bucket string, factor float64, decay time.Duration) {
	g.mu.RLock()
	b := g.bucket[bucket]
	g.mu.RUnlock()
	if b != nil {
		b.Feedback429(factor, decay)
	}
}

// Snapshot all buckets
func (g *Governor) Snapshot() map[string]BucketSnapshot {
	g.mu.RLock()
	defer g.mu.RUnlock()
	out := make(map[string]BucketSnapshot, len(g.bucket))
	for k, b := range g.bucket {
		out[k] = b.Snapshot()
	}
	return out
}

// ------------------------ Binance Endpoint Weight Helpers ------------------------

// FuturesKlinesCost returns weight mapping for /fapi/v1/klines (example common tiers)
func FuturesKlinesCost(limit int) float64 {
	switch {
	case limit <= 0:
		return 1
	case limit < 100:
		return 1
	case limit < 500:
		return 2
	case limit <= 1000:
		return 5
	default:
		// Extreme tier, treat as 10; refer to official weight for specifics
		return 10
	}
}

// SpotKlinesCost: common weight for Spot /api/v3/klines (mostly 1; if uiKlines is enabled, can use 2)
func SpotKlinesCost(limit int) float64 {
	if limit <= 0 {
		return 1
	}
	// Default 1 here; if you enable uiKlines, can customize to 2
	return 1
}

// SafeJitter adds slight jitter to wait time, reducing thundering herd/synchronization
func SafeJitter(d time.Duration, ratio float64) time.Duration {
	if ratio <= 0 {
		return d
	}
	j := float64(d) * ratio
	return time.Duration(float64(d) - j/2 + (safeRandFloat64() * j))
}

var rndOnce sync.Once
var rndSrc *randSource

type randSource struct {
	mu sync.Mutex
	x  uint64
}

func (r *randSource) Float64() float64 {
	r.mu.Lock()
	r.x = r.x*2862933555777941757 + 3037000493
	v := (r.x >> 33) + 1
	r.mu.Unlock()
	// 0..1
	return float64(v%1_000_000) / 1_000_000
}

func safeRandFloat64() float64 {
	rndOnce.Do(func() { rndSrc = &randSource{x: uint64(time.Now().UnixNano())} })
	return rndSrc.Float64()
}
