package websocket

import (
	"log"
)

// Message represents a WebSocket message
type MessageStruct struct {
	Type    string      `json:"type"`
	Payload interface{} `json:"payload,omitempty"`
}

// Connection represents a WebSocket connection
type Connection struct {
	ID      string
	UserID  string
	AppID   string
	RoomIDs map[string]bool
	Send    chan MessageStruct
	Hub     *Hub
}

// Hub maintains active connections and handles message broadcasting
type Hub struct {
	// Registered connections by user ID
	Connections map[string]map[*Connection]bool

	// Room subscriptions: room_id -> connections
	RoomSubscriptions map[string]map[*Connection]bool

	// Register connection
	Register chan *Connection

	// Unregister connection
	Unregister chan *Connection

	// Broadcast message to room
	Broadcast chan struct {
		RoomID  string
		Message MessageStruct
	}
}

// NewHub creates a new WebSocket hub
func NewHub() *Hub {
	return &Hub{
		Connections:       make(map[string]map[*Connection]bool),
		RoomSubscriptions: make(map[string]map[*Connection]bool),
		Register:          make(chan *Connection),
		Unregister:        make(chan *Connection),
		Broadcast: make(chan struct {
			RoomID  string
			Message MessageStruct
		}),
	}
}

// Run starts the hub's main loop
func (h *Hub) Run() {
	for {
		select {
		case conn := <-h.Register:
			h.registerConnection(conn)
			log.Printf("Connection registered: %s (user: %s)", conn.ID, conn.UserID)

		case conn := <-h.Unregister:
			h.unregisterConnection(conn)
			log.Printf("Connection unregistered: %s (user: %s)", conn.ID, conn.UserID)

		case broadcast := <-h.Broadcast:
			h.broadcastToRoom(broadcast.RoomID, broadcast.Message)
		}
	}
}

// registerConnection registers a new connection
func (h *Hub) registerConnection(conn *Connection) {
	if h.Connections[conn.UserID] == nil {
		h.Connections[conn.UserID] = make(map[*Connection]bool)
	}
	h.Connections[conn.UserID][conn] = true
}

// unregisterConnection removes a connection
func (h *Hub) unregisterConnection(conn *Connection) {
	// Remove from user connections
	if userConns, ok := h.Connections[conn.UserID]; ok {
		delete(userConns, conn)
		if len(userConns) == 0 {
			delete(h.Connections, conn.UserID)
		}
	}

	// Remove from all room subscriptions
	for roomID := range conn.RoomIDs {
		h.unsubscribeFromRoom(conn, roomID)
	}

	close(conn.Send)
}

// SubscribeToRoom subscribes a connection to a room
func (h *Hub) SubscribeToRoom(conn *Connection, roomID string) {
	if h.RoomSubscriptions[roomID] == nil {
		h.RoomSubscriptions[roomID] = make(map[*Connection]bool)
	}
	h.RoomSubscriptions[roomID][conn] = true
	conn.RoomIDs[roomID] = true
}

// UnsubscribeFromRoom unsubscribes a connection from a room
func (h *Hub) UnsubscribeFromRoom(conn *Connection, roomID string) {
	h.unsubscribeFromRoom(conn, roomID)
}

func (h *Hub) unsubscribeFromRoom(conn *Connection, roomID string) {
	if roomConns, ok := h.RoomSubscriptions[roomID]; ok {
		delete(roomConns, conn)
		if len(roomConns) == 0 {
			delete(h.RoomSubscriptions, roomID)
		}
	}
	delete(conn.RoomIDs, roomID)
}

// broadcastToRoom sends a message to all connections in a room
func (h *Hub) broadcastToRoom(roomID string, message MessageStruct) {
	if roomConns, ok := h.RoomSubscriptions[roomID]; ok {
		for conn := range roomConns {
			select {
			case conn.Send <- message:
			default:
				// Connection is blocked or closed, remove it
				h.unregisterConnection(conn)
			}
		}
	}
}

// GetRoomConnections returns the number of connections in a room
func (h *Hub) GetRoomConnections(roomID string) int {
	if roomConns, ok := h.RoomSubscriptions[roomID]; ok {
		return len(roomConns)
	}
	return 0
}
