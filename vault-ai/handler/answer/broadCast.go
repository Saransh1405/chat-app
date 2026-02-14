package answer

import (
	"sync"

	"github.com/google/uuid"
)

type SessionBroadcaster struct {
	Clients map[uuid.UUID][]chan string
	mutex   sync.RWMutex
}

var GlobalBroadcaster = &SessionBroadcaster{
	Clients: make(map[uuid.UUID][]chan string),
}

// register a client to the session
func (sb *SessionBroadcaster) RegisterClient(session uuid.UUID, ch chan string) {
	sb.mutex.Lock()
	defer sb.mutex.Unlock()

	sb.Clients[session] = append(sb.Clients[session], ch)
}

// unregister a cleint to the session
func (sb *SessionBroadcaster) UnRegisterClient(session uuid.UUID, ch chan string) {
	sb.mutex.Lock()
	defer sb.mutex.Unlock()

	clients := sb.Clients[session]
	for i, client := range clients {
		if client == ch {
			sb.Clients[session] = append(clients[:i], clients[i+1:]...)
			close(ch)
			break
		}
	}

	// Clean up if no clients left
	if len(sb.Clients[session]) == 0 {
		delete(sb.Clients, session)
	}
}

// broadcast to all clients in the session
func (sb *SessionBroadcaster) BroadcastToSession(session uuid.UUID, message string) {
	sb.mutex.RLock()
	defer sb.mutex.RUnlock()

	clients := sb.Clients[session]
	for _, ch := range clients {
		select {
		case ch <- message:
			// Message sent successfully
		default:
			// Channel is full, skip
		}
	}
}
