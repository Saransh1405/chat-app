package handler

import (
	"fmt"
	"log"
	"net/http"
	"strconv"
	"time"

	websockets "chat-app/white-board/library"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		log.Printf("WebSocket CheckOrigin called. Origin: %s", r.Header.Get("Origin"))
		return true
	},
}

func HandleWebsocket(hub *websockets.Hub, c *gin.Context, roomID string) gin.HandlerFunc {
	return func(c *gin.Context) {
		ws, err := websockets.Upgrader.Upgrade(c.Writer, c.Request, nil)
		if err != nil {
			fmt.Println("Error upgrading to websocket:", err)
			return
		}

		room := hub.FindOrCreateRoom(roomID)

		clientID := fmt.Sprintf("user_%d", time.Now().Unix())
		client := websockets.Client{
			ID:   clientID,
			Conn: ws,
			Send: make(chan []byte, 256),
			Hub:  hub,
			Room: room,
		}

		hub.Register <- &client

		go client.Read()
		go client.Write()
	}
}

func GetRoomsList(hub *websockets.Hub) gin.HandlerFunc {
	return func(c *gin.Context) {
		fmt.Println("GET /api/v1/rooms - Listing rooms")

		page := 1
		pageSize := 20

		if pageStr := c.Query("page"); pageStr != "" {
			if p, err := strconv.Atoi(pageStr); err == nil && p > 0 {
				page = p
			}
		}

		if sizeStr := c.Query("size"); sizeStr != "" {
			if s, err := strconv.Atoi(sizeStr); err == nil && s > 0 && s <= 100 {
				pageSize = s
			}
		}

		hub.Mutex.RLock()
		rooms := make([]websockets.RoomResponse, 0, len(hub.Rooms))

		for _, room := range hub.Rooms {
			room.Mutex.RLock()
			activeUsers := len(room.Clients)
			room.Mutex.RUnlock()

			rooms = append(rooms, websockets.RoomResponse{
				ID:          room.ID,
				Name:        room.ID,
				ActiveUsers: activeUsers,
				CreatedAt:   time.Now().Unix(),
				Status:      "active",
			})
		}
		hub.Mutex.RUnlock()

		total := len(rooms)
		start := (page - 1) * pageSize
		end := start + pageSize

		if start >= total {
			start = total
		}
		if end > total {
			end = total
		}

		var paginatedRooms []websockets.RoomResponse
		if start < total {
			paginatedRooms = rooms[start:end]
		}

		response := websockets.RoomListResponse{
			Rooms:    paginatedRooms,
			Total:    total,
			Page:     page,
			PageSize: pageSize,
			HasMore:  end < total,
		}

		fmt.Printf("Returning %d rooms (total: %d)\n", len(paginatedRooms), total)
		c.JSON(http.StatusOK, response)
	}
}

func GetRoomInfo(hub *websockets.Hub) gin.HandlerFunc {
	return func(c *gin.Context) {
		roomID := c.Param("roomId")
		fmt.Printf("GET /api/v1/rooms/%s - Getting room info\n", roomID)

		if roomID == "" {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": gin.H{
					"code":    "INVALID_ROOM_ID",
					"message": "Room ID is required",
					"details": []map[string]string{
						{"field": "roomId", "issue": "Room ID parameter is missing"},
					},
				},
			})
			return
		}

		hub.Mutex.RLock()
		room, exists := hub.Rooms[roomID]
		hub.Mutex.RUnlock()

		if !exists {
			fmt.Printf("Room %s not found\n", roomID)
			c.JSON(http.StatusNotFound, gin.H{
				"error": gin.H{
					"code":    "ROOM_NOT_FOUND",
					"message": "Room not found",
					"details": []map[string]string{
						{"field": "roomId", "issue": "Room with ID " + roomID + " does not exist"},
					},
				},
			})
			return
		}

		room.Mutex.RLock()
		activeUsers := len(room.Clients)
		users := make([]websockets.RoomUserResponse, 0, activeUsers)

		for client := range room.Clients {
			users = append(users, websockets.RoomUserResponse{
				ID:       client.ID,
				Username: client.Username,
				JoinedAt: client.JoinedAt.Unix(),
				LastSeen: client.LastSeen.Unix(),
				IsOnline: true,
			})
		}
		room.Mutex.RUnlock()

		response := websockets.RoomResponse{
			ID:          room.ID,
			Name:        room.ID,
			ActiveUsers: activeUsers,
			CreatedAt:   time.Now().Unix(),
			Status:      "active",
			Users:       users,
		}

		fmt.Printf("Room %s found with %d active users\n", roomID, activeUsers)
		c.JSON(http.StatusOK, response)
	}
}
