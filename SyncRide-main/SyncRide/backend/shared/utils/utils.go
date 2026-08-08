package utils

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"net/http"
	"time"

	"github.com/google/uuid"
)

// GenerateUUID generates a random RFC 4122 UUID v4
func GenerateUUID() string {
	return uuid.New().String()
}

// GenerateTraceID generates a 32-character hex trace ID
func GenerateTraceID() string {
	bytes := make([]byte, 16)
	_, _ = rand.Read(bytes)
	return hex.EncodeToString(bytes)
}

// Standard Response Envelope
type APIResponse struct {
	Success   bool        `json:"success"`
	Data      interface{} `json:"data,omitempty"`
	Message   string      `json:"message,omitempty"`
	Timestamp string      `json:"timestamp"`
	Meta      interface{} `json:"meta,omitempty"`
}

// WriteJSON writes a standardized JSON success response
func WriteJSON(w http.ResponseWriter, statusCode int, data interface{}, message string, meta ...interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	var metaData interface{}
	if len(meta) > 0 {
		metaData = meta[0]
	}

	response := APIResponse{
		Success:   statusCode >= 200 && statusCode < 300,
		Data:      data,
		Message:   message,
		Timestamp: time.Now().UTC().Format(time.RFC3339),
		Meta:      metaData,
	}

	_ = json.NewEncoder(w).Encode(response)
}
