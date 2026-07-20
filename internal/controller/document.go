package controller

import (
	"errors"
	"net/http"

	"ai-docs-generator/internal/dtos"
	"ai-docs-generator/internal/repository"
	"ai-docs-generator/internal/service"
	"ai-docs-generator/internal/models" 

	"github.com/gin-gonic/gin"
)

type DocumentController struct {
	docService service.DocumentService
}

func NewDocumentController(docService service.DocumentService) *DocumentController {
	return &DocumentController{docService: docService}
}

func (ctrl *DocumentController) Create(c *gin.Context) {
	var req dtos.CreateDocumentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	res, err := ctrl.docService.CreateDocument(c.Request.Context(), &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, res)
}


func (ctrl *DocumentController) List(c *gin.Context) {
	userID := c.Query("user_id")

	docs, err := ctrl.docService.GetDocumentsByUserID(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch documents"})
		return
	}

	if docs == nil {
		docs = []*models.Document{}
	}

	c.JSON(http.StatusOK, docs)
}


func (ctrl *DocumentController) GetByID(c *gin.Context) {
	id := c.Param("id")

	// Get document metadata along with all ordered child blocks
	res, err := ctrl.docService.GetDocumentWithBlocks(c.Request.Context(), id)
	if err != nil {
		if errors.Is(err, repository.ErrDocumentNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "document not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, res)
}

func (ctrl *DocumentController) Update(c *gin.Context) {
	id := c.Param("id")

	var req dtos.UpdateDocumentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	res, err := ctrl.docService.UpdateDocument(c.Request.Context(), id, &req)
	if err != nil {
		if errors.Is(err, repository.ErrDocumentNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "document not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, res)
}

func (ctrl *DocumentController) Delete(c *gin.Context) {
	id := c.Param("id")

	if err := ctrl.docService.DeleteDocument(c.Request.Context(), id); err != nil {
		if errors.Is(err, repository.ErrDocumentNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "document not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "document deleted successfully"})
}