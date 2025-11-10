package kline

import (
	"context"
	"log"
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
	client             client
	store              Store
	rest               RestReconciler
	subs               *SubManager
	ingestor           *WSIngestor
	rate               *Governor
	symbolSync         *SymbolSync
	intervals          []string
	backfillWindow     int
	restMaxConcurrency int

	// stream -> cancel (returned by WSIngestor.Attach)
	subCancels map[string]context.CancelFunc

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
)

type Options struct {
	BatchSize             int
	Intervals             []string // e.g. []string{"3m","4h"}
	BackfillWindow        int
	WSEndpoint            string
	RestMaxConcurrency    int
	SymbolRefreshInterval time.Duration
}

func DefaultOptions() Options {
	return Options{
		BatchSize:             150,
		Intervals:             DefaultIntervals,
		BackfillWindow:        100,
		RestMaxConcurrency:    20,
		SymbolRefreshInterval: time.Duration(DefaultSymbolRefreshInterval),
	}
}

func NewService(opts Options) *Service {
	if len(opts.Intervals) == 0 {
		opts = DefaultOptions()
	}

	cl := NewRawWSClient(opts.WSEndpoint)
	rate := NewGovernor()
	memStore := NewMemoryStore(DefaultStoreCapacity)
	reconciler := NewRestReconciler(rate, memStore)
	subs := NewSubManager(cl)
	ingestor := NewWSIngestor(memStore)

	svc := &Service{
		client:             cl,
		store:              memStore,
		rest:               reconciler,
		subs:               subs,
		ingestor:           ingestor,
		rate:               rate,
		intervals:          opts.Intervals,
		backfillWindow:     opts.BackfillWindow,
		restMaxConcurrency: opts.RestMaxConcurrency,
		subCancels:         make(map[string]context.CancelFunc),
		symbolSync:         NewSymbolSync(opts.SymbolRefreshInterval, NewFilteredBinanceFuturesFetcher(NewKlineHTTPClient())),
	}
	return svc
}

// Start: background WS connection & periodic reconcile (does not block caller)
func (s *Service) Start(ctx context.Context) error {
	// Background connection (Connect has infinite reconnect loop internally)
	go func() {
		defer recoverGuard("Service.Connect")
		if err := s.client.Connect(); err != nil {
			log.Printf("[Service] WS connect exited: %v", err)
		}
	}()

	// Periodic reconcile
	go s.runPeriodicReconcile(ctx)
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

		// Register subscriber and attach for each symbol
		for _, sym := range upSyms {
			stream := buildKlineStream(sym, ivLower)
			// Avoid duplicate attach
			if _, exists := s.subCancels[stream]; exists {
				continue
			}
			ch := s.client.AddSubscriber(stream, DefaultSubscriberBuffer)
			cancel, ok := s.ingestor.Attach(stream, sym, ivLower, ch)
			if ok {
				s.subCancels[stream] = cancel
			}
		}

		// Initial small window backfill (entire batch at once)
		go func(iv string) {
			defer recoverGuard("Service.initialBackfill")
			s.restBackfillWindow(upSyms, iv, s.backfillWindow)
		}(ivLower)
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

		for _, sym := range upSyms {
			stream := buildKlineStream(sym, ivLower)

			// Stop ingestion goroutine
			if cancel, ok := s.subCancels[stream]; ok {
				cancel()
				delete(s.subCancels, stream)
			}

			// Let WS client remove channel (will close ch)
			s.client.RemoveSubscriber(stream)

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

// GetRecentKlines returns the most recent klines as exported Kline type.
func (s *Service) GetRecentKlines(symbol, interval string, limit int) ([]Kline, bool) {
	internal, ready := s.GetRecent(symbol, interval, limit)
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

// EnsureReady: triggers a small window emergency backfill when minimum N bars not satisfied
func (s *Service) EnsureReady(symbol, interval string, minBars int) {
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
	// Stop all ingestion
	s.mu.Lock()
	for stream, cancel := range s.subCancels {
		cancel()
		delete(s.subCancels, stream)
	}
	s.mu.Unlock()

	// Close WS
	s.client.Close()

	// Close ingestor
	s.ingestor.Close()
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
			s.rest.ReconcileWindow(sym, interval, window)
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
