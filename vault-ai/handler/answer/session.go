package answer

import (
	"chat-app/internal/database"
	"chat-app/internal/models"
	"time"

	"github.com/google/uuid"
)

type SessionManager struct {
	db *database.DB
}

func NewSessionManager(db *database.DB) *SessionManager {
	return &SessionManager{db: db}
}

func (sm *SessionManager) CreateSession(userId uuid.UUID, firstQuestion string) (*models.ChatSession, error) {
	// create the user session
	session := models.ChatSession{
		UserID: userId,
		Title:  firstQuestion,
	}

	query := `INSERT INTO chat_sessions (user_id, title, created_at, updated_at)
	          VALUES ($1, $2, $3, $4)
	          RETURNING id, created_at, updated_at`

	err := sm.db.QueryRow(query, session.UserID, session.Title, session.CreatedAt, session.UpdatedAt).Scan(&session.ID, &session.CreatedAt, &session.UpdatedAt)
	if err != nil {
		return nil, err
	}

	// add user first message
	userMessage := &models.ChatMessage{
		SessionID: session.ID,
		Role:      "user",
		Content:   firstQuestion,
		Timestamp: time.Now().UnixMilli(),
	}

	messageQuery := `INSERT INTO chat_messages (session_id, role, content, timestamp)
	          VALUES ($1, $2, $3, $4)
	          RETURNING id`

	err = sm.db.QueryRow(messageQuery, userMessage.SessionID, userMessage.Role, userMessage.Content, userMessage.Timestamp).Scan(&userMessage.ID)
	if err != nil {
		return nil, err
	}

	return &session, nil
}

func (sm *SessionManager) GetSession(userId uuid.UUID, sessionID uuid.UUID) (*models.ChatSession, error) {
	var chatSession models.ChatSession

	query := `SELECT * FROM chat_sessions WHERE user_id = $1 AND id = $2`

	err := sm.db.QueryRow(query, userId, sessionID).Scan(&chatSession.ID, &chatSession.UserID, &chatSession.Title, &chatSession.CreatedAt, &chatSession.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return &chatSession, nil
}

func (sm *SessionManager) AddMessage(sessionID uuid.UUID, role, content string) error {
	message := models.ChatMessage{
		SessionID: sessionID,
		Role:      role,
		Content:   content,
		Timestamp: time.Now().UnixMilli(),
	}

	messageQuery := `INSERT INTO chat_messages (session_id, role, content, timestamp)
	          VALUES ($1, $2, $3, $4)
	          RETURNING id`

	err := sm.db.QueryRow(messageQuery, message.SessionID, message.Role, message.Content, message.Timestamp).Scan(&message.ID)
	if err != nil {
		return err
	}
	return nil
}
