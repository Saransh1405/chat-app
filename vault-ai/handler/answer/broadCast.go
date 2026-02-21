package answer

import (
	"sync"

	"fmt"

	"github.com/google/uuid"
)

type SessionBroadcaster struct {
	Clients map[uuid.UUID][]chan string
	mutex   sync.RWMutex
}

var GlobalBroadcaster = &SessionBroadcaster{
	Clients: make(map[uuid.UUID][]chan string),
}

func (sb *SessionBroadcaster) RegisterClient(session uuid.UUID, ch chan string) {
	sb.mutex.Lock()
	defer sb.mutex.Unlock()

	sb.Clients[session] = append(sb.Clients[session], ch)
}

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

	if len(sb.Clients[session]) == 0 {
		delete(sb.Clients, session)
	}
}

func (sb *SessionBroadcaster) BroadcastToSession(session uuid.UUID, message string) {
	sb.mutex.RLock()
	defer sb.mutex.RUnlock()

	fmt.Println("Broadcasting message to session: ", session)

	clients := sb.Clients[session]
	for _, ch := range clients {
		select {
		case ch <- message:
		default:
		}
	}
}
