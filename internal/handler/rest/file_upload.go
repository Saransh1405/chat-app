package rest

import (
	"fmt"
	"net/http"

	"chat-app/internal/database"
	"chat-app/internal/models"
	"chat-app/internal/utils/errors"
	"chat-app/internal/utils/logger"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	imagekit "github.com/imagekit-developer/imagekit-go/v2"
)

type FileUploadHandler struct {
	db *database.DB
	ik imagekit.Client
}

func NewFileUploadHandler(db *database.DB, ik imagekit.Client) *FileUploadHandler {
	return &FileUploadHandler{
		db: db,
		ik: ik,
	}
}

func (h *FileUploadHandler) Upload(c *gin.Context) {
	file, header, err := c.Request.FormFile("file")
	if err != nil {
		if err == http.ErrMissingFile {
			errors.RespondWithError(c, http.StatusBadRequest, errors.ErrCodeValidationError,
				"File is required", []errors.ErrorDetail{
					{Field: "file", Issue: "No file provided in request"},
				})
			return
		}
		logger.Error("Failed to retrieve file from form", err, logger.Fields{
			"client_ip": c.ClientIP(),
		})
		errors.RespondWithError(c, http.StatusBadRequest, errors.ErrCodeValidationError,
			"Failed to retrieve file from request", nil)
		return
	}
	defer file.Close()

	if header.Size <= 0 {
		errors.RespondWithError(c, http.StatusBadRequest, errors.ErrCodeValidationError,
			"Invalid file size", []errors.ErrorDetail{
				{Field: "file", Issue: "File size must be greater than 0"},
			})
		return
	}

	var req models.FileUploadRequest
	if err := c.ShouldBind(&req); err != nil {
		details := []errors.ErrorDetail{
			{
				Field: "request_body",
				Issue: "Invalid form data format",
			},
		}
		errors.RespondWithError(c, http.StatusBadRequest, errors.ErrCodeValidationError,
			"Invalid request body", details)
		return
	}

	fileName := req.FileName
	if fileName == "" {
		fileName = header.Filename
	}

	folder := req.Folder
	if folder == "" {
		folder = "/uploads"
	}

	mimeType := header.Header.Get("Content-Type")
	if mimeType == "" {
		mimeType = "application/octet-stream"
	}

	uploadParams := imagekit.FileUploadParams{
		File:     file,
		FileName: fileName,
	}

	if folder != "" {
		uploadParams.Folder = imagekit.String(folder)
	}

	if len(req.Tags) > 0 {
		uploadParams.Tags = req.Tags
	}

	if req.CustomMetadata != nil {
		uploadParams.CustomMetadata = req.CustomMetadata
	}

	resp, err := h.ik.Files.Upload(
		c.Request.Context(),
		uploadParams,
	)
	if err != nil {
		logger.Error("Failed to upload file to ImageKit", err, logger.Fields{
			"filename": fileName,
			"folder":   folder,
			"size":     header.Size,
		})

		fmt.Printf("*****************--------------err: %v\n", err)
		errors.RespondWithError(c, http.StatusInternalServerError, errors.ErrCodeInternalError,
			"Failed to upload file", nil)
		return
	}

	logger.Info("File uploaded successfully", logger.Fields{
		"file_id":  resp.FileID,
		"filename": fileName,
		"url":      resp.URL,
		"size":     header.Size,
	})

	fileResponse := gin.H{
		"file_id":   resp.FileID,
		"url":       resp.URL,
		"filename":  fileName,
		"size":      header.Size,
		"mime_type": mimeType,
		"file_path": resp.URL,
	}

	if appID := c.Query("application_id"); appID != "" {
		appIDUUID, err := uuid.Parse(appID)
		if err == nil {
			fileResponse["application_id"] = appIDUUID
		}
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "File uploaded successfully",
		"data":    fileResponse,
		"tags":    req.Tags,
	})
}
