package ws

import (
	"encoding/json"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"github.com/redis/go-redis/v9"
)

type Hub struct {
	clients    map[*Client]bool
	register   chan *Client
	unregister chan *Client
	broadcast  chan []byte
	rdb        *redis.Client
	sync.RWMutex
}

type Client struct {
	conn *websocket.Conn
	uid  string
	send chan []byte
}

func NewHub(rdb *redis.Client) *Hub {
	return &Hub{
		clients:    make(map[*Client]bool),
		register:   make(chan *Client),
		unregister: make(chan *Client),
		broadcast:  make(chan []byte),
		rdb:        rdb,
	}
}

func (h *Hub) Run() {
	ticker := time.NewTicker(5 * time.Second)
	for {
		select {
		case c := <-h.register:
			h.clients[c] = true
		case c := <-h.unregister:
			if _, ok := h.clients[c]; ok {
				delete(h.clients, c)
				close(c.send)
			}
		case message := <-h.broadcast:
			for client := range h.clients {
				select {
				case client.send <- message:
				default:
					delete(h.clients, client)
					close(client.send)
				}
			}
		case <-ticker.C:
			go h.publishStats()
		}
	}
}

func (h *Hub) publishStats() {
	// gather stats: placeholder example
	stats := map[string]interface{}{
		"router":   map[string]interface{}{"uptime": time.Now().Format(time.RFC3339)},
		"tunnels":  []string{"tun1", "tun2"},
		"i2psnark": map[string]int{"torrents": 3},
	}
	b, err := json.Marshal(stats)
	if err != nil {
		log.Println("ws marshal:", err)
		return
	}
	h.broadcast <- b
}

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool { return true },
}

func ServeWS(h *Hub, w http.ResponseWriter, r *http.Request, uid string) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		return
	}
	client := &Client{conn: conn, uid: uid, send: make(chan []byte, 256)}
	h.register <- client
	go client.writePump()
	go client.readPump(h)
}

func (c *Client) readPump(h *Hub) {
	defer func() {
		h.unregister <- c
		c.conn.Close()
	}()
	c.conn.SetReadLimit(512)
	for {
		_, _, err := c.conn.ReadMessage()
		if err != nil {
			return
		}
	}
}

func (c *Client) writePump() {
	for {
		select {
		case msg, ok := <-c.send:
			if !ok {
				c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}
			c.conn.WriteMessage(websocket.TextMessage, msg)
		}
	}
}
