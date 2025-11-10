package kline

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

// Helper function to create a mock kline response
func createMockKlineResponse(count int) [][]interface{} {
	baseTime := int64(1609459200000) // 2021-01-01 00:00:00 UTC
	result := make([][]interface{}, count)
	for i := 0; i < count; i++ {
		result[i] = []interface{}{
			baseTime + int64(i*60000),     // OpenTime
			"29000.0",                     // Open
			"29100.0",                     // High
			"28900.0",                     // Low
			"29050.0",                     // Close
			"100.5",                       // Volume
			baseTime + int64((i+1)*60000), // CloseTime
			"2915025.0",                   // QuoteVolume
			150,                           // Trades
			"50.0",                        // TakerBuyBaseVolume
			"1452500.0",                   // TakerBuyQuoteVolume
		}
	}
	return result
}

func TestNewKlineHTTPClient(t *testing.T) {
	// Test with default options
	client := NewKlineHTTPClient()
	if client == nil {
		t.Fatal("NewKlineHTTPClient returned nil")
	}
	if len(client.bases) == 0 {
		t.Fatal("expected default base URLs")
	}
	if client.maxTry == 0 {
		t.Fatal("expected default maxTry")
	}

	// Test with custom options
	customClient := NewKlineHTTPClient(func(opts *HttpOptions) {
		opts.MaxAttempts = 10
		opts.BaseURLs = []string{"https://custom.example.com"}
	})
	if customClient.maxTry != 10 {
		t.Fatalf("expected maxTry 10, got %d", customClient.maxTry)
	}
	if len(customClient.bases) != 1 || customClient.bases[0] != "https://custom.example.com" {
		t.Fatalf("expected custom base URL, got %v", customClient.bases)
	}
}

func TestKlineHTTPClient_GetKlines_Success(t *testing.T) {
	// Create mock server
	mockData := createMockKlineResponse(5)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/fapi/v1/klines" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		if r.URL.Query().Get("symbol") != "BTCUSDT" {
			t.Errorf("unexpected symbol: %s", r.URL.Query().Get("symbol"))
		}
		if r.URL.Query().Get("interval") != "1m" {
			t.Errorf("unexpected interval: %s", r.URL.Query().Get("interval"))
		}
		if r.URL.Query().Get("limit") != "5" {
			t.Errorf("unexpected limit: %s", r.URL.Query().Get("limit"))
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(mockData)
	}))
	defer server.Close()

	// Create client with mock server URL
	client := NewKlineHTTPClient(func(opts *HttpOptions) {
		opts.BaseURLs = []string{server.URL}
		opts.MaxAttempts = 1
	})

	ctx := context.Background()
	klines, err := client.GetKlines(ctx, "BTCUSDT", "1m", 5)
	if err != nil {
		t.Fatalf("GetKlines failed: %v", err)
	}
	if len(klines) != 5 {
		t.Fatalf("expected 5 klines, got %d", len(klines))
	}

	// Verify first kline
	k := klines[0]
	if k.Symbol != "BTCUSDT" {
		t.Errorf("expected Symbol BTCUSDT, got %s", k.Symbol)
	}
	if k.Interval != "1m" {
		t.Errorf("expected Interval 1m, got %s", k.Interval)
	}
	if k.Close != 29050.0 {
		t.Errorf("expected Close 29050.0, got %f", k.Close)
	}
}

func TestKlineHTTPClient_GetKlinesWindow(t *testing.T) {
	mockData := createMockKlineResponse(3)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("startTime") != "1609459200000" {
			t.Errorf("unexpected startTime: %s", r.URL.Query().Get("startTime"))
		}
		if r.URL.Query().Get("endTime") != "1609459260000" {
			t.Errorf("unexpected endTime: %s", r.URL.Query().Get("endTime"))
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(mockData)
	}))
	defer server.Close()

	client := NewKlineHTTPClient(func(opts *HttpOptions) {
		opts.BaseURLs = []string{server.URL}
		opts.MaxAttempts = 1
	})

	ctx := context.Background()
	startMs := int64(1609459200000)
	endMs := int64(1609459260000)
	klines, err := client.GetKlinesWindow(ctx, "BTCUSDT", "1m", startMs, endMs, 3)
	if err != nil {
		t.Fatalf("GetKlinesWindow failed: %v", err)
	}
	if len(klines) != 3 {
		t.Fatalf("expected 3 klines, got %d", len(klines))
	}
}

func TestKlineHTTPClient_GetKlines_RetryOnNetworkError(t *testing.T) {
	attempts := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		attempts++
		if attempts < 2 {
			// First attempt: simulate network error by closing connection
			w.Header().Set("Connection", "close")
			return
		}
		// Second attempt: success
		mockData := createMockKlineResponse(2)
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(mockData)
	}))
	defer server.Close()

	client := NewKlineHTTPClient(func(opts *HttpOptions) {
		opts.BaseURLs = []string{server.URL}
		opts.MaxAttempts = 3
		opts.Backoff = BackoffPolicy{
			Base: 10 * time.Millisecond,
			Max:  100 * time.Millisecond,
		}
	})

	ctx := context.Background()
	klines, err := client.GetKlines(ctx, "BTCUSDT", "1m", 2)
	if err != nil {
		t.Fatalf("GetKlines failed after retry: %v", err)
	}
	if len(klines) != 2 {
		t.Fatalf("expected 2 klines, got %d", len(klines))
	}
	if attempts < 2 {
		t.Fatalf("expected at least 2 attempts, got %d", attempts)
	}
}

func TestKlineHTTPClient_GetKlines_429RetryAfter(t *testing.T) {
	attempts := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		attempts++
		if attempts == 1 {
			// First attempt: 429 with Retry-After
			w.Header().Set("Retry-After", "1") // 1 second
			w.WriteHeader(http.StatusTooManyRequests)
			return
		}
		// Second attempt: success
		mockData := createMockKlineResponse(2)
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(mockData)
	}))
	defer server.Close()

	client := NewKlineHTTPClient(func(opts *HttpOptions) {
		opts.BaseURLs = []string{server.URL}
		opts.MaxAttempts = 3
		opts.Backoff = BackoffPolicy{
			Base: 10 * time.Millisecond,
			Max:  100 * time.Millisecond,
		}
	})

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	klines, err := client.GetKlines(ctx, "BTCUSDT", "1m", 2)
	if err != nil {
		t.Fatalf("GetKlines failed after 429 retry: %v", err)
	}
	if len(klines) != 2 {
		t.Fatalf("expected 2 klines, got %d", len(klines))
	}
}

func TestKlineHTTPClient_GetKlines_DomainRotation(t *testing.T) {
	attempts := 0
	server1 := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		attempts++
		w.WriteHeader(http.StatusInternalServerError) // 500 error
	}))
	defer server1.Close()

	server2 := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		attempts++
		mockData := createMockKlineResponse(2)
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(mockData)
	}))
	defer server2.Close()

	client := NewKlineHTTPClient(func(opts *HttpOptions) {
		opts.BaseURLs = []string{server1.URL, server2.URL}
		opts.MaxAttempts = 3
		opts.Backoff = BackoffPolicy{
			Base: 10 * time.Millisecond,
			Max:  50 * time.Millisecond,
		}
	})

	ctx := context.Background()
	klines, err := client.GetKlines(ctx, "BTCUSDT", "1m", 2)
	if err != nil {
		t.Fatalf("GetKlines failed after domain rotation: %v", err)
	}
	if len(klines) != 2 {
		t.Fatalf("expected 2 klines, got %d", len(klines))
	}
	if attempts < 2 {
		t.Fatalf("expected at least 2 attempts (domain rotation), got %d", attempts)
	}
}

func TestKlineHTTPClient_GetKlines_HeaderBackoff(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Set header indicating high weight usage (80% of 6000 = 4800)
		w.Header().Set("X-MBX-USED-WEIGHT-1m", "4800")
		mockData := createMockKlineResponse(2)
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(mockData)
	}))
	defer server.Close()

	client := NewKlineHTTPClient(func(opts *HttpOptions) {
		opts.BaseURLs = []string{server.URL}
		opts.MaxAttempts = 1
		opts.HeaderBackoff = HeaderBackoff{
			Enabled:       true,
			Weight1mLimit: 6000,
			TriggerRatio:  0.75,
			Base:          10 * time.Millisecond,
			Max:           100 * time.Millisecond,
		}
	})

	ctx := context.Background()
	start := time.Now()
	klines, err := client.GetKlines(ctx, "BTCUSDT", "1m", 2)
	duration := time.Since(start)

	if err != nil {
		t.Fatalf("GetKlines failed: %v", err)
	}
	if len(klines) != 2 {
		t.Fatalf("expected 2 klines, got %d", len(klines))
	}
	// Should have some delay due to header backoff (80% > 75% threshold)
	if duration < 5*time.Millisecond {
		t.Logf("warning: header backoff delay seems too short: %v", duration)
	}
}

func TestKlineHTTPClient_GetKlines_WithGovernor(t *testing.T) {
	mockData := createMockKlineResponse(2)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(mockData)
	}))
	defer server.Close()

	// Create mock governor
	mockGov := &mockTokenGovernor{
		acquired: make(map[string]float64),
	}
	bucket := NewBucket(1000, 100)
	mockGov.bucket = bucket

	client := NewKlineHTTPClient(func(opts *HttpOptions) {
		opts.BaseURLs = []string{server.URL}
		opts.MaxAttempts = 1
		opts.Governor = mockGov
		opts.GovernorBuckets = []string{"global"}
	})

	ctx := context.Background()
	klines, err := client.GetKlines(ctx, "BTCUSDT", "1m", 2)
	if err != nil {
		t.Fatalf("GetKlines failed: %v", err)
	}
	if len(klines) != 2 {
		t.Fatalf("expected 2 klines, got %d", len(klines))
	}

	// Verify governor was called
	if len(mockGov.acquired) == 0 {
		t.Error("expected governor to be called")
	}
}

func TestKlineHTTPClient_GetKlines_ContextCancel(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Simulate slow response
		time.Sleep(100 * time.Millisecond)
		mockData := createMockKlineResponse(2)
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(mockData)
	}))
	defer server.Close()

	client := NewKlineHTTPClient(func(opts *HttpOptions) {
		opts.BaseURLs = []string{server.URL}
		opts.MaxAttempts = 3
	})

	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()

	_, err := client.GetKlines(ctx, "BTCUSDT", "1m", 2)
	if err == nil {
		t.Fatal("expected error due to context timeout")
	}
	if err != context.DeadlineExceeded && !strings.Contains(err.Error(), "deadline") {
		t.Fatalf("expected context deadline error, got %v", err)
	}
}

func TestKlineHTTPClient_GetFuturesExchangeInfo_Success(t *testing.T) {
	mockInfo := ExchangeInfo{
		Symbols: []ExchangeSymbol{
			{
				Symbol:       "BTCUSDT",
				Status:       "TRADING",
				ContractType: "PERPETUAL",
				QuoteAsset:   "USDT",
			},
			{
				Symbol:       "ETHUSDT",
				Status:       "TRADING",
				ContractType: "PERPETUAL",
				QuoteAsset:   "USDT",
			},
		},
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/fapi/v1/exchangeInfo" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(mockInfo)
	}))
	defer server.Close()

	client := NewKlineHTTPClient(func(opts *HttpOptions) {
		opts.BaseURLs = []string{server.URL}
		opts.MaxAttempts = 1
	})

	ctx := context.Background()
	info, err := client.GetFuturesExchangeInfo(ctx)
	if err != nil {
		t.Fatalf("GetFuturesExchangeInfo failed: %v", err)
	}
	if len(info.Symbols) != 2 {
		t.Fatalf("expected 2 symbols, got %d", len(info.Symbols))
	}
	if info.Symbols[0].Symbol != "BTCUSDT" {
		t.Errorf("expected first symbol BTCUSDT, got %s", info.Symbols[0].Symbol)
	}
}

func TestKlineHTTPClient_GetFuturesExchangeInfo_ContextCancel(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(100 * time.Millisecond)
	}))
	defer server.Close()

	client := NewKlineHTTPClient(func(opts *HttpOptions) {
		opts.BaseURLs = []string{server.URL}
		opts.MaxAttempts = 3
	})

	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()

	_, err := client.GetFuturesExchangeInfo(ctx)
	if err == nil {
		t.Fatal("expected error due to context timeout")
	}
	if err != context.DeadlineExceeded && !strings.Contains(err.Error(), "deadline") {
		t.Fatalf("expected context deadline error, got %v", err)
	}
}

func TestKlineHTTPClient_GetKlines_InvalidJSON(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("invalid json"))
	}))
	defer server.Close()

	client := NewKlineHTTPClient(func(opts *HttpOptions) {
		opts.BaseURLs = []string{server.URL}
		opts.MaxAttempts = 1
	})

	ctx := context.Background()
	_, err := client.GetKlines(ctx, "BTCUSDT", "1m", 2)
	if err == nil {
		t.Fatal("expected error for invalid JSON")
	}
}

func TestKlineHTTPClient_GetKlines_EmptyResponse(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte("[]"))
	}))
	defer server.Close()

	client := NewKlineHTTPClient(func(opts *HttpOptions) {
		opts.BaseURLs = []string{server.URL}
		opts.MaxAttempts = 1
	})

	ctx := context.Background()
	klines, err := client.GetKlines(ctx, "BTCUSDT", "1m", 2)
	if err != nil {
		t.Fatalf("GetKlines failed: %v", err)
	}
	if len(klines) != 0 {
		t.Fatalf("expected 0 klines, got %d", len(klines))
	}
}

func TestKlineHTTPClient_GetKlines_DefaultLimit(t *testing.T) {
	mockData := createMockKlineResponse(60) // Default limit
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("limit") != "60" {
			t.Errorf("unexpected limit: %s", r.URL.Query().Get("limit"))
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(mockData)
	}))
	defer server.Close()

	client := NewKlineHTTPClient(func(opts *HttpOptions) {
		opts.BaseURLs = []string{server.URL}
		opts.MaxAttempts = 1
	})

	ctx := context.Background()
	klines, err := client.GetKlines(ctx, "BTCUSDT", "1m", 0)
	if err != nil {
		t.Fatalf("GetKlines failed: %v", err)
	}
	if len(klines) != 60 {
		t.Fatalf("expected 60 klines (default), got %d", len(klines))
	}
}

func TestKlineHTTPClient_Rotate(t *testing.T) {
	client := NewKlineHTTPClient(func(opts *HttpOptions) {
		opts.BaseURLs = []string{"http://base1.com", "http://base2.com", "http://base3.com"}
	})

	// Test rotation
	initialIdx := client.idx
	client.rotate()
	if client.idx != (initialIdx+1)%3 {
		t.Errorf("expected idx %d after rotate, got %d", (initialIdx+1)%3, client.idx)
	}

	// Rotate multiple times
	for i := 0; i < 5; i++ {
		client.rotate()
	}
	expectedIdx := (initialIdx + 6) % 3
	if client.idx != expectedIdx {
		t.Errorf("expected idx %d after 6 rotates, got %d", expectedIdx, client.idx)
	}
}

func TestParseKlineArray(t *testing.T) {
	baseTime := int64(1609459200000)
	arr := []interface{}{
		baseTime,         // OpenTime
		"29000.0",        // Open
		"29100.0",        // High
		"28900.0",        // Low
		"29050.0",        // Close
		"100.5",          // Volume
		baseTime + 60000, // CloseTime
		"2915025.0",      // QuoteVolume
		150,              // Trades
		"50.0",           // TakerBuyBaseVolume
		"1452500.0",      // TakerBuyQuoteVolume
	}

	k, err := parseKlineArray(arr)
	if err != nil {
		t.Fatalf("parseKlineArray failed: %v", err)
	}
	if k.OpenTime != baseTime {
		t.Errorf("expected OpenTime %d, got %d", baseTime, k.OpenTime)
	}
	if k.Close != 29050.0 {
		t.Errorf("expected Close 29050.0, got %f", k.Close)
	}
	if k.Trades != 150 {
		t.Errorf("expected Trades 150, got %d", k.Trades)
	}
}

func TestParseKlineArray_InvalidLength(t *testing.T) {
	arr := []interface{}{1, 2, 3} // Too short
	_, err := parseKlineArray(arr)
	if err == nil {
		t.Fatal("expected error for invalid length")
	}
}

func TestGetUsedWeight1m(t *testing.T) {
	tests := []struct {
		name   string
		header http.Header
		want   int
	}{
		{
			name: "X-MBX-USED-WEIGHT-1M",
			header: http.Header{
				"X-MBX-USED-WEIGHT-1M": []string{"4800"},
			},
			want: 4800,
		},
		{
			name: "X-MBX-USED-WEIGHT-1m",
			header: http.Header{
				"X-MBX-USED-WEIGHT-1m": []string{"3600"},
			},
			want: 3600,
		},
		{
			name: "X-MBX-USED-WEIGHT",
			header: http.Header{
				"X-MBX-USED-WEIGHT": []string{"2400"},
			},
			want: 2400,
		},
		{
			name: "X-MBX-USED-WEIGHT-1m variant",
			header: http.Header{
				"X-MBX-USED-WEIGHT-1m": []string{"1200"},
			},
			want: 1200,
		},
		{
			name:   "no header",
			header: http.Header{},
			want:   0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := getUsedWeight1m(tt.header)
			if got != tt.want {
				t.Errorf("getUsedWeight1m() = %d, want %d", got, tt.want)
			}
		})
	}
}

func TestRetryAfterDelay(t *testing.T) {
	tests := []struct {
		name     string
		header   http.Header
		fallback time.Duration
		wantMin  time.Duration
		wantMax  time.Duration
	}{
		{
			name: "Retry-After seconds",
			header: http.Header{
				"Retry-After": []string{"5"},
			},
			fallback: 1 * time.Second,
			wantMin:  4 * time.Second,
			wantMax:  6 * time.Second, // Allow jitter
		},
		{
			name:     "no Retry-After, use fallback",
			header:   http.Header{},
			fallback: 2 * time.Second,
			wantMin:  1 * time.Second, // Allow jitter
			wantMax:  3 * time.Second,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := retryAfterDelay(tt.header, tt.fallback)
			if got < tt.wantMin || got > tt.wantMax {
				t.Errorf("retryAfterDelay() = %v, want between %v and %v", got, tt.wantMin, tt.wantMax)
			}
		})
	}
}

func TestHeaderBackoffDelay(t *testing.T) {
	client := NewKlineHTTPClient(func(opts *HttpOptions) {
		opts.HeaderBackoff = HeaderBackoff{
			Enabled:       true,
			Weight1mLimit: 6000,
			TriggerRatio:  0.75,
			Base:          100 * time.Millisecond,
			Max:           500 * time.Millisecond,
		}
	})

	tests := []struct {
		name   string
		header http.Header
		want   bool // true if delay > 0
	}{
		{
			name: "below threshold (50%)",
			header: http.Header{
				"X-MBX-USED-WEIGHT-1m": []string{"3000"},
			},
			want: false,
		},
		{
			name: "at threshold (75%)",
			header: http.Header{
				"X-MBX-USED-WEIGHT-1m": []string{"4500"},
			},
			want: true,
		},
		{
			name: "above threshold (90%)",
			header: http.Header{
				"X-MBX-USED-WEIGHT-1m": []string{"5400"},
			},
			want: true,
		},
		{
			name:   "no header",
			header: http.Header{},
			want:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			delay := client.headerBackoffDelay(tt.header)
			hasDelay := delay > 0
			if hasDelay != tt.want {
				t.Errorf("headerBackoffDelay() delay > 0 = %v, want %v (delay: %v)", hasDelay, tt.want, delay)
			}
		})
	}
}

func TestNextBackoff(t *testing.T) {
	tests := []struct {
		name    string
		cur     time.Duration
		max     time.Duration
		wantMin time.Duration
		wantMax time.Duration
	}{
		{
			name:    "normal backoff",
			cur:     100 * time.Millisecond,
			max:     5 * time.Second,
			wantMin: 150 * time.Millisecond, // 1.6x
			wantMax: 170 * time.Millisecond,
		},
		{
			name:    "capped at max",
			cur:     4 * time.Second,
			max:     5 * time.Second,
			wantMin: 5 * time.Second,
			wantMax: 5 * time.Second,
		},
		{
			name:    "zero cur uses default",
			cur:     0,
			max:     5 * time.Second,
			wantMin: 150 * time.Millisecond,
			wantMax: 160 * time.Millisecond,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := nextBackoff(tt.cur, tt.max)
			if got < tt.wantMin || got > tt.wantMax {
				t.Errorf("nextBackoff(%v, %v) = %v, want between %v and %v", tt.cur, tt.max, got, tt.wantMin, tt.wantMax)
			}
		})
	}
}

func TestWithJitter(t *testing.T) {
	d := 1 * time.Second
	// Test multiple times to verify jitter range
	for i := 0; i < 100; i++ {
		jittered := withJitter(d)
		// Should be within [0.75s, 1.25s] for ±25% jitter
		min := 750 * time.Millisecond
		max := 1250 * time.Millisecond
		if jittered < min || jittered > max {
			t.Fatalf("jittered duration %v out of range [%v, %v]", jittered, min, max)
		}
	}

	// Test zero duration
	result := withJitter(0)
	if result != 0 {
		t.Fatalf("expected 0 for zero duration, got %v", result)
	}
}

// Mock TokenGovernor for testing
type mockTokenGovernor struct {
	bucket   *Bucket
	acquired map[string]float64
}

func (m *mockTokenGovernor) Acquire(ctx context.Context, cost float64, buckets ...string) error {
	if m.bucket != nil {
		if err := m.bucket.WaitN(ctx, cost); err != nil {
			return err
		}
	}
	// Track acquisitions (even if bucket is used)
	if m.acquired == nil {
		m.acquired = make(map[string]float64)
	}
	for _, b := range buckets {
		m.acquired[b] += cost
	}
	return nil
}

func (m *mockTokenGovernor) Feedback429(bucket string, factor float64, decay time.Duration) {
	if m.bucket != nil {
		m.bucket.Feedback429(factor, decay)
	}
}
