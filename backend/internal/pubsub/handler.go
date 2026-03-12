package pubsub

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"
)

// Handler handles incoming Pub/Sub push messages
type Handler struct {
	processor *Processor
	logger    *slog.Logger
}

// NewHandler creates a new Pub/Sub push handler
func NewHandler(processor *Processor, logger *slog.Logger) *Handler {
	return &Handler{
		processor: processor,
		logger:    logger,
	}
}

// HandlePush handles incoming Pub/Sub push messages
// POST /pubsub/bulk-operations
func (h *Handler) HandlePush(c *gin.Context) {
	body, err := io.ReadAll(c.Request.Body)
	if err != nil {
		h.logger.Error("Failed to read request body", "error", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "failed to read request body"})
		return
	}

	var pushMsg PubSubPushMessage
	if err := json.Unmarshal(body, &pushMsg); err != nil {
		h.logger.Error("Failed to parse push message", "error", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "failed to parse push message"})
		return
	}

	// Decode base64 message data
	data, err := base64.StdEncoding.DecodeString(pushMsg.Message.Data)
	if err != nil {
		h.logger.Error("Failed to decode message data", "error", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "failed to decode message data"})
		return
	}

	// Parse the bulk operation message
	var opMsg BulkOperationMessage
	if err := json.Unmarshal(data, &opMsg); err != nil {
		h.logger.Error("Failed to parse operation message", "error", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "failed to parse operation message"})
		return
	}

	h.logger.Info("Received bulk operation message",
		"job_id", opMsg.JobID,
		"operation_type", opMsg.OperationType,
		"message_id", pushMsg.Message.MessageID,
	)

	// Process the message
	ctx := c.Request.Context()
	if err := h.processor.Process(ctx, opMsg); err != nil {
		h.logger.Error("Failed to process bulk operation",
			"job_id", opMsg.JobID,
			"operation_type", opMsg.OperationType,
			"error", err,
		)
		// Return 500 to trigger retry
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("failed to process: %v", err)})
		return
	}

	h.logger.Info("Successfully processed bulk operation",
		"job_id", opMsg.JobID,
		"operation_type", opMsg.OperationType,
	)

	// Return 200 to acknowledge the message
	c.JSON(http.StatusOK, gin.H{"status": "processed"})
}
