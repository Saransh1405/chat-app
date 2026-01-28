package rest

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"chat-app/internal/config"
	"chat-app/internal/database"
	"chat-app/internal/models"
	"chat-app/internal/utils/errors"
	"chat-app/internal/utils/helperfunctions"
	"chat-app/internal/utils/jwt"
	"chat-app/internal/utils/logger"
	"chat-app/internal/utils/validator"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type AuthHandler struct {
	config *config.Config
	db     *database.DB
}

func NewAuthHandler(cfg *config.Config, db *database.DB) *AuthHandler {
	return &AuthHandler{
		config: cfg,
		db:     db,
	}
}

func (h *AuthHandler) Register(c *gin.Context) {
	var req models.RegisterRequest
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

	if req.Email == nil || *req.Email == "" {
		details := []errors.ErrorDetail{
			{
				Field: "email",
				Issue: "Email is required",
			},
		}
		errors.RespondWithError(c, http.StatusBadRequest, errors.ErrCodeValidationError,
			"Email is required", details)
		return
	}

	email := strings.TrimSpace(*req.Email)
	if !validator.ValidateEmail(email) {
		details := []errors.ErrorDetail{
			{
				Field: "email",
				Issue: "Invalid email format",
				Value: email,
			},
		}
		errors.RespondWithError(c, http.StatusBadRequest, errors.ErrCodeValidationError,
			"Invalid email format", details)
		return
	}

	username := strings.TrimSpace(req.Username)
	if !validator.ValidateLength(username, 3, 255) {
		details := []errors.ErrorDetail{
			{
				Field: "username",
				Issue: "Username must be between 3 and 255 characters",
				Value: username,
			},
		}
		errors.RespondWithError(c, http.StatusBadRequest, errors.ErrCodeValidationError,
			"Invalid username", details)
		return
	}

	user := models.User{
		Username:   username,
		Email:      &email,
		AvatarURL:  req.AvatarURL,
		ExternalID: req.ExternalID,
		Metadata:   req.Metadata,
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
				"Invalid application ID format", details)
			return
		}
		user.ApplicationID = *req.ApplicationID
	}

	checkUser := models.User{Email: &email}
	if err := helperfunctions.CheckIfUserExistsWithEmail(c, h.db, &checkUser, appID); err == nil {
		logger.Warn("User registration failed: email already exists", logger.Fields{
			"email":          email,
			"application_id": appID,
			"client_ip":      c.ClientIP(),
		})
		errors.RespondWithError(c, http.StatusConflict, errors.ErrCodeConflict,
			"User with this email already exists", nil)
		return
	}

	checkUser = models.User{Username: username}
	if err := helperfunctions.CheckIfUserExistsWithUsername(c, h.db, &checkUser, appID); err == nil {
		logger.Warn("User registration failed: username already exists", logger.Fields{
			"username":       username,
			"application_id": appID,
			"client_ip":      c.ClientIP(),
		})
		errors.RespondWithError(c, http.StatusConflict, errors.ErrCodeConflict,
			"User with this username already exists", nil)
		return
	}

	startTime := time.Now()
	err := helperfunctions.CreateUser(c, h.db, &user, appID)
	if err != nil {
		logger.Error("Failed to create user", err, logger.Fields{
			"email":          email,
			"username":       username,
			"application_id": appID,
			"duration_ms":    time.Since(startTime).Milliseconds(),
		})
		errors.RespondWithError(c, http.StatusInternalServerError, errors.ErrCodeDatabaseError,
			"Failed to create user", nil)
		return
	}

	logger.Info("User created successfully", logger.Fields{
		"user_id":        user.ID.String(),
		"email":          email,
		"username":       username,
		"application_id": appID,
		"duration_ms":    time.Since(startTime).Milliseconds(),
	})

	accessToken, err := jwt.GenerateToken(
		user.ID.String(),
		user.ApplicationID.String(),
		"",
		h.config.JWT.Secret,
		h.config.JWT.AccessExpiry,
	)
	if err != nil {
		logger.Error("Failed to generate access token", err, logger.Fields{
			"user_id": user.ID.String(),
		})
		errors.RespondWithError(c, http.StatusInternalServerError, errors.ErrCodeInternalError,
			"Failed to generate access token", nil)
		return
	}

	refreshToken, err := jwt.GenerateToken(
		user.ID.String(),
		user.ApplicationID.String(),
		"",
		h.config.JWT.Secret,
		h.config.JWT.RefreshExpiry,
	)
	if err != nil {
		logger.Error("Failed to generate refresh token", err, logger.Fields{
			"user_id": user.ID.String(),
		})
		errors.RespondWithError(c, http.StatusInternalServerError, errors.ErrCodeInternalError,
			"Failed to generate refresh token", nil)
		return
	}

	logger.LogAuthEvent("register", user.ID.String(), true, logger.Fields{
		"email":          email,
		"username":       username,
		"application_id": appID,
	})

	c.JSON(http.StatusCreated, gin.H{
		"message":       "User registered successfully",
		"access_token":  accessToken,
		"refresh_token": refreshToken,
		"user":          user,
	})
}

func (h *AuthHandler) Login(c *gin.Context) {
	var req models.LoginRequest
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

	email := strings.TrimSpace(req.Email)
	if email == "" {
		details := []errors.ErrorDetail{
			{
				Field: "email",
				Issue: "Email is required",
			},
		}
		errors.RespondWithError(c, http.StatusBadRequest, errors.ErrCodeValidationError,
			"Email is required", details)
		return
	}

	if !validator.ValidateEmail(email) {
		details := []errors.ErrorDetail{
			{
				Field: "email",
				Issue: "Invalid email format",
				Value: email,
			},
		}
		errors.RespondWithError(c, http.StatusBadRequest, errors.ErrCodeValidationError,
			"Invalid email format", details)
		return
	}

	user, err := helperfunctions.GetUserByEmail(c, h.db, email)
	if err != nil {
		if err.Error() == "user not found" {

			logger.LogAuthEvent("login", "", false, logger.Fields{
				"email":     email,
				"reason":    "user not found",
				"client_ip": c.ClientIP(),
			})
			errors.RespondWithError(c, http.StatusNotFound, errors.ErrCodeNotFound,
				"User not found", nil)
			return
		}
		logger.Error("Failed to retrieve user for login", err, logger.Fields{
			"email":     email,
			"client_ip": c.ClientIP(),
		})
		errors.RespondWithError(c, http.StatusInternalServerError, errors.ErrCodeDatabaseError,
			"Failed to retrieve user", nil)
		return
	}

	accessToken, err := jwt.GenerateToken(
		user.ID.String(),
		user.ApplicationID.String(),
		"",
		h.config.JWT.Secret,
		h.config.JWT.AccessExpiry,
	)
	if err != nil {
		logger.Error("Failed to generate access token for login", err, logger.Fields{
			"user_id": user.ID.String(),
			"email":   email,
		})
		errors.RespondWithError(c, http.StatusInternalServerError, errors.ErrCodeInternalError,
			"Failed to generate access token", nil)
		return
	}

	refreshToken, err := jwt.GenerateToken(
		user.ID.String(),
		user.ApplicationID.String(),
		"",
		h.config.JWT.Secret,
		h.config.JWT.RefreshExpiry,
	)
	if err != nil {
		logger.Error("Failed to generate refresh token for login", err, logger.Fields{
			"user_id": user.ID.String(),
			"email":   email,
		})
		errors.RespondWithError(c, http.StatusInternalServerError, errors.ErrCodeInternalError,
			"Failed to generate refresh token", nil)
		return
	}

	logger.LogAuthEvent("login", user.ID.String(), true, logger.Fields{
		"email":     email,
		"client_ip": c.ClientIP(),
	})

	c.JSON(http.StatusOK, gin.H{
		"message":       "Login successful",
		"access_token":  accessToken,
		"refresh_token": refreshToken,
		"user":          user,
	})
}

func (h *AuthHandler) RefreshToken(c *gin.Context) {
	var req models.RefreshTokenRequest
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

	refreshToken := strings.TrimSpace(req.RefreshToken)
	if refreshToken == "" {
		details := []errors.ErrorDetail{
			{
				Field: "refresh_token",
				Issue: "Refresh token is required",
			},
		}
		errors.RespondWithError(c, http.StatusBadRequest, errors.ErrCodeValidationError,
			"Refresh token is required", details)
		return
	}

	claims, err := jwt.ValidateToken(refreshToken, h.config.JWT.Secret)
	if err != nil {
		errors.RespondWithError(c, http.StatusUnauthorized, errors.ErrCodeUnauthorized,
			"Invalid or expired refresh token", nil)
		return
	}

	userID := claims.UserID
	if !validator.ValidateUUID(userID) {
		errors.RespondWithError(c, http.StatusUnauthorized, errors.ErrCodeUnauthorized,
			"Invalid token: user ID format is invalid", nil)
		return
	}

	userIDUUID, err := uuid.Parse(userID)
	if err != nil {
		errors.RespondWithError(c, http.StatusUnauthorized, errors.ErrCodeUnauthorized,
			"Invalid token: user ID format is invalid", nil)
		return
	}

	query := `SELECT id, application_id, external_id, username, email, avatar_url, metadata, created_at, updated_at, deleted_at
	          FROM users
	          WHERE id = $1 AND deleted_at IS NULL`

	user := &models.User{}
	var metadataJSON []byte
	err = h.db.QueryRow(query, userIDUUID).Scan(
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
				"User associated with token no longer exists", nil)
			return
		}
		errors.RespondWithError(c, http.StatusInternalServerError, errors.ErrCodeDatabaseError,
			"Failed to retrieve user", nil)
		return
	}

	if len(metadataJSON) > 0 {
		if err := json.Unmarshal(metadataJSON, &user.Metadata); err != nil {
		}
	}

	accessToken, err := jwt.GenerateToken(
		user.ID.String(),
		user.ApplicationID.String(),
		"",
		h.config.JWT.Secret,
		h.config.JWT.AccessExpiry,
	)
	if err != nil {
		errors.RespondWithError(c, http.StatusInternalServerError, errors.ErrCodeInternalError,
			"Failed to generate access token", nil)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":      "Token refreshed successfully",
		"access_token": accessToken,
	})
}
