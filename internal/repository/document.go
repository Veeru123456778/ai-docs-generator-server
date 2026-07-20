// Declare that this file belongs to the 'repository' package
package repository

import (
	// Import context package for query deadlines and cancellation
	"context"
	// Import errors package to return standardized database errors
	"errors"
	// Import fmt package for error formatting
	"fmt"

	// Import models package using 'ai-docs-generator' module path
	"ai-docs-generator/internal/models"

	// Import pgx/v5 package for low-level postgres error checking
	"github.com/jackc/pgx/v5"
	// Import pgxpool package for managing connection pool
	"github.com/jackc/pgx/v5/pgxpool"
)

// Standardized error returned when a requested document does not exist
var ErrDocumentNotFound = errors.New("document not found")

// DocumentRepository defines the contract for document database operations
type DocumentRepository interface {
	// Creates a new document record in the database
	Create(ctx context.Context, doc *models.Document) error
	// Retrieves a single document by its unique ID
	GetByID(ctx context.Context, id string) (*models.Document, error)
	GetByUserID(ctx context.Context, userID string) ([]*models.Document, error) 
	// Updates title and block order for an existing document
	Update(ctx context.Context, doc *models.Document) error
	// Deletes a document and its associated blocks
	Delete(ctx context.Context, id string) error
	// Updates only the block_order array for a document
	UpdateBlockOrder(ctx context.Context, id string, blockOrder []string) error
}

// PostgresDocumentRepository implements DocumentRepository using pgxpool
type PostgresDocumentRepository struct {
	// Database connection pool instance
	pool *pgxpool.Pool
}

// NewPostgresDocumentRepository constructs a new PostgresDocumentRepository
func NewPostgresDocumentRepository(pool *pgxpool.Pool) *PostgresDocumentRepository {
	// Return initialized repository struct containing connection pool
	return &PostgresDocumentRepository{pool: pool}
}

// Create inserts a new document row into the database
func (r *PostgresDocumentRepository) Create(ctx context.Context, doc *models.Document) error {
	// Define SQL INSERT query targeting documents table
	query := `
		INSERT INTO documents (id, user_id, title, block_order, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6)
	`
	// Execute INSERT query with parameters from document model struct
	_, err := r.pool.Exec(ctx, query,
		doc.ID,
		doc.UserID,
		doc.Title,
		doc.BlockOrder,
		doc.CreatedAt,
		doc.UpdatedAt,
	)
	// Check if database execution failed
	if err != nil {
		// Return wrapped error with operational context
		return fmt.Errorf("failed to insert document: %w", err)
	}
	// Return nil indicating successful insertion
	return nil
}



func (r *PostgresDocumentRepository) GetByUserID(ctx context.Context, userID string) ([]*models.Document, error) {
	query := `
		SELECT id, user_id, title, block_order, created_at, updated_at
		FROM documents
		WHERE user_id = $1
		ORDER BY updated_at DESC
	`

	// Use r.pool.Query instead of r.db.QueryContext
	rows, err := r.pool.Query(ctx, query, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to query documents by user_id: %w", err)
	}
	defer rows.Close()

	var docs []*models.Document
	for rows.Next() {
		var doc models.Document
		err := rows.Scan(
			&doc.ID,
			&doc.UserID,
			&doc.Title,
			&doc.BlockOrder,
			&doc.CreatedAt,
			&doc.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan document row: %w", err)
		}
		docs = append(docs, &doc)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error during row iteration: %w", err)
	}

	return docs, nil
}


// GetByID retrieves a document by ID from the database
func (r *PostgresDocumentRepository) GetByID(ctx context.Context, id string) (*models.Document, error) {
	// Define SQL SELECT query targeting single document row
	query := `
		SELECT id, user_id, title, block_order, created_at, updated_at
		FROM documents
		WHERE id = $1
	`
	// Instantiate empty Document struct to receive queried columns
	var doc models.Document

	// Execute query row and scan column values directly into struct fields
	err := r.pool.QueryRow(ctx, query, id).Scan(
		&doc.ID,
		&doc.UserID,
		&doc.Title,
		&doc.BlockOrder,
		&doc.CreatedAt,
		&doc.UpdatedAt,
	)
	// Check if query failed or row was not found
	if err != nil {
		// Check if error is specifically pgx.ErrNoRows
		if errors.Is(err, pgx.ErrNoRows) {
			// Return custom sentinel ErrDocumentNotFound error
			return nil, ErrDocumentNotFound
		}
		// Return wrapped query execution error
		return nil, fmt.Errorf("failed to query document by id: %w", err)
	}

	// Return pointer to populated Document model
	return &doc, nil
}

// Update updates an existing document record's title, block_order, and updated_at
func (r *PostgresDocumentRepository) Update(ctx context.Context, doc *models.Document) error {
	// Define SQL UPDATE query modifying matching document record
	query := `
		UPDATE documents
		SET title = $1, block_order = $2, updated_at = $3
		WHERE id = $4
	`
	// Execute SQL update statement passing updated values
	result, err := r.pool.Exec(ctx, query, doc.Title, doc.BlockOrder, doc.UpdatedAt, doc.ID)
	// Check if SQL statement execution failed
	if err != nil {
		// Return wrapped execution error
		return fmt.Errorf("failed to update document: %w", err)
	}
	// Check if any database rows were actually modified by update
	if result.RowsAffected() == 0 {
		// Return custom ErrDocumentNotFound if document ID didn't exist
		return ErrDocumentNotFound
	}
	// Return nil indicating successful update
	return nil
}

// Delete removes a document record from the database
func (r *PostgresDocumentRepository) Delete(ctx context.Context, id string) error {
	// Define SQL DELETE query targeting specific document ID
	query := `DELETE FROM documents WHERE id = $1`

	// Execute SQL delete query
	result, err := r.pool.Exec(ctx, query, id)
	// Check if execution failed
	if err != nil {
		// Return wrapped error message
		return fmt.Errorf("failed to delete document: %w", err)
	}
	// Verify that at least one row was deleted
	if result.RowsAffected() == 0 {
		// Return not found error if ID matched no existing rows
		return ErrDocumentNotFound
	}
	// Return nil for successful deletion
	return nil
}

// UpdateBlockOrder updates only the block_order array column for a document
func (r *PostgresDocumentRepository) UpdateBlockOrder(ctx context.Context, id string, blockOrder []string) error {
	// Define SQL query targeting block_order column update
	query := `
		UPDATE documents
		SET block_order = $1, updated_at = NOW()
		WHERE id = $2
	`
	// Execute block_order update query
	result, err := r.pool.Exec(ctx, query, blockOrder, id)
	// Check for SQL query errors
	if err != nil {
		// Return formatted wrapped error
		return fmt.Errorf("failed to update block order: %w", err)
	}
	// Check if document was found to be updated
	if result.RowsAffected() == 0 {
		// Return custom document not found error
		return ErrDocumentNotFound
	}
	// Return nil on successful execution
	return nil
}