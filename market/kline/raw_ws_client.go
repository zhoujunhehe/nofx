package kline

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/gorilla/websocket"
)

const (
	// Connection & heartbeat
	WSHandshakeTimeout = 10 * time.Second
	WSReadTimeout      = 60 * time.Second // Read timeout, used with pong handler
	WSPingInterval     = 30 * time.Second // Active ping (safety)
	WSReconnectBase    = 150 * time.Millisecond
	WSReconnectMax     = 5 * time.Second
	WSReadIdleSleep    = 1 * time.Second

	// Control message rate limit (SUB/UNSUB)
	// Spot 5 msg/s; USDM 10 msg/s. Here uses a conservative value, adjust as needed.
	WSCtrlMsgsPerSec = 5

	// Maximum streams per SUBSCRIBE/UNSUBSCRIBE request
	// Binance may reject requests with too many streams (policy violation).
	// A smaller batch size (50) is safer to avoid "Invalid request" errors.
	WSSubscribeBatchSize = 50
)

type RawWSClient struct {
	endpoint string

	// Connection and concurrency control
	conn    *websocket.Conn
	mu      sync.RWMutex  // Protects conn, subs, wantSubs
	writeMu sync.Mutex    // Serializes write operations (WriteJSON/WriteMessage)
	closed  atomic.Bool   // Whether permanently closed (Close call)
	done    chan struct{} // Read loop/reconnect loop exit signal

	// Subscriptions
	subs     map[string]chan []byte // stream(lowercase) -> chan
	wantSubs map[string]struct{}    // Expected subscription set (for reconnection recovery)

	// Control message rate limiting
	ctrlTicker *time.Ticker

	// Request ID
	reqID int64
}

// NewRawWSClient creates client; if endpoint is empty, uses default USDM /ws
func NewRawWSClient(endpoint string) *RawWSClient {
	if endpoint == "" {
		endpoint = WSEndpointFutures
	}
	return &RawWSClient{
		endpoint:   endpoint,
		subs:       make(map[string]chan []byte),
		wantSubs:   make(map[string]struct{}),
		done:       make(chan struct{}),
		ctrlTicker: time.NewTicker(time.Second / time.Duration(WSCtrlMsgsPerSec)),
	}
}

func (c *RawWSClient) nextID() int64 { return atomic.AddInt64(&c.reqID, 1) }

// Connect blocks to establish connection and start read loop; enters reconnect loop on disconnect until Close()
func (c *RawWSClient) Connect() error {
	if c.closed.Load() {
		return errors.New("client already closed")
	}

	backoff := WSReconnectBase
	for {
		if c.closed.Load() {
			return nil
		}

		d := websocket.Dialer{HandshakeTimeout: WSHandshakeTimeout}
		conn, _, err := d.Dial(c.endpoint, nil)
		if err != nil {
			// Connection failed: exponential backoff + jitter
			sleep := SafeJitter(backoff, 0.25)
			log.Printf("[WS] dial failed: %v; retry in %v", err, sleep)
			select {
			case <-time.After(sleep):
			case <-c.done:
				return nil
			}
			backoff = nextBackoff(backoff, WSReconnectMax)
			continue
		}

		// Successfully established connection, reset backoff
		backoff = WSReconnectBase

		// Configure ping/pong and read timeout
		_ = conn.SetReadDeadline(time.Now().Add(WSReadTimeout))
		conn.SetPongHandler(func(appData string) error {
			// Received pong, extend read deadline
			return conn.SetReadDeadline(time.Now().Add(WSReadTimeout))
		})
		conn.SetPingHandler(func(appData string) error {
			// Received ping, immediately reply with pong
			c.writeMu.Lock()
			defer c.writeMu.Unlock()
			return conn.WriteControl(websocket.PongMessage, []byte(appData), time.Now().Add(5*time.Second))
		})

		// Register connection
		c.mu.Lock()
		c.conn = conn
		c.mu.Unlock()
		log.Printf("[WS] connected: %s", c.endpoint)

		// Start read loop and ping loop
		readDone := make(chan struct{})
		go c.readLoop(readDone)

		pingCtx, pingCancel := context.WithCancel(context.Background())
		go c.pingLoop(pingCtx)

		// Resubscribe (batched + rate limited)
		c.resubscribeAll()

		// Wait for read loop to exit (error/close)
		<-readDone
		pingCancel()

		// Cleanup connection
		c.mu.Lock()
		if c.conn != nil {
			_ = c.conn.Close()
			c.conn = nil
		}
		c.mu.Unlock()

		if c.closed.Load() {
			return nil
		}

		// Enter reconnect wait
		sleep := SafeJitter(backoff, 0.25)
		log.Printf("[WS] will reconnect in %v ...", sleep)
		select {
		case <-time.After(sleep):
		case <-c.done:
			return nil
		}
		backoff = nextBackoff(backoff, WSReconnectMax)
	}
}

func (c *RawWSClient) readLoop(done chan struct{}) {
	defer close(done)

	for {
		// Check if closed before attempting to read
		if c.closed.Load() {
			return
		}

		c.mu.RLock()
		conn := c.conn
		c.mu.RUnlock()
		if conn == nil {
			select {
			case <-time.After(WSReadIdleSleep):
				continue
			case <-c.done:
				return
			}
		}

		// Check closed again before blocking read
		if c.closed.Load() {
			return
		}

		_, message, err := conn.ReadMessage()
		if err != nil {
			// If closed, don't log error (expected)
			if !c.closed.Load() {
				log.Printf("[WS] read error: %v", err)
			}
			return
		}
		// Successfully read message, refresh read deadline
		_ = conn.SetReadDeadline(time.Now().Add(WSReadTimeout))

		c.routeMessage(message)
	}
}

// pingLoop actively pings, as a safety measure (some network paths may not receive ping from server)
func (c *RawWSClient) pingLoop(ctx context.Context) {
	t := time.NewTicker(WSPingInterval)
	defer t.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-t.C:
			c.mu.RLock()
			conn := c.conn
			c.mu.RUnlock()
			if conn == nil {
				continue
			}
			c.writeMu.Lock()
			_ = conn.WriteControl(websocket.PingMessage, []byte("hb"), time.Now().Add(5*time.Second))
			c.writeMu.Unlock()
		}
	}
}

// SubscribeStreams sends SUBSCRIBE (converts to lowercase); records to wantSubs for reconnection recovery.
// Note: only sends control message, does not create subscription channel. You can call AddSubscriber before/after.
func (c *RawWSClient) SubscribeStreams(streams []string) error {
	params := compactLower(streams)
	if len(params) == 0 {
		return nil
	}
	c.mu.Lock()
	for _, s := range params {
		c.wantSubs[s] = struct{}{}
	}
	conn := c.conn
	c.mu.Unlock()
	if conn == nil {
		return fmt.Errorf("websocket not connected")
	}
	// Control message rate limiting
	<-c.ctrlTicker.C

	msg := map[string]interface{}{
		"method": WSMethodSubscribe,
		"params": params,
		"id":     c.nextID(),
	}
	log.Printf("[WS] SUB %v", params)

	c.writeMu.Lock()
	err := conn.WriteJSON(msg)
	c.writeMu.Unlock()
	return err
}

// UnsubscribeStreams sends UNSUBSCRIBE and removes from wantSubs
func (c *RawWSClient) UnsubscribeStreams(streams []string) error {
	params := compactLower(streams)
	if len(params) == 0 {
		return nil
	}
	c.mu.Lock()
	for _, s := range params {
		delete(c.wantSubs, s)
	}
	conn := c.conn
	c.mu.Unlock()
	if conn == nil {
		return fmt.Errorf("websocket not connected")
	}
	<-c.ctrlTicker.C

	msg := map[string]interface{}{
		"method": WSMethodUnsubscribe,
		"params": params,
		"id":     c.nextID(),
	}
	log.Printf("[WS] UNSUB %v", params)

	c.writeMu.Lock()
	err := conn.WriteJSON(msg)
	c.writeMu.Unlock()
	return err
}

// AddSubscriber registers a consumption channel (with buffer) for a stream (e.g., btcusdt@kline_1m)
func (c *RawWSClient) AddSubscriber(stream string, bufferSize int) <-chan []byte {
	key := strings.ToLower(strings.TrimSpace(stream))
	ch := make(chan []byte, bufferSize)
	c.mu.Lock()
	c.subs[key] = ch
	c.mu.Unlock()
	return ch
}

func (c *RawWSClient) RemoveSubscriber(stream string) {
	key := strings.ToLower(strings.TrimSpace(stream))
	c.mu.Lock()
	if ch, ok := c.subs[key]; ok {
		close(ch)
		delete(c.subs, key)
	}
	c.mu.Unlock()
}

// Routing: prioritize ack recognition, then kline events (/ws raw payload)
// ack example: {"result":null,"id":1}
// kline example: {"e":"kline","E":..., "s":"BTCUSDT", "k":{ "i":"1m", ... }}
// If wrapped format encountered: {"stream":"...","data":{...}} also compatible (though /ws normally won't)
func (c *RawWSClient) routeMessage(message []byte) {
	// 1) ack fast path: only treat as ack if "id" field exists and contains "result"
	var ack struct {
		ID     *int64      `json:"id"`
		Result interface{} `json:"result"`
	}
	if json.Unmarshal(message, &ack) == nil && ack.ID != nil {
		// If has id, treat as control plane ACK, ignore
		return
	}

	// 2) Possible wrapper layer
	var wrap struct {
		Stream string          `json:"stream"`
		Data   json.RawMessage `json:"data"`
	}
	if json.Unmarshal(message, &wrap) == nil && len(wrap.Data) > 0 {
		// Unwrap data then parse as kline
		c.routeKline(wrap.Data, wrap.Stream)
		return
	}

	// 3) Parse directly as kline
	c.routeKline(message, "")
}

// Evt represents a Binance WebSocket kline event payload.
type Evt struct {
	E  string `json:"e"` // Event type (e.g., "kline")
	E1 int64  `json:"E"` // Event time (milliseconds timestamp)
	S  string `json:"s"` // Symbol (e.g., "BTCUSDT")
	K  struct {
		T  int64  `json:"t"` // Kline start time (milliseconds timestamp)
		T1 int64  `json:"T"` // Kline close time (milliseconds timestamp)
		S  string `json:"s"` // Symbol (e.g., "BTCUSDT")
		I  string `json:"i"` // Interval (e.g., "1m", "3m", "4h")
		F  int64  `json:"f"` // First trade ID
		L  int64  `json:"L"` // Last trade ID
		O  string `json:"o"` // Open price
		C  string `json:"c"` // Close price
		H  string `json:"h"` // High price
		L1 string `json:"l"` // Low price
		V  string `json:"v"` // Base asset volume
		N  int    `json:"n"` // Number of trades
		X  bool   `json:"x"` // Is this kline final? (true = finalized, false = still updating)
		Q  string `json:"q"` // Quote asset volume
		V1 string `json:"V"` // Taker buy base asset volume
		Q1 string `json:"Q"` // Taker buy quote asset volume
		B  string `json:"B"` // Ignore (reserved field)
	} `json:"k"` // Kline data
}

func (c *RawWSClient) routeKline(payload []byte, streamName string) {
	var evt Evt
	if err := json.Unmarshal(payload, &evt); err != nil || evt.E == "" || evt.K.I == "" {
		return
	}

	key := ""
	if streamName != "" {
		key = strings.ToLower(streamName) // Compatible with wrapper layer directly using stream
	} else {
		key = buildKlineStream(evt.S, evt.K.I) // Construct locally
	}

	c.mu.RLock()
	ch, ok := c.subs[key]
	c.mu.RUnlock()
	if !ok {
		// No local channel subscribed to this stream, ignore
		return
	}

	// Non-blocking delivery (drop if full, avoid blocking read loop)
	select {
	case ch <- payload:
	default:
	}
}

// resubscribeAll restores wantSubs after connection rebuild, batched + rate limited
// Groups streams by interval for better log readability
func (c *RawWSClient) resubscribeAll() {
	c.mu.RLock()
	if c.conn == nil {
		c.mu.RUnlock()
		return
	}
	// Copy current wantSubs and group by interval
	byInterval := make(map[string][]string)
	for s := range c.wantSubs {
		// Extract interval from stream name (e.g., "btcusdt@kline_1m" -> "1m")
		parts := strings.Split(s, "@kline_")
		if len(parts) != 2 {
			// Invalid format, add to a catch-all group
			byInterval[""] = append(byInterval[""], s)
			continue
		}
		interval := parts[1]
		byInterval[interval] = append(byInterval[interval], s)
	}
	c.mu.RUnlock()

	if len(byInterval) == 0 {
		return
	}

	// Send grouped by interval, then batched within each interval
	for interval, streams := range byInterval {
		for i := 0; i < len(streams); i += WSSubscribeBatchSize {
			j := i + WSSubscribeBatchSize
			if j > len(streams) {
				j = len(streams)
			}
			if err := c.SubscribeStreams(streams[i:j]); err != nil {
				log.Printf("[WS] resubscribe batch failed (interval=%s): %v", interval, err)
				// Don't return immediately on failure, let upper layer retry or wait for next reconnection
			}
			// Add delay between batches to avoid rate limiting
			time.Sleep(500 * time.Millisecond)
		}
	}
}

// Close closes client: sends close frame, closes all channels, stops reconnection
func (c *RawWSClient) Close() {
	if !c.closed.CompareAndSwap(false, true) {
		return
	}
	close(c.done)

	c.mu.Lock()
	// Close subscription channels first (so consumeLoop can exit)
	for k, ch := range c.subs {
		close(ch)
		delete(c.subs, k)
	}
	// Then close connection (this will unblock ReadMessage())
	if c.conn != nil {
		// Set read deadline to past to unblock ReadMessage() immediately
		_ = c.conn.SetReadDeadline(time.Now().Add(-time.Second))
		_ = c.conn.WriteControl(websocket.CloseMessage,
			websocket.FormatCloseMessage(websocket.CloseNormalClosure, "bye"),
			time.Now().Add(2*time.Second))
		_ = c.conn.Close()
		c.conn = nil
	}
	c.mu.Unlock()

	if c.ctrlTicker != nil {
		c.ctrlTicker.Stop()
	}
}

// ----------------- Utilities -----------------

func compactLower(in []string) []string {
	out := make([]string, 0, len(in))
	for _, s := range in {
		s = strings.ToLower(strings.TrimSpace(s))
		if s != "" {
			out = append(out, s)
		}
	}
	return out
}

func buildKlineStream(symbol, interval string) string {
	// Binance stream names must be lowercase
	return strings.ToLower(symbol) + "@kline_" + strings.ToLower(interval)
}
