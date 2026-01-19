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
	// Get application ID from path
	appID := c.Param("app_id")

	// Get user data from request body
	user := models.User{}
	err := c.ShouldBindBodyWithJSON(&user)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": gin.H{
				"code":    "BAD_REQUEST",
				"message": "Invalid request body",
			},
		})
		return
	}

	if !validator.ValidateUUID(appID) {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": gin.H{
				"code":    "BAD_REQUEST",
				"message": "Invalid application ID",
			},
		})
		return
	}
	user.ApplicationID = uuid.MustParse(appID)

	// validate user data
	if !validator.ValidateEmail(*user.Email) {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": gin.H{
				"code":    "BAD_REQUEST",
				"message": "Invalid email",
			},
		})
		return
	}

	if !validator.ValidateLength(user.Username, 3, 255) {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": gin.H{
				"code":    "BAD_REQUEST",
				"message": "Invalid username",
			},
		})
		return
	}

	if err := helperfunctions.CheckIfUserExistsWithEmail(h.db, &user, appID); err == nil {
		c.JSON(http.StatusConflict, gin.H{
			"error": gin.H{
				"code":    "USER_ALREADY_EXISTS",
				"message": "User with this email already exists",
			},
		})
		return
	}

	if err := helperfunctions.CheckIfUserExistsWithUsername(h.db, &user, appID); err == nil {
		c.JSON(http.StatusConflict, gin.H{
			"error": gin.H{
				"code":    "USER_ALREADY_EXISTS",
				"message": "User with this username already exists",
			},
		})
		return
	}

	err = helperfunctions.CreateUser(h.db, &user, appID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": gin.H{
				"code":    "INTERNAL_SERVER_ERROR",
				"message": "Failed to create user",
			},
		})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "User created successfully, move to login",
		"user":    user,
	})
}

func (h *UserHandler) DirectCreate(c *gin.Context) {
	// Get user data from request body
	user := models.User{}
	err := c.ShouldBindBodyWithJSON(&user)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": gin.H{
				"code":    "BAD_REQUEST",
				"message": "Invalid request body",
			},
		})
		return
	}

	// validate user data
	if !validator.ValidateEmail(*user.Email) {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": gin.H{
				"code":    "BAD_REQUEST",
				"message": "Invalid email",
			},
		})
		return
	}

	if !validator.ValidateLength(user.Username, 3, 255) {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": gin.H{
				"code":    "BAD_REQUEST",
				"message": "Invalid username",
			},
		})
		return
	}

	if err := helperfunctions.CheckIfUserExistsWithEmail(h.db, &user, ""); err == nil {
		c.JSON(http.StatusConflict, gin.H{
			"error": gin.H{
				"code":    "USER_ALREADY_EXISTS",
				"message": "User with this email already exists",
			},
		})
		return
	}

	if err := helperfunctions.CheckIfUserExistsWithUsername(h.db, &user, ""); err == nil {
		c.JSON(http.StatusConflict, gin.H{
			"error": gin.H{
				"code":    "USER_ALREADY_EXISTS",
				"message": "User with this username already exists",
			},
		})
		return
	}

	err = helperfunctions.CreateUser(h.db, &user, "")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": gin.H{
				"code":    "INTERNAL_SERVER_ERROR",
				"message": "Failed to create user",
			},
		})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "User created successfully, move to login",
		"user":    user,
	})
}

func (h *UserHandler) Get(c *gin.Context) {
	userID := c.Param("id")

	var query string
	var args []interface{}

	if !validator.ValidateUUID(userID) {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": gin.H{
				"code":    "BAD_REQUEST",
				"message": "Invalid user ID format",
			},
		})
		return
	}

	appID := c.Param("app_id")

	if appID != "" {
		if !validator.ValidateUUID(appID) {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": gin.H{
					"code":    "BAD_REQUEST",
					"message": "Invalid application ID",
				},
			})
			return
		}
		if userID != "" {
			query = `SELECT id, application_id, external_id, username, email, avatar_url, metadata, created_at, updated_at, deleted_at 
		         FROM users 
		         WHERE id = $1 AND application_id = $2 AND deleted_at IS NULL`
			args = []interface{}{userID, appID}
		} else {
			query = `SELECT id, application_id, external_id, username, email, avatar_url, metadata, created_at, updated_at, deleted_at 
		         FROM users 
		         WHERE application_id = $1 AND deleted_at IS NULL`
			args = []interface{}{appID}
		}
	} else {
		if userID != "" {
			query = `SELECT id, application_id, external_id, username, email, avatar_url, metadata, created_at, updated_at, deleted_at 
		         FROM users 
		         WHERE id = $1 AND deleted_at IS NULL`
			args = []interface{}{userID}
		} else {
			query = `SELECT id, application_id, external_id, username, email, avatar_url, metadata, created_at, updated_at, deleted_at 
			         FROM users 
			         WHERE deleted_at IS NULL`
			args = []interface{}{}
		}
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
			c.JSON(http.StatusNotFound, gin.H{
				"error": gin.H{
					"code":    "USER_NOT_FOUND",
					"message": "User not found",
				},
			})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": gin.H{
				"code":    "INTERNAL_SERVER_ERROR",
				"message": "Failed to retrieve user",
			},
		})
		return
	}

	if len(metadataJSON) > 0 {
		if err := json.Unmarshal(metadataJSON, &user.Metadata); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": gin.H{
					"code":    "INTERNAL_SERVER_ERROR",
					"message": "Failed to parse user metadata",
				},
			})
			return
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"user": user,
	})
}

func (h *UserHandler) Update(c *gin.Context) {
	userID := c.Param("id")
	if userID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": gin.H{
				"code":    "BAD_REQUEST",
				"message": "User ID is required",
			},
		})
		return
	}

	if !validator.ValidateUUID(userID) {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": gin.H{
				"code":    "BAD_REQUEST",
				"message": "Invalid user ID format",
			},
		})
		return
	}

	appID := c.Param("app_id")

	var updateData models.User
	err := c.ShouldBindBodyWithJSON(&updateData)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": gin.H{
				"code":    "BAD_REQUEST",
				"message": "Invalid request body",
			},
		})
		return
	}

	if updateData.Email != nil && *updateData.Email != "" {
		if !validator.ValidateEmail(*updateData.Email) {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": gin.H{
					"code":    "BAD_REQUEST",
					"message": "Invalid email",
				},
			})
			return
		}
	}

	if updateData.Username != "" {
		if !validator.ValidateLength(updateData.Username, 3, 255) {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": gin.H{
					"code":    "BAD_REQUEST",
					"message": "Invalid username",
				},
			})
			return
		}
	}

	var checkQuery string
	var checkArgs []interface{}
	if appID != "" {
		if !validator.ValidateUUID(appID) {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": gin.H{
					"code":    "BAD_REQUEST",
					"message": "Invalid application ID",
				},
			})
			return
		}
		checkQuery = `SELECT id FROM users WHERE id = $1 AND application_id = $2 AND deleted_at IS NULL`
		checkArgs = []interface{}{userID, appID}
	} else {
		checkQuery = `SELECT id FROM users WHERE id = $1 AND deleted_at IS NULL`
		checkArgs = []interface{}{userID}
	}

	var existingID uuid.UUID
	err = h.db.QueryRow(checkQuery, checkArgs...).Scan(&existingID)
	if err != nil {
		if err == sql.ErrNoRows {
			c.JSON(http.StatusNotFound, gin.H{
				"error": gin.H{
					"code":    "USER_NOT_FOUND",
					"message": "User not found",
				},
			})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": gin.H{
				"code":    "INTERNAL_SERVER_ERROR",
				"message": "Failed to check user existence",
			},
		})
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
			c.JSON(http.StatusBadRequest, gin.H{
				"error": gin.H{
					"code":    "USER_ALREADY_EXISTS",
					"message": "User with this email already exists",
				},
			})
			return
		} else if err != sql.ErrNoRows {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": gin.H{
					"code":    "INTERNAL_SERVER_ERROR",
					"message": "Failed to check email uniqueness",
				},
			})
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
			c.JSON(http.StatusBadRequest, gin.H{
				"error": gin.H{
					"code":    "USER_ALREADY_EXISTS",
					"message": "User with this username already exists",
				},
			})
			return
		} else if err != sql.ErrNoRows {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": gin.H{
					"code":    "INTERNAL_SERVER_ERROR",
					"message": "Failed to check username uniqueness",
				},
			})
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
			c.JSON(http.StatusBadRequest, gin.H{
				"error": gin.H{
					"code":    "BAD_REQUEST",
					"message": "Invalid metadata format",
				},
			})
			return
		}
		updateFields = append(updateFields, fmt.Sprintf("metadata = $%d", argIndex))
		args = append(args, metadataJSON)
		argIndex++
	}

	if len(updateFields) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": gin.H{
				"code":    "BAD_REQUEST",
				"message": "No fields to update",
			},
		})
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
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": gin.H{
				"code":    "INTERNAL_SERVER_ERROR",
				"message": "Failed to update user",
			},
		})
		return
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": gin.H{
				"code":    "INTERNAL_SERVER_ERROR",
				"message": "Failed to update user",
			},
		})
		return
	}

	if rowsAffected == 0 {
		c.JSON(http.StatusNotFound, gin.H{
			"error": gin.H{
				"code":    "USER_NOT_FOUND",
				"message": "User not found",
			},
		})
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
	userID := c.Param("id")
	if userID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": gin.H{
				"code":    "BAD_REQUEST",
				"message": "User ID is required",
			},
		})
		return
	}

	if !validator.ValidateUUID(userID) {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": gin.H{
				"code":    "BAD_REQUEST",
				"message": "Invalid user ID format",
			},
		})
		return
	}

	appID := c.Param("app_id")

	var checkQuery string
	var checkArgs []interface{}
	if appID != "" {
		if !validator.ValidateUUID(appID) {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": gin.H{
					"code":    "BAD_REQUEST",
					"message": "Invalid application ID",
				},
			})
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
			c.JSON(http.StatusNotFound, gin.H{
				"error": gin.H{
					"code":    "USER_NOT_FOUND",
					"message": "User not found",
				},
			})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": gin.H{
				"code":    "INTERNAL_SERVER_ERROR",
				"message": "Failed to check user existence",
			},
		})
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
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": gin.H{
				"code":    "INTERNAL_SERVER_ERROR",
				"message": "Failed to delete user",
			},
		})
		return
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": gin.H{
				"code":    "INTERNAL_SERVER_ERROR",
				"message": "Failed to delete user",
			},
		})
		return
	}

	if rowsAffected == 0 {
		c.JSON(http.StatusNotFound, gin.H{
			"error": gin.H{
				"code":    "USER_NOT_FOUND",
				"message": "User not found",
			},
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "User deleted successfully",
	})
}
