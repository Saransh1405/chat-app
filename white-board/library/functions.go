package library

import (
	"encoding/json"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"github.com/redis/go-redis/v9"
)

var Upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

type Hub struct {
	Rooms      map[string]*Room
	Clients    map[*Client]bool
	Register   chan *Client
	Unregister chan *Client
	Mutex      sync.RWMutex
	redis      *redis.Client
}

type Client struct {
	ID       string
	Room     *Room
	Conn     *websocket.Conn
	Send     chan []byte
	Hub      *Hub
	UserID   string `json:"user_id"`
	Username string `json:"username"`
	JoinedAt time.Time
	LastSeen time.Time
}

type Message struct {
	Type      string
	Sender    *Client
	Content   string
	Data      DrawData
	RoomId    string
	Timestamp int64
}

type DrawData struct {
	X         float64 `json:"x,omitempty"`
	Y         float64 `json:"y,omitempty"`
	LastX     float64 `json:"lastX,omitempty"`
	LastY     float64 `json:"lastY,omitempty"`
	Color     string  `json:"color,omitempty"`
	LineWidth float64 `json:"lineWidth,omitempty"`
	Tool      string  `json:"tool,omitempty"`

	Width   float64 `json:"width,omitempty"`
	Height  float64 `json:"height,omitempty"`
	CursorX float64 `json:"cursorX,omitempty"`
	CursorY float64 `json:"cursorY,omitempty"`
}

type RoomResponse struct {
	ID          string             `json:"id"`
	Name        string             `json:"name"`
	ActiveUsers int                `json:"active_users"`
	CreatedAt   int64              `json:"created_at"`
	Status      string             `json:"status"`
	Users       []RoomUserResponse `json:"users,omitempty"`
	CanvasData  []DrawData         `json:"canvas_data,omitempty"`
}

type RoomUserResponse struct {
	ID       string `json:"id"`
	Username string `json:"username"`
	JoinedAt int64  `json:"joined_at"`
	LastSeen int64  `json:"last_seen"`
	IsOnline bool   `json:"is_online"`
}

type Room struct {
	ID        string
	Clients   map[*Client]bool
	Mutex     sync.RWMutex
	RedisKey  string
	Broadcast chan []byte
	Send      chan []byte
}

type RoomListResponse struct {
	Rooms    []RoomResponse `json:"rooms"`
	Total    int            `json:"total"`
	Page     int            `json:"page"`
	PageSize int            `json:"page_size"`
	HasMore  bool           `json:"has_more"`
}

const (
	MessageTypeJoinRoom   = "join_room"
	MessageTypeLeaveRoom  = "leave_room"
	MessageTypeUserJoined = "user_joined"
	MessageTypeUserLeft   = "user_left"

	MessageTypeDrawingStart     = "draw_start"
	MessageTypeDrawingContinues = "draw_move"
	MessageTypeDrawingStop      = "draw_end"

	MessageTypeShape      = "shape"
	MessageTypeErase      = "erase"
	MessageTypeClear      = "clear"
	MessageTypeCursorMove = "cursor_move"

	MessageTypeChatMessage = "chat_message"

	MessageTypeError = "error"
	MessageTypeAck   = "ack"
)

func NewHub() *Hub {
	return &Hub{
		Rooms:      make(map[string]*Room),
		Clients:    make(map[*Client]bool),
		Register:   make(chan *Client),
		Unregister: make(chan *Client),
	}
}

func (h *Hub) Run() {
	for {
		select {
		case client := <-h.Register:
			h.Clients[client] = true
			if client.Room != nil {
				if _, ok := h.Rooms[client.Room.ID]; !ok {
					h.Rooms[client.Room.ID] = client.Room

					if client.Room.Clients == nil {
						client.Room.Clients = make(map[*Client]bool)
					}

					log.Printf("Created new room: %s", client.Room.ID)
				}

				client.Room.Clients[client] = true
			}
			log.Println("Client registered", client)

		case client := <-h.Unregister:
			if _, ok := h.Clients[client]; ok {
				if client.Room != nil {
					if room, exists := h.Rooms[client.Room.ID]; exists {
						delete(room.Clients, client)

						if len(room.Clients) == 0 {
							delete(h.Rooms, client.Room.ID)
							log.Printf("Deleted empty room: %s", client.Room.ID)
						}
					}
				}

				delete(h.Clients, client)
				close(client.Send)
				log.Printf("Client %s disconnected. Total clients: %d", client.ID, len(h.Clients))
			}
		}
	}
}

func (c *Client) Read() {
	defer func() {
		c.Hub.Unregister <- c
		c.Conn.Close()
	}()

	c.Conn.SetReadDeadline(time.Now().Add(60 * time.Second))
	c.Conn.SetPongHandler(func(string) error {
		c.Conn.SetReadDeadline(time.Now().Add(60 * time.Second))
		return nil
	})

	for {
		_, raw, err := c.Conn.ReadMessage()
		if err != nil {
			log.Println("Error reading message:", err)
			break
		}

		var incomingMsg struct {
			Type string          `json:"type"`
			Data json.RawMessage `json:"data"`
		}

		if err := json.Unmarshal(raw, &incomingMsg); err != nil {
			log.Println("Error unmarshaling:", err)
			continue
		}

		msg := Message{
			Type: incomingMsg.Type,
			// Sender:    c,
			// RoomId:    c.Room.ID,
			// Timestamp: time.Now().UnixMilli(),
		}

		switch msg.Type {
		case "draw_start", "draw_move", "draw_end", "shape", "erase":
			var drawData DrawData
			err := json.Unmarshal(incomingMsg.Data, &drawData)
			if err != nil {
				log.Println("Error unmarshaling:", err)
				return
			}
			msg.Data = drawData

		case "cursor_move":
			var drawData DrawData
			err := json.Unmarshal(incomingMsg.Data, &drawData)
			if err != nil {
				log.Println("Error unmarshaling:", err)
				return
			}
			msg.Data = drawData

		case "clear":
		}

		broadcastMsg := map[string]interface{}{
			"type": msg.Type,
			"data": msg.Data,
		}

		b, err := json.Marshal(broadcastMsg)
		if err != nil {
			log.Println("marshal error:", err)
			continue
		}

		c.Room.Broadcast <- b
	}
}

func (c *Client) Write() {
	ticker := time.NewTicker(5 * time.Second)
	defer func() {
		ticker.Stop()
		c.Conn.Close()
	}()

	for {
		select {
		case message, ok := <-c.Send:
			if !ok {
				c.Conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			c.Conn.SetWriteDeadline(time.Now().Add(10 * time.Second))

			w, err := c.Conn.NextWriter(websocket.TextMessage)
			if err != nil {
				return
			}

			w.Write(message)

			n := len(c.Send)
			for i := 0; i < n; i++ {
				w.Write(<-c.Send)
			}

			if err := w.Close(); err != nil {
				return
			}

		case <-ticker.C:
			c.Conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if err := c.Conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

func (r *Room) Run() {
	for {
		select {
		case message := <-r.Broadcast:
			for client := range r.Clients {
				select {
				case client.Send <- message:
					log.Println("Sending message to all clients in room")

				default:
					close(client.Send)
					delete(r.Clients, client)
				}
			}

		}
	}
}

func (h *Hub) FindOrCreateRoom(id string) *Room {
	h.Mutex.Lock()
	defer h.Mutex.Unlock()

	if room, ok := h.Rooms[id]; ok {
		return room
	}
	room := &Room{
		ID:        id,
		Clients:   make(map[*Client]bool),
		Broadcast: make(chan []byte),
	}
	h.Rooms[id] = room

	go room.Run()

	return room
}
