package websocket

import (
	"log"
	"net/http"
	"time"

	"chat-app/internal/config"
	"chat-app/internal/database"
	"chat-app/internal/middleware"
	"chat-app/internal/utils/helperfunctions"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		// allow all origins for now
		allowedOrigins := []string{"localhost:3000"}
		for _, origin := range allowedOrigins {
			if origin == r.Header.Get("Origin") {
				return true
			}
		}
		return false
	},
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

type Handler struct {
	hub    *Hub
	config *config.Config
	db     *database.DB
}

func NewHandler(hub *Hub, cfg *config.Config, db *database.DB) *Handler {
	return &Handler{
		hub:    hub,
		config: cfg,
		db:     db,
	}
}

// WSConnection represents a WebSocket connection
type WSConnection struct {
	*Connection
	wsConn *websocket.Conn
}

func (h *Handler) HandleConnection(c *gin.Context) {
	middleware.Auth(h.config.JWT.Secret)(c)
	userId, ok := c.Get("user_id")
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": gin.H{
				"code":    "UNAUTHORIZED",
				"message": "Unauthorized",
			},
		})
		return
	}

	appID, ok := c.Get("app_id")
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": gin.H{
				"code":    "UNAUTHORIZED",
				"message": "Unauthorized",
			},
		})
		return
	}

	wsConn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Printf("WebSocket upgrade error: %v", err)
		return
	}

	// Create connection object
	connection := &Connection{
		ID:      uuid.New().String(),
		UserID:  userId.(string),
		AppID:   appID.(string),
		RoomIDs: make(map[string]bool),
		Send:    make(chan MessageStruct, 256),
		Hub:     h.hub,
	}

	// Create WebSocket connection wrapper
	wsConnection := &WSConnection{
		Connection: connection,
		wsConn:     wsConn,
	}

	// Register connection
	h.hub.Register <- connection

	// Start goroutines for reading and writing
	go wsConnection.writePump()
	go wsConnection.readPump()
}

// readPump pumps messages from the WebSocket connection to the hub
func (ws *WSConnection) readPump() {
	defer func() {
		ws.Connection.Hub.Unregister <- ws.Connection
		ws.wsConn.Close()
	}()

	// Set read deadline and pong handler
	ws.wsConn.SetReadDeadline(time.Now().Add(60 * time.Second))
	ws.wsConn.SetPongHandler(func(string) error {
		ws.wsConn.SetReadDeadline(time.Now().Add(60 * time.Second))
		return nil
	})

	for {
		var msg struct {
			Type    string                 `json:"type"`
			RoomID  string                 `json:"room_id,omitempty"`
			Payload map[string]interface{} `json:"payload,omitempty"`
		}

		err := ws.wsConn.ReadJSON(&msg)
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("WebSocket error: %v", err)
			}
			break
		}

		// Handle different message types
		switch msg.Type {
		case "subscribe":
			if msg.RoomID != "" {
				isMember, err := helperfunctions.ValidateUserIsMemberOfRoom(ws.UserID, msg.RoomID)
				if err != nil {
					ws.Send <- MessageStruct{
						Type: "error",
						Payload: map[string]interface{}{
							"message": "error validating user is member of room",
						},
					}
					break
				}
				if !isMember {
					ws.Hub.SubscribeToRoom(ws.Connection, msg.RoomID)
					log.Printf("User %s subscribed to room %s", ws.UserID, msg.RoomID)
				}
			}
		case "unsubscribe":
			if msg.RoomID != "" {
				isMember, err := helperfunctions.ValidateUserIsMemberOfRoom(ws.UserID, msg.RoomID)
				if err != nil {
					log.Printf("Error validating user is member of room: %v", err)
					ws.Send <- MessageStruct{
						Type: "error",
						Payload: map[string]interface{}{
							"message": "user is not a member of the room or room does not exist",
						},
					}
					break
				}
				if isMember {
					ws.Hub.UnsubscribeFromRoom(ws.Connection, msg.RoomID)
					log.Printf("User %s unsubscribed from room %s", ws.UserID, msg.RoomID)
				}
			}
		case "ping":
			ws.Send <- MessageStruct{Type: "pong"}
			break
		case "message":
			if msg.RoomID == "" {
				ws.Send <- MessageStruct{
					Type: "error",
					Payload: map[string]interface{}{
						"message": "room_id is required for messages",
					},
				}
				break
			}

			isMember, err := helperfunctions.ValidateUserIsMemberOfRoom(ws.UserID, msg.RoomID)
			if err != nil {
				log.Printf("Error validating user is member of room: %v", err)
				ws.Send <- MessageStruct{
					Type: "error",
					Payload: map[string]interface{}{
						"message": "user is not a member of the room or room does not exist",
					},
				}
				break
			}
			
			// Save message to database

			if !isMember {
				ws.Hub.Broadcast <- struct {
					RoomID  string
					Message MessageStruct
				}{
					RoomID: msg.RoomID,
					Message: MessageStruct{
						Type:    "message",
						Payload: msg.Payload,
					},
				}
			}
		default:
			log.Printf("Unknown message type: %s", msg.Type)
		}
	}
}

// writePump pumps messages from the hub to the WebSocket connection
func (ws *WSConnection) writePump() {
	ticker := time.NewTicker(54 * time.Second)
	defer func() {
		ticker.Stop()
		ws.wsConn.Close()
	}()

	for {
		select {
		case message, ok := <-ws.Send:
			ws.wsConn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if !ok {
				// Hub closed the channel
				ws.wsConn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			if err := ws.wsConn.WriteJSON(message); err != nil {
				log.Printf("WebSocket write error: %v", err)
				return
			}

		case <-ticker.C:
			// Send ping to keep connection alive
			ws.wsConn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if err := ws.wsConn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}
