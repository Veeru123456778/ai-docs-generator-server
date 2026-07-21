// Declare that this file belongs to the 'dtos' package
package dtos

// Import time package for formatted timestamp outputs
import "time"

// CreateDocumentRequest defines incoming JSON payload when creating a new document
type CreateDocumentRequest struct {
	// Optional custom document ID provided by client (server generates UUID if empty)
	ID string `json:"id"`

	// Title of the new document (required field validated by Gin binding)
	Title string `json:"title" binding:"required"`

	// Optional user ID owning the document (defaults to system default if empty)
	UserID string `json:"user_id"`
}

// UpdateDocumentRequest defines incoming JSON payload when updating document metadata
type UpdateDocumentRequest struct {
	// New title to assign to the document (required field)
	Title string `json:"title" binding:"required"`
}

// DocumentResponse defines outgoing JSON payload returned to API clients
type DocumentResponse struct {
	// Unique document ID
	ID string `json:"id"`

	// Owner user ID
	UserID string `json:"user_id"`
    
	// Direct URL to view/edit the document in the Next.js frontend
	URL string `json:"url"`

	// Document title
	Title string `json:"title"`

	// Ordered list of block IDs representing document layout structure
	BlockOrder []string `json:"block_order"`

	// Timestamp when document was created
	CreatedAt time.Time `json:"created_at"`

	// Timestamp when document was last modified
	UpdatedAt time.Time `json:"updated_at"`
}

// DocumentWithBlocksResponse delivers the complete document along with all child blocks in order
type DocumentWithBlocksResponse struct {
	// The core document metadata and block ID sequence
	Document DocumentResponse `json:"document"`

	// List of all hydrated block objects belonging to the document
	Blocks []BlockResponse `json:"blocks"`
}