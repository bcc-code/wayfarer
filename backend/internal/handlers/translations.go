package handlers

import (
	"io"
	"net/http"
	"strings"

	"github.com/bcc-media/wayfarer/translations"
	"github.com/gin-gonic/gin"
)

// TranslationsHandler handles translation export and webhook endpoints
type TranslationsHandler struct {
	service   *translations.Service
	ExportKey string // Static key for export endpoint authentication (like SSFHandler.SyncKey)
}

// NewTranslationsHandler creates a new translations handler
func NewTranslationsHandler(service *translations.Service, exportKey string) *TranslationsHandler {
	return &TranslationsHandler{
		service:   service,
		ExportKey: exportKey,
	}
}

// HandleWebhook processes Phrase translation webhooks
func (h *TranslationsHandler) HandleWebhook(c *gin.Context) {
	// Read webhook payload
	body, err := io.ReadAll(c.Request.Body)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "failed to read request body"})
		return
	}

	// Process webhook using provider
	collection, data, err := h.service.ProcessWebhook(c.Request.Context(), c.Request, body)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if collection == nil {
		// Webhook was not for our project or was ignored
		c.JSON(http.StatusOK, gin.H{"status": "ignored"})
		return
	}

	// Update translations in database
	errs := h.service.UpdateTranslations(c.Request.Context(), collection, data)
	if len(errs) > 0 {
		errorMessages := make([]string, len(errs))
		for i, e := range errs {
			errorMessages[i] = e.Error()
		}
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "failed to update some translations",
			"details": errorMessages,
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":     "success",
		"collection": collection.Value,
		"count":      len(data),
	})
}

// HandleExport exports a specific collection to Phrase
func (h *TranslationsHandler) HandleExport(c *gin.Context) {
	// Validate export key
	key := c.GetHeader("X-Export-Key")
	if key == "" {
		auth := c.GetHeader("Authorization")
		if strings.HasPrefix(auth, "Bearer ") {
			key = strings.TrimPrefix(auth, "Bearer ")
		}
	}
	if key == "" || key != h.ExportKey {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid export key"})
		return
	}

	collectionName := c.Param("collection")
	if collectionName == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "collection parameter required"})
		return
	}

	// Validate collection
	collection := translations.TranslatableCollection{Value: collectionName}
	if !translations.TranslatableCollections.Contains(collection) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid collection"})
		return
	}

	// Send to translation
	err := h.service.SendCollectionToTranslation(c.Request.Context(), collection)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":     "success",
		"collection": collectionName,
	})
}

// HandleExportAll exports all collections to Phrase
func (h *TranslationsHandler) HandleExportAll(c *gin.Context) {
	// Validate export key
	key := c.GetHeader("X-Export-Key")
	if key == "" {
		auth := c.GetHeader("Authorization")
		if strings.HasPrefix(auth, "Bearer ") {
			key = strings.TrimPrefix(auth, "Bearer ")
		}
	}
	if key == "" || key != h.ExportKey {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid export key"})
		return
	}

	// Send all collections to translation
	errs := h.service.SendAllToTranslation(c.Request.Context())
	if len(errs) > 0 {
		errorMessages := make([]string, len(errs))
		for i, e := range errs {
			errorMessages[i] = e.Error()
		}
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "failed to export some collections",
			"details": errorMessages,
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":  "success",
		"message": "all collections exported",
		"count":   len(translations.TranslatableCollections.Members()),
	})
}
