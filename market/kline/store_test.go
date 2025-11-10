package kline

import (
	"sort"
	"sync"
	"testing"
	"time"
)

// Helper function to create a test kline
func createKline(openTime int64, close float64) marketKline {
	return marketKline{
		OpenTime:  openTime,
		CloseTime: openTime + 60000, // 1 minute later
		Open:      close - 10,
		High:      close + 10,
		Low:       close - 20,
		Close:     close,
		Volume:    100.0,
	}
}

func TestMemoryStore_NewMemoryStore(t *testing.T) {
	// Test with valid capacity
	store := NewMemoryStore(100)
	if store == nil {
		t.Fatal("NewMemoryStore returned nil")
	}
	if store.cap != 100 {
		t.Fatalf("expected capacity 100, got %d", store.cap)
	}

	// Test with zero capacity (should use default)
	store2 := NewMemoryStore(0)
	if store2.cap != DefaultStoreCapacity {
		t.Fatalf("expected default capacity %d, got %d", DefaultStoreCapacity, store2.cap)
	}

	// Test with negative capacity (should use default)
	store3 := NewMemoryStore(-10)
	if store3.cap != DefaultStoreCapacity {
		t.Fatalf("expected default capacity %d, got %d", DefaultStoreCapacity, store3.cap)
	}
}

func TestMemoryStore_UpsertFinal_Append(t *testing.T) {
	store := NewMemoryStore(100)
	baseTime := time.Now().UnixMilli()

	// Append klines in order
	for i := 0; i < 5; i++ {
		k := createKline(baseTime+int64(i*60000), 100.0+float64(i))
		store.UpsertFinal("BTCUSDT", "1m", k)
	}

	klines, ok := store.GetRecent("BTCUSDT", "1m", 0)
	if !ok {
		t.Fatal("expected klines to exist")
	}
	if len(klines) != 5 {
		t.Fatalf("expected 5 klines, got %d", len(klines))
	}

	// Verify order
	for i := 0; i < 5; i++ {
		expectedTime := baseTime + int64(i*60000)
		if klines[i].OpenTime != expectedTime {
			t.Fatalf("expected OpenTime %d at index %d, got %d", expectedTime, i, klines[i].OpenTime)
		}
	}
}

func TestMemoryStore_UpsertFinal_Update(t *testing.T) {
	store := NewMemoryStore(100)
	baseTime := time.Now().UnixMilli()

	// Insert a kline
	k1 := createKline(baseTime, 100.0)
	store.UpsertFinal("BTCUSDT", "1m", k1)

	// Update the same kline
	k2 := createKline(baseTime, 200.0) // Same OpenTime, different Close
	store.UpsertFinal("BTCUSDT", "1m", k2)

	klines, ok := store.GetRecent("BTCUSDT", "1m", 0)
	if !ok {
		t.Fatal("expected klines to exist")
	}
	if len(klines) != 1 {
		t.Fatalf("expected 1 kline, got %d", len(klines))
	}
	if klines[0].Close != 200.0 {
		t.Fatalf("expected Close 200.0, got %f", klines[0].Close)
	}
}

func TestMemoryStore_UpsertFinal_InsertMiddle(t *testing.T) {
	store := NewMemoryStore(100)
	baseTime := time.Now().UnixMilli()

	// Insert klines: 0, 2, 3, 4
	for i := 0; i < 5; i++ {
		if i == 1 {
			continue // Skip 1
		}
		k := createKline(baseTime+int64(i*60000), 100.0+float64(i))
		store.UpsertFinal("BTCUSDT", "1m", k)
	}

	// Insert missing kline at position 1
	k1 := createKline(baseTime+60000, 101.0)
	store.UpsertFinal("BTCUSDT", "1m", k1)

	klines, ok := store.GetRecent("BTCUSDT", "1m", 0)
	if !ok {
		t.Fatal("expected klines to exist")
	}
	if len(klines) != 5 {
		t.Fatalf("expected 5 klines, got %d", len(klines))
	}

	// Verify order
	for i := 0; i < 5; i++ {
		expectedTime := baseTime + int64(i*60000)
		if klines[i].OpenTime != expectedTime {
			t.Fatalf("expected OpenTime %d at index %d, got %d", expectedTime, i, klines[i].OpenTime)
		}
	}
}

func TestMemoryStore_UpsertFinal_CapacityLimit(t *testing.T) {
	store := NewMemoryStore(10)
	baseTime := time.Now().UnixMilli()

	// Insert more than capacity
	for i := 0; i < 15; i++ {
		k := createKline(baseTime+int64(i*60000), 100.0+float64(i))
		store.UpsertFinal("BTCUSDT", "1m", k)
	}

	klines, ok := store.GetRecent("BTCUSDT", "1m", 0)
	if !ok {
		t.Fatal("expected klines to exist")
	}
	if len(klines) != 10 {
		t.Fatalf("expected 10 klines (capacity limit), got %d", len(klines))
	}

	// Verify only the latest 10 are kept
	expectedFirstTime := baseTime + int64(5*60000)
	if klines[0].OpenTime != expectedFirstTime {
		t.Fatalf("expected first OpenTime %d, got %d", expectedFirstTime, klines[0].OpenTime)
	}
}

func TestMemoryStore_UpsertFinal_ClearsProvisional(t *testing.T) {
	store := NewMemoryStore(100)
	baseTime := time.Now().UnixMilli()

	// Insert provisional
	kProv := createKline(baseTime, 100.0)
	store.UpsertProvisional("BTCUSDT", "1m", kProv)

	// Insert final with same OpenTime
	kFinal := createKline(baseTime, 200.0)
	store.UpsertFinal("BTCUSDT", "1m", kFinal)

	// Provisional should be cleared (we can't directly check, but final should be there)
	klines, ok := store.GetRecent("BTCUSDT", "1m", 0)
	if !ok {
		t.Fatal("expected klines to exist")
	}
	if len(klines) != 1 {
		t.Fatalf("expected 1 kline, got %d", len(klines))
	}
	if klines[0].Close != 200.0 {
		t.Fatalf("expected Close 200.0, got %f", klines[0].Close)
	}
}

func TestMemoryStore_UpsertFinalBatch(t *testing.T) {
	store := NewMemoryStore(100)
	baseTime := time.Now().UnixMilli()

	// Create batch of klines
	klines := make([]marketKline, 5)
	for i := 0; i < 5; i++ {
		klines[i] = createKline(baseTime+int64(i*60000), 100.0+float64(i))
	}

	store.UpsertFinalBatch("BTCUSDT", "1m", klines)

	result, ok := store.GetRecent("BTCUSDT", "1m", 0)
	if !ok {
		t.Fatal("expected klines to exist")
	}
	if len(result) != 5 {
		t.Fatalf("expected 5 klines, got %d", len(result))
	}

	// Verify order
	for i := 0; i < 5; i++ {
		expectedTime := baseTime + int64(i*60000)
		if result[i].OpenTime != expectedTime {
			t.Fatalf("expected OpenTime %d at index %d, got %d", expectedTime, i, result[i].OpenTime)
		}
	}
}

func TestMemoryStore_UpsertFinalBatch_WithDuplicates(t *testing.T) {
	store := NewMemoryStore(100)
	baseTime := time.Now().UnixMilli()

	// Create batch with duplicates (same OpenTime)
	klines := []marketKline{
		createKline(baseTime, 100.0),
		createKline(baseTime, 200.0), // Same OpenTime, should replace first
		createKline(baseTime+60000, 101.0),
	}

	store.UpsertFinalBatch("BTCUSDT", "1m", klines)

	result, ok := store.GetRecent("BTCUSDT", "1m", 0)
	if !ok {
		t.Fatal("expected klines to exist")
	}
	if len(result) != 2 {
		t.Fatalf("expected 2 klines (duplicate removed), got %d", len(result))
	}
	// The last occurrence should be kept
	if result[0].Close != 200.0 {
		t.Fatalf("expected Close 200.0 (last duplicate), got %f", result[0].Close)
	}
}

func TestMemoryStore_UpsertFinalBatch_Merge(t *testing.T) {
	store := NewMemoryStore(100)
	baseTime := time.Now().UnixMilli()

	// Insert initial klines: 0, 2, 4
	initial := []marketKline{
		createKline(baseTime, 100.0),
		createKline(baseTime+2*60000, 102.0),
		createKline(baseTime+4*60000, 104.0),
	}
	store.UpsertFinalBatch("BTCUSDT", "1m", initial)

	// Insert batch with: 1, 2 (update), 3
	batch := []marketKline{
		createKline(baseTime+60000, 101.0),
		createKline(baseTime+2*60000, 202.0), // Update existing
		createKline(baseTime+3*60000, 103.0),
	}
	store.UpsertFinalBatch("BTCUSDT", "1m", batch)

	result, ok := store.GetRecent("BTCUSDT", "1m", 0)
	if !ok {
		t.Fatal("expected klines to exist")
	}
	if len(result) != 5 {
		t.Fatalf("expected 5 klines, got %d", len(result))
	}

	// Verify order and values
	expected := []float64{100.0, 101.0, 202.0, 103.0, 104.0}
	for i, exp := range expected {
		if result[i].Close != exp {
			t.Fatalf("expected Close %f at index %d, got %f", exp, i, result[i].Close)
		}
	}
}

func TestMemoryStore_UpsertFinalBatch_CapacityLimit(t *testing.T) {
	store := NewMemoryStore(10)
	baseTime := time.Now().UnixMilli()

	// Insert batch larger than capacity
	klines := make([]marketKline, 15)
	for i := 0; i < 15; i++ {
		klines[i] = createKline(baseTime+int64(i*60000), 100.0+float64(i))
	}

	store.UpsertFinalBatch("BTCUSDT", "1m", klines)

	result, ok := store.GetRecent("BTCUSDT", "1m", 0)
	if !ok {
		t.Fatal("expected klines to exist")
	}
	if len(result) != 10 {
		t.Fatalf("expected 10 klines (capacity limit), got %d", len(result))
	}

	// Verify only the latest 10 are kept
	expectedFirstTime := baseTime + int64(5*60000)
	if result[0].OpenTime != expectedFirstTime {
		t.Fatalf("expected first OpenTime %d, got %d", expectedFirstTime, result[0].OpenTime)
	}
}

func TestMemoryStore_UpsertProvisional(t *testing.T) {
	store := NewMemoryStore(100)
	baseTime := time.Now().UnixMilli()

	// Insert provisional
	k1 := createKline(baseTime, 100.0)
	store.UpsertProvisional("BTCUSDT", "1m", k1)

	// Update provisional
	k2 := createKline(baseTime, 200.0)
	store.UpsertProvisional("BTCUSDT", "1m", k2)

	// Provisional should not appear in GetRecent
	klines, ok := store.GetRecent("BTCUSDT", "1m", 0)
	if ok {
		t.Fatal("provisional should not appear in GetRecent")
	}
	if len(klines) != 0 {
		t.Fatalf("expected no klines, got %d", len(klines))
	}
}

func TestMemoryStore_GetRecent(t *testing.T) {
	store := NewMemoryStore(100)
	baseTime := time.Now().UnixMilli()

	// Insert 10 klines
	for i := 0; i < 10; i++ {
		k := createKline(baseTime+int64(i*60000), 100.0+float64(i))
		store.UpsertFinal("BTCUSDT", "1m", k)
	}

	// Get all
	klines, ok := store.GetRecent("BTCUSDT", "1m", 0)
	if !ok {
		t.Fatal("expected klines to exist")
	}
	if len(klines) != 10 {
		t.Fatalf("expected 10 klines, got %d", len(klines))
	}

	// Get last 5
	klines5, ok := store.GetRecent("BTCUSDT", "1m", 5)
	if !ok {
		t.Fatal("expected klines to exist")
	}
	if len(klines5) != 5 {
		t.Fatalf("expected 5 klines, got %d", len(klines5))
	}

	// Verify they are the last 5
	expectedFirstTime := baseTime + int64(5*60000)
	if klines5[0].OpenTime != expectedFirstTime {
		t.Fatalf("expected first OpenTime %d, got %d", expectedFirstTime, klines5[0].OpenTime)
	}
}

func TestMemoryStore_GetRecent_NonExistent(t *testing.T) {
	store := NewMemoryStore(100)

	klines, ok := store.GetRecent("BTCUSDT", "1m", 10)
	if ok {
		t.Fatal("expected no klines")
	}
	if len(klines) != 0 {
		t.Fatalf("expected empty slice, got %d", len(klines))
	}
}

func TestMemoryStore_GetRecent_LimitLargerThanData(t *testing.T) {
	store := NewMemoryStore(100)
	baseTime := time.Now().UnixMilli()

	// Insert 5 klines
	for i := 0; i < 5; i++ {
		k := createKline(baseTime+int64(i*60000), 100.0+float64(i))
		store.UpsertFinal("BTCUSDT", "1m", k)
	}

	// Request more than available
	klines, ok := store.GetRecent("BTCUSDT", "1m", 100)
	if !ok {
		t.Fatal("expected klines to exist")
	}
	if len(klines) != 5 {
		t.Fatalf("expected 5 klines, got %d", len(klines))
	}
}

func TestMemoryStore_MultipleSymbols(t *testing.T) {
	store := NewMemoryStore(100)
	baseTime := time.Now().UnixMilli()

	// Insert klines for different symbols
	k1 := createKline(baseTime, 100.0)
	store.UpsertFinal("BTCUSDT", "1m", k1)

	k2 := createKline(baseTime, 200.0)
	store.UpsertFinal("ETHUSDT", "1m", k2)

	// Verify they are separate
	btc, ok := store.GetRecent("BTCUSDT", "1m", 0)
	if !ok || len(btc) != 1 || btc[0].Close != 100.0 {
		t.Fatal("BTCUSDT kline incorrect")
	}

	eth, ok := store.GetRecent("ETHUSDT", "1m", 0)
	if !ok || len(eth) != 1 || eth[0].Close != 200.0 {
		t.Fatal("ETHUSDT kline incorrect")
	}
}

func TestMemoryStore_MultipleIntervals(t *testing.T) {
	store := NewMemoryStore(100)
	baseTime := time.Now().UnixMilli()

	// Insert klines for different intervals
	k1 := createKline(baseTime, 100.0)
	store.UpsertFinal("BTCUSDT", "1m", k1)

	k2 := createKline(baseTime, 200.0)
	store.UpsertFinal("BTCUSDT", "5m", k2)

	// Verify they are separate
	k1m, ok := store.GetRecent("BTCUSDT", "1m", 0)
	if !ok || len(k1m) != 1 || k1m[0].Close != 100.0 {
		t.Fatal("1m kline incorrect")
	}

	k5m, ok := store.GetRecent("BTCUSDT", "5m", 0)
	if !ok || len(k5m) != 1 || k5m[0].Close != 200.0 {
		t.Fatal("5m kline incorrect")
	}
}

func TestMemoryStore_CaseInsensitive(t *testing.T) {
	store := NewMemoryStore(100)
	baseTime := time.Now().UnixMilli()

	// Insert with uppercase
	k1 := createKline(baseTime, 100.0)
	store.UpsertFinal("BTCUSDT", "1M", k1)

	// Get with lowercase
	klines, ok := store.GetRecent("btcusdt", "1m", 0)
	if !ok || len(klines) != 1 {
		t.Fatal("case insensitive key matching failed")
	}
	if klines[0].Close != 100.0 {
		t.Fatalf("expected Close 100.0, got %f", klines[0].Close)
	}
}

func TestMemoryStore_ConcurrentAccess(t *testing.T) {
	store := NewMemoryStore(100)
	baseTime := time.Now().UnixMilli()

	var wg sync.WaitGroup
	// Concurrent writes
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			for j := 0; j < 10; j++ {
				k := createKline(baseTime+int64(id*100+j*60000), 100.0+float64(id*100+j))
				store.UpsertFinal("BTCUSDT", "1m", k)
			}
		}(i)
	}
	wg.Wait()

	// Concurrent reads
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			klines, ok := store.GetRecent("BTCUSDT", "1m", 50)
			if ok {
				_ = len(klines) // Just verify it doesn't panic
			}
		}()
	}
	wg.Wait()

	// Verify final state
	klines, ok := store.GetRecent("BTCUSDT", "1m", 0)
	if !ok {
		t.Fatal("expected klines to exist")
	}
	// Should have at least some klines (exact count depends on timing)
	if len(klines) == 0 {
		t.Fatal("expected at least some klines after concurrent writes")
	}

	// Verify they are sorted
	if !sort.SliceIsSorted(klines, func(i, j int) bool {
		return klines[i].OpenTime < klines[j].OpenTime
	}) {
		t.Error("klines should be sorted by OpenTime")
	}
}

func TestMemoryStore_DedupeByOpenTimeSorted(t *testing.T) {
	baseTime := time.Now().UnixMilli()

	// Create array with duplicates
	arr := []marketKline{
		createKline(baseTime, 100.0),
		createKline(baseTime, 200.0), // Duplicate
		createKline(baseTime+60000, 101.0),
		createKline(baseTime+60000, 201.0), // Duplicate
		createKline(baseTime+2*60000, 102.0),
	}

	result := dedupeByOpenTimeSorted(arr)
	if len(result) != 3 {
		t.Fatalf("expected 3 unique klines, got %d", len(result))
	}

	// Verify last occurrence is kept
	if result[0].Close != 200.0 {
		t.Fatalf("expected Close 200.0 (last duplicate), got %f", result[0].Close)
	}
	if result[1].Close != 201.0 {
		t.Fatalf("expected Close 201.0 (last duplicate), got %f", result[1].Close)
	}
}

func TestMemoryStore_MergeReplaceByOpenTime(t *testing.T) {
	baseTime := time.Now().UnixMilli()

	old := []marketKline{
		createKline(baseTime, 100.0),
		createKline(baseTime+2*60000, 102.0),
		createKline(baseTime+4*60000, 104.0),
	}

	add := []marketKline{
		createKline(baseTime+60000, 101.0),
		createKline(baseTime+2*60000, 202.0), // Should replace old[1]
		createKline(baseTime+3*60000, 103.0),
	}

	result := mergeReplaceByOpenTime(old, add)
	if len(result) != 5 {
		t.Fatalf("expected 5 klines, got %d", len(result))
	}

	// Verify order
	expected := []float64{100.0, 101.0, 202.0, 103.0, 104.0}
	for i, exp := range expected {
		if result[i].Close != exp {
			t.Fatalf("expected Close %f at index %d, got %f", exp, i, result[i].Close)
		}
	}
}

func TestMemoryStore_EmptyBatch(t *testing.T) {
	store := NewMemoryStore(100)

	// Insert empty batch should not panic
	store.UpsertFinalBatch("BTCUSDT", "1m", []marketKline{})

	klines, ok := store.GetRecent("BTCUSDT", "1m", 0)
	if ok {
		t.Fatal("expected no klines after empty batch")
	}
	if len(klines) != 0 {
		t.Fatalf("expected empty slice, got %d", len(klines))
	}
}

func TestMemoryStore_OutOfOrderInsert(t *testing.T) {
	store := NewMemoryStore(100)
	baseTime := time.Now().UnixMilli()

	// Insert out of order
	store.UpsertFinal("BTCUSDT", "1m", createKline(baseTime+2*60000, 102.0))
	store.UpsertFinal("BTCUSDT", "1m", createKline(baseTime, 100.0))
	store.UpsertFinal("BTCUSDT", "1m", createKline(baseTime+1*60000, 101.0))

	klines, ok := store.GetRecent("BTCUSDT", "1m", 0)
	if !ok {
		t.Fatal("expected klines to exist")
	}
	if len(klines) != 3 {
		t.Fatalf("expected 3 klines, got %d", len(klines))
	}

	// Verify they are sorted
	for i := 0; i < 3; i++ {
		expectedTime := baseTime + int64(i*60000)
		if klines[i].OpenTime != expectedTime {
			t.Fatalf("expected OpenTime %d at index %d, got %d", expectedTime, i, klines[i].OpenTime)
		}
	}
}
