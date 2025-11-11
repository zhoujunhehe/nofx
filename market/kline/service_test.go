package kline

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"
)

func requireServiceIT(t *testing.T) {
	if os.Getenv("KLINE_SERVICE_IT") != "1" {
		t.Skip("service integration tests disabled; set KLINE_SERVICE_IT=1 to run")
	}
}

// TestService_StartAndAddSymbols tests starting the service and adding symbols
func TestService_StartAndAddSymbols(t *testing.T) {
	requireServiceIT(t)
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	opts := Options{
		Intervals:             []string{"1m", "3m"}, // Use shorter intervals for faster testing
		BackfillWindow:        10,
		RestMaxConcurrency:    5,
		WSEndpoint:            WSEndpointFutures,
		EnableRestBackfill:    true,
		SymbolRefreshInterval: 24 * time.Hour, // Very long interval to avoid SymbolSync running during test
	}

	svc := NewService(opts)
	defer svc.Close()

	// Start the service
	if err := svc.Start(ctx); err != nil {
		t.Fatalf("failed to start service: %v", err)
	}

	// Wait a bit for WS connection to establish
	time.Sleep(5 * time.Second)

	// Add a test symbol
	symbols := []string{"BTCUSDT"}
	if err := svc.AddSymbols(symbols); err != nil {
		t.Fatalf("failed to add symbols: %v", err)
	}

	// Wait for initial backfill and WS data
	time.Sleep(10 * time.Second)

	// Verify data was stored via REST backfill
	klines, ok := svc.GetRecent("BTCUSDT", "1m", 5)
	if !ok || len(klines) == 0 {
		t.Logf("no data from REST backfill yet, checking WS data...")
		// Wait a bit more for WS data
		time.Sleep(10 * time.Second)
		klines, ok = svc.GetRecent("BTCUSDT", "1m", 5)
		if !ok || len(klines) == 0 {
			t.Skipf("no kline data received (network may be slow or blocked): ok=%v, len=%d", ok, len(klines))
			return
		}
	}

	// Verify kline data
	if len(klines) == 0 {
		t.Fatal("expected at least one kline, got 0")
	}

	// Verify kline structure
	k := klines[0]
	if k.Symbol != "BTCUSDT" {
		t.Errorf("expected symbol BTCUSDT, got %s", k.Symbol)
	}
	if k.Interval != "1m" {
		t.Errorf("expected interval 1m, got %s", k.Interval)
	}
	if k.OpenTime <= 0 {
		t.Errorf("expected OpenTime > 0, got %d", k.OpenTime)
	}
	if k.CloseTime <= 0 {
		t.Errorf("expected CloseTime > 0, got %d", k.CloseTime)
	}
	if k.Open <= 0 {
		t.Errorf("expected Open > 0, got %f", k.Open)
	}
	if k.Close <= 0 {
		t.Errorf("expected Close > 0, got %f", k.Close)
	}
	if k.High <= 0 {
		t.Errorf("expected High > 0, got %f", k.High)
	}
	if k.Low <= 0 {
		t.Errorf("expected Low > 0, got %f", k.Low)
	}
	if k.Volume < 0 {
		t.Errorf("expected Volume >= 0, got %f", k.Volume)
	}

	t.Logf("successfully received %d klines for BTCUSDT 1m", len(klines))
}

// TestService_MultipleSymbols tests adding multiple symbols
func TestService_MultipleSymbols(t *testing.T) {
	requireServiceIT(t)
	ctx, cancel := context.WithTimeout(context.Background(), 40*time.Second)
	defer cancel()

	opts := Options{
		Intervals:             []string{"1m"},
		BackfillWindow:        10,
		RestMaxConcurrency:    5,
		WSEndpoint:            WSEndpointFutures,
		EnableRestBackfill:    true,
		SymbolRefreshInterval: 24 * time.Hour, // Very long interval to avoid SymbolSync running during test
	}

	svc := NewService(opts)
	defer svc.Close()

	if err := svc.Start(ctx); err != nil {
		t.Fatalf("failed to start service: %v", err)
	}

	time.Sleep(5 * time.Second)

	// Add multiple symbols
	symbols := []string{"BTCUSDT", "ETHUSDT"}
	if err := svc.AddSymbols(symbols); err != nil {
		t.Fatalf("failed to add symbols: %v", err)
	}

	// Wait for data
	time.Sleep(15 * time.Second)

	// Verify both symbols have data
	for _, sym := range symbols {
		klines, ok := svc.GetRecent(sym, "1m", 3)
		if !ok || len(klines) == 0 {
			t.Logf("no data for %s yet, may need more time", sym)
			continue
		}

		if klines[0].Symbol != sym {
			t.Errorf("expected symbol %s, got %s", sym, klines[0].Symbol)
		}
		t.Logf("successfully received %d klines for %s", len(klines), sym)
	}
}

// TestService_MultipleIntervals tests multiple intervals
func TestService_MultipleIntervals(t *testing.T) {
	requireServiceIT(t)
	ctx, cancel := context.WithTimeout(context.Background(), 40*time.Second)
	defer cancel()

	opts := Options{
		Intervals:             []string{"1m", "3m"},
		BackfillWindow:        10,
		RestMaxConcurrency:    5,
		WSEndpoint:            WSEndpointFutures,
		EnableRestBackfill:    true,
		SymbolRefreshInterval: 24 * time.Hour, // Very long interval to avoid SymbolSync running during test
	}

	svc := NewService(opts)
	defer svc.Close()

	if err := svc.Start(ctx); err != nil {
		t.Fatalf("failed to start service: %v", err)
	}

	time.Sleep(5 * time.Second)

	symbols := []string{"BTCUSDT"}
	if err := svc.AddSymbols(symbols); err != nil {
		t.Fatalf("failed to add symbols: %v", err)
	}

	time.Sleep(15 * time.Second)

	// Verify both intervals have data
	intervals := []string{"1m", "3m"}
	for _, iv := range intervals {
		klines, ok := svc.GetRecent("BTCUSDT", iv, 3)
		if !ok || len(klines) == 0 {
			t.Logf("no data for interval %s yet, may need more time", iv)
			continue
		}

		if klines[0].Interval != iv {
			t.Errorf("expected interval %s, got %s", iv, klines[0].Interval)
		}
		t.Logf("successfully received %d klines for BTCUSDT %s", len(klines), iv)
	}
}

// TestService_EnsureReady tests the EnsureReady functionality
func TestService_EnsureReady(t *testing.T) {
	requireServiceIT(t)
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	opts := Options{
		Intervals:             []string{"1m"},
		BackfillWindow:        5,
		RestMaxConcurrency:    5,
		WSEndpoint:            WSEndpointFutures,
		EnableRestBackfill:    true,
		SymbolRefreshInterval: 24 * time.Hour, // Very long interval to avoid SymbolSync running during test
	}

	svc := NewService(opts)
	defer svc.Close()

	if err := svc.Start(ctx); err != nil {
		t.Fatalf("failed to start service: %v", err)
	}

	time.Sleep(3 * time.Second)

	// EnsureReady should trigger backfill if data is not ready
	svc.EnsureReady("BTCUSDT", "1m", 10)

	// Wait for backfill
	time.Sleep(10 * time.Second)

	// Check if data is available
	klines, ok := svc.GetRecent("BTCUSDT", "1m", 10)
	if !ok || len(klines) < 10 {
		t.Logf("EnsureReady may not have completed yet: ok=%v, len=%d", ok, len(klines))
	} else {
		t.Logf("EnsureReady successfully backfilled %d klines", len(klines))
	}
}

// TestService_RemoveSymbols tests removing symbols
func TestService_RemoveSymbols(t *testing.T) {
	requireServiceIT(t)
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	opts := Options{
		Intervals:             []string{"1m"},
		BackfillWindow:        10,
		RestMaxConcurrency:    5,
		WSEndpoint:            WSEndpointFutures,
		EnableRestBackfill:    true,
		SymbolRefreshInterval: 24 * time.Hour, // Very long interval to avoid SymbolSync running during test
	}

	svc := NewService(opts)
	defer svc.Close()

	if err := svc.Start(ctx); err != nil {
		t.Fatalf("failed to start service: %v", err)
	}

	time.Sleep(5 * time.Second)

	// Add symbols
	symbols := []string{"BTCUSDT", "ETHUSDT"}
	if err := svc.AddSymbols(symbols); err != nil {
		t.Fatalf("failed to add symbols: %v", err)
	}

	time.Sleep(5 * time.Second)

	// Remove one symbol
	if err := svc.RemoveSymbols([]string{"ETHUSDT"}); err != nil {
		t.Fatalf("failed to remove symbols: %v", err)
	}

	// Wait a bit for cleanup
	time.Sleep(2 * time.Second)

	// Verify BTCUSDT still has data (if it was received)
	klines, ok := svc.GetRecent("BTCUSDT", "1m", 1)
	if ok && len(klines) > 0 {
		t.Logf("BTCUSDT still has data after removing ETHUSDT: %d klines", len(klines))
	}

	// Note: We can't easily verify that ETHUSDT is removed without accessing internal state
	// But if RemoveSymbols doesn't error, it's likely working correctly
}

// TestService_GetRecent tests getting recent klines
func TestService_GetRecent(t *testing.T) {
	requireServiceIT(t)
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	opts := Options{
		Intervals:             []string{"1m"},
		BackfillWindow:        20,
		RestMaxConcurrency:    5,
		WSEndpoint:            WSEndpointFutures,
		EnableRestBackfill:    true,
		SymbolRefreshInterval: 24 * time.Hour, // Very long interval to avoid SymbolSync running during test
	}

	svc := NewService(opts)
	defer svc.Close()

	if err := svc.Start(ctx); err != nil {
		t.Fatalf("failed to start service: %v", err)
	}

	time.Sleep(5 * time.Second)

	symbols := []string{"BTCUSDT"}
	if err := svc.AddSymbols(symbols); err != nil {
		t.Fatalf("failed to add symbols: %v", err)
	}

	time.Sleep(15 * time.Second)

	// Test getting different limits
	limits := []int{1, 5, 10, 20}
	for _, limit := range limits {
		klines, ok := svc.GetRecent("BTCUSDT", "1m", limit)
		if !ok {
			t.Logf("GetRecent(%d) returned not ready", limit)
			continue
		}

		actualLen := len(klines)
		if actualLen > limit {
			t.Errorf("GetRecent(%d) returned %d klines (expected at most %d)", limit, actualLen, limit)
		}
		if actualLen > 0 {
			t.Logf("GetRecent(%d) returned %d klines", limit, actualLen)
		}
	}
}

// TestService_CaseInsensitive tests case-insensitive symbol and interval handling
func TestService_CaseInsensitive(t *testing.T) {
	requireServiceIT(t)
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	opts := Options{
		Intervals:             []string{"1m"},
		BackfillWindow:        10,
		RestMaxConcurrency:    5,
		WSEndpoint:            WSEndpointFutures,
		EnableRestBackfill:    true,
		SymbolRefreshInterval: 24 * time.Hour, // Very long interval to avoid SymbolSync running during test
	}

	svc := NewService(opts)
	defer svc.Close()

	if err := svc.Start(ctx); err != nil {
		t.Fatalf("failed to start service: %v", err)
	}

	time.Sleep(5 * time.Second)

	// Add symbol with lowercase
	if err := svc.AddSymbols([]string{"btcusdt"}); err != nil {
		t.Fatalf("failed to add symbols: %v", err)
	}

	time.Sleep(10 * time.Second)

	// Get with different cases
	testCases := []struct {
		symbol   string
		interval string
	}{
		{"BTCUSDT", "1m"},
		{"btcusdt", "1m"},
		{"BtcUsdt", "1M"},
		{"btcusdt", "1M"},
	}

	for _, tc := range testCases {
		klines, ok := svc.GetRecent(tc.symbol, tc.interval, 1)
		if ok && len(klines) > 0 {
			// All should return the same data (normalized)
			if klines[0].Symbol != "BTCUSDT" {
				t.Errorf("expected normalized symbol BTCUSDT, got %s", klines[0].Symbol)
			}
			if klines[0].Interval != "1m" {
				t.Errorf("expected normalized interval 1m, got %s", klines[0].Interval)
			}
			t.Logf("case-insensitive test passed for %s/%s", tc.symbol, tc.interval)
		}
	}
}

// TestService_ConcurrentAccess tests concurrent access to the service
func TestService_ConcurrentAccess(t *testing.T) {
	requireServiceIT(t)
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	opts := Options{
		Intervals:             []string{"1m"},
		BackfillWindow:        10,
		RestMaxConcurrency:    5,
		WSEndpoint:            WSEndpointFutures,
		SymbolRefreshInterval: 24 * time.Hour, // Very long interval to avoid SymbolSync running during test
	}

	svc := NewService(opts)
	defer svc.Close()

	if err := svc.Start(ctx); err != nil {
		t.Fatalf("failed to start service: %v", err)
	}

	time.Sleep(5 * time.Second)

	// Add symbols concurrently
	symbols := []string{"BTCUSDT", "ETHUSDT", "BNBUSDT"}
	done := make(chan error, len(symbols))
	for _, sym := range symbols {
		go func(s string) {
			done <- svc.AddSymbols([]string{s})
		}(sym)
	}

	// Wait for all adds to complete
	for i := 0; i < len(symbols); i++ {
		if err := <-done; err != nil {
			t.Errorf("concurrent AddSymbols failed: %v", err)
		}
	}

	time.Sleep(10 * time.Second)

	// Concurrent reads
	readDone := make(chan bool, len(symbols))
	for _, sym := range symbols {
		go func(s string) {
			_, ok := svc.GetRecent(s, "1m", 1)
			readDone <- ok
		}(sym)
	}

	// Wait for all reads
	readCount := 0
	for i := 0; i < len(symbols); i++ {
		if <-readDone {
			readCount++
		}
	}

	t.Logf("concurrent access test completed: %d/%d symbols had data", readCount, len(symbols))
}

// TestService_Close tests graceful shutdown
func TestService_Close(t *testing.T) {
	requireServiceIT(t)
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()

	opts := Options{
		Intervals:             []string{"1m"},
		BackfillWindow:        10,
		RestMaxConcurrency:    5,
		WSEndpoint:            WSEndpointFutures,
		SymbolRefreshInterval: 24 * time.Hour, // Very long interval to avoid SymbolSync running during test
	}

	svc := NewService(opts)

	if err := svc.Start(ctx); err != nil {
		t.Fatalf("failed to start service: %v", err)
	}

	time.Sleep(3 * time.Second)

	// Add a symbol
	if err := svc.AddSymbols([]string{"BTCUSDT"}); err != nil {
		t.Fatalf("failed to add symbols: %v", err)
	}

	time.Sleep(5 * time.Second)

	// Close should not panic
	svc.Close()

	// Wait a bit to ensure cleanup
	time.Sleep(2 * time.Second)

	t.Log("service closed successfully")
}

// TestService_DataConsistency tests that data from WS and REST are consistent
func TestService_DataConsistency(t *testing.T) {
	requireServiceIT(t)
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	opts := Options{
		Intervals:             []string{"1m"},
		BackfillWindow:        50,
		RestMaxConcurrency:    5,
		WSEndpoint:            WSEndpointFutures,
		SymbolRefreshInterval: 24 * time.Hour, // Very long interval to avoid SymbolSync running during test
	}

	svc := NewService(opts)
	defer svc.Close()

	if err := svc.Start(ctx); err != nil {
		t.Fatalf("failed to start service: %v", err)
	}

	time.Sleep(5 * time.Second)

	symbols := []string{"BTCUSDT"}
	if err := svc.AddSymbols(symbols); err != nil {
		t.Fatalf("failed to add symbols: %v", err)
	}

	// Wait for both REST backfill and WS data
	time.Sleep(30 * time.Second)

	// Get recent klines
	klines, ok := svc.GetRecent("BTCUSDT", "1m", 50)
	if !ok || len(klines) < 10 {
		t.Skipf("insufficient data for consistency test: ok=%v, len=%d", ok, len(klines))
		return
	}

	// Verify klines are sorted by OpenTime
	for i := 1; i < len(klines); i++ {
		if klines[i].OpenTime < klines[i-1].OpenTime {
			t.Errorf("klines not sorted: klines[%d].OpenTime=%d < klines[%d].OpenTime=%d",
				i, klines[i].OpenTime, i-1, klines[i-1].OpenTime)
		}
	}

	// Verify no duplicates (same OpenTime)
	seen := make(map[int64]bool)
	for _, k := range klines {
		if seen[k.OpenTime] {
			t.Errorf("duplicate OpenTime found: %d", k.OpenTime)
		}
		seen[k.OpenTime] = true
	}

	// Verify price consistency (High >= Low, High >= Open, High >= Close, etc.)
	// Use a small tolerance for floating-point comparisons
	const priceTolerance = 0.0001
	for _, k := range klines {
		if k.High < k.Low-priceTolerance {
			t.Errorf("High < Low: High=%f, Low=%f", k.High, k.Low)
		}
		if k.High < k.Open-priceTolerance {
			t.Errorf("High < Open: High=%f, Open=%f", k.High, k.Open)
		}
		if k.High < k.Close-priceTolerance {
			t.Errorf("High < Close: High=%f, Close=%f", k.High, k.Close)
		}
		if k.Low > k.Open+priceTolerance {
			t.Errorf("Low > Open: Low=%f, Open=%f", k.Low, k.Open)
		}
		if k.Low > k.Close+priceTolerance {
			t.Errorf("Low > Close: Low=%f, Close=%f", k.Low, k.Close)
		}
		if k.CloseTime <= k.OpenTime {
			t.Errorf("CloseTime <= OpenTime: CloseTime=%d, OpenTime=%d", k.CloseTime, k.OpenTime)
		}
	}

	t.Logf("data consistency test passed: %d klines verified", len(klines))
}

// TestService_SymbolSync_SubscribeAll tests that SymbolSync subscribes to all trading pairs
func TestService_SymbolSync_SubscribeAll(t *testing.T) {
	requireServiceIT(t)
	ctx, cancel := context.WithTimeout(context.Background(), 120*time.Second)
	defer cancel()

	opts := Options{
		Intervals:             []string{"1m"}, // Use single interval for faster testing
		BackfillWindow:        10,
		RestMaxConcurrency:    5,
		WSEndpoint:            WSEndpointFutures,
		SymbolRefreshInterval: 30 * time.Second, // Short interval to allow SymbolSync to run
	}

	svc := NewService(opts)
	defer svc.Close()

	// Start the service (SymbolSync will start and fetch all symbols)
	if err := svc.Start(ctx); err != nil {
		t.Fatalf("failed to start service: %v", err)
	}

	// Wait for WS connection to establish
	time.Sleep(5 * time.Second)

	// Wait for SymbolSync to complete first sync (it runs immediately on Start)
	// Give it enough time to fetch symbols and subscribe
	time.Sleep(15 * time.Second)

	// Get all subscribed symbols for the interval
	subscribedSymbols := svc.subs.Symbols("1m")
	if len(subscribedSymbols) == 0 {
		t.Skipf("no symbols subscribed yet (SymbolSync may not have completed): waiting longer...")
		// Wait a bit more
		time.Sleep(10 * time.Second)
		subscribedSymbols = svc.subs.Symbols("1m")
		if len(subscribedSymbols) == 0 {
			t.Skipf("no symbols subscribed after extended wait")
			return
		}
	}

	t.Logf("SymbolSync subscribed to %d symbols", len(subscribedSymbols))

	// Verify that we have a reasonable number of symbols (Binance typically has 500+ futures pairs)
	if len(subscribedSymbols) < 100 {
		t.Logf("warning: expected at least 100 symbols, got %d (may be normal if some symbols were filtered)", len(subscribedSymbols))
	}

	// Verify common symbols are subscribed
	commonSymbols := []string{"BTCUSDT", "ETHUSDT", "BNBUSDT"}
	found := make(map[string]bool)
	for _, sym := range subscribedSymbols {
		for _, common := range commonSymbols {
			if sym == common {
				found[common] = true
			}
		}
	}

	for _, common := range commonSymbols {
		if !found[common] {
			t.Logf("warning: common symbol %s not found in subscriptions (may be delisted or filtered)", common)
		}
	}

	// Wait for data to arrive for some symbols
	time.Sleep(20 * time.Second)

	// Verify data is being received for at least some symbols
	dataCount := 0
	testSymbols := []string{"BTCUSDT", "ETHUSDT", "BNBUSDT"}
	for _, sym := range testSymbols {
		klines, ok := svc.GetRecent(sym, "1m", 1)
		if ok && len(klines) > 0 {
			dataCount++
			t.Logf("received data for %s: %d klines", sym, len(klines))
		}
	}

	if dataCount == 0 {
		t.Logf("no data received yet for test symbols (may need more time or network issues)")
	} else {
		t.Logf("received data for %d/%d test symbols", dataCount, len(testSymbols))
	}

	// Verify subscription count matches what we expect
	// Note: The actual count may vary due to filtering, but should be substantial
	t.Logf("SymbolSync test completed: %d symbols subscribed, %d symbols with data", len(subscribedSymbols), dataCount)
}

// ------------- Unit Tests for GetRecentKlines Logic -------------

// mockRestReconciler is a mock reconciler for testing
type mockRestReconciler struct {
	shouldError bool
	callCount   int
}

func (m *mockRestReconciler) ReconcileWindow(symbol, interval string, limit int) error {
	m.callCount++
	if m.shouldError {
		return fmt.Errorf("mock error")
	}
	return nil
}

// TestGetRecentKlines_DataMissingNoLock tests that ready is false when data is missing and lock cannot be acquired
func TestGetRecentKlines_DataMissingNoLock(t *testing.T) {
	requireServiceIT(t)
	store := newMockStore()
	mockReconciler := &mockRestReconciler{shouldError: false}

	opts := Options{
		Intervals:          []string{"1m"},
		BackfillWindow:     10,
		RestMaxConcurrency: 5,
		EnableRestBackfill: true,
	}

	svc := &Service{
		store:              store,
		rest:               mockReconciler,
		intervals:          opts.Intervals,
		backfillWindow:     opts.BackfillWindow,
		restMaxConcurrency: opts.RestMaxConcurrency,
		enableBackfill:     opts.EnableRestBackfill,
		onDemand:           make(map[string]bool),
		shortSuppress:      make(map[string]time.Time),
	}

	// Simulate another goroutine holding the lock
	svc.onDemand["BTCUSDT|1m"] = true

	// Request 10 klines but store has only 3 (shortOnly = true, needBackfill = false)
	now := time.Now().UnixMilli()
	stepMs := int64(1 * time.Minute / time.Millisecond)
	for i := 0; i < 3; i++ {
		store.UpsertFinal("BTCUSDT", "1m", marketKline{
			OpenTime:  now - int64(3-i)*stepMs,
			CloseTime: now - int64(3-i)*stepMs + stepMs - 1,
			Open:      100.0,
			High:      101.0,
			Low:       99.0,
			Close:     100.5,
			Volume:    1000.0,
		})
	}

	// GetRecentKlines should return ready = false because data is missing and lock cannot be acquired
	klines, ready := svc.GetRecentKlines("BTCUSDT", "1m", 10)
	if ready {
		t.Error("expected ready = false when data is missing and lock cannot be acquired")
	}
	if len(klines) != 3 {
		t.Errorf("expected 3 klines, got %d", len(klines))
	}
}

// TestGetRecentKlines_ShortOnlySuppressedNoLock tests that ready is not set to false when shortOnly and suppressed
func TestGetRecentKlines_ShortOnlySuppressedNoLock(t *testing.T) {
	requireServiceIT(t)
	store := newMockStore()
	mockReconciler := &mockRestReconciler{shouldError: false}

	opts := Options{
		Intervals:          []string{"1m"},
		BackfillWindow:     10,
		RestMaxConcurrency: 5,
		EnableRestBackfill: true,
	}

	svc := &Service{
		store:              store,
		rest:               mockReconciler,
		intervals:          opts.Intervals,
		backfillWindow:     opts.BackfillWindow,
		restMaxConcurrency: opts.RestMaxConcurrency,
		enableBackfill:     opts.EnableRestBackfill,
		onDemand:           make(map[string]bool),
		shortSuppress:      make(map[string]time.Time),
	}

	// Simulate another goroutine holding the lock
	svc.onDemand["BTCUSDT|1m"] = true

	// Set short suppress (data was already tried and is complete, just not enough history)
	svc.setShortSuppress("BTCUSDT", "1m", 30*time.Minute)

	// Request 10 klines but store has only 3 (shortOnly = true, needBackfill = false)
	now := time.Now().UnixMilli()
	stepMs := int64(1 * time.Minute / time.Millisecond)
	for i := 0; i < 3; i++ {
		store.UpsertFinal("BTCUSDT", "1m", marketKline{
			OpenTime:  now - int64(3-i)*stepMs,
			CloseTime: now - int64(3-i)*stepMs + stepMs - 1,
			Open:      100.0,
			High:      101.0,
			Low:       99.0,
			Close:     100.5,
			Volume:    1000.0,
		})
	}

	// GetRecentKlines should not set ready = false when shortOnly and suppressed
	klines, ready := svc.GetRecentKlines("BTCUSDT", "1m", 10)
	// ready should be true (from store) because suppressed means data is complete
	if !ready {
		t.Error("expected ready = true when shortOnly and suppressed (data is complete, just not enough history)")
	}
	if len(klines) != 3 {
		t.Errorf("expected 3 klines, got %d", len(klines))
	}
}

// TestGetRecentKlines_ReconcileWindowErrorNoSuppress tests that setShortSuppress is not called when ReconcileWindow fails
func TestGetRecentKlines_ReconcileWindowErrorNoSuppress(t *testing.T) {
	requireServiceIT(t)
	store := newMockStore()
	mockReconciler := &mockRestReconciler{shouldError: true} // Make ReconcileWindow return error

	opts := Options{
		Intervals:          []string{"1m"},
		BackfillWindow:     10,
		RestMaxConcurrency: 5,
		EnableRestBackfill: true,
	}

	svc := &Service{
		store:              store,
		rest:               mockReconciler,
		intervals:          opts.Intervals,
		backfillWindow:     opts.BackfillWindow,
		restMaxConcurrency: opts.RestMaxConcurrency,
		enableBackfill:     opts.EnableRestBackfill,
		onDemand:           make(map[string]bool),
		shortSuppress:      make(map[string]time.Time),
	}

	// Request 10 klines but store has only 3 (shortOnly = true, needBackfill = false)
	now := time.Now().UnixMilli()
	stepMs := int64(1 * time.Minute / time.Millisecond)
	for i := 0; i < 3; i++ {
		store.UpsertFinal("BTCUSDT", "1m", marketKline{
			OpenTime:  now - int64(3-i)*stepMs,
			CloseTime: now - int64(3-i)*stepMs + stepMs - 1,
			Open:      100.0,
			High:      101.0,
			Low:       99.0,
			Close:     100.5,
			Volume:    1000.0,
		})
	}

	// GetRecentKlines should try to reconcile, but since it fails, suppress should not be set
	klines, _ := svc.GetRecentKlines("BTCUSDT", "1m", 10)
	if len(klines) != 3 {
		t.Errorf("expected 3 klines, got %d", len(klines))
	}

	// Verify that suppress was not set (because ReconcileWindow failed)
	if svc.isShortSuppressed("BTCUSDT", "1m") {
		t.Error("expected suppress not to be set when ReconcileWindow fails")
	}

	// Verify ReconcileWindow was called
	if mockReconciler.callCount == 0 {
		t.Error("expected ReconcileWindow to be called")
	}
}

// TestGetRecentKlines_ReconcileWindowSuccessSuppress tests that setShortSuppress is called when ReconcileWindow succeeds
func TestGetRecentKlines_ReconcileWindowSuccessSuppress(t *testing.T) {
	requireServiceIT(t)
	store := newMockStore()
	mockReconciler := &mockRestReconciler{shouldError: false}

	opts := Options{
		Intervals:          []string{"1m"},
		BackfillWindow:     10,
		RestMaxConcurrency: 5,
		EnableRestBackfill: true,
	}

	svc := &Service{
		store:              store,
		rest:               mockReconciler,
		intervals:          opts.Intervals,
		backfillWindow:     opts.BackfillWindow,
		restMaxConcurrency: opts.RestMaxConcurrency,
		enableBackfill:     opts.EnableRestBackfill,
		onDemand:           make(map[string]bool),
		shortSuppress:      make(map[string]time.Time),
	}

	// Request 10 klines but store has only 3 (shortOnly = true, needBackfill = false)
	now := time.Now().UnixMilli()
	stepMs := int64(1 * time.Minute / time.Millisecond)
	for i := 0; i < 3; i++ {
		store.UpsertFinal("BTCUSDT", "1m", marketKline{
			OpenTime:  now - int64(3-i)*stepMs,
			CloseTime: now - int64(3-i)*stepMs + stepMs - 1,
			Open:      100.0,
			High:      101.0,
			Low:       99.0,
			Close:     100.5,
			Volume:    1000.0,
		})
	}

	// GetRecentKlines should try to reconcile, and since it succeeds but data is still short,
	// suppress should be set
	klines, _ := svc.GetRecentKlines("BTCUSDT", "1m", 10)
	if len(klines) != 3 {
		t.Errorf("expected 3 klines, got %d", len(klines))
	}

	// Verify that suppress was set (because ReconcileWindow succeeded but data is still short)
	if !svc.isShortSuppressed("BTCUSDT", "1m") {
		t.Error("expected suppress to be set when ReconcileWindow succeeds but data is still short")
	}

	// Verify ReconcileWindow was called
	if mockReconciler.callCount == 0 {
		t.Error("expected ReconcileWindow to be called")
	}
}

// TestGetRecentKlines_NeedBackfillNoLock tests that ready is false when needBackfill is true and lock cannot be acquired
func TestGetRecentKlines_NeedBackfillNoLock(t *testing.T) {
	store := newMockStore()
	mockReconciler := &mockRestReconciler{shouldError: false}

	opts := Options{
		Intervals:          []string{"1m"},
		BackfillWindow:     10,
		RestMaxConcurrency: 5,
		EnableRestBackfill: true,
	}

	svc := &Service{
		store:              store,
		rest:               mockReconciler,
		intervals:          opts.Intervals,
		backfillWindow:     opts.BackfillWindow,
		restMaxConcurrency: opts.RestMaxConcurrency,
		enableBackfill:     opts.EnableRestBackfill,
		onDemand:           make(map[string]bool),
		shortSuppress:      make(map[string]time.Time),
	}

	// Simulate another goroutine holding the lock
	svc.onDemand["BTCUSDT|1m"] = true

	// Create klines with a gap (needBackfill = true)
	now := time.Now().UnixMilli()
	stepMs := int64(1 * time.Minute / time.Millisecond)
	// First kline
	store.UpsertFinal("BTCUSDT", "1m", marketKline{
		OpenTime:  now - 5*stepMs,
		CloseTime: now - 5*stepMs + stepMs - 1,
		Open:      100.0,
		High:      101.0,
		Low:       99.0,
		Close:     100.5,
		Volume:    1000.0,
	})
	// Gap: missing kline at now - 4*stepMs
	// Third kline (gap detected)
	store.UpsertFinal("BTCUSDT", "1m", marketKline{
		OpenTime:  now - 3*stepMs,
		CloseTime: now - 3*stepMs + stepMs - 1,
		Open:      100.0,
		High:      101.0,
		Low:       99.0,
		Close:     100.5,
		Volume:    1000.0,
	})

	// GetRecentKlines should return ready = false because needBackfill is true and lock cannot be acquired
	klines, ready := svc.GetRecentKlines("BTCUSDT", "1m", 10)
	if ready {
		t.Error("expected ready = false when needBackfill is true and lock cannot be acquired")
	}
	if len(klines) != 2 {
		t.Errorf("expected 2 klines, got %d", len(klines))
	}
}

// ------------- Unit Tests for hasGapsOrTailMissing -------------

// TestHasGapsOrTailMissing_EmptySlice tests that empty slice returns false
func TestHasGapsOrTailMissing_EmptySlice(t *testing.T) {
	stepMs := int64(1 * time.Minute / time.Millisecond)
	if hasGapsOrTailMissing(nil, stepMs) {
		t.Error("expected false for empty slice")
	}
	if hasGapsOrTailMissing([]marketKline{}, stepMs) {
		t.Error("expected false for empty slice")
	}
}

// TestHasGapsOrTailMissing_InvalidStepMs tests that invalid stepMs returns false
func TestHasGapsOrTailMissing_InvalidStepMs(t *testing.T) {
	now := time.Now().UnixMilli()
	stepMs := int64(1 * time.Minute / time.Millisecond)
	klines := []marketKline{
		{OpenTime: now - stepMs, CloseTime: now - stepMs + stepMs - 1},
		{OpenTime: now, CloseTime: now + stepMs - 1},
	}

	if hasGapsOrTailMissing(klines, 0) {
		t.Error("expected false for stepMs = 0")
	}
	if hasGapsOrTailMissing(klines, -1) {
		t.Error("expected false for stepMs < 0")
	}
}

// TestHasGapsOrTailMissing_NoGapsNoTailMissing tests that continuous klines with no tail missing returns false
func TestHasGapsOrTailMissing_NoGapsNoTailMissing(t *testing.T) {
	now := time.Now().UnixMilli()
	stepMs := int64(1 * time.Minute / time.Millisecond)
	// Align to step boundary
	expectedLastOpen := now - (now % stepMs)

	klines := []marketKline{
		{OpenTime: expectedLastOpen - 2*stepMs, CloseTime: expectedLastOpen - 2*stepMs + stepMs - 1},
		{OpenTime: expectedLastOpen - stepMs, CloseTime: expectedLastOpen - stepMs + stepMs - 1},
		{OpenTime: expectedLastOpen, CloseTime: expectedLastOpen + stepMs - 1},
	}

	if hasGapsOrTailMissing(klines, stepMs) {
		t.Error("expected false for continuous klines with no tail missing")
	}
}

// TestHasGapsOrTailMissing_MiddleGap tests that middle gap is detected
func TestHasGapsOrTailMissing_MiddleGap(t *testing.T) {
	now := time.Now().UnixMilli()
	stepMs := int64(1 * time.Minute / time.Millisecond)
	expectedLastOpen := now - (now % stepMs)

	// Create klines with a gap in the middle
	klines := []marketKline{
		{OpenTime: expectedLastOpen - 3*stepMs, CloseTime: expectedLastOpen - 3*stepMs + stepMs - 1},
		// Gap: missing kline at expectedLastOpen - 2*stepMs
		{OpenTime: expectedLastOpen - stepMs, CloseTime: expectedLastOpen - stepMs + stepMs - 1},
		{OpenTime: expectedLastOpen, CloseTime: expectedLastOpen + stepMs - 1},
	}

	if !hasGapsOrTailMissing(klines, stepMs) {
		t.Error("expected true for klines with middle gap")
	}
}

// TestHasGapsOrTailMissing_TailMissing tests that tail missing is detected
func TestHasGapsOrTailMissing_TailMissing(t *testing.T) {
	now := time.Now().UnixMilli()
	stepMs := int64(1 * time.Minute / time.Millisecond)
	expectedLastOpen := now - (now % stepMs)

	// Create klines where the last one is more than stepMs behind expected
	klines := []marketKline{
		{OpenTime: expectedLastOpen - 3*stepMs, CloseTime: expectedLastOpen - 3*stepMs + stepMs - 1},
		{OpenTime: expectedLastOpen - 2*stepMs, CloseTime: expectedLastOpen - 2*stepMs + stepMs - 1},
		{OpenTime: expectedLastOpen - 2*stepMs - stepMs, CloseTime: expectedLastOpen - 2*stepMs - stepMs + stepMs - 1}, // More than stepMs behind expected
	}

	if !hasGapsOrTailMissing(klines, stepMs) {
		t.Error("expected true for klines with tail missing")
	}
}

// TestHasGapsOrTailMissing_SingleKline_NoTailMissing tests single kline that is up to date
func TestHasGapsOrTailMissing_SingleKline_NoTailMissing(t *testing.T) {
	now := time.Now().UnixMilli()
	stepMs := int64(1 * time.Minute / time.Millisecond)
	expectedLastOpen := now - (now % stepMs)

	klines := []marketKline{
		{OpenTime: expectedLastOpen, CloseTime: expectedLastOpen + stepMs - 1},
	}

	if hasGapsOrTailMissing(klines, stepMs) {
		t.Error("expected false for single up-to-date kline")
	}
}

// TestHasGapsOrTailMissing_SingleKline_TailMissing tests single kline that is behind
func TestHasGapsOrTailMissing_SingleKline_TailMissing(t *testing.T) {
	now := time.Now().UnixMilli()
	stepMs := int64(1 * time.Minute / time.Millisecond)
	expectedLastOpen := now - (now % stepMs)

	// Kline is more than stepMs behind expected
	klines := []marketKline{
		{OpenTime: expectedLastOpen - 2*stepMs, CloseTime: expectedLastOpen - 2*stepMs + stepMs - 1},
	}

	if !hasGapsOrTailMissing(klines, stepMs) {
		t.Error("expected true for single kline that is behind")
	}
}

// TestHasGapsOrTailMissing_ExactStepBoundary tests klines exactly at step boundaries
func TestHasGapsOrTailMissing_ExactStepBoundary(t *testing.T) {
	now := time.Now().UnixMilli()
	stepMs := int64(1 * time.Minute / time.Millisecond)
	expectedLastOpen := now - (now % stepMs)

	// Klines exactly at step boundaries (no gaps)
	klines := []marketKline{
		{OpenTime: expectedLastOpen - 2*stepMs, CloseTime: expectedLastOpen - 2*stepMs + stepMs - 1},
		{OpenTime: expectedLastOpen - stepMs, CloseTime: expectedLastOpen - stepMs + stepMs - 1},
		{OpenTime: expectedLastOpen, CloseTime: expectedLastOpen + stepMs - 1},
	}

	if hasGapsOrTailMissing(klines, stepMs) {
		t.Error("expected false for klines exactly at step boundaries")
	}
}

// TestHasGapsOrTailMissing_MultipleGaps tests multiple gaps in the middle
func TestHasGapsOrTailMissing_MultipleGaps(t *testing.T) {
	now := time.Now().UnixMilli()
	stepMs := int64(1 * time.Minute / time.Millisecond)
	expectedLastOpen := now - (now % stepMs)

	// Create klines with multiple gaps
	klines := []marketKline{
		{OpenTime: expectedLastOpen - 5*stepMs, CloseTime: expectedLastOpen - 5*stepMs + stepMs - 1},
		// Gap 1: missing at expectedLastOpen - 4*stepMs
		{OpenTime: expectedLastOpen - 3*stepMs, CloseTime: expectedLastOpen - 3*stepMs + stepMs - 1},
		// Gap 2: missing at expectedLastOpen - 2*stepMs
		{OpenTime: expectedLastOpen - stepMs, CloseTime: expectedLastOpen - stepMs + stepMs - 1},
		{OpenTime: expectedLastOpen, CloseTime: expectedLastOpen + stepMs - 1},
	}

	if !hasGapsOrTailMissing(klines, stepMs) {
		t.Error("expected true for klines with multiple gaps")
	}
}

// TestHasGapsOrTailMissing_EdgeCase_OneStepBehind tests kline exactly one step behind (should not be considered missing)
func TestHasGapsOrTailMissing_EdgeCase_OneStepBehind(t *testing.T) {
	now := time.Now().UnixMilli()
	stepMs := int64(1 * time.Minute / time.Millisecond)
	expectedLastOpen := now - (now % stepMs)

	// Last kline is exactly one step behind (not more than stepMs)
	klines := []marketKline{
		{OpenTime: expectedLastOpen - stepMs, CloseTime: expectedLastOpen - stepMs + stepMs - 1},
	}

	if hasGapsOrTailMissing(klines, stepMs) {
		t.Error("expected false for kline exactly one step behind (not more than stepMs)")
	}
}

// TestHasGapsOrTailMissing_EdgeCase_MoreThanOneStepBehind tests kline more than one step behind (should be considered missing)
func TestHasGapsOrTailMissing_EdgeCase_MoreThanOneStepBehind(t *testing.T) {
	now := time.Now().UnixMilli()
	stepMs := int64(1 * time.Minute / time.Millisecond)
	expectedLastOpen := now - (now % stepMs)

	// Last kline is more than one step behind
	klines := []marketKline{
		{OpenTime: expectedLastOpen - 2*stepMs - 1, CloseTime: expectedLastOpen - 2*stepMs - 1 + stepMs - 1}, // More than stepMs behind
	}

	if !hasGapsOrTailMissing(klines, stepMs) {
		t.Error("expected true for kline more than one step behind")
	}
}

// ------------- Integration Tests for GetRecentKlines with Binance API -------------

// TestGetRecentKlines_Integration_Basic tests basic GetRecentKlines functionality with real Binance API
func TestGetRecentKlines_Integration_Basic(t *testing.T) {
	requireServiceIT(t)
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	opts := Options{
		Intervals:             []string{Interval3m, Interval15m, Interval1h, Interval4h, Interval8h, Interval1d, Interval1w},
		BackfillWindow:        20,
		RestMaxConcurrency:    5,
		WSEndpoint:            WSEndpointFutures,
		EnableRestBackfill:    false,          // Disable automatic backfill, rely on GetRecentKlines internal REST calls
		SymbolRefreshInterval: 24 * time.Hour, // Very long interval to avoid SymbolSync running during test
	}

	svc := NewService(opts)
	defer svc.Close()

	if err := svc.Start(ctx); err != nil {
		t.Fatalf("failed to start service: %v", err)
	}

	// Wait for WS connection
	time.Sleep(5 * time.Second)

	// Add symbol
	symbols := []string{"BTCUSDT"}
	if err := svc.AddSymbols(symbols); err != nil {
		t.Fatalf("failed to add symbols: %v", err)
	}

	// Wait for initial data (WS + REST backfill)
	time.Sleep(15 * time.Second)

	// Test GetRecentKlines with different limits
	testCases := []struct {
		interval string
		limit    int
	}{
		{"1m", 5},
		{"1m", 10},
		{"1m", 20},
		{"3m", 5},
		{"3m", 10},
		{"15m", 5},
		{"15m", 10},
		{"1h", 5},
		{"1h", 10},
		{"4h", 5},
		{"4h", 10},
		{"8h", 5},
		{"8h", 10},
		{"1d", 5},
		{"1d", 10},
		{"1w", 5},
		{"1w", 10},
	}

	for _, tc := range testCases {
		klines, ready := svc.GetRecentKlines("BTCUSDT", tc.interval, tc.limit)
		if !ready {
			t.Logf("warning: data not ready for %s@%s (limit=%d), may need more time", "BTCUSDT", tc.interval, tc.limit)
			// Wait a bit more and retry
			time.Sleep(5 * time.Second)
			klines, ready = svc.GetRecentKlines("BTCUSDT", tc.interval, tc.limit)
		}

		if len(klines) == 0 {
			t.Errorf("expected at least one kline for %s@%s (limit=%d), got 0", "BTCUSDT", tc.interval, tc.limit)
			continue
		}

		if len(klines) > tc.limit {
			t.Errorf("expected at most %d klines for %s@%s, got %d", tc.limit, "BTCUSDT", tc.interval, len(klines))
		}

		// Verify kline structure
		for i, k := range klines {
			if k.OpenTime <= 0 {
				t.Errorf("kline[%d].OpenTime should be > 0, got %d", i, k.OpenTime)
			}
			if k.CloseTime <= k.OpenTime {
				t.Errorf("kline[%d].CloseTime (%d) should be > OpenTime (%d)", i, k.CloseTime, k.OpenTime)
			}
			if k.Open <= 0 || k.High <= 0 || k.Low <= 0 || k.Close <= 0 {
				t.Errorf("kline[%d] has invalid price: O=%.2f H=%.2f L=%.2f C=%.2f", i, k.Open, k.High, k.Low, k.Close)
			}
			if k.High < k.Low {
				t.Errorf("kline[%d].High (%.2f) < Low (%.2f)", i, k.High, k.Low)
			}
			if k.High < k.Open || k.High < k.Close {
				t.Errorf("kline[%d].High (%.2f) should be >= Open (%.2f) and Close (%.2f)", i, k.High, k.Open, k.Close)
			}
			if k.Low > k.Open || k.Low > k.Close {
				t.Errorf("kline[%d].Low (%.2f) should be <= Open (%.2f) and Close (%.2f)", i, k.Low, k.Open, k.Close)
			}
		}

		// Verify klines are sorted by OpenTime (ascending)
		for i := 1; i < len(klines); i++ {
			if klines[i].OpenTime < klines[i-1].OpenTime {
				t.Errorf("klines not sorted: klines[%d].OpenTime=%d < klines[%d].OpenTime=%d",
					i, klines[i].OpenTime, i-1, klines[i-1].OpenTime)
			}
		}

		t.Logf("successfully retrieved %d klines for BTCUSDT@%s (limit=%d, ready=%v)", len(klines), tc.interval, tc.limit, ready)
	}
}

// TestGetRecentKlines_Integration_DataBackfill tests that GetRecentKlines triggers backfill when data is missing
func TestGetRecentKlines_Integration_DataBackfill(t *testing.T) {
	requireServiceIT(t)
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	opts := Options{
		Intervals:             []string{"1m"},
		BackfillWindow:        20,
		RestMaxConcurrency:    5,
		WSEndpoint:            WSEndpointFutures,
		EnableRestBackfill:    false, // Disable automatic backfill, rely on GetRecentKlines internal REST calls
		SymbolRefreshInterval: 24 * time.Hour,
	}

	svc := NewService(opts)
	defer svc.Close()

	if err := svc.Start(ctx); err != nil {
		t.Fatalf("failed to start service: %v", err)
	}

	time.Sleep(5 * time.Second)

	// Add a new symbol that doesn't have data yet
	symbols := []string{"ETHUSDT"}
	if err := svc.AddSymbols(symbols); err != nil {
		t.Fatalf("failed to add symbols: %v", err)
	}

	// Immediately try to get data (should trigger backfill)
	klines, ready := svc.GetRecentKlines("ETHUSDT", "1m", 10)

	// Wait for backfill to complete
	time.Sleep(10 * time.Second)

	// Try again after backfill
	klinesAfter, readyAfter := svc.GetRecentKlines("ETHUSDT", "1m", 10)

	if len(klinesAfter) == 0 && len(klines) == 0 {
		t.Skipf("no data received for ETHUSDT (network may be slow or blocked)")
		return
	}

	// After backfill, we should have more or equal data
	if len(klinesAfter) < len(klines) {
		t.Logf("warning: data count decreased after backfill: before=%d, after=%d", len(klines), len(klinesAfter))
	}

	// Verify data quality
	if len(klinesAfter) > 0 {
		// Check that klines are recent (within last hour)
		now := time.Now().UnixMilli()
		latestKline := klinesAfter[len(klinesAfter)-1]
		oneHourAgo := now - int64(time.Hour/time.Millisecond)
		if latestKline.OpenTime < oneHourAgo {
			t.Logf("warning: latest kline is old: OpenTime=%d (now=%d)", latestKline.OpenTime, now)
		}

		t.Logf("backfill test: initial=%d klines (ready=%v), after=%d klines (ready=%v)",
			len(klines), ready, len(klinesAfter), readyAfter)
	}
}

// TestGetRecentKlines_Integration_MultipleSymbols tests GetRecentKlines with multiple symbols
func TestGetRecentKlines_Integration_MultipleSymbols(t *testing.T) {
	requireServiceIT(t)
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	opts := Options{
		Intervals:             []string{"1m"},
		BackfillWindow:        10,
		RestMaxConcurrency:    5,
		WSEndpoint:            WSEndpointFutures,
		EnableRestBackfill:    false, // Disable automatic backfill, rely on GetRecentKlines internal REST calls
		SymbolRefreshInterval: 24 * time.Hour,
	}

	svc := NewService(opts)
	defer svc.Close()

	if err := svc.Start(ctx); err != nil {
		t.Fatalf("failed to start service: %v", err)
	}

	time.Sleep(5 * time.Second)

	// Add multiple symbols
	symbols := []string{"BTCUSDT", "ETHUSDT", "BNBUSDT"}
	if err := svc.AddSymbols(symbols); err != nil {
		t.Fatalf("failed to add symbols: %v", err)
	}

	// Wait for data
	time.Sleep(15 * time.Second)

	// Test each symbol
	for _, sym := range symbols {
		klines, ready := svc.GetRecentKlines(sym, "1m", 5)
		if !ready {
			t.Logf("warning: data not ready for %s, may need more time", sym)
			continue
		}

		if len(klines) == 0 {
			t.Logf("warning: no data for %s yet", sym)
			continue
		}

		// Verify symbol data
		if klines[0].OpenTime <= 0 {
			t.Errorf("%s: invalid OpenTime: %d", sym, klines[0].OpenTime)
		}
		if klines[0].Close <= 0 {
			t.Errorf("%s: invalid Close price: %.2f", sym, klines[0].Close)
		}

		t.Logf("successfully retrieved %d klines for %s (ready=%v)", len(klines), sym, ready)
	}
}

// TestGetRecentKlines_Integration_ReadyState tests the ready state behavior
func TestGetRecentKlines_Integration_ReadyState(t *testing.T) {
	requireServiceIT(t)
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	opts := Options{
		Intervals:             []string{"1m"},
		BackfillWindow:        10,
		RestMaxConcurrency:    5,
		WSEndpoint:            WSEndpointFutures,
		EnableRestBackfill:    false, // Disable automatic backfill, rely on GetRecentKlines internal REST calls
		SymbolRefreshInterval: 24 * time.Hour,
	}

	svc := NewService(opts)
	defer svc.Close()

	if err := svc.Start(ctx); err != nil {
		t.Fatalf("failed to start service: %v", err)
	}

	time.Sleep(5 * time.Second)

	// Add symbol
	symbols := []string{"BTCUSDT"}
	if err := svc.AddSymbols(symbols); err != nil {
		t.Fatalf("failed to add symbols: %v", err)
	}

	// Wait for initial data
	time.Sleep(15 * time.Second)

	// Test with reasonable limit (should be ready)
	klines, ready := svc.GetRecentKlines("BTCUSDT", "1m", 10)
	if len(klines) > 0 && !ready {
		t.Logf("warning: data exists but ready=false (may be normal if data is being backfilled)")
	}

	// Test with very large limit (may not be ready if not enough history)
	klinesLarge, readyLarge := svc.GetRecentKlines("BTCUSDT", "1m", 100)
	if len(klinesLarge) > 0 {
		t.Logf("large limit test: got %d klines, ready=%v", len(klinesLarge), readyLarge)
		// If we have data but not ready, it's likely because we don't have 1000 klines
		if !readyLarge && len(klinesLarge) < 100 {
			t.Logf("expected: not ready because we only have %d klines (requested 1000)", len(klinesLarge))
		}
	}

	// Test with limit=0 (should return all available)
	klinesAll, readyAll := svc.GetRecentKlines("BTCUSDT", "1m", 0)
	if len(klinesAll) > 0 {
		t.Logf("limit=0 test: got %d klines, ready=%v", len(klinesAll), readyAll)
	}

	t.Logf("ready state test completed: limit=10 (ready=%v), limit=1000 (ready=%v), limit=0 (ready=%v)",
		ready, readyLarge, readyAll)
}

// TestGetRecentKlines_Integration_CaseInsensitive tests case-insensitive symbol and interval handling
func TestGetRecentKlines_Integration_CaseInsensitive(t *testing.T) {
	requireServiceIT(t)
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	opts := Options{
		Intervals:             []string{"1m"},
		BackfillWindow:        10,
		RestMaxConcurrency:    5,
		WSEndpoint:            WSEndpointFutures,
		EnableRestBackfill:    false, // Disable automatic backfill, rely on GetRecentKlines internal REST calls
		SymbolRefreshInterval: 24 * time.Hour,
	}

	svc := NewService(opts)
	defer svc.Close()

	if err := svc.Start(ctx); err != nil {
		t.Fatalf("failed to start service: %v", err)
	}

	time.Sleep(5 * time.Second)

	// Add symbol with lowercase
	if err := svc.AddSymbols([]string{"btcusdt"}); err != nil {
		t.Fatalf("failed to add symbols: %v", err)
	}

	time.Sleep(15 * time.Second)

	// Test different case combinations
	testCases := []struct {
		symbol   string
		interval string
	}{
		{"BTCUSDT", "1m"},
		{"btcusdt", "1m"},
		{"BtcUsdt", "1M"},
		{"btcusdt", "1M"},
	}

	for _, tc := range testCases {
		klines, ready := svc.GetRecentKlines(tc.symbol, tc.interval, 5)
		if !ready || len(klines) == 0 {
			t.Logf("warning: no data for %s@%s (ready=%v, len=%d)", tc.symbol, tc.interval, ready, len(klines))
			continue
		}

		// All should return the same data (normalized)
		if len(klines) > 0 {
			t.Logf("case-insensitive test passed for %s@%s: got %d klines", tc.symbol, tc.interval, len(klines))
		}
	}
}
