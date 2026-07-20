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

// DocumentService defines high-level business operations for documents
type DocumentService interface {
	CreateDocument(ctx context.Context, req *dtos.CreateDocumentRequest) (*dtos.DocumentResponse, error)
	GetDocumentByID(ctx context.Context, id string) (*dtos.DocumentResponse, error)
	GetDocumentsByUserID(ctx context.Context, userID string) ([]*models.Document, error) 
	GetDocumentWithBlocks(ctx context.Context, id string) (*dtos.DocumentWithBlocksResponse, error)
	UpdateDocument(ctx context.Context, id string, req *dtos.UpdateDocumentRequest) (*dtos.DocumentResponse, error)
	DeleteDocument(ctx context.Context, id string) error
}

// documentService implements the DocumentService interface
type documentService struct {
	docRepo repository.DocumentRepository
	blockRepo repository.BlockRepository
}

// NewDocumentService constructs and returns a new documentService
func NewDocumentService(docRepo repository.DocumentRepository, blockRepo repository.BlockRepository) DocumentService {
	return &documentService{
		docRepo:   docRepo,
		blockRepo: blockRepo,
	}
}

// CreateDocument handles business logic for creating a new document
func (s *documentService) CreateDocument(ctx context.Context, req *dtos.CreateDocumentRequest) (*dtos.DocumentResponse, error) {
	docID := req.ID
	if docID == "" {
		docID = uuid.New().String()
	}

	// Use provided owner user ID or assign default system user
	userID := req.UserID
	if userID == "" {
		userID = "default_user"
	}

	// Capture current UTC timestamp for record creation
	now := time.Now().UTC()

	// Instantiate new database model object
	doc := &models.Document{
		ID:         docID,
		UserID:     userID,
		Title:      req.Title,
		BlockOrder: []string{}, // Initialize with an empty block sequence
		CreatedAt:  now,
		UpdatedAt:  now,
	}

	// Persist new document model into database via repository
	if err := s.docRepo.Create(ctx, doc); err != nil {
		return nil, fmt.Errorf("service failed to create document: %w", err)
	}

	// Map and return response DTO payload
	return &dtos.DocumentResponse{
		ID:         doc.ID,
		UserID:     doc.UserID,
		Title:      doc.Title,
		BlockOrder: doc.BlockOrder,
		CreatedAt:  doc.CreatedAt,
		UpdatedAt:  doc.UpdatedAt,
	}, nil
}

func (s *documentService) GetDocumentsByUserID(ctx context.Context, userID string) ([]*models.Document, error) {
	if userID == "" {
		// Fallback to default user if none provided in query param
		userID = "usr_123"
	}
	return s.docRepo.GetByUserID(ctx, userID)
}

func (s *documentService) GetDocumentByID(ctx context.Context, id string) (*dtos.DocumentResponse, error) {
	doc, err := s.docRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	return &dtos.DocumentResponse{
		ID:         doc.ID,
		UserID:     doc.UserID,
		Title:      doc.Title,
		BlockOrder: doc.BlockOrder,
		CreatedAt:  doc.CreatedAt,
		UpdatedAt:  doc.UpdatedAt,
	}, nil
}

// GetDocumentWithBlocks fetches complete document metadata along with all child blocks
func (s *documentService) GetDocumentWithBlocks(ctx context.Context, id string) (*dtos.DocumentWithBlocksResponse, error) {
	doc, err := s.docRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	// Fetch all blocks associated with this document ID
	blocks, err := s.blockRepo.GetByDocumentID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("service failed to fetch document blocks: %w", err)
	}

	// Index blocks by their unique block ID in a temporary lookup map
	blockMap := make(map[string]*models.Block)
	for _, b := range blocks {
		blockMap[b.ID] = b
	}

	// Order blocks strictly according to document's BlockOrder sequence array
	orderedBlocks := make([]dtos.BlockResponse, 0, len(blocks))
	for _, blockID := range doc.BlockOrder {
		if b, exists := blockMap[blockID]; exists {
			orderedBlocks = append(orderedBlocks, dtos.BlockResponse{
				ID:         b.ID,
				DocumentID: b.DocumentID,
				Version:    b.Version,
				Content:    b.Content,
				CreatedAt:  b.CreatedAt,
				UpdatedAt:  b.UpdatedAt,
			})
			// Delete processed block from map to track orphan blocks
			delete(blockMap, blockID)
		}
	}

	// Append any leftover blocks missing from block_order array to prevent data loss
	for _, b := range blockMap {
		orderedBlocks = append(orderedBlocks, dtos.BlockResponse{
			ID:         b.ID,
			DocumentID: b.DocumentID,
			Version:    b.Version,
			Content:    b.Content,
			CreatedAt:  b.CreatedAt,
			UpdatedAt:  b.UpdatedAt,
		})
	}

	// Assemble and return composite DocumentWithBlocks DTO payload
	return &dtos.DocumentWithBlocksResponse{
		Document: dtos.DocumentResponse{
			ID:         doc.ID,
			UserID:     doc.UserID,
			Title:      doc.Title,
			BlockOrder: doc.BlockOrder,
			CreatedAt:  doc.CreatedAt,
			UpdatedAt:  doc.UpdatedAt,
		},
		Blocks: orderedBlocks,
	}, nil
}

// UpdateDocument updates document metadata such as title
func (s *documentService) UpdateDocument(ctx context.Context, id string, req *dtos.UpdateDocumentRequest) (*dtos.DocumentResponse, error) {
	doc, err := s.docRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	doc.Title = req.Title
	doc.UpdatedAt = time.Now().UTC()

	if err := s.docRepo.Update(ctx, doc); err != nil {
		return nil, fmt.Errorf("service failed to update document: %w", err)
	}

	return &dtos.DocumentResponse{
		ID:         doc.ID,
		UserID:     doc.UserID,
		Title:      doc.Title,
		BlockOrder: doc.BlockOrder,
		CreatedAt:  doc.CreatedAt,
		UpdatedAt:  doc.UpdatedAt,
	}, nil
}


// DeleteDocument removes a document AND all its associated blocks
func (s *documentService) DeleteDocument(ctx context.Context, id string) error {
	// 1. Fetch all blocks belonging to the document
	blocks, err := s.blockRepo.GetByDocumentID(ctx, id)
	if err == nil {
		// 2. Loop through and delete each child block
		for _, b := range blocks {
			_ = s.blockRepo.Delete(ctx, b.ID)
		}
	}

	// 3. Delete the parent document record
	return s.docRepo.Delete(ctx, id)
}