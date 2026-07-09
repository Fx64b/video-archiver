package download

import (
	"net/http"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	log "github.com/sirupsen/logrus"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool { return true },
}

const (
	// pongWait is how long a connection may stay silent before the read side
	// gives up; pings are sent at 90% of it.
	pongWait   = 60 * time.Second
	pingPeriod = (pongWait * 9) / 10
	writeWait  = 10 * time.Second
	// clientSendBuffer is how many undelivered messages a client may queue
	// before further broadcasts are skipped for it. Skipping beats stalling:
	// one stuck client must never back-pressure download/tools progress for
	// everyone else.
	clientSendBuffer = 64
)

// wsClient pairs a connection with its outbound queue. writePump is the
// connection's ONLY writer — gorilla/websocket forbids concurrent writers, so
// data messages and pings are serialized through the same goroutine.
type wsClient struct {
	conn *websocket.Conn
	send chan interface{}
	hub  *WebSocketHub
}

func (c *wsClient) writePump() {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		// A write failure means the connection is dead; make sure the hub
		// forgets it. Unregister is idempotent and safe after Stop.
		c.hub.Unregister(c.conn)
		c.conn.Close()
	}()

	for {
		select {
		case msg, ok := <-c.send:
			if !ok {
				// Hub closed the queue (unregister/stop/slow drop).
				c.conn.WriteControl(websocket.CloseMessage,
					websocket.FormatCloseMessage(websocket.CloseGoingAway, ""),
					time.Now().Add(writeWait))
				return
			}
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := c.conn.WriteJSON(msg); err != nil {
				log.WithError(err).Debug("WebSocket write failed")
				return
			}
		case <-ticker.C:
			if err := c.conn.WriteControl(websocket.PingMessage, []byte{},
				time.Now().Add(writeWait)); err != nil {
				log.WithError(err).Debug("Failed to send ping")
				return
			}
		}
	}
}

// WebSocketHub fans progress updates out to connected clients. All client-map
// access happens inside Run's goroutine; other goroutines interact only
// through channels, so there is no shared mutable state to race on.
type WebSocketHub struct {
	clients    map[*websocket.Conn]*wsClient
	broadcast  chan interface{}
	register   chan *wsClient
	unregister chan *websocket.Conn

	stop     chan struct{}
	stopOnce sync.Once
	stopped  chan struct{}
}

func NewWebSocketHub() *WebSocketHub {
	return &WebSocketHub{
		clients:    make(map[*websocket.Conn]*wsClient),
		broadcast:  make(chan interface{}),
		register:   make(chan *wsClient),
		unregister: make(chan *websocket.Conn),
		stop:       make(chan struct{}),
		stopped:    make(chan struct{}),
	}
}

func (h *WebSocketHub) Run() {
	defer close(h.stopped)
	for {
		select {
		case <-h.stop:
			for conn, client := range h.clients {
				delete(h.clients, conn)
				close(client.send)
			}
			return

		case client := <-h.register:
			h.clients[client.conn] = client
			go client.writePump()
			log.Info("New client connected")

		case conn := <-h.unregister:
			if client, ok := h.clients[conn]; ok {
				delete(h.clients, conn)
				close(client.send)
				log.Debug("Client disconnected")
			}

		case update := <-h.broadcast:
			for _, client := range h.clients {
				select {
				case client.send <- update:
				default:
					// Queue full: skip this update for the lagging client
					// instead of stalling the hub or dropping the connection.
					// Progress messages are snapshots, so missed
					// intermediates are harmless; genuinely dead connections
					// are reaped when their pings fail.
				}
			}
		}
	}
}

// Stop disconnects all clients and terminates Run. Safe to call more than
// once; Register/Unregister/Broadcast become no-ops afterwards.
func (h *WebSocketHub) Stop() {
	h.stopOnce.Do(func() { close(h.stop) })
	<-h.stopped
}

func GetUpgrader() *websocket.Upgrader {
	return &upgrader
}

func (h *WebSocketHub) Register(conn *websocket.Conn) {
	client := &wsClient{
		conn: conn,
		send: make(chan interface{}, clientSendBuffer),
		hub:  h,
	}
	select {
	case h.register <- client:
	case <-h.stop:
		conn.Close()
	}
}

func (h *WebSocketHub) Unregister(conn *websocket.Conn) {
	select {
	case h.unregister <- conn:
	case <-h.stop:
	}
}

// Broadcast sends an arbitrary message to every connected client. It lets other
// services (e.g. the tools service) push updates over the same WebSocket the
// frontend is already connected to.
func (h *WebSocketHub) Broadcast(update interface{}) {
	select {
	case h.broadcast <- update:
	case <-h.stop:
	}
}
