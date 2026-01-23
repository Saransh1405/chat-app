package rest

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"chat-app/internal/database"
	"chat-app/internal/models"
	"chat-app/internal/utils/errors"
	"chat-app/internal/utils/helperfunctions"
	"chat-app/internal/utils/validator"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type UserHandler struct {
	db *database.DB
}

func NewUserHandler(db *database.DB) *UserHandler {
	return &UserHandler{db: db}
}

func (h *UserHandler) Create(c *gin.Context) {
	var req models.UserCreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		details := []errors.ErrorDetail{
			{
				Field: "request_body",
				Issue: "Invalid JSON format or missing required fields",
			},
		}
		errors.RespondWithError(c, http.StatusBadRequest, errors.ErrCodeValidationError,
			"Invalid request body", details)
		return
	}

	appID := ""
	if req.ApplicationID != nil {
		appID = req.ApplicationID.String()
		if !validator.ValidateUUID(appID) {
			details := []errors.ErrorDetail{
				{
					Field: "application_id",
					Issue: "Invalid UUID format",
					Value: appID,
				},
			}
			errors.RespondWithError(c, http.StatusBadRequest, errors.ErrCodeValidationError,
				"Invalid application ID", details)
			return
		}
	}

	user := models.User{
		Username:   req.Username,
		Email:      req.Email,
		AvatarURL:  req.AvatarURL,
		ExternalID: req.ExternalID,
		Metadata:   req.Metadata,
	}

	if req.ApplicationID != nil {
		user.ApplicationID = *req.ApplicationID
	}

	if !validator.ValidateEmail(*user.Email) {
		details := []errors.ErrorDetail{
			{
				Field: "email",
				Issue: "Invalid email format",
				Value: *user.Email,
			},
		}
		errors.RespondWithError(c, http.StatusBadRequest, errors.ErrCodeValidationError,
			"Invalid email", details)
		return
	}

	if !validator.ValidateLength(user.Username, 3, 255) {
		details := []errors.ErrorDetail{
			{
				Field: "username",
				Issue: "Username must be between 3 and 255 characters",
				Value: user.Username,
			},
		}
		errors.RespondWithError(c, http.StatusBadRequest, errors.ErrCodeValidationError,
			"Invalid username", details)
		return
	}

	if err := helperfunctions.CheckIfUserExistsWithEmail(c, h.db, &user, appID); err == nil {
		errors.RespondWithError(c, http.StatusConflict, errors.ErrCodeConflict,
			"User with this email already exists", nil)
		return
	}

	if err := helperfunctions.CheckIfUserExistsWithUsername(c, h.db, &user, appID); err == nil {
		errors.RespondWithError(c, http.StatusConflict, errors.ErrCodeConflict,
			"User with this username already exists", nil)
		return
	}

	err := helperfunctions.CreateUser(c, h.db, &user, appID)
	if err != nil {
		errors.RespondWithError(c, http.StatusInternalServerError, errors.ErrCodeDatabaseError,
			"Failed to create user", nil)
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "User created successfully, move to login",
		"user":    user,
	})
}

func (h *UserHandler) DirectCreate(c *gin.Context) {
	var req models.UserCreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		details := []errors.ErrorDetail{
			{
				Field: "request_body",
				Issue: "Invalid JSON format or missing required fields",
			},
		}
		errors.RespondWithError(c, http.StatusBadRequest, errors.ErrCodeValidationError,
			"Invalid request body", details)
		return
	}

	user := models.User{
		Username:   req.Username,
		Email:      req.Email,
		AvatarURL:  req.AvatarURL,
		ExternalID: req.ExternalID,
		Metadata:   req.Metadata,
	}

	if !validator.ValidateEmail(*user.Email) {
		details := []errors.ErrorDetail{
			{
				Field: "email",
				Issue: "Invalid email format",
				Value: *user.Email,
			},
		}
		errors.RespondWithError(c, http.StatusBadRequest, errors.ErrCodeValidationError,
			"Invalid email", details)
		return
	}

	if !validator.ValidateLength(user.Username, 3, 255) {
		details := []errors.ErrorDetail{
			{
				Field: "username",
				Issue: "Username must be between 3 and 255 characters",
				Value: user.Username,
			},
		}
		errors.RespondWithError(c, http.StatusBadRequest, errors.ErrCodeValidationError,
			"Invalid username", details)
		return
	}

	appID := ""
	if req.ApplicationID != nil {
		appID = req.ApplicationID.String()
	}

	if err := helperfunctions.CheckIfUserExistsWithEmail(c, h.db, &user, appID); err == nil {
		errors.RespondWithError(c, http.StatusConflict, errors.ErrCodeConflict,
			"User with this email already exists", nil)
		return
	}

	if err := helperfunctions.CheckIfUserExistsWithUsername(c, h.db, &user, appID); err == nil {
		errors.RespondWithError(c, http.StatusConflict, errors.ErrCodeConflict,
			"User with this username already exists", nil)
		return
	}

	err := helperfunctions.CreateUser(c, h.db, &user, appID)
	if err != nil {
		errors.RespondWithError(c, http.StatusInternalServerError, errors.ErrCodeDatabaseError,
			"Failed to create user", nil)
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "User created successfully, move to login",
		"user":    user,
	})
}

func (h *UserHandler) Get(c *gin.Context) {
	userID := c.Query("id")
	appID := c.Query("application_id")

	var query string
	var args []interface{}

	if userID != "" {
		if !validator.ValidateUUID(userID) {
			details := []errors.ErrorDetail{
				{
					Field: "id",
					Issue: "Invalid UUID format",
					Value: userID,
				},
			}
			errors.RespondWithError(c, http.StatusBadRequest, errors.ErrCodeValidationError,
				"Invalid user ID format", details)
			return
		}
		query = `SELECT id, application_id, external_id, username, email, avatar_url, metadata, created_at, updated_at, deleted_at 
		         FROM users 
		         WHERE id = $1 AND deleted_at IS NULL`
		args = []interface{}{userID}
	} else if appID != "" {
		if !validator.ValidateUUID(appID) {
			details := []errors.ErrorDetail{
				{
					Field: "application_id",
					Issue: "Invalid UUID format",
					Value: appID,
				},
			}
			errors.RespondWithError(c, http.StatusBadRequest, errors.ErrCodeValidationError,
				"Invalid application ID", details)
			return
		}
		query = `SELECT id, application_id, external_id, username, email, avatar_url, metadata, created_at, updated_at, deleted_at 
		         FROM users 
		         WHERE application_id = $1 AND deleted_at IS NULL`
		args = []interface{}{appID}
	} else {
		details := []errors.ErrorDetail{
			{
				Field: "id",
				Issue: "Either user ID or application ID is required",
			},
			{
				Field: "application_id",
				Issue: "Either user ID or application ID is required",
			},
		}
		errors.RespondWithError(c, http.StatusBadRequest, errors.ErrCodeValidationError,
			"Either user ID or application ID is required", details)
		return
	}

	user := models.User{}
	var metadataJSON []byte
	err := h.db.QueryRow(query, args...).Scan(
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

	if err != nil {
		if err == sql.ErrNoRows {
			errors.RespondWithError(c, http.StatusNotFound, errors.ErrCodeNotFound,
				"User not found", nil)
			return
		}
		errors.RespondWithError(c, http.StatusInternalServerError, errors.ErrCodeDatabaseError,
			"Failed to retrieve user", nil)
		return
	}

	if len(metadataJSON) > 0 {
		if err := json.Unmarshal(metadataJSON, &user.Metadata); err != nil {
			errors.RespondWithError(c, http.StatusInternalServerError, errors.ErrCodeInternalError,
				"Failed to parse user metadata", nil)
			return
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"user": user,
	})
}

func (h *UserHandler) Update(c *gin.Context) {
	var req models.UserUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		details := []errors.ErrorDetail{
			{
				Field: "request_body",
				Issue: "Invalid JSON format or missing required fields",
			},
		}
		errors.RespondWithError(c, http.StatusBadRequest, errors.ErrCodeValidationError,
			"Invalid request body", details)
		return
	}

	userID := req.ID.String()
	appID := ""
	if req.ApplicationID != nil {
		appID = req.ApplicationID.String()
		if !validator.ValidateUUID(appID) {
			details := []errors.ErrorDetail{
				{
					Field: "application_id",
					Issue: "Invalid UUID format",
					Value: appID,
				},
			}
			errors.RespondWithError(c, http.StatusBadRequest, errors.ErrCodeValidationError,
				"Invalid application ID", details)
			return
		}
	}

	updateData := models.User{
		Username:  req.Username,
		Email:     req.Email,
		AvatarURL: req.AvatarURL,
		Metadata:  req.Metadata,
	}

	if updateData.Email != nil && *updateData.Email != "" {
		if !validator.ValidateEmail(*updateData.Email) {
			details := []errors.ErrorDetail{
				{
					Field: "email",
					Issue: "Invalid email format",
					Value: *updateData.Email,
				},
			}
			errors.RespondWithError(c, http.StatusBadRequest, errors.ErrCodeValidationError,
				"Invalid email", details)
			return
		}
	}

	if updateData.Username != "" {
		if !validator.ValidateLength(updateData.Username, 3, 255) {
			details := []errors.ErrorDetail{
				{
					Field: "username",
					Issue: "Username must be between 3 and 255 characters",
					Value: updateData.Username,
				},
			}
			errors.RespondWithError(c, http.StatusBadRequest, errors.ErrCodeValidationError,
				"Invalid username", details)
			return
		}
	}

	var checkQuery string
	var checkArgs []interface{}
	if appID != "" {
		if !validator.ValidateUUID(appID) {
			details := []errors.ErrorDetail{
				{
					Field: "application_id",
					Issue: "Invalid UUID format",
					Value: appID,
				},
			}
			errors.RespondWithError(c, http.StatusBadRequest, errors.ErrCodeValidationError,
				"Invalid application ID", details)
			return
		}
		checkQuery = `SELECT id FROM users WHERE id = $1 AND application_id = $2 AND deleted_at IS NULL`
		checkArgs = []interface{}{userID, appID}
	} else {
		checkQuery = `SELECT id FROM users WHERE id = $1 AND deleted_at IS NULL`
		checkArgs = []interface{}{userID}
	}

	var existingID uuid.UUID
	err := h.db.QueryRow(checkQuery, checkArgs...).Scan(&existingID)
	if err != nil {
		if err == sql.ErrNoRows {
			errors.RespondWithError(c, http.StatusNotFound, errors.ErrCodeNotFound,
				"User not found", nil)
			return
		}
		errors.RespondWithError(c, http.StatusInternalServerError, errors.ErrCodeDatabaseError,
			"Failed to check user existence", nil)
		return
	}

	if updateData.Email != nil && *updateData.Email != "" {
		var emailCheckQuery string
		var emailCheckArgs []interface{}
		if appID != "" {
			emailCheckQuery = `SELECT id FROM users WHERE email = $1 AND application_id = $2 AND id != $3 AND deleted_at IS NULL`
			emailCheckArgs = []interface{}{*updateData.Email, appID, userID}
		} else {
			emailCheckQuery = `SELECT id FROM users WHERE email = $1 AND id != $2 AND deleted_at IS NULL`
			emailCheckArgs = []interface{}{*updateData.Email, userID}
		}

		var duplicateID uuid.UUID
		err = h.db.QueryRow(emailCheckQuery, emailCheckArgs...).Scan(&duplicateID)
		if err == nil {
			errors.RespondWithError(c, http.StatusConflict, errors.ErrCodeConflict,
				"User with this email already exists", nil)
			return
		} else if err != sql.ErrNoRows {
			errors.RespondWithError(c, http.StatusInternalServerError, errors.ErrCodeDatabaseError,
				"Failed to check email uniqueness", nil)
			return
		}
	}

	if updateData.Username != "" {
		var usernameCheckQuery string
		var usernameCheckArgs []interface{}
		if appID != "" {
			usernameCheckQuery = `SELECT id FROM users WHERE username = $1 AND application_id = $2 AND id != $3 AND deleted_at IS NULL`
			usernameCheckArgs = []interface{}{updateData.Username, appID, userID}
		} else {
			usernameCheckQuery = `SELECT id FROM users WHERE username = $1 AND id != $2 AND deleted_at IS NULL`
			usernameCheckArgs = []interface{}{updateData.Username, userID}
		}

		var duplicateID uuid.UUID
		err = h.db.QueryRow(usernameCheckQuery, usernameCheckArgs...).Scan(&duplicateID)
		if err == nil {
			errors.RespondWithError(c, http.StatusConflict, errors.ErrCodeConflict,
				"User with this username already exists", nil)
			return
		} else if err != sql.ErrNoRows {
			errors.RespondWithError(c, http.StatusInternalServerError, errors.ErrCodeDatabaseError,
				"Failed to check username uniqueness", nil)
			return
		}
	}

	updateFields := []string{}
	args := []interface{}{}
	argIndex := 1

	if updateData.Username != "" {
		updateFields = append(updateFields, fmt.Sprintf("username = $%d", argIndex))
		args = append(args, updateData.Username)
		argIndex++
	}

	if updateData.Email != nil {
		updateFields = append(updateFields, fmt.Sprintf("email = $%d", argIndex))
		args = append(args, *updateData.Email)
		argIndex++
	}

	if updateData.AvatarURL != nil {
		updateFields = append(updateFields, fmt.Sprintf("avatar_url = $%d", argIndex))
		args = append(args, *updateData.AvatarURL)
		argIndex++
	}

	if updateData.Metadata != nil {
		metadataJSON, err := json.Marshal(updateData.Metadata)
		if err != nil {
			details := []errors.ErrorDetail{
				{
					Field: "metadata",
					Issue: "Invalid JSON format",
				},
			}
			errors.RespondWithError(c, http.StatusBadRequest, errors.ErrCodeValidationError,
				"Invalid metadata format", details)
			return
		}
		updateFields = append(updateFields, fmt.Sprintf("metadata = $%d", argIndex))
		args = append(args, metadataJSON)
		argIndex++
	}

	if len(updateFields) == 0 {
		errors.RespondWithError(c, http.StatusBadRequest, errors.ErrCodeValidationError,
			"No fields to update", nil)
		return
	}

	updateFields = append(updateFields, fmt.Sprintf("updated_at = $%d", argIndex))
	args = append(args, time.Now())
	argIndex++

	var whereClause string
	if appID != "" {
		whereClause = fmt.Sprintf("WHERE id = $%d AND application_id = $%d AND deleted_at IS NULL", argIndex, argIndex+1)
		args = append(args, userID, appID)
	} else {
		whereClause = fmt.Sprintf("WHERE id = $%d AND deleted_at IS NULL", argIndex)
		args = append(args, userID)
	}

	updateQuery := fmt.Sprintf("UPDATE users SET %s %s", strings.Join(updateFields, ", "), whereClause)

	result, err := h.db.Exec(updateQuery, args...)
	if err != nil {
		errors.RespondWithError(c, http.StatusInternalServerError, errors.ErrCodeDatabaseError,
			"Failed to update user", nil)
		return
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		errors.RespondWithError(c, http.StatusInternalServerError, errors.ErrCodeDatabaseError,
			"Failed to update user", nil)
		return
	}

	if rowsAffected == 0 {
		errors.RespondWithError(c, http.StatusNotFound, errors.ErrCodeNotFound,
			"User not found", nil)
		return
	}

	var fetchQuery string
	var fetchArgs []interface{}
	if appID != "" {
		fetchQuery = `SELECT id, application_id, external_id, username, email, avatar_url, metadata, created_at, updated_at, deleted_at 
		              FROM users 
		              WHERE id = $1 AND application_id = $2 AND deleted_at IS NULL`
		fetchArgs = []interface{}{userID, appID}
	} else {
		fetchQuery = `SELECT id, application_id, external_id, username, email, avatar_url, metadata, created_at, updated_at, deleted_at 
		              FROM users 
		              WHERE id = $1 AND deleted_at IS NULL`
		fetchArgs = []interface{}{userID}
	}

	updatedUser := models.User{}
	var metadataJSON []byte
	err = h.db.QueryRow(fetchQuery, fetchArgs...).Scan(
		&updatedUser.ID,
		&updatedUser.ApplicationID,
		&updatedUser.ExternalID,
		&updatedUser.Username,
		&updatedUser.Email,
		&updatedUser.AvatarURL,
		&metadataJSON,
		&updatedUser.CreatedAt,
		&updatedUser.UpdatedAt,
		&updatedUser.DeletedAt,
	)

	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"message": "User updated successfully",
		})
		return
	}

	if len(metadataJSON) > 0 {
		if err := json.Unmarshal(metadataJSON, &updatedUser.Metadata); err != nil {
			c.JSON(http.StatusOK, gin.H{
				"message": "User updated successfully",
				"user":    updatedUser,
			})
			return
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "User updated successfully",
		"user":    updatedUser,
	})
}

func (h *UserHandler) Delete(c *gin.Context) {
	var req models.UserDeleteRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		details := []errors.ErrorDetail{
			{
				Field: "request_body",
				Issue: "Invalid JSON format or missing required fields",
			},
		}
		errors.RespondWithError(c, http.StatusBadRequest, errors.ErrCodeValidationError,
			"Invalid request body", details)
		return
	}

	userID := req.ID.String()
	appID := ""
	if req.ApplicationID != nil {
		appID = req.ApplicationID.String()
		if !validator.ValidateUUID(appID) {
			details := []errors.ErrorDetail{
				{
					Field: "application_id",
					Issue: "Invalid UUID format",
					Value: appID,
				},
			}
			errors.RespondWithError(c, http.StatusBadRequest, errors.ErrCodeValidationError,
				"Invalid application ID", details)
			return
		}
	}

	var checkQuery string
	var checkArgs []interface{}
	if appID != "" {
		if !validator.ValidateUUID(appID) {
			details := []errors.ErrorDetail{
				{
					Field: "application_id",
					Issue: "Invalid UUID format",
					Value: appID,
				},
			}
			errors.RespondWithError(c, http.StatusBadRequest, errors.ErrCodeValidationError,
				"Invalid application ID", details)
			return
		}
		checkQuery = `SELECT id FROM users WHERE id = $1 AND application_id = $2 AND deleted_at IS NULL`
		checkArgs = []interface{}{userID, appID}
	} else {
		checkQuery = `SELECT id FROM users WHERE id = $1 AND deleted_at IS NULL`
		checkArgs = []interface{}{userID}
	}

	var existingID uuid.UUID
	err := h.db.QueryRow(checkQuery, checkArgs...).Scan(&existingID)
	if err != nil {
		if err == sql.ErrNoRows {
			errors.RespondWithError(c, http.StatusNotFound, errors.ErrCodeNotFound,
				"User not found", nil)
			return
		}
		errors.RespondWithError(c, http.StatusInternalServerError, errors.ErrCodeDatabaseError,
			"Failed to check user existence", nil)
		return
	}

	var deleteQuery string
	var deleteArgs []interface{}
	now := time.Now()

	if appID != "" {
		deleteQuery = `UPDATE users SET deleted_at = $1, updated_at = $2 WHERE id = $3 AND application_id = $4 AND deleted_at IS NULL`
		deleteArgs = []interface{}{now, now, userID, appID}
	} else {
		deleteQuery = `UPDATE users SET deleted_at = $1, updated_at = $2 WHERE id = $3 AND deleted_at IS NULL`
		deleteArgs = []interface{}{now, now, userID}
	}

	result, err := h.db.Exec(deleteQuery, deleteArgs...)
	if err != nil {
		errors.RespondWithError(c, http.StatusInternalServerError, errors.ErrCodeDatabaseError,
			"Failed to delete user", nil)
		return
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		errors.RespondWithError(c, http.StatusInternalServerError, errors.ErrCodeDatabaseError,
			"Failed to delete user", nil)
		return
	}

	if rowsAffected == 0 {
		errors.RespondWithError(c, http.StatusNotFound, errors.ErrCodeNotFound,
			"User not found", nil)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "User deleted successfully",
	})
}
