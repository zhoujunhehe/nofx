package kline

import (
	"strings"
	"testing"
	"time"
)

// Integration tests against Binance WS.
func TestRawWSClient_Subscribe_Integration(t *testing.T) {
	c := NewRawWSClient(WSEndpointFutures)

	// Connect() blocks until connection is lost, so start it in background
	connectErr := make(chan error, 1)
	go func() {
		connectErr <- c.Connect()
	}()
	defer c.Close()

	// Wait a bit for connection to establish (Connect() logs when connected)
	// Then try to subscribe - if it fails, connection likely didn't establish
	time.Sleep(3 * time.Second)

	stream := "btcusdt@kline_1m"
	ch := c.AddSubscriber(stream, 10)
	if err := c.SubscribeStreams([]string{stream}); err != nil {
		t.Skipf("subscribe failed (connection may not be established): %v", err)
		return
	}
	// wait for any kline event up to 65s (one full minute window)
	timer := time.NewTimer(65 * time.Second)
	defer timer.Stop()
	select {
	case msg := <-ch:
		if !strings.Contains(strings.ToLower(string(msg)), `"k"`) {
			t.Fatalf("unexpected payload: %s", string(msg))
		}
		println(string(msg))
		return
	case <-timer.C:
		t.Skip("timeout waiting for kline event (environment may be rate-limited or blocked)")
	}
}
