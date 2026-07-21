package controller

import (
	"errors"
	"net/http"

	"ai-docs-generator/internal/dtos"
	"ai-docs-generator/internal/repository"
	"ai-docs-generator/internal/service"

	"github.com/gin-gonic/gin"
)

type BlockController struct {
	blockService service.BlockService
}

func NewBlockController(blockService service.BlockService) *BlockController {
	return &BlockController{blockService: blockService}
}

func (ctrl *BlockController) Create(c *gin.Context) {
	var req dtos.CreateBlockRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	res, err := ctrl.blockService.CreateBlock(c.Request.Context(), &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, res)
}

// BatchCreate handles batch creation of blocks
func (ctrl *BlockController) BatchCreate(c *gin.Context) {
	var req dtos.BatchCreateBlocksRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body: " + err.Error()})
		return
	}

	// Sanity check: Ensure at least one block was supplied
	if len(req.Blocks) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "blocks array cannot be empty"})
		return
	}

	res, err := ctrl.blockService.BatchCreateBlocks(c.Request.Context(), &req)
	if err != nil {
		// Handle non-existent parent document explicitly
		if errors.Is(err, service.ErrDocumentNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, res)
}

func (ctrl *BlockController) GetByID(c *gin.Context) {
	id := c.Param("id")

	res, err := ctrl.blockService.GetBlockByID(c.Request.Context(), id)
	if err != nil {
		if errors.Is(err, repository.ErrBlockNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "block not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, res)
}

func (ctrl *BlockController) Update(c *gin.Context) {
	id := c.Param("id")

	var req dtos.UpdateBlockRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	res, err := ctrl.blockService.UpdateBlock(c.Request.Context(), id, &req)
	if err != nil {
		// Return 409 Conflict if optimistic concurrency version mismatch occurs
		if errors.Is(err, repository.ErrVersionConflict) {
			c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
			return
		}
		if errors.Is(err, repository.ErrBlockNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "block not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, res)
}

func (ctrl *BlockController) Delete(c *gin.Context) {
	id := c.Param("id")

	if err := ctrl.blockService.DeleteBlock(c.Request.Context(), id); err != nil {
		if errors.Is(err, repository.ErrBlockNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "block not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "block deleted successfully"})
}