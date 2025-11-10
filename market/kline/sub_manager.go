package kline

import (
	"fmt"
	"log"
	"maps"
	"slices"
	"strings"
	"sync"
)

type wsClient interface {
	SubscribeStreams([]string) error
	UnsubscribeStreams([]string) error
}

// SubManager manages dynamic subscription/unsubscription of kline streams (maintains symbol sets per interval)
type SubManager struct {
	c  wsClient
	mu sync.RWMutex
	// interval(lower) -> set(symbol UPPER)
	subs map[string]map[string]struct{}
	// Maximum number of streams per control message (avoid params being too long at once)
	maxBatch int
}

func NewSubManager(c wsClient) *SubManager {
	return &SubManager{
		c: c,
		// 100~200 is generally safe; Spot/USDM control messages also have rate limits handled internally by wsClient
		maxBatch: 150,
		subs:     make(map[string]map[string]struct{}),
	}
}

// -------- Public API --------

// AddSymbols subscribes a batch of symbols (already subscribed ones are automatically skipped); returns actually added symbols (uppercase)
func (sm *SubManager) AddSymbols(interval string, symbols []string) ([]string, error) {
	iv := strings.ToLower(strings.TrimSpace(interval))
	if len(symbols) == 0 || iv == "" {
		return nil, nil
	}

	// Normalize + compute what needs to be added
	toAdd := make([]string, 0, len(symbols))
	sm.mu.Lock()
	set := sm.ensure(iv)
	for _, s := range symbols {
		su := strings.ToUpper(strings.TrimSpace(s))
		if su == "" {
			continue
		}
		if _, ok := set[su]; ok {
			continue
		}
		set[su] = struct{}{} // Put in first (optimistic memory state update)
		toAdd = append(toAdd, su)
	}
	sm.mu.Unlock()

	if len(toAdd) == 0 {
		return nil, nil
	}

	// Send subscription (batched), rollback on failure
	streams := sm.buildStreams(iv, toAdd)
	if err := sm.callBatched(sm.c.SubscribeStreams, streams); err != nil {
		// Rollback
		sm.mu.Lock()
		set = sm.ensure(iv)
		for _, su := range toAdd {
			delete(set, su)
		}
		sm.mu.Unlock()
		log.Printf("SubscribeStreams failed: %v", err)
		return nil, err
	}
	return toAdd, nil
}

// RemoveSymbols unsubscribes a batch of symbols; returns actually removed symbols (uppercase)
func (sm *SubManager) RemoveSymbols(interval string, symbols []string) ([]string, error) {
	iv := strings.ToLower(strings.TrimSpace(interval))
	if len(symbols) == 0 || iv == "" {
		return nil, nil
	}

	// Compute what needs to be removed (remove from memory first, add back on failure)
	toRemove := make([]string, 0, len(symbols))
	sm.mu.Lock()
	set := sm.ensure(iv)
	for _, s := range symbols {
		su := strings.ToUpper(strings.TrimSpace(s))
		if su == "" {
			continue
		}
		if _, ok := set[su]; ok {
			delete(set, su) // Delete first
			toRemove = append(toRemove, su)
		}
	}
	sm.mu.Unlock()

	if len(toRemove) == 0 {
		return nil, nil
	}

	streams := sm.buildStreams(iv, toRemove)
	if err := sm.callBatched(sm.c.UnsubscribeStreams, streams); err != nil {
		// Rollback
		sm.mu.Lock()
		set = sm.ensure(iv)
		for _, su := range toRemove {
			set[su] = struct{}{}
		}
		sm.mu.Unlock()
		log.Printf("UnsubscribeStreams failed: %v", err)
		return nil, err
	}
	return toRemove, nil
}

// Sync synchronizes an interval's subscription set to desired (minimizes SUB/UNSUB)
func (sm *SubManager) Sync(interval string, desired []string) (added, removed []string, err error) {
	iv := strings.ToLower(strings.TrimSpace(interval))
	if iv == "" {
		return nil, nil, fmt.Errorf("invalid interval")
	}
	// Normalize desired
	desiredSet := make(map[string]struct{}, len(desired))
	for _, s := range desired {
		su := strings.ToUpper(strings.TrimSpace(s))
		if su != "" {
			desiredSet[su] = struct{}{}
		}
	}

	// Compute diff
	sm.mu.RLock()
	current := sm.copySet(iv)
	sm.mu.RUnlock()

	for su := range desiredSet {
		if _, ok := current[su]; !ok {
			added = append(added, su)
		}
	}
	for su := range current {
		if _, ok := desiredSet[su]; !ok {
			removed = append(removed, su)
		}
	}

	// Execute
	if len(added) > 0 {
		if _, e := sm.AddSymbols(iv, added); e != nil {
			err = e
		}
	}
	if len(removed) > 0 {
		if _, e := sm.RemoveSymbols(iv, removed); e != nil {
			if err != nil {
				err = fmt.Errorf("%v; %w", err, e)
			} else {
				err = e
			}
		}
	}
	return added, removed, err
}

// Symbols returns subscription snapshot for an interval (uppercase)
func (sm *SubManager) Symbols(interval string) []string {
	iv := strings.ToLower(strings.TrimSpace(interval))
	sm.mu.RLock()
	defer sm.mu.RUnlock()
	set := sm.subs[iv]
	out := make([]string, 0, len(set))
	for s := range set {
		out = append(out, s)
	}
	// Sort for comparison/testing
	slices.Sort(out)
	return out
}

// StreamsFor returns all stream names for an interval (lowercase ws stream names)
func (sm *SubManager) StreamsFor(interval string) []string {
	iv := strings.ToLower(strings.TrimSpace(interval))
	sm.mu.RLock()
	defer sm.mu.RUnlock()
	set := sm.subs[iv]
	out := make([]string, 0, len(set))
	for s := range set {
		out = append(out, buildKlineStream(s, iv))
	}
	slices.Sort(out)
	return out
}

// Has checks if already subscribed (symbol case insensitive)
func (sm *SubManager) Has(interval, symbol string) bool {
	iv := strings.ToLower(strings.TrimSpace(interval))
	su := strings.ToUpper(strings.TrimSpace(symbol))
	sm.mu.RLock()
	defer sm.mu.RUnlock()
	set := sm.subs[iv]
	_, ok := set[su]
	return ok
}

// Count returns number of subscribed symbols under interval
func (sm *SubManager) Count(interval string) int {
	iv := strings.ToLower(strings.TrimSpace(interval))
	sm.mu.RLock()
	defer sm.mu.RUnlock()
	return len(sm.subs[iv])
}

// All returns full copy: interval -> set(symbol UPPER)
func (sm *SubManager) All() map[string]map[string]struct{} {
	sm.mu.RLock()
	defer sm.mu.RUnlock()
	out := make(map[string]map[string]struct{}, len(sm.subs))
	for iv, set := range sm.subs {
		out[iv] = maps.Clone(set)
	}
	return out
}

// -------- Internal Utilities --------
// buildStreams combines a set of symbols (uppercase) with interval (lowercase) into ws stream names (lowercase).
func (sm *SubManager) buildStreams(intervalLower string, symbolsUpper []string) []string {
	if intervalLower == "" || len(symbolsUpper) == 0 {
		return nil
	}
	out := make([]string, 0, len(symbolsUpper))
	for _, su := range symbolsUpper {
		// buildKlineStream will convert final stream name to lowercase, e.g., btcusdt@kline_1m
		out = append(out, buildKlineStream(su, intervalLower))
	}
	return out
}

func (sm *SubManager) ensure(iv string) map[string]struct{} {
	if m := sm.subs[iv]; m != nil {
		return m
	}
	m := make(map[string]struct{})
	sm.subs[iv] = m
	return m
}

func (sm *SubManager) copySet(iv string) map[string]struct{} {
	src := sm.subs[iv]
	dst := make(map[string]struct{}, len(src))
	for k := range src {
		dst[k] = struct{}{}
	}
	return dst
}

func (sm *SubManager) callBatched(fn func([]string) error, streams []string) error {
	if len(streams) == 0 {
		return nil
	}
	batch := sm.maxBatch
	for i := 0; i < len(streams); i += batch {
		j := i + batch
		if j > len(streams) {
			j = len(streams)
		}
		if err := fn(streams[i:j]); err != nil {
			return err
		}
	}
	return nil
}
