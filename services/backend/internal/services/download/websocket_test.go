package download

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/gorilla/websocket"
)

// hubTestServer upgrades incoming connections and registers them with the hub,
// mirroring what HandleWebSocket does in the API layer.
func hubTestServer(t *testing.T, hub *WebSocketHub) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			return
		}
		hub.Register(conn)
		defer hub.Unregister(conn)
		for {
			if _, _, err := conn.ReadMessage(); err != nil {
				return
			}
		}
	}))
}

func dialHub(t *testing.T, server *httptest.Server) *websocket.Conn {
	t.Helper()
	url := "ws" + strings.TrimPrefix(server.URL, "http")
	conn, _, err := websocket.DefaultDialer.Dial(url, nil)
	if err != nil {
		t.Fatalf("dial: %v", err)
	}
	return conn
}

// TestHubConcurrentBroadcast hammers the hub from many producers while clients
// connect and disconnect. Run under -race this covers the old data races
// (map mutation under RLock, concurrent writers per connection).
func TestHubConcurrentBroadcast(t *testing.T) {
	hub := NewWebSocketHub()
	go hub.Run()
	defer hub.Stop()

	server := hubTestServer(t, hub)
	defer server.Close()

	// A reading client that consumes everything.
	reader := dialHub(t, server)
	defer reader.Close()
	received := make(chan struct{}, 1024)
	go func() {
		for {
			var msg map[string]any
			if err := reader.ReadJSON(&msg); err != nil {
				return
			}
			select {
			case received <- struct{}{}:
			default:
			}
		}
	}()

	// Churn: clients connecting and dropping while broadcasts are in flight.
	var wg sync.WaitGroup
	for i := 0; i < 4; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < 10; j++ {
				c := dialHub(t, server)
				time.Sleep(time.Millisecond)
				c.Close()
			}
		}()
	}

	// Many concurrent producers, as download workers + tools workers would be.
	for i := 0; i < 8; i++ {
		wg.Add(1)
		go func(n int) {
			defer wg.Done()
			for j := 0; j < 50; j++ {
				hub.Broadcast(map[string]any{"producer": n, "seq": j})
			}
		}(i)
	}

	wg.Wait()

	select {
	case <-received:
	case <-time.After(5 * time.Second):
		t.Fatal("reading client never received a broadcast")
	}
}

// TestHubSlowClientDoesNotStall verifies that a client that stops reading
// neither back-pressures broadcasts nor takes healthy clients down with it —
// updates are simply skipped for it while its queue is full.
func TestHubSlowClientDoesNotStall(t *testing.T) {
	hub := NewWebSocketHub()
	go hub.Run()
	defer hub.Stop()

	server := hubTestServer(t, hub)
	defer server.Close()

	// The slow client never reads.
	slow := dialHub(t, server)
	defer slow.Close()

	// The healthy client reads everything.
	healthy := dialHub(t, server)
	defer healthy.Close()
	sentinelSeen := make(chan struct{}, 1)
	go func() {
		for {
			var msg map[string]any
			if err := healthy.ReadJSON(&msg); err != nil {
				return
			}
			if msg["sentinel"] == true {
				select {
				case sentinelSeen <- struct{}{}:
				default:
				}
			}
		}
	}()

	// Push well past the slow client's buffer. If a slow client could stall
	// the hub, this send loop would block long before finishing.
	const total = clientSendBuffer * 20
	done := make(chan struct{})
	go func() {
		for i := 0; i < total; i++ {
			hub.Broadcast(map[string]any{"seq": i})
		}
		close(done)
	}()

	select {
	case <-done:
	case <-time.After(10 * time.Second):
		t.Fatal("broadcasting stalled — slow client back-pressured the hub")
	}

	// After the burst, the healthy client must still be connected and
	// receiving: keep sending sentinels until one arrives.
	deadline := time.After(5 * time.Second)
	for {
		hub.Broadcast(map[string]any{"sentinel": true})
		select {
		case <-sentinelSeen:
			return
		case <-time.After(50 * time.Millisecond):
		case <-deadline:
			t.Fatal("healthy client stopped receiving after slow client lagged")
		}
	}
}

// TestHubStop verifies Stop terminates promptly and later calls are no-ops.
func TestHubStop(t *testing.T) {
	hub := NewWebSocketHub()
	go hub.Run()

	server := hubTestServer(t, hub)
	defer server.Close()

	conn := dialHub(t, server)
	defer conn.Close()

	stopped := make(chan struct{})
	go func() {
		hub.Stop()
		hub.Stop() // must be idempotent
		close(stopped)
	}()

	select {
	case <-stopped:
	case <-time.After(5 * time.Second):
		t.Fatal("hub.Stop did not return")
	}

	// After Stop, these must not block or panic.
	hub.Broadcast(map[string]any{"seq": 1})
	hub.Unregister(conn)
}
