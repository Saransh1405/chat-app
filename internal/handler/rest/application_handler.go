package rest

import (
	"database/sql"
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

type ApplicationHandler struct {
	db *database.DB
}

func NewApplicationHandler(db *database.DB) *ApplicationHandler {
	return &ApplicationHandler{db: db}
}

func (h *ApplicationHandler) Create(c *gin.Context) {
	// Get application data from request body
	var app models.Application
	err := c.ShouldBindBodyWithJSON(&app)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": gin.H{
				"code":    "BAD_REQUEST",
				"message": "Invalid request body",
			},
		})
		return
	}

	// Validate name
	if !validator.ValidateNotEmpty(app.Name) {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": gin.H{
				"code":    "BAD_REQUEST",
				"message": "Application name is required",
			},
		})
		return
	}

	if !validator.ValidateLength(app.Name, 3, 255) {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": gin.H{
				"code":    "BAD_REQUEST",
				"message": "Application name must be between 3 and 255 characters",
			},
		})
		return
	}

	// Check if application with same name already exists
	var existingID uuid.UUID
	checkQuery := `SELECT id FROM applications WHERE name = $1 AND deleted_at IS NULL`
	err = h.db.QueryRow(checkQuery, app.Name).Scan(&existingID)
	if err == nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": gin.H{
				"code":    "APPLICATION_ALREADY_EXISTS",
				"message": "Application with this name already exists",
			},
		})
		return
	} else if err != sql.ErrNoRows {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": gin.H{
				"code":    "INTERNAL_SERVER_ERROR",
				"message": "Failed to check application existence",
			},
		})
		return
	}

	// Generate API key and secret key
	apiKey, err := helperfunctions.GenerateAPIKey()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": gin.H{
				"code":    "INTERNAL_SERVER_ERROR",
				"message": "Failed to generate API key",
			},
		})
		return
	}

	secretKey, err := helperfunctions.GenerateSecretKey()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": gin.H{
				"code":    "INTERNAL_SERVER_ERROR",
				"message": "Failed to generate secret key",
			},
		})
		return
	}

	// Ensure API key is unique (retry if collision, though very unlikely)
	maxRetries := 5
	for i := 0; i < maxRetries; i++ {
		var existingAPIKey string
		apiKeyCheckQuery := `SELECT api_key FROM applications WHERE api_key = $1 AND deleted_at IS NULL`
		err = h.db.QueryRow(apiKeyCheckQuery, apiKey).Scan(&existingAPIKey)
		if err == sql.ErrNoRows {
			break // API key is unique, proceed
		} else if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": gin.H{
					"code":    "INTERNAL_SERVER_ERROR",
					"message": "Failed to check API key uniqueness",
				},
			})
			return
		}
		// Collision detected, generate new API key
		if i < maxRetries-1 {
			apiKey, err = helperfunctions.GenerateAPIKey()
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{
					"error": gin.H{
						"code":    "INTERNAL_SERVER_ERROR",
						"message": "Failed to generate API key",
					},
				})
				return
			}
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": gin.H{
					"code":    "INTERNAL_SERVER_ERROR",
					"message": "Failed to generate unique API key",
				},
			})
			return
		}
	}

	// Insert application into database
	insertQuery := `INSERT INTO applications (name, api_key, secret_key, created_at, updated_at) 
	                VALUES ($1, $2, $3, $4, $5) 
	                RETURNING id, created_at, updated_at`

	now := time.Now()
	var appID uuid.UUID
	err = h.db.QueryRow(insertQuery, app.Name, apiKey, secretKey, now, now).Scan(
		&appID,
		&app.CreatedAt,
		&app.UpdatedAt,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": gin.H{
				"code":    "INTERNAL_SERVER_ERROR",
				"message": "Failed to create application",
			},
		})
		return
	}

	// Fetch created application to return (excluding secret_key from response)
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
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": gin.H{
				"code":    "INTERNAL_SERVER_ERROR",
				"message": "Failed to retrieve created application",
			},
		})
		return
	}

	// Return response with secret key only on creation (security: client should store this immediately)
	c.JSON(http.StatusCreated, gin.H{
		"message": "Application created successfully",
		"application": gin.H{
			"id":         createdApp.ID,
			"name":       createdApp.Name,
			"api_key":    createdApp.APIKey,
			"secret_key": secretKey, // Only returned on creation
			"created_at": createdApp.CreatedAt,
			"updated_at": createdApp.UpdatedAt,
		},
	})
}

func (h *ApplicationHandler) Get(c *gin.Context) {
	// Get application ID from path parameter
	appID := c.Param("id")
	if appID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": gin.H{
				"code":    "BAD_REQUEST",
				"message": "Application ID is required",
			},
		})
		return
	}

	// Validate UUID format
	if !validator.ValidateUUID(appID) {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": gin.H{
				"code":    "BAD_REQUEST",
				"message": "Invalid application ID format",
			},
		})
		return
	}

	// Query application from database
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
			c.JSON(http.StatusNotFound, gin.H{
				"error": gin.H{
					"code":    "APPLICATION_NOT_FOUND",
					"message": "Application not found",
				},
			})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": gin.H{
				"code":    "INTERNAL_SERVER_ERROR",
				"message": "Failed to retrieve application",
			},
		})
		return
	}

	// Return application (secret_key is not included in response)
	c.JSON(http.StatusOK, gin.H{
		"application": gin.H{
			"id":         app.ID,
			"name":       app.Name,
			"api_key":    app.APIKey,
			"created_at": app.CreatedAt,
			"updated_at": app.UpdatedAt,
		},
	})
}

func (h *ApplicationHandler) List(c *gin.Context) {
	// Query all applications from database
	query := `SELECT id, name, api_key, created_at, updated_at, deleted_at 
	          FROM applications 
	          WHERE deleted_at IS NULL 
	          ORDER BY created_at DESC`

	rows, err := h.db.Query(query)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": gin.H{
				"code":    "INTERNAL_SERVER_ERROR",
				"message": "Failed to retrieve applications",
			},
		})
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
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": gin.H{
					"code":    "INTERNAL_SERVER_ERROR",
					"message": "Failed to scan application",
				},
			})
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
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": gin.H{
				"code":    "INTERNAL_SERVER_ERROR",
				"message": "Failed to retrieve applications",
			},
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"applications": applications,
		"count":        len(applications),
	})
}

func (h *ApplicationHandler) Update(c *gin.Context) {
	// Get application ID from path parameter
	appID := c.Param("id")
	if appID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": gin.H{
				"code":    "BAD_REQUEST",
				"message": "Application ID is required",
			},
		})
		return
	}

	// Validate UUID format
	if !validator.ValidateUUID(appID) {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": gin.H{
				"code":    "BAD_REQUEST",
				"message": "Invalid application ID format",
			},
		})
		return
	}

	// Get application data from request body
	var updateData models.Application
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

	// Validate name if provided
	if updateData.Name != "" {
		if !validator.ValidateLength(updateData.Name, 3, 255) {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": gin.H{
					"code":    "BAD_REQUEST",
					"message": "Application name must be between 3 and 255 characters",
				},
			})
			return
		}
	}

	// Check if application exists
	var existingID uuid.UUID
	checkQuery := `SELECT id FROM applications WHERE id = $1 AND deleted_at IS NULL`
	err = h.db.QueryRow(checkQuery, appID).Scan(&existingID)
	if err != nil {
		if err == sql.ErrNoRows {
			c.JSON(http.StatusNotFound, gin.H{
				"error": gin.H{
					"code":    "APPLICATION_NOT_FOUND",
					"message": "Application not found",
				},
			})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": gin.H{
				"code":    "INTERNAL_SERVER_ERROR",
				"message": "Failed to check application existence",
			},
		})
		return
	}

	// Check for duplicate name if name is being updated
	if updateData.Name != "" {
		var duplicateID uuid.UUID
		nameCheckQuery := `SELECT id FROM applications WHERE name = $1 AND id != $2 AND deleted_at IS NULL`
		err = h.db.QueryRow(nameCheckQuery, updateData.Name, appID).Scan(&duplicateID)
		if err == nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": gin.H{
					"code":    "APPLICATION_ALREADY_EXISTS",
					"message": "Application with this name already exists",
				},
			})
			return
		} else if err != sql.ErrNoRows {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": gin.H{
					"code":    "INTERNAL_SERVER_ERROR",
					"message": "Failed to check name uniqueness",
				},
			})
			return
		}
	}

	// Build update query dynamically based on provided fields
	updateFields := []string{}
	args := []interface{}{}
	argIndex := 1

	if updateData.Name != "" {
		updateFields = append(updateFields, fmt.Sprintf("name = $%d", argIndex))
		args = append(args, updateData.Name)
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

	// Add updated_at
	updateFields = append(updateFields, fmt.Sprintf("updated_at = $%d", argIndex))
	args = append(args, time.Now())
	argIndex++

	// Build WHERE clause
	whereClause := fmt.Sprintf("WHERE id = $%d AND deleted_at IS NULL", argIndex)
	args = append(args, appID)

	// Build final query
	updateQuery := fmt.Sprintf("UPDATE applications SET %s %s", strings.Join(updateFields, ", "), whereClause)

	// Execute update
	result, err := h.db.Exec(updateQuery, args...)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": gin.H{
				"code":    "INTERNAL_SERVER_ERROR",
				"message": "Failed to update application",
			},
		})
		return
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": gin.H{
				"code":    "INTERNAL_SERVER_ERROR",
				"message": "Failed to update application",
			},
		})
		return
	}

	if rowsAffected == 0 {
		c.JSON(http.StatusNotFound, gin.H{
			"error": gin.H{
				"code":    "APPLICATION_NOT_FOUND",
				"message": "Application not found",
			},
		})
		return
	}

	// Fetch updated application to return
	fetchQuery := `SELECT id, name, api_key, created_at, updated_at, deleted_at 
	              FROM applications 
	              WHERE id = $1 AND deleted_at IS NULL`

	updatedApp := models.Application{}
	err = h.db.QueryRow(fetchQuery, appID).Scan(
		&updatedApp.ID,
		&updatedApp.Name,
		&updatedApp.APIKey,
		&updatedApp.CreatedAt,
		&updatedApp.UpdatedAt,
		&updatedApp.DeletedAt,
	)

	if err != nil {
		// If we can't fetch the updated application, still return success but without application data
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
	// Get application ID from path parameter
	appID := c.Param("id")
	if appID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": gin.H{
				"code":    "BAD_REQUEST",
				"message": "Application ID is required",
			},
		})
		return
	}

	// Validate UUID format
	if !validator.ValidateUUID(appID) {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": gin.H{
				"code":    "BAD_REQUEST",
				"message": "Invalid application ID format",
			},
		})
		return
	}

	// Check if application exists before soft delete
	var existingID uuid.UUID
	checkQuery := `SELECT id FROM applications WHERE id = $1 AND deleted_at IS NULL`
	err := h.db.QueryRow(checkQuery, appID).Scan(&existingID)
	if err != nil {
		if err == sql.ErrNoRows {
			c.JSON(http.StatusNotFound, gin.H{
				"error": gin.H{
					"code":    "APPLICATION_NOT_FOUND",
					"message": "Application not found",
				},
			})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": gin.H{
				"code":    "INTERNAL_SERVER_ERROR",
				"message": "Failed to check application existence",
			},
		})
		return
	}

	// Perform soft delete (set deleted_at timestamp)
	now := time.Now()
	deleteQuery := `UPDATE applications SET deleted_at = $1, updated_at = $2 WHERE id = $3 AND deleted_at IS NULL`

	result, err := h.db.Exec(deleteQuery, now, now, appID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": gin.H{
				"code":    "INTERNAL_SERVER_ERROR",
				"message": "Failed to delete application",
			},
		})
		return
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": gin.H{
				"code":    "INTERNAL_SERVER_ERROR",
				"message": "Failed to delete application",
			},
		})
		return
	}

	if rowsAffected == 0 {
		c.JSON(http.StatusNotFound, gin.H{
			"error": gin.H{
				"code":    "APPLICATION_NOT_FOUND",
				"message": "Application not found",
			},
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Application deleted successfully",
	})
}
