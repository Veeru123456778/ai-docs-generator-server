// Declare that this file belongs to the 'repository' package
package repository

import (
	"context"
	"errors"
	"fmt"

	"ai-docs-generator/internal/models"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Standardized error returned when a block is not found
var ErrBlockNotFound = errors.New("block not found")

// Standardized error returned when version tag does not match (concurrency collision)
var ErrVersionConflict = errors.New("optimistic concurrency conflict: block version mismatch")

// BlockRepository defines database interactions for individual document blocks
type BlockRepository interface {
	// Inserts a new block record into database
	Create(ctx context.Context, block *models.Block) error

	BatchCreate(ctx context.Context, blocks []*models.Block) error

	// Retrieves a single block by block ID
	GetByID(ctx context.Context, id string) (*models.Block, error)
	// Retrieves all blocks matching a parent document ID
	GetByDocumentID(ctx context.Context, documentID string) ([]*models.Block, error)
	// Updates a block using optimistic locking version check
	Update(ctx context.Context, block *models.Block) error
	// Deletes a single block by ID
	Delete(ctx context.Context, id string) error
}

// PostgresBlockRepository implements BlockRepository interface
type PostgresBlockRepository struct {
	// Holds reference to postgres database connection pool
	pool *pgxpool.Pool
}

// NewPostgresBlockRepository initializes and returns a PostgresBlockRepository
func NewPostgresBlockRepository(pool *pgxpool.Pool) *PostgresBlockRepository {
	// Construct and return repository struct pointer
	return &PostgresBlockRepository{pool: pool}
}

// Create inserts a new block into the blocks table
func (r *PostgresBlockRepository) Create(ctx context.Context, block *models.Block) error {
	// Define SQL INSERT statement for block creation
	query := `
		INSERT INTO blocks (id, document_id,type, version, content, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
	`
	// Execute SQL INSERT passing model struct values
	_, err := r.pool.Exec(ctx, query,
		block.ID,
		block.DocumentID,
		block.Type,
		block.Version,
		block.Content,
		block.CreatedAt,
		block.UpdatedAt,
	)
	// Check if database insertion query failed
	if err != nil {
		// Return wrapped error message
		return fmt.Errorf("failed to insert block: %w", err)
	}
	// Return nil indicating successful block insertion
	return nil
}

// BatchCreate inserts multiple blocks into the blocks table
func (r *PostgresBlockRepository) BatchCreate(ctx context.Context, blocks []*models.Block) error {
	query := `
        INSERT INTO blocks (id,document_id, type, version, content, created_at, updated_at)
        VALUES ($1, $2, $3, $4, $5, $6, $7)
    `
	batch := &pgx.Batch{}

	for _, block := range blocks {
		batch.Queue(query, block.ID, block.DocumentID, block.Type, block.Version, block.Content, block.CreatedAt, block.UpdatedAt)
	}

	br := r.pool.SendBatch(ctx, batch)
	defer br.Close()

	for range blocks {
		if _, err := br.Exec(); err != nil {
			return fmt.Errorf("failed to batch insert blocks: %w", err)
		}
	}
	return nil
}

// GetByID fetches a single block record by ID
func (r *PostgresBlockRepository) GetByID(ctx context.Context, id string) (*models.Block, error) {
	// Define SQL SELECT statement targeting single block by ID
	query := `
		SELECT id, document_id, type, version, content, created_at, updated_at
		FROM blocks
		WHERE id = $1
	`
	// Prepare empty Block struct variable
	var block models.Block

	// Execute single row query and scan values into block fields
	err := r.pool.QueryRow(ctx, query, id).Scan(
		&block.ID,
		&block.DocumentID,
		&block.Type,
		&block.Version,
		&block.Content,
		&block.CreatedAt,
		&block.UpdatedAt,
	)
	// Check if query returned error
	if err != nil {
		// Check if error matches no rows found condition
		if errors.Is(err, pgx.ErrNoRows) {
			// Return custom block not found error
			return nil, ErrBlockNotFound
		}
		// Return wrapped error context
		return nil, fmt.Errorf("failed to query block by id: %w", err)
	}

	// Return pointer to populated Block model
	return &block, nil
}

// GetByDocumentID fetches all blocks associated with a document
func (r *PostgresBlockRepository) GetByDocumentID(ctx context.Context, documentID string) ([]*models.Block, error) {
	// Query to select all blocks matching document_id
	query := `
		SELECT id, document_id,type, version, content, created_at, updated_at
		FROM blocks
		WHERE document_id = $1
	`
	// Execute multi-row SQL query
	rows, err := r.pool.Query(ctx, query, documentID)
	// Check if executing query failed
	if err != nil {
		// Return wrapped error message
		return nil, fmt.Errorf("failed to query blocks by document id: %w", err)
	}
	// Guarantee SQL query rows reader is closed when function finishes
	defer rows.Close()

	// Prepare slice to hold all scanned block pointers
	var blocks []*models.Block

	// Loop through each returned database row
	for rows.Next() {
		// Instantiate new empty Block struct for current iteration
		var block models.Block
		// Scan row column values into current block struct
		if err := rows.Scan(
			&block.ID,
			&block.DocumentID,
			&block.Type,
			&block.Version,
			&block.Content,
			&block.CreatedAt,
			&block.UpdatedAt,
		); err != nil {
			// Return wrapped error if row scanning fails
			return nil, fmt.Errorf("failed to scan block row: %w", err)
		}
		// Append populated block pointer to slice
		blocks = append(blocks, &block)
	}

	// Check if row iteration encountered any errors
	if err := rows.Err(); err != nil {
		// Return wrapped iteration error
		return nil, fmt.Errorf("error during blocks row iteration: %w", err)
	}

	// Return full slice of fetched block pointers
	return blocks, nil
}

// Update updates block content and increments version using optimistic locking
func (r *PostgresBlockRepository) Update(ctx context.Context, block *models.Block) error {
	// Define UPDATE query enforcing matching version for optimistic concurrency control
	query := `
		UPDATE blocks
		SET content = $1, version = version + 1, updated_at = $2
		WHERE id = $3 AND version = $4
	`
	// Execute query using current block.Version parameter for atomic comparison
	result, err := r.pool.Exec(ctx, query, block.Content, block.UpdatedAt, block.ID, block.Version)
	// Check if database statement execution failed
	if err != nil {
		// Return wrapped update error
		return fmt.Errorf("failed to update block: %w", err)
	}
	// Check if zero rows were affected (indicates either ID missing or Version changed concurrently)
	if result.RowsAffected() == 0 {
		// Check if the block even exists in database
		_, fetchErr := r.GetByID(ctx, block.ID)
		if errors.Is(fetchErr, ErrBlockNotFound) {
			// Return block not found error if ID is missing
			return ErrBlockNotFound
		}
		// Return version conflict error if ID exists but version did not match
		return ErrVersionConflict
	}
	// Return nil on successful block update
	return nil
}

// Delete removes a block from the database
func (r *PostgresBlockRepository) Delete(ctx context.Context, id string) error {
	// Define SQL DELETE query statement
	query := `DELETE FROM blocks WHERE id = $1`

	// Execute delete statement
	result, err := r.pool.Exec(ctx, query, id)
	// Check if query execution failed
	if err != nil {
		// Return wrapped error context
		return fmt.Errorf("failed to delete block: %w", err)
	}
	// Check if zero rows were modified
	if result.RowsAffected() == 0 {
		// Return block not found error if ID didn't match existing row
		return ErrBlockNotFound
	}
	// Return nil indicating successful deletion
	return nil
}
