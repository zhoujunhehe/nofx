package kline

import (
	"context"
	"math"
	"sync"
	"testing"
	"time"
)

// Helper function to check token count with tolerance
func checkTokens(t *testing.T, actual, expected, tolerance float64, msg string) {
	if math.Abs(actual-expected) > tolerance {
		t.Fatalf("%s: expected around %f, got %f (tolerance: %f)", msg, expected, actual, tolerance)
	}
}

func TestNewBucket(t *testing.T) {
	// Test with valid parameters
	b := NewBucket(1000, 100)
	if b == nil {
		t.Fatal("NewBucket returned nil")
	}
	if b.capacity != 1000 {
		t.Fatalf("expected capacity 1000, got %f", b.capacity)
	}
	if b.ratePerSec != 100 {
		t.Fatalf("expected ratePerSec 100, got %f", b.ratePerSec)
	}
	if b.tokens != 1000 {
		t.Fatalf("expected tokens 1000 (full bucket), got %f", b.tokens)
	}
	if b.scale != 1.0 {
		t.Fatalf("expected scale 1.0, got %f", b.scale)
	}

	// Test with zero capacity (should use default)
	b2 := NewBucket(0, 100)
	if b2.capacity != 1000 {
		t.Fatalf("expected default capacity 1000, got %f", b2.capacity)
	}

	// Test with zero rate (should use default)
	b3 := NewBucket(1000, 0)
	if b3.ratePerSec != 100 {
		t.Fatalf("expected default ratePerSec 100, got %f", b3.ratePerSec)
	}
}

func TestBucket_TryTake(t *testing.T) {
	b := NewBucket(1000, 100)

	// Test taking tokens when available
	if !b.TryTake(100) {
		t.Fatal("TryTake should succeed when tokens available")
	}

	// Verify tokens were deducted (allow small tolerance for refill during test)
	snap := b.Snapshot()
	checkTokens(t, snap.Tokens, 900.0, 0.1, "tokens after TryTake(100)")

	// Test taking more than available
	if b.TryTake(1000) {
		t.Fatal("TryTake should fail when tokens insufficient")
	}

	// Test taking zero cost
	if !b.TryTake(0) {
		t.Fatal("TryTake should succeed with zero cost")
	}

	// Test taking negative cost
	if !b.TryTake(-10) {
		t.Fatal("TryTake should succeed with negative cost")
	}
}

func TestBucket_TryTake_Refill(t *testing.T) {
	b := NewBucket(1000, 100) // 100 tokens per second

	// Exhaust tokens
	b.TryTake(1000)

	// Wait for refill (1 second should give 100 tokens)
	time.Sleep(1100 * time.Millisecond)

	// Should be able to take some tokens now
	if !b.TryTake(50) {
		t.Fatal("TryTake should succeed after refill")
	}
}

func TestBucket_TryTake_CostLargerThanCapacity(t *testing.T) {
	b := NewBucket(100, 10)

	// Try to take more than capacity
	if b.TryTake(200) {
		t.Fatal("TryTake should fail when cost > capacity")
	}
}

func TestBucket_Reserve(t *testing.T) {
	b := NewBucket(1000, 100) // 100 tokens per second

	// Reserve when tokens available
	wait := b.Reserve(100)
	if wait != 0 {
		t.Fatalf("expected 0 wait time, got %v", wait)
	}

	// Exhaust tokens
	b.TryTake(1000)

	// Reserve should calculate wait time
	wait = b.Reserve(100)
	if wait <= 0 {
		t.Fatalf("expected positive wait time, got %v", wait)
	}
	// Should be approximately 1 second (100 tokens / 100 per second)
	if wait < 900*time.Millisecond || wait > 1100*time.Millisecond {
		t.Fatalf("expected wait time around 1s, got %v", wait)
	}
}

func TestBucket_Reserve_ZeroCost(t *testing.T) {
	b := NewBucket(1000, 100)
	wait := b.Reserve(0)
	if wait != 0 {
		t.Fatalf("expected 0 wait time for zero cost, got %v", wait)
	}
}

func TestBucket_WaitN(t *testing.T) {
	b := NewBucket(1000, 100)

	// Wait when tokens available (should succeed immediately)
	ctx := context.Background()
	if err := b.WaitN(ctx, 100); err != nil {
		t.Fatalf("WaitN should succeed when tokens available: %v", err)
	}

	// Exhaust tokens
	b.TryTake(900)

	// Wait for small amount (should succeed after short wait)
	ctx2, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	if err := b.WaitN(ctx2, 50); err != nil {
		t.Fatalf("WaitN should succeed after refill: %v", err)
	}
}

func TestBucket_WaitN_ContextTimeout(t *testing.T) {
	b := NewBucket(1000, 1) // Very slow refill: 1 token per second

	// Exhaust tokens
	b.TryTake(1000)

	// Wait with short timeout
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	err := b.WaitN(ctx, 100)
	if err != ErrContextTimeout {
		t.Fatalf("expected ErrContextTimeout, got %v", err)
	}
}

func TestBucket_WaitN_ZeroCost(t *testing.T) {
	b := NewBucket(1000, 100)
	ctx := context.Background()
	if err := b.WaitN(ctx, 0); err != nil {
		t.Fatalf("WaitN should succeed with zero cost: %v", err)
	}
}

func TestBucket_Feedback429(t *testing.T) {
	b := NewBucket(1000, 100)

	// Apply 429 feedback (reduce to 50% rate)
	b.Feedback429(0.5, 100*time.Millisecond)

	// Verify scale is reduced
	snap := b.Snapshot()
	if snap.Scale != 0.5 {
		t.Fatalf("expected scale 0.5, got %f", snap.Scale)
	}

	// Wait for decay
	time.Sleep(150 * time.Millisecond)

	// Verify scale is restored
	snap2 := b.Snapshot()
	if snap2.Scale != 1.0 {
		t.Fatalf("expected scale 1.0 after decay, got %f", snap2.Scale)
	}
}

func TestBucket_Feedback429_InvalidFactor(t *testing.T) {
	b := NewBucket(1000, 100)

	// Test with factor > 1 (should clamp to 0.8)
	b.Feedback429(2.0, 100*time.Millisecond)
	snap := b.Snapshot()
	if snap.Scale != 0.8 {
		t.Fatalf("expected scale 0.8 (default for invalid), got %f", snap.Scale)
	}

	// Test with factor <= 0 (should clamp to 0.8)
	b.Feedback429(-1.0, 100*time.Millisecond)
	snap2 := b.Snapshot()
	if snap2.Scale != 0.8 {
		t.Fatalf("expected scale 0.8 (default for invalid), got %f", snap2.Scale)
	}
}

func TestBucket_SetRatePerSec(t *testing.T) {
	b := NewBucket(1000, 100)

	// Change rate
	b.SetRatePerSec(200)

	snap := b.Snapshot()
	if snap.RatePerSec != 200 {
		t.Fatalf("expected ratePerSec 200, got %f", snap.RatePerSec)
	}

	// Test with zero rate (should not change)
	b.SetRatePerSec(0)
	snap2 := b.Snapshot()
	if snap2.RatePerSec != 200 {
		t.Fatalf("expected ratePerSec to remain 200, got %f", snap2.RatePerSec)
	}
}

func TestBucket_SetCapacity(t *testing.T) {
	b := NewBucket(1000, 100)

	// Take some tokens
	b.TryTake(500)

	// Reduce capacity
	b.SetCapacity(300)

	snap := b.Snapshot()
	if snap.Capacity != 300 {
		t.Fatalf("expected capacity 300, got %f", snap.Capacity)
	}
	// Tokens should be capped to new capacity
	if snap.Tokens > 300 {
		t.Fatalf("expected tokens <= 300, got %f", snap.Tokens)
	}

	// Test with zero capacity (should not change)
	b.SetCapacity(0)
	snap2 := b.Snapshot()
	if snap2.Capacity != 300 {
		t.Fatalf("expected capacity to remain 300, got %f", snap2.Capacity)
	}
}

func TestBucket_Snapshot(t *testing.T) {
	b := NewBucket(1000, 100)

	// Take some tokens
	b.TryTake(200)

	snap := b.Snapshot()
	if snap.Capacity != 1000 {
		t.Fatalf("expected capacity 1000, got %f", snap.Capacity)
	}
	if snap.RatePerSec != 100 {
		t.Fatalf("expected ratePerSec 100, got %f", snap.RatePerSec)
	}
	checkTokens(t, snap.Tokens, 800.0, 0.1, "tokens after Snapshot()")
	if snap.Scale != 1.0 {
		t.Fatalf("expected scale 1.0, got %f", snap.Scale)
	}
}

func TestNewGovernor(t *testing.T) {
	g := NewGovernor()
	if g == nil {
		t.Fatal("NewGovernor returned nil")
	}
	if g.bucket == nil {
		t.Fatal("bucket map should be initialized")
	}
	if len(g.bucket) != 0 {
		t.Fatalf("expected empty bucket map, got %d buckets", len(g.bucket))
	}
}

func TestGovernor_AddBucket(t *testing.T) {
	g := NewGovernor()
	b := NewBucket(1000, 100)

	g.AddBucket("global", b)

	if len(g.bucket) != 1 {
		t.Fatalf("expected 1 bucket, got %d", len(g.bucket))
	}
	if g.bucket["global"] != b {
		t.Fatal("bucket not stored correctly")
	}
}

func TestGovernor_SetOrder(t *testing.T) {
	g := NewGovernor()
	order := []string{"global", "backfill", "reconcile"}
	g.SetOrder(order)

	if len(g.order) != 3 {
		t.Fatalf("expected order length 3, got %d", len(g.order))
	}
	for i, name := range order {
		if g.order[i] != name {
			t.Fatalf("expected order[%d] = %s, got %s", i, name, g.order[i])
		}
	}
}

func TestGovernor_Acquire_NoBuckets(t *testing.T) {
	g := NewGovernor()
	ctx := context.Background()

	// Acquire with no buckets should succeed
	if err := g.Acquire(ctx, 100); err != nil {
		t.Fatalf("Acquire should succeed with no buckets: %v", err)
	}
}

func TestGovernor_Acquire_SingleBucket(t *testing.T) {
	g := NewGovernor()
	b := NewBucket(1000, 100)
	g.AddBucket("global", b)
	g.SetOrder([]string{"global"}) // Set order so Acquire knows which bucket to use

	ctx := context.Background()

	// Acquire should succeed
	if err := g.Acquire(ctx, 100); err != nil {
		t.Fatalf("Acquire should succeed: %v", err)
	}

	// Verify tokens were deducted (allow small tolerance for refill during test)
	snap := b.Snapshot()
	checkTokens(t, snap.Tokens, 900.0, 0.1, "tokens after Acquire(100)")
}

func TestGovernor_Acquire_MultipleBuckets(t *testing.T) {
	g := NewGovernor()
	b1 := NewBucket(1000, 100)
	b2 := NewBucket(500, 50)
	g.AddBucket("global", b1)
	g.AddBucket("backfill", b2)
	g.SetOrder([]string{"global", "backfill"})

	ctx := context.Background()

	// Acquire should succeed on both buckets
	if err := g.Acquire(ctx, 100); err != nil {
		t.Fatalf("Acquire should succeed: %v", err)
	}

	// Verify both buckets had tokens deducted
	snap1 := b1.Snapshot()
	snap2 := b2.Snapshot()
	checkTokens(t, snap1.Tokens, 900.0, 0.1, "tokens after Acquire(100)")
	checkTokens(t, snap2.Tokens, 400.0, 0.1, "tokens after Acquire(100)")
}

func TestGovernor_Acquire_WithOrder(t *testing.T) {
	g := NewGovernor()
	b1 := NewBucket(1000, 100)
	b2 := NewBucket(500, 50)
	g.AddBucket("backfill", b2)
	g.AddBucket("global", b1)
	g.SetOrder([]string{"global", "backfill"})

	ctx := context.Background()

	// Acquire without specifying buckets should use order
	if err := g.Acquire(ctx, 100); err != nil {
		t.Fatalf("Acquire should succeed: %v", err)
	}

	// Verify both buckets were checked in order
	snap1 := b1.Snapshot()
	snap2 := b2.Snapshot()
	checkTokens(t, snap1.Tokens, 900.0, 0.1, "tokens after Acquire(100)")
	checkTokens(t, snap2.Tokens, 400.0, 0.1, "tokens after Acquire(100)")
}

func TestGovernor_Acquire_SpecificBuckets(t *testing.T) {
	g := NewGovernor()
	b1 := NewBucket(1000, 100)
	b2 := NewBucket(500, 50)
	b3 := NewBucket(200, 20)
	g.AddBucket("global", b1)
	g.AddBucket("backfill", b2)
	g.AddBucket("reconcile", b3)

	ctx := context.Background()

	// Acquire from specific buckets only
	if err := g.Acquire(ctx, 50, "global", "reconcile"); err != nil {
		t.Fatalf("Acquire should succeed: %v", err)
	}

	// Verify only specified buckets were used
	snap1 := b1.Snapshot()
	snap2 := b2.Snapshot()
	snap3 := b3.Snapshot()
	checkTokens(t, snap1.Tokens, 950.0, 0.1, "tokens after Acquire(50)")
	checkTokens(t, snap2.Tokens, 500.0, 0.1, "tokens after Acquire(50)")
	checkTokens(t, snap3.Tokens, 150.0, 0.1, "tokens after Acquire(50)")
}

func TestGovernor_Acquire_ContextTimeout(t *testing.T) {
	g := NewGovernor()
	b := NewBucket(100, 1) // Very slow refill
	g.AddBucket("global", b)
	g.SetOrder([]string{"global"}) // Set order so Acquire knows which bucket to use

	// Exhaust tokens
	b.TryTake(100)

	// Acquire with short timeout
	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()

	err := g.Acquire(ctx, 50)
	if err != ErrContextTimeout {
		t.Fatalf("expected ErrContextTimeout, got %v", err)
	}
}

func TestGovernor_Acquire_ZeroCost(t *testing.T) {
	g := NewGovernor()
	b := NewBucket(1000, 100)
	g.AddBucket("global", b)

	ctx := context.Background()
	if err := g.Acquire(ctx, 0); err != nil {
		t.Fatalf("Acquire should succeed with zero cost: %v", err)
	}
}

func TestGovernor_Feedback429(t *testing.T) {
	g := NewGovernor()
	b := NewBucket(1000, 100)
	g.AddBucket("global", b)

	// Apply 429 feedback
	g.Feedback429("global", 0.5, 100*time.Millisecond)

	// Verify scale is reduced
	snap := b.Snapshot()
	if snap.Scale != 0.5 {
		t.Fatalf("expected scale 0.5, got %f", snap.Scale)
	}

	// Wait for decay
	time.Sleep(150 * time.Millisecond)

	// Verify scale is restored
	snap2 := b.Snapshot()
	if snap2.Scale != 1.0 {
		t.Fatalf("expected scale 1.0 after decay, got %f", snap2.Scale)
	}
}

func TestGovernor_Feedback429_NonExistentBucket(t *testing.T) {
	g := NewGovernor()

	// Should not panic
	g.Feedback429("nonexistent", 0.5, 100*time.Millisecond)
}

func TestGovernor_Snapshot(t *testing.T) {
	g := NewGovernor()
	b1 := NewBucket(1000, 100)
	b2 := NewBucket(500, 50)
	g.AddBucket("global", b1)
	g.AddBucket("backfill", b2)

	// Take some tokens
	b1.TryTake(200)
	b2.TryTake(100)

	snap := g.Snapshot()
	if len(snap) != 2 {
		t.Fatalf("expected 2 buckets in snapshot, got %d", len(snap))
	}

	checkTokens(t, snap["global"].Tokens, 800.0, 0.1, "tokens after Snapshot()")
	checkTokens(t, snap["backfill"].Tokens, 400.0, 0.1, "tokens after Snapshot()")
}

func TestFuturesKlinesCost(t *testing.T) {
	tests := []struct {
		limit int
		want  float64
	}{
		{0, 1},
		{-10, 1},
		{50, 1},
		{99, 1},
		{100, 2},
		{499, 2},
		{500, 5},
		{1000, 5},
		{1500, 10},
		{2000, 10},
	}

	for _, tt := range tests {
		got := FuturesKlinesCost(tt.limit)
		if got != tt.want {
			t.Errorf("FuturesKlinesCost(%d) = %f, want %f", tt.limit, got, tt.want)
		}
	}
}

func TestSpotKlinesCost(t *testing.T) {
	tests := []struct {
		limit int
		want  float64
	}{
		{0, 1},
		{-10, 1},
		{50, 1},
		{100, 1},
		{500, 1},
		{1000, 1},
	}

	for _, tt := range tests {
		got := SpotKlinesCost(tt.limit)
		if got != tt.want {
			t.Errorf("SpotKlinesCost(%d) = %f, want %f", tt.limit, got, tt.want)
		}
	}
}

func TestSafeJitter(t *testing.T) {
	d := 1 * time.Second
	ratio := 0.1 // 10% jitter

	// Test multiple times to verify jitter range
	for i := 0; i < 100; i++ {
		jittered := SafeJitter(d, ratio)
		// Should be within [d*0.95, d*1.05] for 10% jitter
		min := time.Duration(float64(d) * 0.95)
		max := time.Duration(float64(d) * 1.05)
		if jittered < min || jittered > max {
			t.Fatalf("jittered duration %v out of range [%v, %v]", jittered, min, max)
		}
	}

	// Test with zero ratio (should return original)
	result := SafeJitter(d, 0)
	if result != d {
		t.Fatalf("expected original duration with zero ratio, got %v", result)
	}

	// Test with negative ratio (should return original)
	result2 := SafeJitter(d, -0.1)
	if result2 != d {
		t.Fatalf("expected original duration with negative ratio, got %v", result2)
	}
}

func TestBucket_ConcurrentAccess(t *testing.T) {
	b := NewBucket(10000, 1000)
	ctx := context.Background()

	// Concurrent acquires
	var wg sync.WaitGroup
	errors := make(chan error, 100)

	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			if err := b.WaitN(ctx, 10); err != nil {
				errors <- err
			}
		}()
	}

	wg.Wait()
	close(errors)

	// Check for errors
	for err := range errors {
		if err != nil {
			t.Errorf("unexpected error in concurrent access: %v", err)
		}
	}

	// Verify tokens were deducted correctly
	snap := b.Snapshot()
	expected := 10000.0 - 100*10.0 // 100 acquires * 10 cost each
	// Allow some tolerance for refill during concurrent access
	if snap.Tokens < expected-100 || snap.Tokens > expected+100 {
		t.Fatalf("expected tokens around %f, got %f", expected, snap.Tokens)
	}
}

func TestGovernor_ConcurrentAccess(t *testing.T) {
	g := NewGovernor()
	b := NewBucket(10000, 1000)
	g.AddBucket("global", b)
	g.SetOrder([]string{"global"}) // Set order so Acquire knows which bucket to use
	ctx := context.Background()

	// Concurrent acquires
	var wg sync.WaitGroup
	errors := make(chan error, 50)

	for i := 0; i < 50; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			if err := g.Acquire(ctx, 10); err != nil {
				errors <- err
			}
		}()
	}

	wg.Wait()
	close(errors)

	// Check for errors
	for err := range errors {
		if err != nil {
			t.Errorf("unexpected error in concurrent access: %v", err)
		}
	}
}
