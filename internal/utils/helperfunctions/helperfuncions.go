package helperfunctions

import "chat-app/internal/models"

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

func ValidateUserIsMemberOfRoom(userID string, roomID string) error {
	return nil
}