package kline

import (
	"sort"
	"strings"
	"sync"
)

const (
	// DefaultStoreCapacity is the default max number of finalized klines kept per symbol×interval.
	DefaultStoreCapacity = 500
)

// Store abstracts persistence for klines.
type Store interface {
	UpsertFinal(symbol, interval string, k marketKline)
	UpsertFinalBatch(symbol, interval string, ks []marketKline)
	UpsertProvisional(symbol, interval string, k marketKline)
	GetRecent(symbol, interval string, limit int) ([]marketKline, bool)
}

// MemoryStore keeps recent klines in-memory per symbol×interval with a size cap.
// data[key] is always sorted by OpenTime ascending and at most one entry per OpenTime.
type MemoryStore struct {
	cap int
	mu  sync.RWMutex
	// key: "SYM|iv" -> sorted by OpenTime ascending
	data map[string][]marketKline
	// provisional buffer: key -> last provisional kline (only for upper layer fallback display; final write will clear it)
	prov map[string]marketKline
}

func NewMemoryStore(capacity int) *MemoryStore {
	if capacity <= 0 {
		capacity = DefaultStoreCapacity
	}
	return &MemoryStore{
		cap:  capacity,
		data: make(map[string][]marketKline),
		prov: make(map[string]marketKline),
	}
}

func (m *MemoryStore) makeKey(symbol, interval string) string {
	return strings.ToUpper(strings.TrimSpace(symbol)) + "|" + strings.ToLower(strings.TrimSpace(interval))
}

// ------- Single upsert (binary insert/replace) -------

func (m *MemoryStore) UpsertFinal(symbol, interval string, k marketKline) {
	key := m.makeKey(symbol, interval)

	m.mu.Lock()
	defer m.mu.Unlock()

	arr := m.data[key]
	// Fast path: most cases append to tail or overwrite tail
	if n := len(arr); n > 0 {
		last := arr[n-1].OpenTime
		if k.OpenTime == last {
			arr[n-1] = k
			m.data[key] = arr
			delete(m.prov, key)
			return
		}
		if k.OpenTime > last {
			arr = append(arr, k)
			// Capacity trimming
			if len(arr) > m.cap {
				arr = arr[len(arr)-m.cap:]
			}
			m.data[key] = arr
			delete(m.prov, key)
			return
		}
	}
	// Less common: insert in middle or replace historical position
	i := sort.Search(len(arr), func(i int) bool { return arr[i].OpenTime >= k.OpenTime })
	if i < len(arr) && arr[i].OpenTime == k.OpenTime {
		arr[i] = k
		m.data[key] = arr
		delete(m.prov, key)
		return
	}
	// Insert in middle: expand and shift right
	arr = append(arr, marketKline{}) // Grow by one
	copy(arr[i+1:], arr[i:])         // Shift right
	arr[i] = k
	// Capacity trimming: keep last cap entries (newest at end)
	if len(arr) > m.cap {
		arr = arr[len(arr)-m.cap:]
	}
	m.data[key] = arr
	delete(m.prov, key)
}

// ------- Batch upsert (merge) -------

func (m *MemoryStore) UpsertFinalBatch(symbol, interval string, ks []marketKline) {
	if len(ks) == 0 {
		return
	}
	key := m.makeKey(symbol, interval)

	// Preprocessing: sort input ks by OpenTime and deduplicate (keep last occurrence)
	sort.Slice(ks, func(i, j int) bool { return ks[i].OpenTime < ks[j].OpenTime })
	ks = dedupeByOpenTimeSorted(ks)

	m.mu.Lock()
	defer m.mu.Unlock()

	old := m.data[key]
	merged := mergeReplaceByOpenTime(old, ks)

	// Capacity trimming (only keep latest cap entries)
	if len(merged) > m.cap {
		merged = merged[len(merged)-m.cap:]
	}
	m.data[key] = merged
	delete(m.prov, key)
}

// dedupe input is already sorted by OpenTime: keep only the last one for same OpenTime
func dedupeByOpenTimeSorted(arr []marketKline) []marketKline {
	if len(arr) <= 1 {
		return arr
	}
	write := 1
	for read := 1; read < len(arr); read++ {
		if arr[read].OpenTime == arr[write-1].OpenTime {
			// Overwrite with the "newer" one
			arr[write-1] = arr[read]
		} else {
			if write != read {
				arr[write] = arr[read]
			}
			write++
		}
	}
	return arr[:write]
}

// Merges old (sorted and unique) with add (sorted and unique) by OpenTime, overwrites with add when same ts.
func mergeReplaceByOpenTime(old, add []marketKline) []marketKline {
	i, j := 0, 0
	// Pre-allocate capacity to reduce reallocation
	out := make([]marketKline, 0, len(old)+len(add))
	for i < len(old) && j < len(add) {
		ot, at := old[i].OpenTime, add[j].OpenTime
		if ot < at {
			out = append(out, old[i])
			i++
		} else if ot > at {
			out = append(out, add[j])
			j++
		} else {
			// Same ts: overwrite with add
			out = append(out, add[j])
			i++
			j++
		}
	}
	if i < len(old) {
		out = append(out, old[i:]...)
	}
	if j < len(add) {
		out = append(out, add[j:]...)
	}
	return out
}

// ------- Provisional Cache -------

func (m *MemoryStore) UpsertProvisional(symbol, interval string, k marketKline) {
	key := m.makeKey(symbol, interval)
	m.mu.Lock()
	m.prov[key] = k
	m.mu.Unlock()
}

// GetRecent returns the most recent limit "final" klines; second return value indicates if data exists.
// Note: does not include provisional; if you need fallback return, can read m.prov externally and combine.
func (m *MemoryStore) GetRecent(symbol, interval string, limit int) ([]marketKline, bool) {
	key := m.makeKey(symbol, interval)
	m.mu.RLock()
	arr := m.data[key]
	n := len(arr)
	m.mu.RUnlock()
	if n == 0 {
		return nil, false
	}
	if limit <= 0 || limit >= n {
		out := make([]marketKline, n)
		copy(out, arr)
		return out, true
	}
	out := make([]marketKline, limit)
	copy(out, arr[n-limit:])
	return out, true
}
