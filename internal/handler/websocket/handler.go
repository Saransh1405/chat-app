package websocket

import (
	"net/http"
	"strings"
	"time"

	"chat-app/internal/config"
	"chat-app/internal/database"
	"chat-app/internal/utils/errors"
	"chat-app/internal/utils/jwt"
	"chat-app/internal/utils/logger"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		origin := r.Header.Get("Origin")
		logger.Debug("WebSocket origin check", logger.Fields{
			"origin":    origin,
			"client_ip": r.RemoteAddr,
		})

		if origin == "" {
			logger.Debug("No origin header, allowing connection (development mode)")
			return true
		}

		allowedPatterns := []string{
			"http://localhost:3000",
			"http://localhost:5173",
			"http://127.0.0.1:3000",
			"http://127.0.0.1:5173",
		}
		for _, pattern := range allowedPatterns {
			if origin == pattern {
				logger.Debug("Origin matched allowed pattern", logger.Fields{"pattern": pattern})
				return true
			}
		}

		if strings.HasPrefix(origin, "http://localhost:") || strings.HasPrefix(origin, "http://127.0.0.1:") {
			logger.Debug("Origin matched localhost pattern", logger.Fields{"origin": origin})
			return true
		}

		logger.Warn("WebSocket origin not allowed", logger.Fields{
			"origin":    origin,
			"client_ip": r.RemoteAddr,
		})
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

type WSConnection struct {
	*Connection
	wsConn *websocket.Conn
	db     *database.DB
}

func (h *Handler) HandleConnection(c *gin.Context) {
	token := c.Query("token")
	if token == "" {
		authHeader := c.GetHeader("Authorization")
		if authHeader != "" {
			parts := strings.Split(authHeader, " ")
			if len(parts) == 2 && parts[0] == "Bearer" {
				token = parts[1]
			}
		}
	}

	if token == "" {
		logger.Warn("WebSocket connection failed: no token provided", logger.Fields{
			"client_ip": c.ClientIP(),
		})
		errors.RespondWithError(c, http.StatusUnauthorized, errors.ErrCodeUnauthorized,
			"Token is required", nil)
		return
	}

	claims, err := jwt.ValidateToken(token, h.config.JWT.Secret)
	if err != nil {
		logger.Warn("WebSocket connection failed: invalid token", logger.Fields{
			"client_ip": c.ClientIP(),
			"error":     err.Error(),
		})
		errors.RespondWithError(c, http.StatusUnauthorized, errors.ErrCodeUnauthorized,
			"Invalid or expired token", nil)
		return
	}

	userId := claims.UserID
	appID := claims.AppID

	wsConn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		logger.Error("WebSocket upgrade error", err, logger.Fields{
			"user_id":   userId,
			"app_id":    appID,
			"client_ip": c.ClientIP(),
		})
		return
	}

	connection := &Connection{
		ID:      uuid.New().String(),
		UserID:  userId,
		AppID:   appID,
		RoomIDs: make(map[string]bool),
		Send:    make(chan MessageStruct, 256),
		Hub:     h.hub,
	}

	wsConnection := &WSConnection{
		Connection: connection,
		wsConn:     wsConn,
		db:         h.db,
	}

	h.hub.Register <- connection

	logger.LogWebSocketEvent("connection_established", connection.UserID, "", logger.Fields{
		"connection_id": connection.ID,
		"app_id":        connection.AppID,
	})

	go wsConnection.writePump()
	go wsConnection.readPump()
}

func (ws *WSConnection) readPump() {
	defer func() {
		ws.Connection.Hub.Unregister <- ws.Connection
		ws.wsConn.Close()
		logger.LogWebSocketEvent("connection_closed", ws.UserID, "", logger.Fields{
			"connection_id": ws.ID,
		})
	}()

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
				logger.Error("WebSocket read error", err, logger.Fields{
					"connection_id": ws.ID,
					"user_id":       ws.UserID,
				})
			}
			break
		}

		switch msg.Type {
		case "subscribe":
			if msg.RoomID != "" {
				// For now, allow subscription without validation (you can uncomment validation later)
				ws.Hub.SubscribeToRoom(ws.Connection, msg.RoomID)
				logger.LogWebSocketEvent("subscribe", ws.UserID, msg.RoomID, logger.Fields{
					"connection_id": ws.ID,
				})
				ws.Send <- MessageStruct{
					Type: "subscribed",
					Payload: map[string]interface{}{
						"room_id": msg.RoomID,
						"message": "Successfully subscribed to room",
					},
				}
				// Uncomment below to enable room membership validation
				// isMember, err := helperfunctions.ValidateUserIsMemberOfRoom(ws.db, ws.UserID, msg.RoomID)
				// if err != nil {
				// 	logger.Error("Failed to validate room membership for WebSocket subscribe", err, logger.Fields{
				// 		"user_id":       ws.UserID,
				// 		"room_id":       msg.RoomID,
				// 		"connection_id": ws.ID,
				// 	})
				// 	ws.Send <- MessageStruct{
				// 		Type: "error",
				// 		Payload: map[string]interface{}{
				// 			"message": "error validating user is member of room",
				// 		},
				// 	}
				// 	break
				// }
				// if isMember {
				// 	ws.Hub.SubscribeToRoom(ws.Connection, msg.RoomID)
				// 	logger.LogWebSocketEvent("subscribe", ws.UserID, msg.RoomID, logger.Fields{
				// 		"connection_id": ws.ID,
				// 	})
				// } else {
				// 	logger.Warn("User attempted to subscribe to room they're not a member of", logger.Fields{
				// 		"user_id":       ws.UserID,
				// 		"room_id":       msg.RoomID,
				// 		"connection_id": ws.ID,
				// 	})
				// 	ws.Send <- MessageStruct{
				// 		Type: "error",
				// 		Payload: map[string]interface{}{
				// 			"message": "user is not a member of the room",
				// 		},
				// 	}
				// }
			} else {
				ws.Send <- MessageStruct{
					Type: "error",
					Payload: map[string]interface{}{
						"message": "room_id is required for subscription",
					},
				}
			}
		case "unsubscribe":
			if msg.RoomID != "" {
				ws.Hub.UnsubscribeFromRoom(ws.Connection, msg.RoomID)
				logger.LogWebSocketEvent("unsubscribe", ws.UserID, msg.RoomID, logger.Fields{
					"connection_id": ws.ID,
				})
				ws.Send <- MessageStruct{
					Type: "unsubscribed",
					Payload: map[string]interface{}{
						"room_id": msg.RoomID,
						"message": "Successfully unsubscribed from room",
					},
				}
			} else {
				ws.Send <- MessageStruct{
					Type: "error",
					Payload: map[string]interface{}{
						"message": "room_id is required for unsubscription",
					},
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

			// isMember, err := helperfunctions.ValidateUserIsMemberOfRoom(ws.db, ws.UserID, msg.RoomID)
			// if err != nil {
			// 	logger.Error("Error validating user is member of room for WebSocket message", err, logger.Fields{
			// 		"user_id":       ws.UserID,
			// 		"room_id":       msg.RoomID,
			// 		"connection_id": ws.ID,
			// 	})
			// 	ws.Send <- MessageStruct{
			// 		Type: "error",
			// 		Payload: map[string]interface{}{
			// 			"message": "error validating user is member of room",
			// 		},
			// 	}
			// 	break
			// }

			// if isMember {
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
			logger.LogWebSocketEvent("message_sent", ws.UserID, msg.RoomID, logger.Fields{
				"connection_id": ws.ID,
			})
			// } else {
			logger.Warn("User attempted to send WebSocket message to room they're not a member of", logger.Fields{
				"user_id":       ws.UserID,
				"room_id":       msg.RoomID,
				"connection_id": ws.ID,
			})
			ws.Send <- MessageStruct{
				Type: "error",
				Payload: map[string]interface{}{
					"message": "user is not a member of the room",
				},
			}
			// }
		default:
			logger.Warn("Unknown WebSocket message type", logger.Fields{
				"message_type":  msg.Type,
				"connection_id": ws.ID,
				"user_id":       ws.UserID,
			})
		}
	}
}

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
				ws.wsConn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			if err := ws.wsConn.WriteJSON(message); err != nil {
				logger.Error("WebSocket write error", err, logger.Fields{
					"connection_id": ws.ID,
					"user_id":       ws.UserID,
					"message_type":  message.Type,
				})
				return
			}

		case <-ticker.C:
			ws.wsConn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if err := ws.wsConn.WriteMessage(websocket.PingMessage, nil); err != nil {
				logger.Error("WebSocket ping error", err, logger.Fields{
					"connection_id": ws.ID,
					"user_id":       ws.UserID,
				})
				return
			}
		}
	}
}
