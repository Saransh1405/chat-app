package helperfunctions

import (
	"chat-app/internal/models"
	"crypto/rand"
	"encoding/base64"
)

func CreateUser(user *models.User, appID string) error {
	return nil
}

func CheckIfUserExistsWithEmail(user *models.User, appID string) error {
	return nil
}

func CheckIfUserExistsWithUsername(user *models.User, appID string) error {
	return nil
}

func GetUserByEmail(email string) (*models.User, error) {
	return nil, nil
}

func ValidateUserIsMemberOfRoom(userID string, roomID string) (bool, error) {
	return true, nil
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
