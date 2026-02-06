package websocket

import (
	"sync"

	"chat-app/internal/utils/logger"
)

type MessageStruct struct {
	Type    string      `json:"type"`
	Payload interface{} `json:"payload,omitempty"`
}

type Connection struct {
	ID      string
	UserID  string
	AppID   string
	RoomIDs map[string]bool
	Send    chan MessageStruct
	Hub     *Hub
}

type Hub struct {
	Connections map[string]map[*Connection]bool

	RoomSubscriptions map[string]map[*Connection]bool

	Register chan *Connection

	Unregister chan *Connection

	Broadcast chan struct {
		RoomID  string
		Message MessageStruct
	}

	mu sync.RWMutex
}

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

func (h *Hub) Run() {
	for {
		select {
		case conn := <-h.Register:
			h.registerConnection(conn)
			logger.LogWebSocketEvent("connection_registered", conn.UserID, "", logger.Fields{
				"connection_id": conn.ID,
				"app_id":        conn.AppID,
			})

		case conn := <-h.Unregister:
			h.unregisterConnection(conn)
			logger.LogWebSocketEvent("connection_unregistered", conn.UserID, "", logger.Fields{
				"connection_id": conn.ID,
				"app_id":        conn.AppID,
			})

		case broadcast := <-h.Broadcast:
			h.broadcastToRoom(broadcast.RoomID, broadcast.Message)
		}
	}
}

func (h *Hub) registerConnection(conn *Connection) {
	if h.Connections[conn.UserID] == nil {
		h.Connections[conn.UserID] = make(map[*Connection]bool)
	}
	h.Connections[conn.UserID][conn] = true
}

func (h *Hub) unregisterConnection(conn *Connection) {
	if userConns, ok := h.Connections[conn.UserID]; ok {
		delete(userConns, conn)
		if len(userConns) == 0 {
			delete(h.Connections, conn.UserID)
		}
	}

	for roomID := range conn.RoomIDs {
		h.unsubscribeFromRoom(conn, roomID)
	}

	close(conn.Send)
}

func (h *Hub) SubscribeToRoom(conn *Connection, roomID string) {
	h.mu.Lock()
	defer h.mu.Unlock()

	if h.RoomSubscriptions[roomID] == nil {
		h.RoomSubscriptions[roomID] = make(map[*Connection]bool)
	}
	h.RoomSubscriptions[roomID][conn] = true
	conn.RoomIDs[roomID] = true
}

func (h *Hub) UnsubscribeFromRoom(conn *Connection, roomID string) {
	h.mu.Lock()
	defer h.mu.Unlock()
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

func (h *Hub) broadcastToRoom(roomID string, message MessageStruct) {
	h.mu.RLock()
	roomConns := make([]*Connection, 0)
	if conns, ok := h.RoomSubscriptions[roomID]; ok {
		for conn := range conns {
			roomConns = append(roomConns, conn)
		}
	}
	h.mu.RUnlock()

	logger.Info("Broadcasting to room", logger.Fields{
		"room_id":          roomID,
		"subscriber_count": len(roomConns),
		"message_type":     message.Type,
	})

	if len(roomConns) == 0 {
		logger.Warn("No subscribers found for room", logger.Fields{
			"room_id": roomID,
			"available_rooms": func() []string {
				h.mu.RLock()
				defer h.mu.RUnlock()
				rooms := make([]string, 0, len(h.RoomSubscriptions))
				for r := range h.RoomSubscriptions {
					rooms = append(rooms, r)
				}
				return rooms
			}(),
		})
		return
	}

	// Send messages outside of lock to avoid blocking
	sentCount := 0
	for _, conn := range roomConns {
		select {
		case conn.Send <- message:
			sentCount++
		default:
			// Channel is full or closed, unregister connection
			logger.Warn("Failed to send message to connection, channel full or closed", logger.Fields{
				"connection_id": conn.ID,
				"user_id":       conn.UserID,
				"room_id":       roomID,
			})
			go func(c *Connection) {
				h.mu.Lock()
				h.unregisterConnection(c)
				h.mu.Unlock()
			}(conn)
		}
	}

	logger.Info("Broadcast completed", logger.Fields{
		"room_id":       roomID,
		"total_sent":    sentCount,
		"total_targets": len(roomConns),
	})
}

func (h *Hub) GetRoomConnections(roomID string) int {
	h.mu.RLock()
	defer h.mu.RUnlock()

	if roomConns, ok := h.RoomSubscriptions[roomID]; ok {
		return len(roomConns)
	}
	return 0
}
