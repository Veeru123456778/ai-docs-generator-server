// Declare that this file belongs to the 'service' package
package service

import (
	"context"
	"fmt"
	"time"

	"ai-docs-generator/internal/dtos"
	"ai-docs-generator/internal/models"
	"ai-docs-generator/internal/repository"

	"github.com/google/uuid"
)

// BlockService defines operations for managing blocks and sequence order
type BlockService interface {
	CreateBlock(ctx context.Context, req *dtos.CreateBlockRequest) (*dtos.BlockResponse, error)
	BatchCreateBlocks(ctx context.Context, req *dtos.BatchCreateBlocksRequest) (*dtos.BatchCreateBlocksResponse,error)
	GetBlockByID(ctx context.Context, id string) (*dtos.BlockResponse, error)
	UpdateBlock(ctx context.Context, id string, req *dtos.UpdateBlockRequest) (*dtos.BlockResponse, error)
	DeleteBlock(ctx context.Context, id string) error
}

// blockService implements BlockService interface
type blockService struct {
	blockRepo repository.BlockRepository
	docRepo repository.DocumentRepository
}

// NewBlockService initializes and returns a new blockService instance
func NewBlockService(blockRepo repository.BlockRepository, docRepo repository.DocumentRepository) BlockService {
	return &blockService{
		blockRepo: blockRepo,
		docRepo:  docRepo,
	}
}

// CreateBlock handles block insertion and atomic block order sequence recalculation
func (s *blockService) CreateBlock(ctx context.Context, req *dtos.CreateBlockRequest) (*dtos.BlockResponse, error) {
	doc, err := s.docRepo.GetByID(ctx, req.DocumentID)
	if err != nil {
		return nil, fmt.Errorf("cannot create block, parent document missing: %w", err)
	}

	blockID := req.ID
	if blockID == "" {
		blockID = uuid.New().String()
	}

	// Set initial block creation attributes
	now := time.Now().UTC()
	block := &models.Block{
		ID:         blockID,
		DocumentID: req.DocumentID,
		Version:    1, // Initial block version starts at 1
		Content:    req.Content,
		CreatedAt:  now,
		UpdatedAt:  now,
	}

	// Insert block model into blocks table
	if err := s.blockRepo.Create(ctx, block); err != nil {
		return nil, fmt.Errorf("service failed to insert block: %w", err)
	}

	// Calculate new block order array inserting new block after target ID if requested
	newOrder := insertBlockID(doc.BlockOrder, blockID, req.TargetAfterID)

	// Update document's block_order column in database
	if err := s.docRepo.UpdateBlockOrder(ctx, doc.ID, newOrder); err != nil {
		return nil, fmt.Errorf("service failed to update parent document block order: %w", err)
	}

	// Return created block response DTO
	return &dtos.BlockResponse{
		ID:         block.ID,
		DocumentID: block.DocumentID,
		Version:    block.Version,
		Content:    block.Content,
		CreatedAt:  block.CreatedAt,
		UpdatedAt:  block.UpdatedAt,
	}, nil
}


// BatchCreateBlocks handles batch creation of blocks
func (s *blockService) BatchCreateBlocks(ctx context.Context, req *dtos.BatchCreateBlocksRequest) (*dtos.BatchCreateBlocksResponse, error) {
    now := time.Now().UTC()
    var blocks []*models.Block
    for _, blockReq := range req.Blocks {
        blockID := blockReq.ID
        if blockID == "" {
            blockID = uuid.New().String()
        }
        blocks = append(blocks, &models.Block{
            ID:         blockID,
            DocumentID: blockReq.DocumentID,
            Version:    1,
            Content:    blockReq.Content,
            CreatedAt:  now,
            UpdatedAt:  now,
        })
    }

    if err := s.blockRepo.BatchCreate(ctx, blocks); err != nil {
        return nil, fmt.Errorf("service failed to batch create blocks: %w", err)
    }

    var blockResponses []dtos.BlockResponse
    for _, block := range blocks {
        blockResponses = append(blockResponses, dtos.BlockResponse{
            ID:         block.ID,
            DocumentID: block.DocumentID,
            Version:    block.Version,
            Content:    block.Content,
            CreatedAt:  block.CreatedAt,
            UpdatedAt:  block.UpdatedAt,
        })
    }

    return &dtos.BatchCreateBlocksResponse{Blocks: blockResponses}, nil
}


// GetBlockByID fetches a single block DTO by ID
func (s *blockService) GetBlockByID(ctx context.Context, id string) (*dtos.BlockResponse, error) {
	block, err := s.blockRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	return &dtos.BlockResponse{
		ID:         block.ID,
		DocumentID: block.DocumentID,
		Version:    block.Version,
		Content:    block.Content,
		CreatedAt:  block.CreatedAt,
		UpdatedAt:  block.UpdatedAt,
	}, nil
}

// UpdateBlock updates block content using optimistic concurrency lock
func (s *blockService) UpdateBlock(ctx context.Context, id string, req *dtos.UpdateBlockRequest) (*dtos.BlockResponse, error) {
	block := &models.Block{
		ID:        id,
		Version:   req.Version,
		Content:   req.Content,
		UpdatedAt: time.Now().UTC(),
	}

	// Attempt atomic update in database repository
	if err := s.blockRepo.Update(ctx, block); err != nil {
		return nil, err
	}

	// Fetch updated block record to retrieve new incremented version number
	updatedBlock, err := s.blockRepo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch updated block details: %w", err)
	}

	// Return updated block DTO
	return &dtos.BlockResponse{
		ID:         updatedBlock.ID,
		DocumentID: updatedBlock.DocumentID,
		Version:    updatedBlock.Version,
		Content:    updatedBlock.Content,
		CreatedAt:  updatedBlock.CreatedAt,
		UpdatedAt:  updatedBlock.UpdatedAt,
	}, nil
}

// DeleteBlock removes a block and removes its ID from the parent document's block order
func (s *blockService) DeleteBlock(ctx context.Context, id string) error {
	block, err := s.blockRepo.GetByID(ctx, id)
	if err != nil {
		return err
	}

	if err := s.blockRepo.Delete(ctx, id); err != nil {
		return fmt.Errorf("service failed to delete block: %w", err)
	}

	// Fetch parent document to update its sequence order
	doc, err := s.docRepo.GetByID(ctx, block.DocumentID)
	if err == nil {
		// Filter out deleted block ID from block order slice
		newOrder := removeBlockID(doc.BlockOrder, id)
		// Update document's block order sequence in database
		_ = s.docRepo.UpdateBlockOrder(ctx, doc.ID, newOrder)
	}

	// Return nil indicating successful deletion
	return nil
}

// Helper function to insert a block ID after a target ID in a slice
func insertBlockID(order []string, newID string, targetAfterID string) []string {
	if targetAfterID == "" {
		return append(order, newID)
	}

	// Find index position of targetAfterID
	targetIdx := -1
	for i, id := range order {
		if id == targetAfterID {
			targetIdx = i
			break
		}
	}

	// If target ID was not found in array, append new ID to end of sequence
	if targetIdx == -1 {
		return append(order, newID)
	}

	// Construct new slice with inserted block ID placed directly after target index
	result := make([]string, 0, len(order)+1)
	result = append(result, order[:targetIdx+1]...)
	result = append(result, newID)
	result = append(result, order[targetIdx+1:]...)
	return result
}

// Helper function to remove a block ID from a slice
func removeBlockID(order []string, targetID string) []string {
	// Initialize slice to store remaining block IDs
	result := make([]string, 0, len(order))
	for _, id := range order {
		// Keep ID if it does not match deleted block ID
		if id != targetID {
			result = append(result, id)
		}
	}
	return result
}