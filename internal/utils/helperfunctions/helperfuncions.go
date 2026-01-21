package helperfunctions

import (
	"database/sql"
	"encoding/json"
	"errors"
	"time"

	"chat-app/internal/database"
	"chat-app/internal/models"
	"crypto/rand"
	"encoding/base64"

	"github.com/google/uuid"
)

func CreateUser(db *database.DB, user *models.User, appID string) error {
	appIDUUID, err := uuid.Parse(appID)
	if err != nil {
		return errors.New("invalid application ID format")
	}

	query := `INSERT INTO users (application_id, external_id, username, email, avatar_url, metadata, created_at, updated_at)
	          VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	          RETURNING id, created_at, updated_at`

	now := time.Now()
	var metadataJSON []byte
	if user.Metadata != nil {
		metadataJSON, err = json.Marshal(user.Metadata)
		if err != nil {
			return errors.New("failed to marshal metadata")
		}
	}

	err = db.QueryRow(query,
		appIDUUID,
		user.ExternalID,
		user.Username,
		user.Email,
		user.AvatarURL,
		metadataJSON,
		now,
		now,
	).Scan(&user.ID, &user.CreatedAt, &user.UpdatedAt)

	if err != nil {
		return err
	}

	user.ApplicationID = appIDUUID
	return nil
}

func CheckIfUserExistsWithEmail(db *database.DB, user *models.User, appID string) error {
	var query string
	var args []interface{}

	if appID != "" {
		appIDUUID, err := uuid.Parse(appID)
		if err != nil {
			return errors.New("invalid application ID format")
		}
		query = `SELECT id, application_id, external_id, username, email, avatar_url, metadata, created_at, updated_at, deleted_at
		         FROM users
		         WHERE email = $1 AND application_id = $2 AND deleted_at IS NULL`
		args = []interface{}{user.Email, appIDUUID}
	} else {
		query = `SELECT id, application_id, external_id, username, email, avatar_url, metadata, created_at, updated_at, deleted_at
		         FROM users
		         WHERE email = $1 AND deleted_at IS NULL`
		args = []interface{}{user.Email}
	}

	var metadataJSON []byte
	err := db.QueryRow(query, args...).Scan(
		&user.ID,
		&user.ApplicationID,
		&user.ExternalID,
		&user.Username,
		&user.Email,
		&user.AvatarURL,
		&metadataJSON,
		&user.CreatedAt,
		&user.UpdatedAt,
		&user.DeletedAt,
	)

	if err == sql.ErrNoRows {
		return errors.New("user not found")
	}
	if err != nil {
		return err
	}

	if len(metadataJSON) > 0 {
		if err := json.Unmarshal(metadataJSON, &user.Metadata); err != nil {
			// Metadata parsing error is not critical, continue
		}
	}

	return nil
}

func CheckIfUserExistsWithUsername(db *database.DB, user *models.User, appID string) error {
	var query string
	var args []interface{}

	if appID != "" {
		appIDUUID, err := uuid.Parse(appID)
		if err != nil {
			return errors.New("invalid application ID format")
		}
		query = `SELECT id, application_id, external_id, username, email, avatar_url, metadata, created_at, updated_at, deleted_at
		         FROM users
		         WHERE username = $1 AND application_id = $2 AND deleted_at IS NULL`
		args = []interface{}{user.Username, appIDUUID}
	} else {
		query = `SELECT id, application_id, external_id, username, email, avatar_url, metadata, created_at, updated_at, deleted_at
		         FROM users
		         WHERE username = $1 AND deleted_at IS NULL`
		args = []interface{}{user.Username}
	}

	var metadataJSON []byte
	err := db.QueryRow(query, args...).Scan(
		&user.ID,
		&user.ApplicationID,
		&user.ExternalID,
		&user.Username,
		&user.Email,
		&user.AvatarURL,
		&metadataJSON,
		&user.CreatedAt,
		&user.UpdatedAt,
		&user.DeletedAt,
	)

	if err == sql.ErrNoRows {
		return errors.New("user not found")
	}
	if err != nil {
		return err
	}

	if len(metadataJSON) > 0 {
		if err := json.Unmarshal(metadataJSON, &user.Metadata); err != nil {
			// Metadata parsing error is not critical, continue
		}
	}

	return nil
}

func GetUserByEmail(db *database.DB, email string) (*models.User, error) {
	query := `SELECT id, application_id, external_id, username, email, avatar_url, metadata, created_at, updated_at, deleted_at
	          FROM users
	          WHERE email = $1 AND deleted_at IS NULL`

	user := &models.User{}
	var metadataJSON []byte

	err := db.QueryRow(query, email).Scan(
		&user.ID,
		&user.ApplicationID,
		&user.ExternalID,
		&user.Username,
		&user.Email,
		&user.AvatarURL,
		&metadataJSON,
		&user.CreatedAt,
		&user.UpdatedAt,
		&user.DeletedAt,
	)

	if err == sql.ErrNoRows {
		return nil, errors.New("user not found")
	}
	if err != nil {
		return nil, err
	}

	if len(metadataJSON) > 0 {
		if err := json.Unmarshal(metadataJSON, &user.Metadata); err != nil {
			// Metadata parsing error is not critical, continue
		}
	}

	return user, nil
}

func ValidateUserIsMemberOfRoom(db *database.DB, userID string, roomID string) (bool, error) {
	userIDUUID, err := uuid.Parse(userID)
	if err != nil {
		return false, errors.New("invalid user ID format")
	}

	roomIDUUID, err := uuid.Parse(roomID)
	if err != nil {
		return false, errors.New("invalid room ID format")
	}

	query := `SELECT COUNT(*) FROM room_members
	          WHERE user_id = $1 AND room_id = $2`

	var count int
	err = db.QueryRow(query, userIDUUID, roomIDUUID).Scan(&count)
	if err != nil {
		return false, err
	}

	return count > 0, nil
}

func GenerateAPIKey() (string, error) {
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(bytes), nil
}

func GenerateSecretKey() (string, error) {
	bytes := make([]byte, 64)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(bytes), nil
}

func SaveMessageToDB(db *database.DB, message *models.Message) (*models.Message, error) {
	query := `INSERT INTO messages (room_id, user_id, content, message_type, reply_to, metadata, created_at, updated_at)
	          VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	          RETURNING id, created_at, updated_at`

	now := time.Now()
	var metadataJSON []byte
	var err error

	if message.Metadata != nil {
		metadataJSON, err = json.Marshal(message.Metadata)
		if err != nil {
			return nil, errors.New("failed to marshal metadata")
		}
	}

	err = db.QueryRow(query,
		message.RoomID,
		message.UserID,
		message.Content,
		message.MessageType,
		message.ReplyTo,
		metadataJSON,
		now,
		now,
	).Scan(&message.ID, &message.CreatedAt, &message.UpdatedAt)

	if err != nil {
		return nil, err
	}

	return message, nil
}

func CheckIfUserExistsWithId(db *database.DB, userId string) bool {
	userIDUUID, err := uuid.Parse(userId)
	if err != nil {
		return false
	}

	query := `SELECT COUNT(*) FROM users WHERE id = $1 AND deleted_at IS NULL`

	var count int
	err = db.QueryRow(query, userIDUUID).Scan(&count)
	if err != nil {
		return false
	}

	return count > 0
}

func CheckIfApplicationExistsWithId(db *database.DB, applicationId string) (bool, *models.Application) {
	appIDUUID, err := uuid.Parse(applicationId)
	if err != nil {
		return false, nil
	}

	query := `SELECT id, name, api_key, secret_key, created_at, updated_at, deleted_at FROM applications WHERE id = $1 AND deleted_at IS NULL`

	app := &models.Application{}
	err = db.QueryRow(query, appIDUUID).Scan(
		&app.ID,
		&app.Name,
		&app.APIKey,
		&app.SecretKey,
		&app.CreatedAt,
		&app.UpdatedAt,
		&app.DeletedAt,
	)
	if err != nil {
		return false, nil
	}

	return true, app
}

func SaveTypingIndicatorToDB(db *database.DB, typing *models.TypingIndicator) (*models.TypingIndicator, error) {
	query := `INSERT INTO typing_indicators (room_id, user_id, expires_at, created_at, updated_at)
	          VALUES ($1, $2, $3, $4, $5)
	          RETURNING id, room_id, user_id, created_at, expires_at`

	now := time.Now()
	var err error

	err = db.QueryRow(query,
		typing.RoomID,
		typing.UserID,
		typing.ExpiresAt,
		now,
		now,
	).Scan(&typing.ID, &typing.RoomID, &typing.UserID, &typing.CreatedAt, &typing.ExpiresAt)

	if err != nil {
		return nil, err
	}

	return typing, nil
}

func ExpireTypingIndicator(db *database.DB, typing *models.TypingIndicator) error {
	return nil
}
