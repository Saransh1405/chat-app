package rest

import (
	"database/sql"
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

type ApplicationHandler struct {
	db *database.DB
}

func NewApplicationHandler(db *database.DB) *ApplicationHandler {
	return &ApplicationHandler{db: db}
}

func (h *ApplicationHandler) Create(c *gin.Context) {
	var req models.ApplicationCreateRequest
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

	if !validator.ValidateNotEmpty(req.Name) {
		details := []errors.ErrorDetail{
			{
				Field: "name",
				Issue: "Application name is required",
			},
		}
		errors.RespondWithError(c, http.StatusBadRequest, errors.ErrCodeValidationError,
			"Application name is required", details)
		return
	}

	if !validator.ValidateLength(req.Name, 3, 255) {
		details := []errors.ErrorDetail{
			{
				Field: "name",
				Issue: "Application name must be between 3 and 255 characters",
				Value: req.Name,
			},
		}
		errors.RespondWithError(c, http.StatusBadRequest, errors.ErrCodeValidationError,
			"Application name must be between 3 and 255 characters", details)
		return
	}

	app := models.Application{Name: req.Name}

	var existingID uuid.UUID
	checkQuery := `SELECT id FROM applications WHERE name = $1 AND deleted_at IS NULL`
	err := h.db.QueryRow(checkQuery, req.Name).Scan(&existingID)
	if err == nil {
		errors.RespondWithError(c, http.StatusConflict, errors.ErrCodeConflict,
			"Application with this name already exists", nil)
		return
	} else if err != sql.ErrNoRows {
		errors.RespondWithError(c, http.StatusInternalServerError, errors.ErrCodeDatabaseError,
			"Failed to check application existence", nil)
		return
	}

	apiKey, err := helperfunctions.GenerateAPIKey()
	if err != nil {
		errors.RespondWithError(c, http.StatusInternalServerError, errors.ErrCodeInternalError,
			"Failed to generate API key", nil)
		return
	}

	secretKey, err := helperfunctions.GenerateSecretKey()
	if err != nil {
		errors.RespondWithError(c, http.StatusInternalServerError, errors.ErrCodeInternalError,
			"Failed to generate secret key", nil)
		return
	}

	maxRetries := 5
	for i := 0; i < maxRetries; i++ {
		var existingAPIKey string
		apiKeyCheckQuery := `SELECT api_key FROM applications WHERE api_key = $1 AND deleted_at IS NULL`
		err = h.db.QueryRow(apiKeyCheckQuery, apiKey).Scan(&existingAPIKey)
		if err == sql.ErrNoRows {
			break
		} else if err != nil {
			errors.RespondWithError(c, http.StatusInternalServerError, errors.ErrCodeDatabaseError,
				"Failed to check API key uniqueness", nil)
			return
		}
		if i < maxRetries-1 {
			apiKey, err = helperfunctions.GenerateAPIKey()
			if err != nil {
				errors.RespondWithError(c, http.StatusInternalServerError, errors.ErrCodeInternalError,
					"Failed to generate API key", nil)
				return
			}
		} else {
			errors.RespondWithError(c, http.StatusInternalServerError, errors.ErrCodeInternalError,
				"Failed to generate unique API key", nil)
			return
		}
	}

	insertQuery := `INSERT INTO applications (name, api_key, secret_key, created_at, updated_at) 
	                VALUES ($1, $2, $3, $4, $5) 
	                RETURNING id, created_at, updated_at`

	now := time.Now()
	var appID uuid.UUID
	err = h.db.QueryRow(insertQuery, req.Name, apiKey, secretKey, now, now).Scan(
		&appID,
		&app.CreatedAt,
		&app.UpdatedAt,
	)
	if err != nil {
		errors.RespondWithError(c, http.StatusInternalServerError, errors.ErrCodeDatabaseError,
			"Failed to create application", nil)
		return
	}

	fetchQuery := `SELECT id, name, api_key, created_at, updated_at, deleted_at 
	              FROM applications 
	              WHERE id = $1 AND deleted_at IS NULL`

	createdApp := models.Application{}
	err = h.db.QueryRow(fetchQuery, appID).Scan(
		&createdApp.ID,
		&createdApp.Name,
		&createdApp.APIKey,
		&createdApp.CreatedAt,
		&createdApp.UpdatedAt,
		&createdApp.DeletedAt,
	)
	if err != nil {
		errors.RespondWithError(c, http.StatusInternalServerError, errors.ErrCodeDatabaseError,
			"Failed to retrieve created application", nil)
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "Application created successfully",
		"application": gin.H{
			"id":         createdApp.ID,
			"name":       createdApp.Name,
			"api_key":    createdApp.APIKey,
			"secret_key": secretKey,
			"created_at": createdApp.CreatedAt,
			"updated_at": createdApp.UpdatedAt,
		},
	})
}

func (h *ApplicationHandler) Get(c *gin.Context) {
	appID := c.Query("id")
	if appID != "" {
		if !validator.ValidateUUID(appID) {
			details := []errors.ErrorDetail{
				{
					Field: "id",
					Issue: "Invalid UUID format",
					Value: appID,
				},
			}
			errors.RespondWithError(c, http.StatusBadRequest, errors.ErrCodeValidationError,
				"Invalid application ID format", details)
			return
		}

		query := `SELECT id, name, api_key, created_at, updated_at, deleted_at 
		          FROM applications 
		          WHERE id = $1 AND deleted_at IS NULL`

		app := models.Application{}
		err := h.db.QueryRow(query, appID).Scan(
			&app.ID,
			&app.Name,
			&app.APIKey,
			&app.CreatedAt,
			&app.UpdatedAt,
			&app.DeletedAt,
		)

		if err != nil {
			if err == sql.ErrNoRows {
				errors.RespondWithError(c, http.StatusNotFound, errors.ErrCodeNotFound,
					"Application not found", nil)
				return
			}
			errors.RespondWithError(c, http.StatusInternalServerError, errors.ErrCodeDatabaseError,
				"Failed to retrieve application", nil)
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"application": gin.H{
				"id":         app.ID,
				"name":       app.Name,
				"api_key":    app.APIKey,
				"created_at": app.CreatedAt,
				"updated_at": app.UpdatedAt,
			},
		})
		return
	}

	h.List(c)
}

func (h *ApplicationHandler) List(c *gin.Context) {
	query := `SELECT id, name, api_key, created_at, updated_at, deleted_at 
	          FROM applications 
	          WHERE deleted_at IS NULL 
	          ORDER BY created_at DESC`

	rows, err := h.db.Query(query)
	if err != nil {
		errors.RespondWithError(c, http.StatusInternalServerError, errors.ErrCodeDatabaseError,
			"Failed to retrieve applications", nil)
		return
	}
	defer rows.Close()

	var applications []gin.H
	for rows.Next() {
		var app models.Application
		err := rows.Scan(
			&app.ID,
			&app.Name,
			&app.APIKey,
			&app.CreatedAt,
			&app.UpdatedAt,
			&app.DeletedAt,
		)
		if err != nil {
			errors.RespondWithError(c, http.StatusInternalServerError, errors.ErrCodeDatabaseError,
				"Failed to scan application", nil)
			return
		}

		applications = append(applications, gin.H{
			"id":         app.ID,
			"name":       app.Name,
			"api_key":    app.APIKey,
			"created_at": app.CreatedAt,
			"updated_at": app.UpdatedAt,
		})
	}

	if err = rows.Err(); err != nil {
		errors.RespondWithError(c, http.StatusInternalServerError, errors.ErrCodeDatabaseError,
			"Failed to retrieve applications", nil)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"applications": applications,
		"count":        len(applications),
	})
}

func (h *ApplicationHandler) Update(c *gin.Context) {
	var req models.ApplicationUpdateRequest
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

	if req.Name != "" {
		if !validator.ValidateLength(req.Name, 3, 255) {
			details := []errors.ErrorDetail{
				{
					Field: "name",
					Issue: "Application name must be between 3 and 255 characters",
					Value: req.Name,
				},
			}
			errors.RespondWithError(c, http.StatusBadRequest, errors.ErrCodeValidationError,
				"Application name must be between 3 and 255 characters", details)
			return
		}
	}

	var existingID uuid.UUID
	checkQuery := `SELECT id FROM applications WHERE id = $1 AND deleted_at IS NULL`
	err := h.db.QueryRow(checkQuery, req.ID).Scan(&existingID)
	if err != nil {
		if err == sql.ErrNoRows {
			errors.RespondWithError(c, http.StatusNotFound, errors.ErrCodeNotFound,
				"Application not found", nil)
			return
		}
		errors.RespondWithError(c, http.StatusInternalServerError, errors.ErrCodeDatabaseError,
			"Failed to check application existence", nil)
		return
	}

	if req.Name != "" {
		var duplicateID uuid.UUID
		nameCheckQuery := `SELECT id FROM applications WHERE name = $1 AND id != $2 AND deleted_at IS NULL`
		err = h.db.QueryRow(nameCheckQuery, req.Name, req.ID).Scan(&duplicateID)
		if err == nil {
			errors.RespondWithError(c, http.StatusConflict, errors.ErrCodeConflict,
				"Application with this name already exists", nil)
			return
		} else if err != sql.ErrNoRows {
			errors.RespondWithError(c, http.StatusInternalServerError, errors.ErrCodeDatabaseError,
				"Failed to check name uniqueness", nil)
			return
		}
	}

	updateFields := []string{}
	args := []interface{}{}
	argIndex := 1

	if req.Name != "" {
		updateFields = append(updateFields, fmt.Sprintf("name = $%d", argIndex))
		args = append(args, req.Name)
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

	whereClause := fmt.Sprintf("WHERE id = $%d AND deleted_at IS NULL", argIndex)
	args = append(args, req.ID)

	updateQuery := fmt.Sprintf("UPDATE applications SET %s %s", strings.Join(updateFields, ", "), whereClause)

	result, err := h.db.Exec(updateQuery, args...)
	if err != nil {
		errors.RespondWithError(c, http.StatusInternalServerError, errors.ErrCodeDatabaseError,
			"Failed to update application", nil)
		return
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		errors.RespondWithError(c, http.StatusInternalServerError, errors.ErrCodeDatabaseError,
			"Failed to update application", nil)
		return
	}

	if rowsAffected == 0 {
		errors.RespondWithError(c, http.StatusNotFound, errors.ErrCodeNotFound,
			"Application not found", nil)
		return
	}

	fetchQuery := `SELECT id, name, api_key, created_at, updated_at, deleted_at 
	              FROM applications 
	              WHERE id = $1 AND deleted_at IS NULL`

	updatedApp := models.Application{}
	err = h.db.QueryRow(fetchQuery, req.ID).Scan(
		&updatedApp.ID,
		&updatedApp.Name,
		&updatedApp.APIKey,
		&updatedApp.CreatedAt,
		&updatedApp.UpdatedAt,
		&updatedApp.DeletedAt,
	)

	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"message": "Application updated successfully",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Application updated successfully",
		"application": gin.H{
			"id":         updatedApp.ID,
			"name":       updatedApp.Name,
			"api_key":    updatedApp.APIKey,
			"created_at": updatedApp.CreatedAt,
			"updated_at": updatedApp.UpdatedAt,
		},
	})
}

func (h *ApplicationHandler) Delete(c *gin.Context) {
	var req models.ApplicationDeleteRequest
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

	appID := req.ID.String()

	var existingID uuid.UUID
	checkQuery := `SELECT id FROM applications WHERE id = $1 AND deleted_at IS NULL`
	err := h.db.QueryRow(checkQuery, appID).Scan(&existingID)
	if err != nil {
		if err == sql.ErrNoRows {
			errors.RespondWithError(c, http.StatusNotFound, errors.ErrCodeNotFound,
				"Application not found", nil)
			return
		}
		errors.RespondWithError(c, http.StatusInternalServerError, errors.ErrCodeDatabaseError,
			"Failed to check application existence", nil)
		return
	}

	now := time.Now()
	deleteQuery := `UPDATE applications SET deleted_at = $1, updated_at = $2 WHERE id = $3 AND deleted_at IS NULL`

	result, err := h.db.Exec(deleteQuery, now, now, req.ID)
	if err != nil {
		errors.RespondWithError(c, http.StatusInternalServerError, errors.ErrCodeDatabaseError,
			"Failed to delete application", nil)
		return
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		errors.RespondWithError(c, http.StatusInternalServerError, errors.ErrCodeDatabaseError,
			"Failed to delete application", nil)
		return
	}

	if rowsAffected == 0 {
		errors.RespondWithError(c, http.StatusNotFound, errors.ErrCodeNotFound,
			"Application not found", nil)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Application deleted successfully",
	})
}
