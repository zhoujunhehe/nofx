package kline

import (
	"context"
	"encoding/json"
	"log"
	"strconv"
	"strings"
	"sync"
	"time"
)

// Only parses fields we care about (compatible with both raw payload and wrapped payload)
type wsKlineEnvelope struct {
	Stream string          `json:"stream,omitempty"`
	Data   json.RawMessage `json:"data,omitempty"`
}

// WSIngestor: reads from each stream's WS channel and writes to Store
type WSIngestor struct {
	store Store

	mu      sync.RWMutex
	streams map[string]*streamState // stream -> state

	closed bool
}

type streamState struct {
	cancel      context.CancelFunc
	lastProTS   int64     // Last provisional open_time
	lastProTime time.Time // Last provisional write time, used for throttling
}

func NewWSIngestor(store Store) *WSIngestor {
	return &WSIngestor{
		store:   store,
		streams: make(map[string]*streamState),
	}
}

// Attach starts processing a stream's channel; returns cancel function (optional).
// - stream: e.g., "btcusdt@kline_1m"
// - symbol/interval used as key for writing; symbol will be uppercased, interval lowercased
// - ch: channel returned from RawWSClient.AddSubscriber
func (w *WSIngestor) Attach(stream, symbol, interval string, ch <-chan []byte) (context.CancelFunc, bool) {
	key := strings.ToLower(strings.TrimSpace(stream))
	if key == "" {
		return func() {}, false
	}

	w.mu.Lock()
	if w.closed {
		w.mu.Unlock()
		return func() {}, false
	}
	if _, exists := w.streams[key]; exists {
		w.mu.Unlock()
		return func() {}, false // Already exists, don't attach again
	}
	ctx, cancel := context.WithCancel(context.Background())
	w.streams[key] = &streamState{cancel: cancel}
	w.mu.Unlock()

	sym := strings.ToUpper(strings.TrimSpace(symbol))
	iv := strings.ToLower(strings.TrimSpace(interval))

	go w.consumeLoop(ctx, key, sym, iv, ch)
	return cancel, true
}

func (w *WSIngestor) Detach(stream string) {
	key := strings.ToLower(strings.TrimSpace(stream))
	w.mu.Lock()
	if st, ok := w.streams[key]; ok {
		if st.cancel != nil {
			st.cancel()
		}
		delete(w.streams, key)
	}
	w.mu.Unlock()
}

// Close stops all stream processing
func (w *WSIngestor) Close() {
	w.mu.Lock()
	if w.closed {
		w.mu.Unlock()
		return
	}
	w.closed = true
	for k, st := range w.streams {
		if st.cancel != nil {
			st.cancel()
		}
		delete(w.streams, k)
	}
	w.mu.Unlock()
}

// ------- Internal Implementation -------
func (w *WSIngestor) consumeLoop(ctx context.Context, streamKey, symbol, interval string, ch <-chan []byte) {
	defer func() {
		// Cleanup: delete state
		w.mu.Lock()
		delete(w.streams, streamKey)
		w.mu.Unlock()
		// Defense: panic recover
		if r := recover(); r != nil {
			log.Printf("WSIngestor panic recovered on %s: %v", streamKey, r)
		}
	}()

	for {
		select {
		case <-ctx.Done():
			return
		case msg, ok := <-ch:
			if !ok {
				return
			}
			w.onMessage(streamKey, symbol, interval, msg)
		}
	}
}

func (w *WSIngestor) onMessage(streamKey, symbol, interval string, msg []byte) {
	// 1) May be wrapped structure {"stream": "...", "data": {...}}
	var env wsKlineEnvelope
	if json.Unmarshal(msg, &env) == nil && len(env.Data) > 0 {
		w.handleKlinePayload(streamKey, symbol, interval, env.Data)
		return
	}
	// 2) Raw kline payload
	w.handleKlinePayload(streamKey, symbol, interval, msg)
}

func (w *WSIngestor) handleKlinePayload(streamKey, symbol, interval string, payload []byte) {
	var ev Evt
	if err := json.Unmarshal(payload, &ev); err != nil {
		log.Printf("kline WS unmarshal failed: %v", err)
		return
	}
	// Basic validation
	if ev.S == "" || ev.K.I == "" || ev.K.T == 0 {
		return
	}

	k := marketKline{
		OpenTime:            ev.K.T,
		CloseTime:           ev.K.T1,
		Open:                parseF(ev.K.O),
		High:                parseF(ev.K.H),
		Low:                 parseF(ev.K.L1),
		Close:               parseF(ev.K.C),
		Volume:              parseF(ev.K.V),
		QuoteVolume:         parseF(ev.K.Q),
		Trades:              ev.K.N,
		TakerBuyBaseVolume:  parseF(ev.K.V1),
		TakerBuyQuoteVolume: parseF(ev.K.Q1),
	}

	if ev.K.X {
		// final: write directly
		w.store.UpsertFinal(symbol, interval, k)
		// After final arrives, clear provisional memory for this stream
		w.mu.Lock()
		if st, ok := w.streams[streamKey]; ok {
			st.lastProTS = 0
			st.lastProTime = time.Time{}
		}
		w.mu.Unlock()
		return
	}

	// provisional: throttle and deduplicate (high-frequency increments with same open_time only kept at certain frequency)
	const proMinInterval = 300 * time.Millisecond // You can adjust up/down based on load
	write := false

	w.mu.Lock()
	if st, ok := w.streams[streamKey]; ok {
		now := time.Now()
		if st.lastProTS != k.OpenTime {
			// New bar (open_time changed), allow write and update state
			write = true
			st.lastProTS = k.OpenTime
			st.lastProTime = now
		} else {
			// Multiple increments of same bar, throttle by time
			if now.Sub(st.lastProTime) >= proMinInterval {
				write = true
				st.lastProTime = now
			}
		}
	}
	w.mu.Unlock()

	if write {
		w.store.UpsertProvisional(symbol, interval, k)
	}
}

func parseF(s string) float64 {
	if s == "" {
		return 0
	}
	f, err := strconv.ParseFloat(s, 64)
	if err != nil {
		return 0
	}
	return f
}
