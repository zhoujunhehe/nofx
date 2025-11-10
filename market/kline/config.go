package kline

import (
	"fmt"
	"strings"
)

const (
	// Default Binance WebSocket endpoints
	WSEndpointSpot    = "wss://stream.binance.com:9443/ws"
	WSEndpointFutures = "wss://fstream.binance.com/ws"

	// WebSocket request methods
	WSMethodSubscribe   = "SUBSCRIBE"
	WSMethodUnsubscribe = "UNSUBSCRIBE"

	// Stream format
	KlineStreamFormat = "%s@kline_%s"

	// Interval constants (used to avoid magic strings)
	Interval3m  = "3m"  // 3-minute kline interval
	Interval15m = "15m" // 15-minute kline interval
	Interval1h  = "1h"  // 1-hour kline interval
	Interval4h  = "4h"  // 4-hour kline interval
	Interval8h  = "8h"  // 8-hour kline interval
	Interval1d  = "1d"  // 1-day kline interval
	Interval1w  = "1w"  // 1-week kline interval
	// Adaptive backoff defaults
	DefaultWeight1mLimit        = 1200      // assumed 1-minute weight cap
	WeightBackoffThresholdRatio = 0.90      // back off when usage exceeds 90%
	BackoffBaseDelay            = 200 * 1e6 // 200ms in nanoseconds (used via time.Duration)
	BackoffMaxDelay             = 3 * 1e9   // 3s in nanoseconds

	// REST paths
	FapiExchangeInfoPath = "/fapi/v1/exchangeInfo"

	// Filters/constants for symbols
	QuoteAssetUSDT        = "USDT"
	ContractTypePerpetual = "PERPETUAL"
	SymbolStatusTrading   = "TRADING"
)

// DefaultIntervals is the default set of intervals maintained by the service.
var DefaultIntervals = []string{Interval3m, Interval15m, Interval1h, Interval4h, Interval8h, Interval1d, Interval1w}

// BuildKlineStream builds a lowercase kline stream name like btcusdt@kline_1m.
func BuildKlineStream(symbol, interval string) string {
	return fmt.Sprintf(KlineStreamFormat, strings.ToLower(symbol), strings.ToLower(interval))
}

// FuturesBaseURLs moved to http_client.go (single source of truth there)
