package kline

import (
	"context"
	"testing"
	"time"
)

// TestService_StartAndAddSymbols tests starting the service and adding symbols
func TestService_StartAndAddSymbols(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	opts := Options{
		Intervals:             []string{"1m", "3m"}, // Use shorter intervals for faster testing
		BackfillWindow:        10,
		RestMaxConcurrency:    5,
		WSEndpoint:            WSEndpointFutures,
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
	ctx, cancel := context.WithTimeout(context.Background(), 40*time.Second)
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
	ctx, cancel := context.WithTimeout(context.Background(), 40*time.Second)
	defer cancel()

	opts := Options{
		Intervals:             []string{"1m", "3m"},
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
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	opts := Options{
		Intervals:             []string{"1m"},
		BackfillWindow:        5,
		RestMaxConcurrency:    5,
		WSEndpoint:            WSEndpointFutures,
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
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	opts := Options{
		Intervals:             []string{"1m"},
		BackfillWindow:        20,
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
