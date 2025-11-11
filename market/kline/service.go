package kline

import (
	"context"
	"log"
	"strconv"
	"strings"
	"sync"
	"time"
)

type client interface {
	Connect() error
	AddSubscriber(string, int) <-chan []byte
	RemoveSubscriber(string)
	Close()
}

type Service struct {
	// interval(lower) -> client (each interval has its own WebSocket connection)
	clients            map[string]client
	store              Store
	rest               RestReconciler
	subs               *SubManager
	ingestor           *WSIngestor
	rate               *Governor
	symbolSync         *SymbolSync
	intervals          []string
	backfillWindow     int
	restMaxConcurrency int
	enableBackfill     bool
	wsEndpoint         string

	// stream -> cancel (returned by WSIngestor.Attach)
	subCancels map[string]context.CancelFunc

	// on-demand backfill in-progress flags keyed by "SYMBOL|interval"
	onDemand map[string]bool

	// suppress repeated backfills when series is inherently short (e.g., late-listed token)
	// key: "SYMBOL|interval" -> expiry time
	shortSuppress map[string]time.Time

	mu sync.RWMutex
}

// Default is the package-level service instance for consumers who prefer a singleton.
var Default *Service

const (
	ReconcileTick3m         = 30 * time.Second
	ReconcileTick15m        = 1 * time.Minute
	ReconcileTick1h         = 10 * time.Minute
	ReconcileTick4h         = 1 * time.Hour
	ReconcileTick8h         = 2 * time.Hour
	ReconcileTick1d         = 12 * time.Hour
	ReconcileTick1w         = 24 * time.Hour
	DefaultSubscriberBuffer = 256
	EnsureReadyMinBars      = 10
	maxRetries              = 3
	retryDelay              = 300 * time.Millisecond
)

type Options struct {
	BatchSize             int
	Intervals             []string // e.g. []string{"3m","4h"}
	BackfillWindow        int
	WSEndpoint            string
	RestMaxConcurrency    int
	SymbolRefreshInterval time.Duration
	// EnableRestBackfill toggles REST-based backfilling and reconciliation. Default: false (off).
	EnableRestBackfill bool
}

func DefaultOptions() Options {
	return Options{
		BatchSize:             150,
		Intervals:             DefaultIntervals,
		BackfillWindow:        100,
		RestMaxConcurrency:    20,
		SymbolRefreshInterval: time.Duration(DefaultSymbolRefreshInterval),
		EnableRestBackfill:    false,
	}
}

func NewService(opts Options) *Service {
	if len(opts.Intervals) == 0 {
		opts = DefaultOptions()
	}

	rate := NewGovernor()
	memStore := NewMemoryStore(DefaultStoreCapacity)
	reconciler := NewRestReconciler(rate, memStore)
	subs := NewSubManager()
	ingestor := NewWSIngestor(memStore)

	// Create clients map for each interval
	clients := make(map[string]client)
	for _, iv := range opts.Intervals {
		ivLower := strings.ToLower(iv)
		cl := NewRawWSClient(opts.WSEndpoint)
		clients[ivLower] = cl
		subs.SetClient(ivLower, cl)
	}

	svc := &Service{
		clients:            clients,
		store:              memStore,
		rest:               reconciler,
		subs:               subs,
		ingestor:           ingestor,
		rate:               rate,
		intervals:          opts.Intervals,
		backfillWindow:     opts.BackfillWindow,
		restMaxConcurrency: opts.RestMaxConcurrency,
		enableBackfill:     opts.EnableRestBackfill,
		wsEndpoint:         opts.WSEndpoint,
		subCancels:         make(map[string]context.CancelFunc),
		symbolSync:         NewSymbolSync(opts.SymbolRefreshInterval, NewFilteredBinanceFuturesFetcher(NewKlineHTTPClient())),
	}
	return svc
}

// Start: background WS connection & periodic reconcile (does not block caller)
func (s *Service) Start(ctx context.Context) error {
	// Start a separate WebSocket connection for each interval
	s.mu.RLock()
	intervals := make([]string, len(s.intervals))
	copy(intervals, s.intervals)
	clients := make(map[string]client)
	for k, v := range s.clients {
		clients[k] = v
	}
	s.mu.RUnlock()

	for _, iv := range intervals {
		ivLower := strings.ToLower(iv)
		cl, ok := clients[ivLower]
		if !ok {
			continue
		}
		// Background connection for this interval (Connect has infinite reconnect loop internally)
		go func(interval string, c client) {
			defer recoverGuard("Service.Connect[" + interval + "]")
			if err := c.Connect(); err != nil {
				log.Printf("[Service] WS connect exited for interval %s: %v", interval, err)
			}
		}(iv, cl)
	}

	// Periodic reconcile
	if s.enableBackfill {
		go s.runPeriodicReconcile(ctx)
	}
	// Start symbol sync (dynamic subscribe/unsubscribe)
	go s.symbolSync.Start(ctx, s.onSymbolsChange)
	return nil
}

// onSymbolsChange handles symbol add/remove callback: uniformly calls AddSymbols/RemoveSymbols
func (s *Service) onSymbolsChange(added, removed []string) {
	if len(added) > 0 {
		_ = s.AddSymbols(added)
	}
	if len(removed) > 0 {
		_ = s.RemoveSymbols(removed)
	}
}

// AddSymbols: dynamically subscribes for all intervals and starts ingestion & initial backfill
func (s *Service) AddSymbols(symbols []string) error {
	if len(symbols) == 0 {
		return nil
	}

	// Normalize
	upSyms := symSliceToUpper(symbols)

	s.mu.Lock()
	defer s.mu.Unlock()

	for _, iv := range s.intervals {
		ivLower := strings.ToLower(iv)

		// Subscribe (SubManager will deduplicate/batch/rate limit)
		if _, err := s.subs.AddSymbols(ivLower, upSyms); err != nil {
			return err
		}

		// Get client for this interval
		cl, ok := s.clients[ivLower]
		if !ok {
			continue
		}

		// Register subscriber and attach for each symbol
		for _, sym := range upSyms {
			stream := buildKlineStream(sym, ivLower)
			// Avoid duplicate attach
			if _, exists := s.subCancels[stream]; exists {
				continue
			}
			ch := cl.AddSubscriber(stream, DefaultSubscriberBuffer)
			cancel, ok := s.ingestor.Attach(stream, sym, ivLower, ch)
			if ok {
				s.subCancels[stream] = cancel
			}
		}

		// Initial small window backfill (entire batch at once) if enabled
		if s.enableBackfill {
			go func(iv string) {
				defer recoverGuard("Service.initialBackfill")
				s.restBackfillWindow(upSyms, iv, s.backfillWindow)
			}(ivLower)
		}
	}
	return nil
}

// RemoveSymbols: unsubscribes for all intervals and stops ingestion
func (s *Service) RemoveSymbols(symbols []string) error {
	if len(symbols) == 0 {
		return nil
	}

	upSyms := symSliceToUpper(symbols)

	s.mu.Lock()
	defer s.mu.Unlock()

	for _, iv := range s.intervals {
		ivLower := strings.ToLower(iv)

		if _, err := s.subs.RemoveSymbols(ivLower, upSyms); err != nil {
			return err
		}

		// Get client for this interval
		cl, ok := s.clients[ivLower]
		if !ok {
			continue
		}

		for _, sym := range upSyms {
			stream := buildKlineStream(sym, ivLower)

			// Stop ingestion goroutine
			if cancel, ok := s.subCancels[stream]; ok {
				cancel()
				delete(s.subCancels, stream)
			}

			// Let WS client remove channel (will close ch)
			cl.RemoveSubscriber(stream)

			// Notify Ingestor to cleanup local state (it will cleanup internally when ch closes; explicit here)
			s.ingestor.Detach(stream)
		}
	}
	return nil
}

// GetRecent: returns the most recent N "final" klines; second return value indicates if ready (per store semantics)
func (s *Service) GetRecent(symbol, interval string, limit int) ([]marketKline, bool) {
	return s.store.GetRecent(strings.ToUpper(symbol), strings.ToLower(interval), limit)
}

// Kline is the exported kline structure for external consumers (e.g., market package).
type Kline struct {
	OpenTime            int64
	Open                float64
	High                float64
	Low                 float64
	Close               float64
	Volume              float64
	CloseTime           int64
	QuoteVolume         float64
	Trades              int
	TakerBuyBaseVolume  float64
	TakerBuyQuoteVolume float64
}

// getRecentKlines returns the most recent klines as exported Kline type (internal implementation).
func (s *Service) getRecentKlines(symbol, interval string, limit int) ([]Kline, bool) {
	internal, ready := s.GetRecent(symbol, interval, limit)

	// On-demand completeness check and backfill (only first caller triggers)
	sym := strings.ToUpper(symbol)
	iv := strings.ToLower(interval)
	needBackfill := false
	stepMs := intervalMillis(iv)

	// Conditions for missing data:
	// 1) Not enough bars compared to requested limit
	shortOnly := limit > 0 && len(internal) < limit
	// 2) Middle gaps (non-contiguous OpenTime deltas)
	if stepMs > 0 && len(internal) >= 2 {
		for i := 1; i < len(internal); i++ {
			if internal[i].OpenTime-internal[i-1].OpenTime > stepMs {
				needBackfill = true
				break
			}
		}
	}
	// 3) Tail missing (latest bar behind expected schedule)
	if !needBackfill && stepMs > 0 && len(internal) > 0 {
		nowMs := time.Now().UnixMilli()
		expectedLastOpen := nowMs - (nowMs % stepMs)
		if expectedLastOpen-internal[len(internal)-1].OpenTime > stepMs {
			needBackfill = true
		}
	}

	// If data is missing and we can't get the lock, mark as not ready
	if needBackfill || shortOnly {
		if s.tryStartOnDemand(sym, iv) {
			defer s.endOnDemand(sym, iv)
			if shortOnly && !needBackfill {
				// If short-only and currently suppressed, skip requesting
				if s.isShortSuppressed(sym, iv) {
					// do nothing, will return current internal
				} else {
					window := max(limit, s.backfillWindow)
					if err := s.rest.ReconcileWindow(sym, iv, window); err == nil {
						// Only update suppress if reconciliation succeeded
						// Refresh view
						after, _ := s.GetRecent(sym, iv, limit)
						internal = after
						// If still short and no gaps/tail-missing, suppress for a cooldown
						if limit > 0 && len(internal) < limit && !hasGapsOrTailMissing(internal, stepMs) {
							s.setShortSuppress(sym, iv, shortSuppressDuration(iv))
						}
					}
				}
			} else {
				// Has gaps or tail missing -> reconcile immediately
				window := max(limit, s.backfillWindow)
				_ = s.rest.ReconcileWindow(sym, iv, window) // Ignore error, just try to reconcile
				internal, ready = s.GetRecent(symbol, interval, limit)
			}
		} else {
			// Data is missing but couldn't get lock (another goroutine is backfilling)
			// Only mark as not ready if there are actual gaps/tail missing, or if short-only but not suppressed
			// (suppressed means we already tried and data might be complete, just not enough history)
			if needBackfill || (shortOnly && !s.isShortSuppressed(sym, iv)) {
				ready = false
			}
		}
	}

	out := make([]Kline, len(internal))
	for i, mk := range internal {
		out[i] = Kline{
			OpenTime:            mk.OpenTime,
			Open:                mk.Open,
			High:                mk.High,
			Low:                 mk.Low,
			Close:               mk.Close,
			Volume:              mk.Volume,
			CloseTime:           mk.CloseTime,
			QuoteVolume:         mk.QuoteVolume,
			Trades:              mk.Trades,
			TakerBuyBaseVolume:  mk.TakerBuyBaseVolume,
			TakerBuyQuoteVolume: mk.TakerBuyQuoteVolume,
		}
	}
	return out, ready
}

// GetRecentKlines returns the most recent klines as exported Kline type.
// If ready is false, it will retry up to 3 times with delays to wait for other goroutines
// to complete data fetching.
func (s *Service) GetRecentKlines(symbol, interval string, limit int) ([]Kline, bool) {
	klines, ready := s.getRecentKlines(symbol, interval, limit)
	// If ready is false, retry up to maxRetries times
	if !ready {
		for i := 0; i < maxRetries; i++ {
			time.Sleep(retryDelay)
			klines, ready = s.getRecentKlines(symbol, interval, limit)
			if ready {
				break
			}
		}
	}

	return klines, ready
}

// EnsureReady: triggers a small window emergency backfill when minimum N bars not satisfied
func (s *Service) EnsureReady(symbol, interval string, minBars int) {
	if !s.enableBackfill {
		return
	}
	sym := strings.ToUpper(symbol)
	iv := strings.ToLower(interval)
	if _, ok := s.store.GetRecent(sym, iv, minBars); ok {
		return
	}
	go func() {
		defer recoverGuard("Service.EnsureReadyBackfill")
		s.restBackfillWindow([]string{sym}, iv, max(minBars, EnsureReadyMinBars))
	}()
}

// InitDefault initializes the package-level Default service and starts it.
// If symbols are provided, they will be added for subscription.
func InitDefault(opts Options, symbols []string) error {
	Default = NewService(opts)
	if err := Default.Start(context.Background()); err != nil {
		return err
	}
	if len(symbols) > 0 {
		_ = Default.AddSymbols(symbols)
	}
	return nil
}

// Close: graceful shutdown
func (s *Service) Close() {
	// First, close all WS clients to close all subscription channels
	// This allows consumeLoop to immediately detect channel closure and exit
	s.mu.RLock()
	clients := make(map[string]client)
	for k, v := range s.clients {
		clients[k] = v
	}
	s.mu.RUnlock()

	// Close all clients (this closes all channels)
	// This is the primary way for consumeLoop to exit
	for _, cl := range clients {
		cl.Close()
	}

	// Cancel all contexts immediately (parallel with channel closure)
	// This provides a secondary exit path for consumeLoop
	s.mu.Lock()
	cancels := make(map[string]context.CancelFunc)
	for k, v := range s.subCancels {
		cancels[k] = v
	}
	s.mu.Unlock()

	// Cancel all contexts in parallel
	for _, cancel := range cancels {
		cancel()
	}

	// Close ingestor (this will cancel all consumeLoop contexts as well)
	// This ensures all contexts are canceled, even if some were missed
	s.ingestor.Close()

	// Clean up subCancels
	s.mu.Lock()
	for k := range cancels {
		delete(s.subCancels, k)
	}
	s.mu.Unlock()

	// Give goroutines time to exit (consumeLoop, readLoop, pingLoop, etc.)
	time.Sleep(200 * time.Millisecond)
}

// ------------- Internal: Periodic Backfill -------------

func (s *Service) runPeriodicReconcile(ctx context.Context) {
	t3m := time.NewTicker(ReconcileTick3m)
	t15m := time.NewTicker(ReconcileTick15m)
	t1h := time.NewTicker(ReconcileTick1h)
	t4h := time.NewTicker(ReconcileTick4h)
	t8h := time.NewTicker(ReconcileTick8h)
	t1d := time.NewTicker(ReconcileTick1d)
	t1w := time.NewTicker(ReconcileTick1w)
	defer t3m.Stop()
	defer t15m.Stop()
	defer t1h.Stop()
	defer t4h.Stop()
	defer t8h.Stop()
	defer t1d.Stop()
	defer t1w.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-t3m.C:
			s.reconcileInterval(Interval3m, 100)
		case <-t15m.C:
			s.reconcileInterval(Interval15m, 100)
		case <-t1h.C:
			s.reconcileInterval(Interval1h, 100)
		case <-t4h.C:
			s.reconcileInterval(Interval4h, 100)
		case <-t8h.C:
			s.reconcileInterval(Interval8h, 100)
		case <-t1d.C:
			s.reconcileInterval(Interval1d, 100)
		case <-t1w.C:
			s.reconcileInterval(Interval1w, 100)
		}
	}
}

func (s *Service) reconcileInterval(interval string, window int) {
	iv := strings.ToLower(interval)
	symbols := s.subs.Symbols(iv)
	if len(symbols) == 0 {
		return
	}
	go func() {
		defer recoverGuard("Service.reconcileInterval")
		s.restBackfillWindow(symbols, iv, window)
	}()
}

func (s *Service) restBackfillWindow(symbols []string, interval string, window int) {
	if window <= 0 || len(symbols) == 0 {
		return
	}
	maxConc := s.restMaxConcurrency
	if maxConc <= 0 {
		maxConc = 1
	}

	var wg sync.WaitGroup
	sem := make(chan struct{}, maxConc)
	for _, sym := range symbols {
		wg.Add(1)
		sem <- struct{}{}
		sym := sym
		go func() {
			defer func() {
				recoverGuard("Service.restBackfillWorker")
				<-sem
				wg.Done()
			}()
			_ = s.rest.ReconcileWindow(sym, interval, window) // Errors are logged in ReconcileWindow
		}()
	}
	wg.Wait()
}

// ------------- Utilities -------------

func symSliceToUpper(in []string) []string {
	out := make([]string, 0, len(in))
	for _, s := range in {
		if su := strings.ToUpper(strings.TrimSpace(s)); su != "" {
			out = append(out, su)
		}
	}
	return out
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func recoverGuard(tag string) {
	if r := recover(); r != nil {
		log.Printf("[%s] panic recovered: %v", tag, r)
	}
}

// hasGapsOrTailMissing returns true if there are middle gaps or tail behind schedule based on stepMs.
func hasGapsOrTailMissing(internal []marketKline, stepMs int64) bool {
	if stepMs <= 0 || len(internal) == 0 {
		return false
	}
	// middle gaps
	for i := 1; i < len(internal); i++ {
		if internal[i].OpenTime-internal[i-1].OpenTime > stepMs {
			return true
		}
	}
	// tail missing
	nowMs := time.Now().UnixMilli()
	expectedLastOpen := nowMs - (nowMs % stepMs)
	return expectedLastOpen-internal[len(internal)-1].OpenTime > stepMs
}

// intervalMillis converts Binance interval string to milliseconds.
// Supports "m","h","d","w" with known defaults and simple parsing fallback.
func intervalMillis(iv string) int64 {
	switch iv {
	case Interval3m:
		return int64(3 * time.Minute / time.Millisecond)
	case Interval15m:
		return int64(15 * time.Minute / time.Millisecond)
	case Interval1h:
		return int64(time.Hour / time.Millisecond)
	case Interval4h:
		return int64(4 * time.Hour / time.Millisecond)
	case Interval8h:
		return int64(8 * time.Hour / time.Millisecond)
	case Interval1d:
		return int64(24 * time.Hour / time.Millisecond)
	case Interval1w:
		return int64(7 * 24 * time.Hour / time.Millisecond)
	}
	// Fallback generic parser: e.g. "5m","2h","2d","1w"
	if len(iv) >= 2 {
		unit := iv[len(iv)-1]
		num := iv[:len(iv)-1]
		if n, err := strconv.Atoi(num); err == nil && n > 0 {
			switch unit {
			case 'm':
				return int64((time.Duration(n) * time.Minute) / time.Millisecond)
			case 'h':
				return int64((time.Duration(n) * time.Hour) / time.Millisecond)
			case 'd':
				return int64((time.Duration(n) * 24 * time.Hour) / time.Millisecond)
			case 'w':
				return int64((time.Duration(n) * 7 * 24 * time.Hour) / time.Millisecond)
			}
		}
	}
	return 0
}

// shortSuppressDuration returns a suppression TTL for short-only backfill based on interval.
// It scales with interval size and is clamped to [10m, 24h] to avoid extremes.
func shortSuppressDuration(iv string) time.Duration {
	stepMs := intervalMillis(iv)
	if stepMs <= 0 {
		return 30 * time.Minute
	}
	// Clamp suppression to [interval, 24h]
	step := time.Duration(stepMs) * time.Millisecond
	d := step
	if d > 24*time.Hour {
		d = 24 * time.Hour
	}
	return d
}

// tryStartOnDemand returns true if caller wins the right to trigger backfill for (symbol, interval).
func (s *Service) tryStartOnDemand(symbol, interval string) bool {
	key := strings.ToUpper(strings.TrimSpace(symbol)) + "|" + strings.ToLower(strings.TrimSpace(interval))
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.onDemand == nil {
		s.onDemand = make(map[string]bool)
	}
	if s.onDemand[key] {
		return false
	}
	s.onDemand[key] = true
	return true
}

// endOnDemand marks the (symbol, interval) on-demand backfill as completed.
func (s *Service) endOnDemand(symbol, interval string) {
	key := strings.ToUpper(strings.TrimSpace(symbol)) + "|" + strings.ToLower(strings.TrimSpace(interval))
	s.mu.Lock()
	delete(s.onDemand, key)
	s.mu.Unlock()
}

// isShortSuppressed returns true if short-only backfill should be suppressed now.
func (s *Service) isShortSuppressed(symbol, interval string) bool {
	key := strings.ToUpper(strings.TrimSpace(symbol)) + "|" + strings.ToLower(strings.TrimSpace(interval))
	s.mu.RLock()
	defer s.mu.RUnlock()
	if s.shortSuppress == nil {
		return false
	}
	exp, ok := s.shortSuppress[key]
	return ok && time.Now().Before(exp)
}

// setShortSuppress suppresses short-only backfills for the given duration.
func (s *Service) setShortSuppress(symbol, interval string, d time.Duration) {
	if d <= 0 {
		return
	}
	key := strings.ToUpper(strings.TrimSpace(symbol)) + "|" + strings.ToLower(strings.TrimSpace(interval))
	s.mu.Lock()
	if s.shortSuppress == nil {
		s.shortSuppress = make(map[string]time.Time)
	}
	s.shortSuppress[key] = time.Now().Add(d)
	s.mu.Unlock()
}
