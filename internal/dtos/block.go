// Declare that this file belongs to the 'dtos' package
package dtos

// Import json package for handling dynamic arbitrary JSON block content
import (
	"encoding/json"
	"time"
)

// CreateBlockRequest defines incoming JSON payload when inserting a new block
type CreateBlockRequest struct {
	// Optional block ID (server generates UUID if omitted)
	ID string `json:"id"`

	// Parent document ID to attach block to (required field)
	DocumentID string `json:"document_id" binding:"required"`

	// Optional target block ID after which this block will be atomically inserted
	TargetAfterID string `json:"target_after_id"`

	// Arbitrary JSON payload representing rich block content (required field)
	Content json.RawMessage `json:"content" binding:"required"`
}

// UpdateBlockRequest defines incoming JSON payload when modifying block content
type UpdateBlockRequest struct {
	// Current version number sent by client to prevent race conditions (required field)
	Version int `json:"version" binding:"required"`

	// New arbitrary JSON block content payload (required field)
	Content json.RawMessage `json:"content" binding:"required"`
}

// BlockResponse defines outgoing JSON payload returned for block operations
type BlockResponse struct {
	// Unique block ID
	ID string `json:"id"`

	// Parent document ID
	DocumentID string `json:"document_id"`

	// Current block version number
	Version int `json:"version"`

	// Structured JSON content of the block
	Content json.RawMessage `json:"content"`

	// Timestamp when block was created
	CreatedAt time.Time `json:"created_at"`

	// Timestamp when block was last modified
	UpdatedAt time.Time `json:"updated_at"`
}