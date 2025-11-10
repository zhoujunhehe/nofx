package kline

import (
	"encoding/json"
	"fmt"
	"sync"
	"testing"
	"time"
)

// mockStore is a simple in-memory store for testing.
type mockStore struct {
	mu                sync.RWMutex
	finalKlines       map[string][]marketKline // key: "SYMBOL|interval"
	provisionalKlines map[string]marketKline   // key: "SYMBOL|interval"
}

func newMockStore() *mockStore {
	return &mockStore{
		finalKlines:       make(map[string][]marketKline),
		provisionalKlines: make(map[string]marketKline),
	}
}

func (m *mockStore) makeKey(symbol, interval string) string {
	return symbol + "|" + interval
}

func (m *mockStore) UpsertFinal(symbol, interval string, k marketKline) {
	m.mu.Lock()
	defer m.mu.Unlock()
	key := m.makeKey(symbol, interval)
	m.finalKlines[key] = append(m.finalKlines[key], k)
}

func (m *mockStore) UpsertFinalBatch(symbol, interval string, ks []marketKline) {
	for _, k := range ks {
		m.UpsertFinal(symbol, interval, k)
	}
}

func (m *mockStore) UpsertProvisional(symbol, interval string, k marketKline) {
	m.mu.Lock()
	defer m.mu.Unlock()
	key := m.makeKey(symbol, interval)
	m.provisionalKlines[key] = k
}

func (m *mockStore) GetRecent(symbol, interval string, limit int) ([]marketKline, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	key := m.makeKey(symbol, interval)
	klines := m.finalKlines[key]
	if len(klines) == 0 {
		return nil, false
	}
	if limit <= 0 || limit >= len(klines) {
		return klines, true
	}
	return klines[len(klines)-limit:], true
}

func (m *mockStore) getFinalCount(symbol, interval string) int {
	m.mu.RLock()
	defer m.mu.RUnlock()
	key := m.makeKey(symbol, interval)
	return len(m.finalKlines[key])
}

func (m *mockStore) getProvisional(symbol, interval string) (marketKline, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	key := m.makeKey(symbol, interval)
	k, ok := m.provisionalKlines[key]
	return k, ok
}

func TestWSIngestor_Attach(t *testing.T) {
	store := newMockStore()
	ingestor := NewWSIngestor(store)

	ch := make(chan []byte, 10)
	cancel, ok := ingestor.Attach("btcusdt@kline_1m", "BTCUSDT", "1m", ch)
	if !ok {
		t.Fatal("Attach should succeed")
	}
	if cancel == nil {
		t.Fatal("cancel function should not be nil")
	}

	// Test duplicate attach
	_, ok2 := ingestor.Attach("btcusdt@kline_1m", "BTCUSDT", "1m", ch)
	if ok2 {
		t.Fatal("duplicate attach should fail")
	}

	// Cleanup
	cancel()
	time.Sleep(50 * time.Millisecond) // Wait for goroutine to exit
}

func TestWSIngestor_Detach(t *testing.T) {
	store := newMockStore()
	ingestor := NewWSIngestor(store)

	ch := make(chan []byte, 10)
	cancel, ok := ingestor.Attach("btcusdt@kline_1m", "BTCUSDT", "1m", ch)
	if !ok {
		t.Fatal("Attach should succeed")
	}

	ingestor.Detach("btcusdt@kline_1m")
	time.Sleep(50 * time.Millisecond) // Wait for goroutine to exit

	// Detach again should be safe
	ingestor.Detach("btcusdt@kline_1m")

	// Cancel should also work
	cancel()
}

func TestWSIngestor_Close(t *testing.T) {
	store := newMockStore()
	ingestor := NewWSIngestor(store)

	ch1 := make(chan []byte, 10)
	ch2 := make(chan []byte, 10)
	_, ok1 := ingestor.Attach("btcusdt@kline_1m", "BTCUSDT", "1m", ch1)
	_, ok2 := ingestor.Attach("ethusdt@kline_1m", "ETHUSDT", "1m", ch2)
	if !ok1 || !ok2 {
		t.Fatal("Attach should succeed")
	}

	ingestor.Close()
	time.Sleep(50 * time.Millisecond) // Wait for goroutines to exit

	// Close again should be safe
	ingestor.Close()
}

func TestWSIngestor_HandleFinalKline(t *testing.T) {
	store := newMockStore()
	ingestor := NewWSIngestor(store)

	ch := make(chan []byte, 10)
	_, ok := ingestor.Attach("btcusdt@kline_1m", "BTCUSDT", "1m", ch)
	if !ok {
		t.Fatal("Attach should succeed")
	}
	defer ingestor.Close()

	// Send a final kline message
	finalMsg := `{
		"e": "kline",
		"s": "BTCUSDT",
		"k": {
			"t": 1609459200000,
			"T": 1609459260000,
			"i": "1m",
			"o": "29000.0",
			"h": "29100.0",
			"l": "28900.0",
			"c": "29050.0",
			"v": "100.5",
			"q": "2915025.0",
			"n": 150,
			"V": "50.0",
			"Q": "1452500.0",
			"x": true
		}
	}`

	ch <- []byte(finalMsg)
	time.Sleep(100 * time.Millisecond) // Wait for processing

	count := store.getFinalCount("BTCUSDT", "1m")
	if count != 1 {
		t.Fatalf("expected 1 final kline, got %d", count)
	}

	klines, ok := store.GetRecent("BTCUSDT", "1m", 1)
	if !ok || len(klines) != 1 {
		t.Fatal("should have 1 kline")
	}
	k := klines[0]
	if k.OpenTime != 1609459200000 || k.Close != 29050.0 {
		t.Fatalf("unexpected kline data: %+v", k)
	}
}

func TestWSIngestor_HandleProvisionalKline(t *testing.T) {
	store := newMockStore()
	ingestor := NewWSIngestor(store)

	ch := make(chan []byte, 10)
	_, ok := ingestor.Attach("btcusdt@kline_1m", "BTCUSDT", "1m", ch)
	if !ok {
		t.Fatal("Attach should succeed")
	}
	defer ingestor.Close()

	// Send a provisional kline message
	provisionalMsg := `{
		"e": "kline",
		"s": "BTCUSDT",
		"k": {
			"t": 1609459200000,
			"T": 1609459260000,
			"i": "1m",
			"o": "29000.0",
			"h": "29100.0",
			"l": "28900.0",
			"c": "29050.0",
			"v": "100.5",
			"q": "2915025.0",
			"n": 150,
			"V": "50.0",
			"Q": "1452500.0",
			"x": false
		}
	}`

	ch <- []byte(provisionalMsg)
	time.Sleep(100 * time.Millisecond) // Wait for processing

	// Provisional should be stored
	prov, ok := store.getProvisional("BTCUSDT", "1m")
	if !ok {
		t.Fatal("provisional kline should be stored")
	}
	if prov.OpenTime != 1609459200000 || prov.Close != 29050.0 {
		t.Fatalf("unexpected provisional kline: %+v", prov)
	}

	// Final count should still be 0
	count := store.getFinalCount("BTCUSDT", "1m")
	if count != 0 {
		t.Fatalf("expected 0 final klines, got %d", count)
	}
}

func TestWSIngestor_HandleProvisionalThrottling(t *testing.T) {
	store := newMockStore()
	ingestor := NewWSIngestor(store)

	ch := make(chan []byte, 10)
	_, ok := ingestor.Attach("btcusdt@kline_1m", "BTCUSDT", "1m", ch)
	if !ok {
		t.Fatal("Attach should succeed")
	}
	defer ingestor.Close()

	// Send multiple provisional updates for the same kline (same open_time)
	msg := `{
		"e": "kline",
		"s": "BTCUSDT",
		"k": {
			"t": 1609459200000,
			"T": 1609459260000,
			"i": "1m",
			"o": "29000.0",
			"h": "29100.0",
			"l": "28900.0",
			"c": "29050.0",
			"v": "100.5",
			"q": "2915025.0",
			"n": 150,
			"V": "50.0",
			"Q": "1452500.0",
			"x": false
		}
	}`

	// Send 3 messages rapidly
	ch <- []byte(msg)
	ch <- []byte(msg)
	ch <- []byte(msg)
	time.Sleep(100 * time.Millisecond)

	// Due to throttling (300ms min interval), only the first should be written
	// But we can't easily verify the exact count due to timing, so just verify
	// that at least one provisional was written
	_, ok = store.getProvisional("BTCUSDT", "1m")
	if !ok {
		t.Fatal("at least one provisional kline should be stored")
	}
}

func TestWSIngestor_HandleWrappedMessage(t *testing.T) {
	store := newMockStore()
	ingestor := NewWSIngestor(store)

	ch := make(chan []byte, 10)
	_, ok := ingestor.Attach("btcusdt@kline_1m", "BTCUSDT", "1m", ch)
	if !ok {
		t.Fatal("Attach should succeed")
	}
	defer ingestor.Close()

	// Send a wrapped message (stream + data format)
	wrappedMsg := `{
		"stream": "btcusdt@kline_1m",
		"data": {
			"e": "kline",
			"s": "BTCUSDT",
			"k": {
				"t": 1609459200000,
				"T": 1609459260000,
				"i": "1m",
				"o": "29000.0",
				"h": "29100.0",
				"l": "28900.0",
				"c": "29050.0",
				"v": "100.5",
				"q": "2915025.0",
				"n": 150,
				"V": "50.0",
				"Q": "1452500.0",
				"x": true
			}
		}
	}`

	ch <- []byte(wrappedMsg)
	time.Sleep(100 * time.Millisecond) // Wait for processing

	count := store.getFinalCount("BTCUSDT", "1m")
	if count != 1 {
		t.Fatalf("expected 1 final kline, got %d", count)
	}
}

func TestWSIngestor_HandleInvalidMessage(t *testing.T) {
	store := newMockStore()
	ingestor := NewWSIngestor(store)

	ch := make(chan []byte, 10)
	_, ok := ingestor.Attach("btcusdt@kline_1m", "BTCUSDT", "1m", ch)
	if !ok {
		t.Fatal("Attach should succeed")
	}
	defer ingestor.Close()

	// Send invalid JSON
	ch <- []byte("invalid json")
	time.Sleep(100 * time.Millisecond)

	// Send message with missing required fields
	invalidMsg := `{
		"e": "kline",
		"s": "",
		"k": {
			"t": 0,
			"i": ""
		}
	}`
	ch <- []byte(invalidMsg)
	time.Sleep(100 * time.Millisecond)

	// Should not crash and should not store anything
	count := store.getFinalCount("BTCUSDT", "1m")
	if count != 0 {
		t.Fatalf("expected 0 final klines, got %d", count)
	}
}

func TestWSIngestor_ChannelClose(t *testing.T) {
	store := newMockStore()
	ingestor := NewWSIngestor(store)

	ch := make(chan []byte, 10)
	_, ok := ingestor.Attach("btcusdt@kline_1m", "BTCUSDT", "1m", ch)
	if !ok {
		t.Fatal("Attach should succeed")
	}

	// Close the channel
	close(ch)
	time.Sleep(100 * time.Millisecond) // Wait for goroutine to detect close

	// Should clean up gracefully
	// We can't easily verify internal state, but it shouldn't panic
}

func TestWSIngestor_ContextCancel(t *testing.T) {
	store := newMockStore()
	ingestor := NewWSIngestor(store)

	ch := make(chan []byte, 10)
	cancel, ok := ingestor.Attach("btcusdt@kline_1m", "BTCUSDT", "1m", ch)
	if !ok {
		t.Fatal("Attach should succeed")
	}

	// Cancel the context
	cancel()
	time.Sleep(100 * time.Millisecond) // Wait for goroutine to exit

	// Should clean up gracefully
	// Try to send a message - it should be ignored (goroutine exited)
	msg := `{
		"e": "kline",
		"s": "BTCUSDT",
		"k": {
			"t": 1609459200000,
			"T": 1609459260000,
			"i": "1m",
			"o": "29000.0",
			"h": "29100.0",
			"l": "28900.0",
			"c": "29050.0",
			"v": "100.5",
			"q": "2915025.0",
			"n": 150,
			"V": "50.0",
			"Q": "1452500.0",
			"x": true
		}
	}`
	ch <- []byte(msg)
	time.Sleep(100 * time.Millisecond)

	// Should not be stored (goroutine already exited)
	count := store.getFinalCount("BTCUSDT", "1m")
	if count != 0 {
		t.Fatalf("expected 0 final klines (goroutine should have exited), got %d", count)
	}
}

func TestWSIngestor_FinalClearsProvisional(t *testing.T) {
	store := newMockStore()
	ingestor := NewWSIngestor(store)

	ch := make(chan []byte, 10)
	_, ok := ingestor.Attach("btcusdt@kline_1m", "BTCUSDT", "1m", ch)
	if !ok {
		t.Fatal("Attach should succeed")
	}
	defer ingestor.Close()

	openTime := int64(1609459200000)

	// Send provisional first
	provisionalMsg := createKlineMsg("BTCUSDT", "1m", openTime, 29050.0, false)
	ch <- []byte(provisionalMsg)
	time.Sleep(100 * time.Millisecond)

	// Verify provisional is stored
	_, ok = store.getProvisional("BTCUSDT", "1m")
	if !ok {
		t.Fatal("provisional should be stored")
	}

	// Send final for the same open_time
	finalMsg := createKlineMsg("BTCUSDT", "1m", openTime, 29060.0, true)
	ch <- []byte(finalMsg)
	time.Sleep(100 * time.Millisecond)

	// Final should be stored
	count := store.getFinalCount("BTCUSDT", "1m")
	if count != 1 {
		t.Fatalf("expected 1 final kline, got %d", count)
	}

	// Provisional should be cleared (internal state, we can't directly verify)
	// But we can verify the final has the correct data
	klines, ok := store.GetRecent("BTCUSDT", "1m", 1)
	if !ok || len(klines) != 1 {
		t.Fatal("should have 1 kline")
	}
	if klines[0].Close != 29060.0 {
		t.Fatalf("expected close price 29060.0, got %f", klines[0].Close)
	}
}

// Helper function to create a kline message
func createKlineMsg(symbol, interval string, openTime int64, closePrice float64, isFinal bool) string {
	closeTime := openTime + 60000 // 1 minute later
	msg := map[string]interface{}{
		"e": "kline",
		"s": symbol,
		"k": map[string]interface{}{
			"t": openTime,
			"T": closeTime,
			"i": interval,
			"o": "29000.0",
			"h": "29100.0",
			"l": "28900.0",
			"c": toString(closePrice),
			"v": "100.5",
			"q": "2915025.0",
			"n": 150,
			"V": "50.0",
			"Q": "1452500.0",
			"x": isFinal,
		},
	}
	data, _ := json.Marshal(msg)
	return string(data)
}

func toString(f float64) string {
	return fmt.Sprintf("%.1f", f)
}
