package handlers

import (
	"bytes"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"path/filepath"
	"strings"

	"github.com/bcc-media/wayfarer/internal/database"
	"github.com/bcc-media/wayfarer/internal/database/sqlc"
	"github.com/bcc-media/wayfarer/internal/services"
	"github.com/bcc-media/wayfarer/internal/ulid"
	"github.com/gin-gonic/gin"
)

const (
	maxFileSize = 30 * 1024 * 1024 // 30MB in bytes
)

// validMimeTypes maps MIME types to their allowed extensions
var validMimeTypes = map[string][]string{
	"image/jpeg":      {".jpg", ".jpeg"},
	"image/png":       {".png"},
	"image/gif":       {".gif"},
	"image/webp":      {".webp"},
	"video/mp4":       {".mp4"},
	"video/webm":      {".webm"},
	"video/quicktime": {".mov"},
	"application/pdf": {".pdf"},
}

// UploadHandler handles file upload requests
type UploadHandler struct {
	DB        *database.DB
	S3Service *services.S3Service
}

type uploadResponse struct {
	ID             string  `json:"id"`
	Filename       string  `json:"filename"`
	StoredFilename string  `json:"storedFilename"`
	FileSize       int     `json:"fileSize"`
	MimeType       string  `json:"mimeType"`
	PublicURL      string  `json:"publicUrl"`
	UploadedBy     string  `json:"uploadedBy"`
	CreatedAt      string  `json:"createdAt"`
	Width          *int32  `json:"width,omitempty"`
	Height         *int32  `json:"height,omitempty"`
	Blurhash       *string `json:"blurhash,omitempty"`
}

// HandleFileUpload handles the file upload endpoint
func (h *UploadHandler) HandleFileUpload(c *gin.Context) {
	ctx := c.Request.Context()

	// Get user ID from context (set by JWT middleware)
	userIDValue, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "authentication required"})
		return
	}
	userID, ok := userIDValue.(string)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid user ID"})
		return
	}

	// Check user roles (must be admin, superadmin, project_admin, or church_admin)
	rolesValue, exists := c.Get("user_roles")
	if !exists {
		c.JSON(http.StatusForbidden, gin.H{"error": "insufficient permissions"})
		return
	}
	roles, ok := rolesValue.([]string)
	if !ok {
		c.JSON(http.StatusForbidden, gin.H{"error": "invalid roles"})
		return
	}

	hasAdminRole := false
	for _, role := range roles {
		if role == "admin" || role == "superadmin" || role == "project_admin" || role == "church_admin" {
			hasAdminRole = true
			break
		}
	}
	if !hasAdminRole {
		c.JSON(http.StatusForbidden, gin.H{"error": "admin, superadmin, project_admin, or church_admin role required"})
		return
	}

	// Parse multipart form with size limit
	if err := c.Request.ParseMultipartForm(maxFileSize); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "failed to parse multipart form"})
		return
	}

	// Get file from form
	file, header, err := c.Request.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "file field is required"})
		return
	}
	defer file.Close()

	// Check file size
	if header.Size > maxFileSize {
		c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("file size exceeds maximum of %d bytes", maxFileSize)})
		return
	}

	// Read file into buffer to detect MIME type
	buffer := &bytes.Buffer{}
	fileSize, err := io.Copy(buffer, file)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to read file"})
		return
	}

	// Detect MIME type from file content
	mimeType := http.DetectContentType(buffer.Bytes())

	// Validate MIME type
	allowedExts, validMime := validMimeTypes[mimeType]
	if !validMime {
		c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("unsupported file type: %s", mimeType)})
		return
	}

	// Validate file extension
	ext := strings.ToLower(filepath.Ext(header.Filename))
	validExt := false
	for _, allowedExt := range allowedExts {
		if ext == allowedExt {
			validExt = true
			break
		}
	}
	if !validExt {
		c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("file extension %s does not match MIME type %s", ext, mimeType)})
		return
	}

	// Generate ULID filename with original extension
	fileID := ulid.NewFileUploadID()
	storedFilename := fileID + ext

	// Upload to S3
	publicURL, err := h.S3Service.UploadFile(ctx, bytes.NewReader(buffer.Bytes()), storedFilename, mimeType, fileSize)
	if err != nil {
		slog.Error("Failed to upload file to S3", "error", err, "filename", storedFilename, "size", fileSize)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to upload file to S3"})
		return
	}

	// Process image metadata if this is an image
	var width, height *int32
	var blurhash *string
	if services.IsImageMimeType(mimeType) {
		imgMeta, err := services.ProcessImage(buffer.Bytes(), mimeType)
		if err != nil {
			slog.Warn("Failed to process image metadata", "error", err, "filename", storedFilename)
			// Continue without image metadata - not a fatal error
		} else {
			w := int32(imgMeta.Width)
			h := int32(imgMeta.Height)
			width = &w
			height = &h
			blurhash = &imgMeta.Blurhash
		}
	}

	// Insert record to database
	uploadRecord, err := h.DB.Queries.CreateFileUpload(ctx, sqlc.CreateFileUploadParams{
		ID:             fileID,
		Filename:       header.Filename,
		StoredFilename: storedFilename,
		FileSize:       int32(fileSize),
		MimeType:       mimeType,
		PublicUrl:      publicURL,
		UploadedBy:     userID,
		Width:          width,
		Height:         height,
		Blurhash:       blurhash,
	})
	if err != nil {
		slog.Error("Failed to save file record to database", "error", err, "fileID", fileID)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to save file record"})
		return
	}

	// Return success response
	c.JSON(http.StatusOK, uploadResponse{
		ID:             uploadRecord.ID,
		Filename:       uploadRecord.Filename,
		StoredFilename: uploadRecord.StoredFilename,
		FileSize:       int(uploadRecord.FileSize),
		MimeType:       uploadRecord.MimeType,
		PublicURL:      uploadRecord.PublicUrl,
		UploadedBy:     uploadRecord.UploadedBy,
		CreatedAt:      uploadRecord.CreatedAt.Time.Format("2006-01-02T15:04:05Z07:00"),
		Width:          uploadRecord.Width,
		Height:         uploadRecord.Height,
		Blurhash:       uploadRecord.Blurhash,
	})
}
