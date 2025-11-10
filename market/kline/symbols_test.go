package kline

import (
	"context"
	"sort"
	"strings"
	"testing"
	"time"
)

// TestFilteredBinanceFuturesFetcher_FetchSymbols tests fetching symbols from Binance API
func TestFilteredBinanceFuturesFetcher_FetchSymbols(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	client := NewKlineHTTPClient()
	fetcher := NewFilteredBinanceFuturesFetcher(client)

	symbols, err := fetcher.FetchSymbols(ctx)
	if err != nil {
		t.Skipf("failed to fetch symbols (likely no network access): %v", err)
		return
	}

	if len(symbols) == 0 {
		t.Fatal("expected at least one symbol, got 0")
	}

	// Verify all symbols are uppercase
	for _, sym := range symbols {
		if sym != strings.ToUpper(sym) {
			t.Errorf("symbol %s is not uppercase", sym)
		}
		if !strings.HasSuffix(sym, "USDT") {
			t.Errorf("symbol %s does not end with USDT", sym)
		}
	}

	// Verify symbols are sorted
	if !sort.StringsAreSorted(symbols) {
		t.Error("symbols are not sorted")
	}

	// Verify common symbols exist
	commonSymbols := []string{"BTCUSDT", "ETHUSDT", "BNBUSDT"}
	found := make(map[string]bool)
	for _, sym := range symbols {
		for _, common := range commonSymbols {
			if sym == common {
				found[common] = true
			}
		}
	}

	for _, common := range commonSymbols {
		if !found[common] {
			t.Logf("warning: common symbol %s not found (may be delisted)", common)
		}
	}

	t.Logf("fetched %d symbols", len(symbols))
}

// TestFilteredBinanceFuturesFetcher_CustomPredicate tests custom filtering
func TestFilteredBinanceFuturesFetcher_CustomPredicate(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	client := NewKlineHTTPClient()
	fetcher := NewFilteredBinanceFuturesFetcher(client)

	// Custom predicate: only BTC and ETH
	fetcher.Predicate = func(s ExchangeSymbol) bool {
		sym := strings.ToUpper(s.Symbol)
		return (sym == "BTCUSDT" || sym == "ETHUSDT") &&
			strings.EqualFold(s.Status, SymbolStatusTrading) &&
			strings.EqualFold(s.ContractType, ContractTypePerpetual) &&
			strings.EqualFold(s.QuoteAsset, QuoteAssetUSDT)
	}

	symbols, err := fetcher.FetchSymbols(ctx)
	if err != nil {
		t.Skipf("failed to fetch symbols (likely no network access): %v", err)
		return
	}

	// Should have at most 2 symbols
	if len(symbols) > 2 {
		t.Errorf("expected at most 2 symbols with custom predicate, got %d: %v", len(symbols), symbols)
	}

	// Verify only BTC and ETH
	for _, sym := range symbols {
		if sym != "BTCUSDT" && sym != "ETHUSDT" {
			t.Errorf("unexpected symbol with custom predicate: %s", sym)
		}
	}
}

// TestSymbolSync_ComputeDiffOnce tests computing diff without updating state
func TestSymbolSync_ComputeDiffOnce(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	client := NewKlineHTTPClient()
	fetcher := NewFilteredBinanceFuturesFetcher(client)
	sync := NewSymbolSync(0, fetcher)

	// First call should return all symbols as "added" (since lastSet is empty)
	added, removed, err := sync.ComputeDiffOnce(ctx)
	if err != nil {
		t.Skipf("failed to compute diff (likely no network access): %v", err)
		return
	}

	if len(added) == 0 {
		t.Error("expected at least one added symbol on first call")
	}
	if len(removed) != 0 {
		t.Errorf("expected no removed symbols on first call, got %d", len(removed))
	}

	// Verify added symbols are sorted
	if !sort.StringsAreSorted(added) {
		t.Error("added symbols are not sorted")
	}

	t.Logf("first diff: %d added, %d removed", len(added), len(removed))
}

// TestSymbolSync_WithInitialSeed tests initial seed functionality
func TestSymbolSync_WithInitialSeed(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	client := NewKlineHTTPClient()
	fetcher := NewFilteredBinanceFuturesFetcher(client)
	sync := NewSymbolSync(0, fetcher)

	// Set initial seed
	initialSeed := []string{"BTCUSDT", "ETHUSDT", "BNBUSDT"}
	sync.WithInitialSeed(initialSeed)

	// Compute diff - should not treat all symbols as "added"
	added, removed, err := sync.ComputeDiffOnce(ctx)
	if err != nil {
		t.Skipf("failed to compute diff (likely no network access): %v", err)
		return
	}

	// Should have fewer added symbols than total (since we seeded with some)
	// But we can't guarantee exact count since Binance may have more symbols
	if len(added) > 0 {
		t.Logf("found %d new symbols not in initial seed", len(added))
	}

	// Verify initial seed symbols are not in "added" (they should be in lastSet)
	for _, seed := range initialSeed {
		for _, a := range added {
			if a == seed {
				t.Errorf("initial seed symbol %s should not be in added list", seed)
			}
		}
	}

	t.Logf("diff with initial seed: %d added, %d removed", len(added), len(removed))
}

// TestSymbolSync_Start tests the Start method with a short interval
func TestSymbolSync_Start(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	client := NewKlineHTTPClient()
	fetcher := NewFilteredBinanceFuturesFetcher(client)
	sync := NewSymbolSync(1*time.Second, fetcher)

	var onChangeCalls int
	var lastAdded, lastRemoved []string

	onChange := func(added, removed []string) {
		onChangeCalls++
		lastAdded = added
		lastRemoved = removed
		t.Logf("onChange called: %d added, %d removed", len(added), len(removed))
	}

	// Start in background
	go sync.Start(ctx, onChange)

	// Wait for at least one onChange call (initial sync)
	time.Sleep(2 * time.Second)

	// Cancel context to stop
	cancel()

	// Wait a bit for goroutine to exit
	time.Sleep(500 * time.Millisecond)

	if onChangeCalls == 0 {
		t.Skip("onChange was not called (likely no network access or timeout)")
		return
	}

	// Verify onChange was called
	if len(lastAdded) == 0 && len(lastRemoved) == 0 && onChangeCalls == 1 {
		// First call with empty lastSet should have added symbols
		t.Log("first sync completed (no changes reported, which is expected with initial seed)")
	}

	t.Logf("onChange was called %d times", onChangeCalls)
}

// TestSymbolSync_ContextCancel tests that context cancellation works
func TestSymbolSync_ContextCancel(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())

	client := NewKlineHTTPClient()
	fetcher := NewFilteredBinanceFuturesFetcher(client)
	sync := NewSymbolSync(1*time.Second, fetcher)

	onChange := func(added, removed []string) {
		t.Logf("onChange: %d added, %d removed", len(added), len(removed))
	}

	// Start in background
	done := make(chan struct{})
	go func() {
		sync.Start(ctx, onChange)
		close(done)
	}()

	// Give it a tiny moment to start, then cancel
	time.Sleep(10 * time.Millisecond)
	cancel()

	// Wait for goroutine to exit (should exit quickly after context cancellation)
	select {
	case <-done:
		// Good, goroutine exited
		t.Log("goroutine exited successfully after context cancellation")
	case <-time.After(5 * time.Second):
		t.Error("goroutine did not exit after context cancellation")
	}
}

// TestUniqKeepUpper tests the uniqKeepUpper helper function
func TestUniqKeepUpper(t *testing.T) {
	tests := []struct {
		name     string
		input    []string
		expected int
	}{
		{
			name:     "no duplicates",
			input:    []string{"BTCUSDT", "ETHUSDT", "BNBUSDT"},
			expected: 3,
		},
		{
			name:     "with duplicates",
			input:    []string{"BTCUSDT", "ETHUSDT", "BTCUSDT", "BNBUSDT", "ETHUSDT"},
			expected: 3,
		},
		{
			name:     "empty strings",
			input:    []string{"BTCUSDT", "", "ETHUSDT", "  ", "BNBUSDT"},
			expected: 3,
		},
		{
			name:     "empty input",
			input:    []string{},
			expected: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			seen := make(map[string]struct{})
			result := uniqKeepUpper(tt.input, seen)
			if len(result) != tt.expected {
				t.Errorf("expected %d unique symbols, got %d: %v", tt.expected, len(result), result)
			}

			// Verify all are uppercase
			for _, s := range result {
				if s != strings.ToUpper(s) {
					t.Errorf("symbol %s is not uppercase", s)
				}
			}

			// Verify no duplicates
			seenMap := make(map[string]bool)
			for _, s := range result {
				if seenMap[s] {
					t.Errorf("duplicate found: %s", s)
				}
				seenMap[s] = true
			}
		})
	}
}

// TestDiffSets tests the diffSets helper function
func TestDiffSets(t *testing.T) {
	tests := []struct {
		name            string
		prev            map[string]struct{}
		curr            map[string]struct{}
		expectedAdded   int
		expectedRemoved int
	}{
		{
			name:            "all new",
			prev:            map[string]struct{}{},
			curr:            map[string]struct{}{"BTCUSDT": {}, "ETHUSDT": {}},
			expectedAdded:   2,
			expectedRemoved: 0,
		},
		{
			name:            "all removed",
			prev:            map[string]struct{}{"BTCUSDT": {}, "ETHUSDT": {}},
			curr:            map[string]struct{}{},
			expectedAdded:   0,
			expectedRemoved: 2,
		},
		{
			name:            "mixed changes",
			prev:            map[string]struct{}{"BTCUSDT": {}, "ETHUSDT": {}},
			curr:            map[string]struct{}{"ETHUSDT": {}, "BNBUSDT": {}},
			expectedAdded:   1,
			expectedRemoved: 1,
		},
		{
			name:            "no changes",
			prev:            map[string]struct{}{"BTCUSDT": {}, "ETHUSDT": {}},
			curr:            map[string]struct{}{"BTCUSDT": {}, "ETHUSDT": {}},
			expectedAdded:   0,
			expectedRemoved: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			added, removed := diffSets(tt.prev, tt.curr)
			if len(added) != tt.expectedAdded {
				t.Errorf("expected %d added, got %d: %v", tt.expectedAdded, len(added), added)
			}
			if len(removed) != tt.expectedRemoved {
				t.Errorf("expected %d removed, got %d: %v", tt.expectedRemoved, len(removed), removed)
			}
		})
	}
}

// TestNormalizeUpper tests the normalizeUpper helper function
func TestNormalizeUpper(t *testing.T) {
	tests := []struct {
		name     string
		input    []string
		expected []string
	}{
		{
			name:     "mixed case",
			input:    []string{"btcusdt", "ETHUSDT", "BnbUsdt"},
			expected: []string{"BNBUSDT", "BTCUSDT", "ETHUSDT"}, // sorted
		},
		{
			name:     "with spaces",
			input:    []string{" BTCUSDT ", "  ETHUSDT  ", "BNBUSDT"},
			expected: []string{"BNBUSDT", "BTCUSDT", "ETHUSDT"}, // sorted
		},
		{
			name:     "empty strings",
			input:    []string{"BTCUSDT", "", "ETHUSDT", "  "},
			expected: []string{"BTCUSDT", "ETHUSDT"}, // empty filtered out
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := normalizeUpper(tt.input)
			if len(result) != len(tt.expected) {
				t.Errorf("expected %d items, got %d: %v", len(tt.expected), len(result), result)
			}
			for i, exp := range tt.expected {
				if i >= len(result) || result[i] != exp {
					t.Errorf("expected %s at index %d, got %v", exp, i, result)
				}
			}
		})
	}
}
