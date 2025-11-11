package kline

import (
	"errors"
	"fmt"
	"sort"
	"strings"
	"sync"
	"testing"
)

// mockWSClient is a mock implementation of wsClient for testing
type mockWSClient struct {
	mu               sync.Mutex
	subscribed       map[string]bool // stream -> subscribed
	subscribeCalls   []string        // all streams that were subscribed
	unsubscribeCalls []string        // all streams that were unsubscribed
	shouldFail       bool            // if true, Subscribe/Unsubscribe will fail
	failOnCall       int             // fail on Nth call (0 = first call)
	callCount        int             // track number of calls
}

func newMockWSClient() *mockWSClient {
	return &mockWSClient{
		subscribed:       make(map[string]bool),
		subscribeCalls:   make([]string, 0),
		unsubscribeCalls: make([]string, 0),
	}
}

func (m *mockWSClient) SubscribeStreams(streams []string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.callCount++
	if m.shouldFail && (m.failOnCall == 0 || m.callCount == m.failOnCall) {
		return errors.New("subscribe failed")
	}
	for _, s := range streams {
		s = strings.ToLower(strings.TrimSpace(s))
		m.subscribed[s] = true
		m.subscribeCalls = append(m.subscribeCalls, s)
	}
	return nil
}

func (m *mockWSClient) UnsubscribeStreams(streams []string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.callCount++
	if m.shouldFail && (m.failOnCall == 0 || m.callCount == m.failOnCall) {
		return errors.New("unsubscribe failed")
	}
	for _, s := range streams {
		s = strings.ToLower(strings.TrimSpace(s))
		m.subscribed[s] = false
		m.unsubscribeCalls = append(m.unsubscribeCalls, s)
	}
	return nil
}

func (m *mockWSClient) reset() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.subscribed = make(map[string]bool)
	m.subscribeCalls = make([]string, 0)
	m.unsubscribeCalls = make([]string, 0)
	m.shouldFail = false
	m.failOnCall = 0
	m.callCount = 0
}

func TestSubManager_AddSymbols(t *testing.T) {
	client := newMockWSClient()
	sm := NewSubManager()
	sm.SetClient("1m", client)

	// Test adding symbols
	added, err := sm.AddSymbols("1m", []string{"BTCUSDT", "ETHUSDT"})
	if err != nil {
		t.Fatalf("AddSymbols failed: %v", err)
	}
	if len(added) != 2 {
		t.Fatalf("expected 2 added symbols, got %d", len(added))
	}

	// Verify symbols are stored
	symbols := sm.Symbols("1m")
	if len(symbols) != 2 {
		t.Fatalf("expected 2 symbols, got %d", len(symbols))
	}
	sort.Strings(symbols)
	if symbols[0] != "BTCUSDT" || symbols[1] != "ETHUSDT" {
		t.Fatalf("unexpected symbols: %v", symbols)
	}

	// Verify streams were subscribed
	streams := sm.StreamsFor("1m")
	if len(streams) != 2 {
		t.Fatalf("expected 2 streams, got %d", len(streams))
	}
	expectedStreams := []string{BuildKlineStream("BTCUSDT", "1m"), BuildKlineStream("ETHUSDT", "1m")}
	sort.Strings(streams)
	sort.Strings(expectedStreams)
	if streams[0] != expectedStreams[0] || streams[1] != expectedStreams[1] {
		t.Fatalf("unexpected streams: %v", streams)
	}

	// Test adding duplicate symbols (should be skipped)
	added2, err := sm.AddSymbols("1m", []string{"BTCUSDT", "BNBUSDT"})
	if err != nil {
		t.Fatalf("AddSymbols failed: %v", err)
	}
	if len(added2) != 1 || added2[0] != "BNBUSDT" {
		t.Fatalf("expected only BNBUSDT to be added, got %v", added2)
	}

	// Test case insensitivity
	added3, err := sm.AddSymbols("1m", []string{"btcusdt", "ethusdt"})
	if err != nil {
		t.Fatalf("AddSymbols failed: %v", err)
	}
	if len(added3) != 0 {
		t.Fatalf("expected no new symbols (already subscribed), got %v", added3)
	}
}

func TestSubManager_AddSymbols_EmptyInput(t *testing.T) {
	client := newMockWSClient()
	sm := NewSubManager()
	sm.SetClient("1m", client)

	// Test empty symbols
	added, err := sm.AddSymbols("1m", []string{})
	if err != nil {
		t.Fatalf("AddSymbols should not fail on empty input: %v", err)
	}
	if len(added) != 0 {
		t.Fatalf("expected no added symbols, got %d", len(added))
	}

	// Test empty interval
	added2, err := sm.AddSymbols("", []string{"BTCUSDT"})
	if err != nil {
		t.Fatalf("AddSymbols should not fail on empty interval: %v", err)
	}
	if len(added2) != 0 {
		t.Fatalf("expected no added symbols, got %d", len(added2))
	}
}

func TestSubManager_AddSymbols_RollbackOnError(t *testing.T) {
	client := newMockWSClient()
	client.shouldFail = true
	client.failOnCall = 1 // Fail on first call
	sm := NewSubManager()
	sm.SetClient("1m", client)

	// Try to add symbols (should fail)
	added, err := sm.AddSymbols("1m", []string{"BTCUSDT", "ETHUSDT"})
	if err == nil {
		t.Fatal("AddSymbols should fail")
	}
	if len(added) != 0 {
		t.Fatalf("expected no added symbols on failure, got %d", len(added))
	}

	// Verify symbols were rolled back (not in store)
	symbols := sm.Symbols("1m")
	if len(symbols) != 0 {
		t.Fatalf("expected no symbols after rollback, got %d: %v", len(symbols), symbols)
	}
}

func TestSubManager_RemoveSymbols(t *testing.T) {
	client := newMockWSClient()
	sm := NewSubManager()
	sm.SetClient("1m", client)

	// First add some symbols
	_, err := sm.AddSymbols("1m", []string{"BTCUSDT", "ETHUSDT", "BNBUSDT"})
	if err != nil {
		t.Fatalf("AddSymbols failed: %v", err)
	}

	// Remove symbols
	removed, err := sm.RemoveSymbols("1m", []string{"BTCUSDT", "ETHUSDT"})
	if err != nil {
		t.Fatalf("RemoveSymbols failed: %v", err)
	}
	if len(removed) != 2 {
		t.Fatalf("expected 2 removed symbols, got %d", len(removed))
	}

	// Verify symbols were removed
	symbols := sm.Symbols("1m")
	if len(symbols) != 1 || symbols[0] != "BNBUSDT" {
		t.Fatalf("expected only BNBUSDT, got %v", symbols)
	}

	// Test removing non-existent symbols
	removed2, err := sm.RemoveSymbols("1m", []string{"DOGEUSDT"})
	if err != nil {
		t.Fatalf("RemoveSymbols should not fail on non-existent symbols: %v", err)
	}
	if len(removed2) != 0 {
		t.Fatalf("expected no removed symbols, got %d", len(removed2))
	}
}

func TestSubManager_RemoveSymbols_RollbackOnError(t *testing.T) {
	client := newMockWSClient()
	sm := NewSubManager()
	sm.SetClient("1m", client)

	// First add some symbols
	_, err := sm.AddSymbols("1m", []string{"BTCUSDT", "ETHUSDT"})
	if err != nil {
		t.Fatalf("AddSymbols failed: %v", err)
	}

	// Reset call count and make unsubscribe fail on next call
	client.mu.Lock()
	client.callCount = 0
	client.shouldFail = true
	client.failOnCall = 1
	client.mu.Unlock()

	// Try to remove symbols (should fail)
	removed, err := sm.RemoveSymbols("1m", []string{"BTCUSDT"})
	if err == nil {
		t.Fatal("RemoveSymbols should fail")
	}
	if len(removed) != 0 {
		t.Fatalf("expected no removed symbols on failure, got %d", len(removed))
	}

	// Verify symbols were rolled back (still in store)
	symbols := sm.Symbols("1m")
	if len(symbols) != 2 {
		t.Fatalf("expected 2 symbols after rollback, got %d: %v", len(symbols), symbols)
	}
}

func TestSubManager_Sync(t *testing.T) {
	client := newMockWSClient()
	sm := NewSubManager()
	sm.SetClient("1m", client)

	// Initial state: add some symbols
	_, err := sm.AddSymbols("1m", []string{"BTCUSDT", "ETHUSDT"})
	if err != nil {
		t.Fatalf("AddSymbols failed: %v", err)
	}

	// Sync to new set (add BNBUSDT, remove ETHUSDT)
	added, removed, err := sm.Sync("1m", []string{"BTCUSDT", "BNBUSDT"})
	if err != nil {
		t.Fatalf("Sync failed: %v", err)
	}
	if len(added) != 1 || added[0] != "BNBUSDT" {
		t.Fatalf("expected BNBUSDT to be added, got %v", added)
	}
	if len(removed) != 1 || removed[0] != "ETHUSDT" {
		t.Fatalf("expected ETHUSDT to be removed, got %v", removed)
	}

	// Verify final state
	symbols := sm.Symbols("1m")
	sort.Strings(symbols)
	expected := []string{"BNBUSDT", "BTCUSDT"}
	sort.Strings(expected)
	if len(symbols) != len(expected) {
		t.Fatalf("expected %d symbols, got %d: %v", len(expected), len(symbols), symbols)
	}
	for i, s := range symbols {
		if s != expected[i] {
			t.Fatalf("expected %s at index %d, got %s", expected[i], i, s)
		}
	}
}

func TestSubManager_Sync_NoChanges(t *testing.T) {
	client := newMockWSClient()
	sm := NewSubManager()
	sm.SetClient("1m", client)

	// Initial state
	_, err := sm.AddSymbols("1m", []string{"BTCUSDT", "ETHUSDT"})
	if err != nil {
		t.Fatalf("AddSymbols failed: %v", err)
	}

	// Sync to same set
	added, removed, err := sm.Sync("1m", []string{"BTCUSDT", "ETHUSDT"})
	if err != nil {
		t.Fatalf("Sync failed: %v", err)
	}
	if len(added) != 0 {
		t.Fatalf("expected no added symbols, got %d", len(added))
	}
	if len(removed) != 0 {
		t.Fatalf("expected no removed symbols, got %d", len(removed))
	}
}

func TestSubManager_Sync_InvalidInterval(t *testing.T) {
	client := newMockWSClient()
	sm := NewSubManager()
	sm.SetClient("1m", client)

	added, removed, err := sm.Sync("", []string{"BTCUSDT"})
	if err == nil {
		t.Fatal("Sync should fail on invalid interval")
	}
	if len(added) != 0 || len(removed) != 0 {
		t.Fatalf("expected no changes on error, got added=%d, removed=%d", len(added), len(removed))
	}
}

func TestSubManager_Symbols(t *testing.T) {
	client := newMockWSClient()
	sm := NewSubManager()
	sm.SetClient("1m", client)

	// Empty interval
	symbols := sm.Symbols("1m")
	if len(symbols) != 0 {
		t.Fatalf("expected no symbols, got %d", len(symbols))
	}

	// Add symbols
	_, err := sm.AddSymbols("1m", []string{"ETHUSDT", "BTCUSDT"})
	if err != nil {
		t.Fatalf("AddSymbols failed: %v", err)
	}

	// Verify symbols are returned and sorted
	symbols = sm.Symbols("1m")
	if len(symbols) != 2 {
		t.Fatalf("expected 2 symbols, got %d", len(symbols))
	}
	if !sort.StringsAreSorted(symbols) {
		t.Error("symbols should be sorted")
	}
	if symbols[0] != "BTCUSDT" || symbols[1] != "ETHUSDT" {
		t.Fatalf("unexpected symbols: %v", symbols)
	}
}

func TestSubManager_StreamsFor(t *testing.T) {
	client := newMockWSClient()
	sm := NewSubManager()
	sm.SetClient("1m", client)

	// Add symbols
	_, err := sm.AddSymbols("1m", []string{"BTCUSDT", "ETHUSDT"})
	if err != nil {
		t.Fatalf("AddSymbols failed: %v", err)
	}

	// Get streams
	streams := sm.StreamsFor("1m")
	if len(streams) != 2 {
		t.Fatalf("expected 2 streams, got %d", len(streams))
	}
	if !sort.StringsAreSorted(streams) {
		t.Error("streams should be sorted")
	}

	expected := []string{BuildKlineStream("BTCUSDT", "1m"), BuildKlineStream("ETHUSDT", "1m")}
	sort.Strings(expected)
	if streams[0] != expected[0] || streams[1] != expected[1] {
		t.Fatalf("unexpected streams: %v", streams)
	}
}

func TestSubManager_Has(t *testing.T) {
	client := newMockWSClient()
	sm := NewSubManager()
	sm.SetClient("1m", client)

	// Add symbol
	_, err := sm.AddSymbols("1m", []string{"BTCUSDT"})
	if err != nil {
		t.Fatalf("AddSymbols failed: %v", err)
	}

	// Test case insensitivity
	if !sm.Has("1m", "BTCUSDT") {
		t.Error("Has should return true for BTCUSDT")
	}
	if !sm.Has("1m", "btcusdt") {
		t.Error("Has should be case insensitive")
	}
	if !sm.Has("1M", "BTCUSDT") {
		t.Error("Has should be case insensitive for interval")
	}

	// Test non-existent symbol
	if sm.Has("1m", "ETHUSDT") {
		t.Error("Has should return false for non-existent symbol")
	}

	// Test non-existent interval
	if sm.Has("5m", "BTCUSDT") {
		t.Error("Has should return false for non-existent interval")
	}
}

func TestSubManager_Count(t *testing.T) {
	client := newMockWSClient()
	sm := NewSubManager()
	sm.SetClient("1m", client)

	// Empty interval
	if count := sm.Count("1m"); count != 0 {
		t.Fatalf("expected count 0, got %d", count)
	}

	// Add symbols
	_, err := sm.AddSymbols("1m", []string{"BTCUSDT", "ETHUSDT", "BNBUSDT"})
	if err != nil {
		t.Fatalf("AddSymbols failed: %v", err)
	}

	if count := sm.Count("1m"); count != 3 {
		t.Fatalf("expected count 3, got %d", count)
	}

	// Remove one
	_, err = sm.RemoveSymbols("1m", []string{"BTCUSDT"})
	if err != nil {
		t.Fatalf("RemoveSymbols failed: %v", err)
	}

	if count := sm.Count("1m"); count != 2 {
		t.Fatalf("expected count 2, got %d", count)
	}
}

func TestSubManager_All(t *testing.T) {
	client1m := newMockWSClient()
	client5m := newMockWSClient()
	sm := NewSubManager()
	sm.SetClient("1m", client1m)
	sm.SetClient("5m", client5m)

	// Add symbols to multiple intervals
	_, err := sm.AddSymbols("1m", []string{"BTCUSDT", "ETHUSDT"})
	if err != nil {
		t.Fatalf("AddSymbols failed: %v", err)
	}
	_, err = sm.AddSymbols("5m", []string{"BNBUSDT"})
	if err != nil {
		t.Fatalf("AddSymbols failed: %v", err)
	}

	// Get all
	all := sm.All()
	if len(all) != 2 {
		t.Fatalf("expected 2 intervals, got %d", len(all))
	}

	// Verify 1m
	if set, ok := all["1m"]; !ok {
		t.Fatal("expected 1m interval")
	} else {
		if len(set) != 2 {
			t.Fatalf("expected 2 symbols in 1m, got %d", len(set))
		}
		if _, ok := set["BTCUSDT"]; !ok {
			t.Error("expected BTCUSDT in 1m")
		}
		if _, ok := set["ETHUSDT"]; !ok {
			t.Error("expected ETHUSDT in 1m")
		}
	}

	// Verify 5m
	if set, ok := all["5m"]; !ok {
		t.Fatal("expected 5m interval")
	} else {
		if len(set) != 1 {
			t.Fatalf("expected 1 symbol in 5m, got %d", len(set))
		}
		if _, ok := set["BNBUSDT"]; !ok {
			t.Error("expected BNBUSDT in 5m")
		}
	}

	// Verify it's a copy (modifying shouldn't affect original)
	delete(all["1m"], "BTCUSDT")
	symbols := sm.Symbols("1m")
	if len(symbols) != 2 {
		t.Fatal("modifying returned map should not affect original")
	}
}

func TestSubManager_Batching(t *testing.T) {
	client := newMockWSClient()
	sm := NewSubManager()
	sm.SetClient("1m", client)

	// Create more than maxBatch symbols (each unique)
	symbols := make([]string, 200)
	for i := 0; i < 200; i++ {
		// Generate unique symbols with numeric suffix
		symbols[i] = strings.ToUpper(fmt.Sprintf("TEST%dUSDT", i))
	}

	// Add all symbols
	added, err := sm.AddSymbols("1m", symbols)
	if err != nil {
		t.Fatalf("AddSymbols failed: %v", err)
	}

	// Verify all were added (all symbols are unique)
	if len(added) != 200 {
		t.Fatalf("expected all 200 symbols to be added, got %d", len(added))
	}

	// Verify subscribe was called multiple times (batched)
	// Since maxBatch is 50, 200 symbols should be split into 4 batches
	client.mu.Lock()
	callCount := client.callCount
	streamsCount := len(client.subscribeCalls)
	client.mu.Unlock()

	// Should have been called 4 times (50 + 50 + 50 + 50)
	if callCount != 4 {
		t.Fatalf("expected SubscribeStreams to be called 4 times (batched), got %d", callCount)
	}
	// All 200 streams should have been subscribed
	if streamsCount != 200 {
		t.Fatalf("expected 200 streams to be subscribed, got %d", streamsCount)
	}
}

func TestSubManager_MultipleIntervals(t *testing.T) {
	client1m := newMockWSClient()
	client5m := newMockWSClient()
	sm := NewSubManager()
	sm.SetClient("1m", client1m)
	sm.SetClient("5m", client5m)

	// Add symbols to different intervals
	_, err := sm.AddSymbols("1m", []string{"BTCUSDT"})
	if err != nil {
		t.Fatalf("AddSymbols failed: %v", err)
	}
	_, err = sm.AddSymbols("5m", []string{"ETHUSDT"})
	if err != nil {
		t.Fatalf("AddSymbols failed: %v", err)
	}

	// Verify they are separate
	if count := sm.Count("1m"); count != 1 {
		t.Fatalf("expected 1 symbol in 1m, got %d", count)
	}
	if count := sm.Count("5m"); count != 1 {
		t.Fatalf("expected 1 symbol in 5m, got %d", count)
	}

	// Verify streams are correct
	streams1m := sm.StreamsFor("1m")
	expected1m := BuildKlineStream("BTCUSDT", "1m")
	if len(streams1m) != 1 || streams1m[0] != expected1m {
		t.Fatalf("unexpected streams for 1m: %v, expected %s", streams1m, expected1m)
	}

	streams5m := sm.StreamsFor("5m")
	expected5m := BuildKlineStream("ETHUSDT", "5m")
	if len(streams5m) != 1 || streams5m[0] != expected5m {
		t.Fatalf("unexpected streams for 5m: %v, expected %s", streams5m, expected5m)
	}
}

func TestSubManager_ConcurrentAccess(t *testing.T) {
	client := newMockWSClient()
	sm := NewSubManager()
	sm.SetClient("1m", client)

	// Concurrent adds
	var wg sync.WaitGroup
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			symbols := []string{strings.ToUpper(strings.Repeat("A", id+1)) + "USDT"}
			_, err := sm.AddSymbols("1m", symbols)
			if err != nil {
				t.Errorf("AddSymbols failed in goroutine %d: %v", id, err)
			}
		}(i)
	}
	wg.Wait()

	// Verify all were added
	count := sm.Count("1m")
	if count != 10 {
		t.Fatalf("expected 10 symbols, got %d", count)
	}
}
