// Declare that this file belongs to the 'models' package
package models

// Import time package for handling database timestamp columns
import "time"

// Document represents a single row in the 'documents' database table
type Document struct {
	// Unique identifier for the document (UUID string)
	ID string `json:"id" db:"id"`

	// ID of the user who owns this document
	UserID string `json:"user_id" db:"user_id"`

	// Title of the document
	Title string `json:"title" db:"title"`

	// Ordered array of block IDs defining the document layout sequence
	BlockOrder []string `json:"block_order" db:"block_order"`

	// Timestamp when the document record was created in the database
	CreatedAt time.Time `json:"created_at" db:"created_at"`

	// Timestamp when the document record was last updated in the database
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
}