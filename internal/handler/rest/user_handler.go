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

	appIDUUIDString := ""
	if req.ApplicationID != nil {
		appIDUUIDString = req.ApplicationID.String()
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

	if err := helperfunctions.CheckIfUserExistsWithEmail(c, h.db, &user, appIDUUIDString); err == nil {
		errors.RespondWithError(c, http.StatusConflict, errors.ErrCodeConflict,
			"User with this email already exists", nil)
		return
	}

	if err := helperfunctions.CheckIfUserExistsWithUsername(c, h.db, &user, appIDUUIDString); err == nil {
		errors.RespondWithError(c, http.StatusConflict, errors.ErrCodeConflict,
			"User with this username already exists", nil)
		return
	}

	err := helperfunctions.CreateUser(c, h.db, &user, appIDUUIDString)
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

	appIDUUIDString := ""
	if req.ApplicationID != nil {
		appIDUUIDString = req.ApplicationID.String()
	}

	if err := helperfunctions.CheckIfUserExistsWithEmail(c, h.db, &user, appIDUUIDString); err == nil {
		errors.RespondWithError(c, http.StatusConflict, errors.ErrCodeConflict,
			"User with this email already exists", nil)
		return
	}

	if err := helperfunctions.CheckIfUserExistsWithUsername(c, h.db, &user, appIDUUIDString); err == nil {
		errors.RespondWithError(c, http.StatusConflict, errors.ErrCodeConflict,
			"User with this username already exists", nil)
		return
	}

	err := helperfunctions.CreateUser(c, h.db, &user, appIDUUIDString)
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
		userIDUUID, err := uuid.Parse(userID)
		if err != nil {
			errors.RespondWithError(c, http.StatusBadRequest, errors.ErrCodeValidationError,
				"Invalid user ID format", nil)
			return
		}
		query = `SELECT id, application_id, external_id, username, email, avatar_url, metadata, created_at, updated_at, deleted_at 
		         FROM users 
		         WHERE id = $1 AND deleted_at IS NULL`
		args = []interface{}{userIDUUID}
	} else if appID != "" {
		appIDUUID, err := uuid.Parse(appID)
		if err != nil {
			errors.RespondWithError(c, http.StatusBadRequest, errors.ErrCodeValidationError,
				"Invalid application ID format", nil)
			return
		}
		query = `SELECT id, application_id, external_id, username, email, avatar_url, metadata, created_at, updated_at, deleted_at 
		         FROM users 
		         WHERE application_id = $1 AND deleted_at IS NULL`
		args = []interface{}{appIDUUID}
	} else {
		errors.RespondWithError(c, http.StatusBadRequest, errors.ErrCodeValidationError,
			"Either user ID or application ID is required", nil)
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

	userIDUUID := req.ID
	var appIDUUID *uuid.UUID
	if req.ApplicationID != nil {
		appIDUUID = req.ApplicationID
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
	if appIDUUID != nil {
		checkQuery = `SELECT id FROM users WHERE id = $1 AND application_id = $2 AND deleted_at IS NULL`
		checkArgs = []interface{}{userIDUUID, *appIDUUID}
	} else {
		checkQuery = `SELECT id FROM users WHERE id = $1 AND deleted_at IS NULL`
		checkArgs = []interface{}{userIDUUID}
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
		if appIDUUID != nil {
			emailCheckQuery = `SELECT id FROM users WHERE email = $1 AND application_id = $2 AND id != $3 AND deleted_at IS NULL`
			emailCheckArgs = []interface{}{*updateData.Email, *appIDUUID, userIDUUID}
		} else {
			emailCheckQuery = `SELECT id FROM users WHERE email = $1 AND id != $2 AND deleted_at IS NULL`
			emailCheckArgs = []interface{}{*updateData.Email, userIDUUID}
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
		if appIDUUID != nil {
			usernameCheckQuery = `SELECT id FROM users WHERE username = $1 AND application_id = $2 AND id != $3 AND deleted_at IS NULL`
			usernameCheckArgs = []interface{}{updateData.Username, *appIDUUID, userIDUUID}
		} else {
			usernameCheckQuery = `SELECT id FROM users WHERE username = $1 AND id != $2 AND deleted_at IS NULL`
			usernameCheckArgs = []interface{}{updateData.Username, userIDUUID}
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
	if appIDUUID != nil {
		whereClause = fmt.Sprintf("WHERE id = $%d AND application_id = $%d AND deleted_at IS NULL", argIndex, argIndex+1)
		args = append(args, userIDUUID, *appIDUUID)
	} else {
		whereClause = fmt.Sprintf("WHERE id = $%d AND deleted_at IS NULL", argIndex)
		args = append(args, userIDUUID)
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
	if appIDUUID != nil {
		fetchQuery = `SELECT id, application_id, external_id, username, email, avatar_url, metadata, created_at, updated_at, deleted_at 
		              FROM users 
		              WHERE id = $1 AND application_id = $2 AND deleted_at IS NULL`
		fetchArgs = []interface{}{userIDUUID, *appIDUUID}
	} else {
		fetchQuery = `SELECT id, application_id, external_id, username, email, avatar_url, metadata, created_at, updated_at, deleted_at 
		              FROM users 
		              WHERE id = $1 AND deleted_at IS NULL`
		fetchArgs = []interface{}{userIDUUID}
	}

	updatedUser := models.User{}
	var metadataJSONScan []byte
	err = h.db.QueryRow(fetchQuery, fetchArgs...).Scan(
		&updatedUser.ID,
		&updatedUser.ApplicationID,
		&updatedUser.ExternalID,
		&updatedUser.Username,
		&updatedUser.Email,
		&updatedUser.AvatarURL,
		&metadataJSONScan,
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

	if len(metadataJSONScan) > 0 {
		if err := json.Unmarshal(metadataJSONScan, &updatedUser.Metadata); err != nil {
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

	userIDUUID := req.ID
	var appIDUUID *uuid.UUID
	if req.ApplicationID != nil {
		appIDUUID = req.ApplicationID
	}

	var checkQuery string
	var checkArgs []interface{}
	if appIDUUID != nil {
		checkQuery = `SELECT id FROM users WHERE id = $1 AND application_id = $2 AND deleted_at IS NULL`
		checkArgs = []interface{}{userIDUUID, *appIDUUID}
	} else {
		checkQuery = `SELECT id FROM users WHERE id = $1 AND deleted_at IS NULL`
		checkArgs = []interface{}{userIDUUID}
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

	if appIDUUID != nil {
		deleteQuery = `UPDATE users SET deleted_at = $1, updated_at = $2 WHERE id = $3 AND application_id = $4 AND deleted_at IS NULL`
		deleteArgs = []interface{}{now, now, userIDUUID, *appIDUUID}
	} else {
		deleteQuery = `UPDATE users SET deleted_at = $1, updated_at = $2 WHERE id = $3 AND deleted_at IS NULL`
		deleteArgs = []interface{}{now, now, userIDUUID}
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

func (h *UserHandler) GetUsers(c *gin.Context) {
	var req models.UserGetRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		details := []errors.ErrorDetail{
			{
				Field: "query_params",
				Issue: "Invalid query parameters",
			},
		}
		errors.RespondWithError(c, http.StatusBadRequest, errors.ErrCodeValidationError,
			"Invalid query parameters", details)
		return
	}

	var appID uuid.UUID
	if req.ApplicationID != nil && *req.ApplicationID != "" {
		appID, err := uuid.Parse(*req.ApplicationID)
		if err != nil {
			errors.RespondWithError(c, http.StatusBadRequest, errors.ErrCodeValidationError,
				"Invalid application_id format", nil)
			return
		}

		if appID == uuid.Nil {
			appID = uuid.Nil
		}
	}

	var query string
	var args []interface{}
	if appID != uuid.Nil {
		query = `SELECT id, application_id, external_id, username, email, avatar_url, metadata, created_at, updated_at, deleted_at 
		          FROM users 
		          WHERE application_id = $1 AND deleted_at IS NULL`
		args = []interface{}{appID}
	} else {
		query = `SELECT id, application_id, external_id, username, email, avatar_url, metadata, created_at, updated_at, deleted_at 
		          FROM users 
		          WHERE deleted_at IS NULL`
		args = []interface{}{}
	}

	fmt.Println(query)
	fmt.Println(args)

	rows, err := h.db.Query(query, args...)
	if err != nil {
		errors.RespondWithError(c, http.StatusInternalServerError, errors.ErrCodeDatabaseError,
			"Failed to retrieve users", nil)
		return
	}
	defer rows.Close()

	var users []models.User
	for rows.Next() {
		var user models.User
		var appID *uuid.UUID 
		err := rows.Scan(
			&user.ID,
			&appID,
			&user.ExternalID,
			&user.Username,
			&user.Email,
			&user.AvatarURL,
			&user.Metadata,
			&user.CreatedAt,
			&user.UpdatedAt,
			&user.DeletedAt,
		)
		if err != nil {
			errors.RespondWithError(c, http.StatusInternalServerError, errors.ErrCodeDatabaseError,
				"Failed to retrieve users", nil)
			return
		}

		if appID != nil {
			user.ApplicationID = *appID
		}

		users = append(users, user)
	}

	if err := rows.Err(); err != nil {
		errors.RespondWithError(c, http.StatusInternalServerError, errors.ErrCodeDatabaseError,
			"Failed to retrieve users", nil)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"users": users,
	})
}
