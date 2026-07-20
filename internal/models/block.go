// Declare that this file belongs to the 'models' package
package models

// Import time package for handling database timestamp columns
import "time"

// Block represents a single row in the 'blocks' database table
type Block struct {
	// Unique identifier for the block (UUID string)
	ID string `json:"id" db:"id"`

	// Foreign key referencing the parent document ID
	DocumentID string `json:"document_id" db:"document_id"`

	// Version number used for optimistic concurrency control during updates
	Version int `json:"version" db:"version"`

	// Flexible JSON structure storing raw text, headings, or AI block metadata
	Content []byte `json:"content" db:"content"`

	// Timestamp when the block record was created in the database
	CreatedAt time.Time `json:"created_at" db:"created_at"`

	// Timestamp when the block record was last updated in the database
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
}