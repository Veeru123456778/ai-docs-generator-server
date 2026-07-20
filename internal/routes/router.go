package routes

import (
	"context"
	"net/http"
	"time"

	"ai-docs-generator/internal/config"
	"ai-docs-generator/internal/controller"
	"ai-docs-generator/internal/database"

	"github.com/gin-gonic/gin"
)

func RegisterRoutes(
	r *gin.Engine,
	cfg *config.Config,
	db *database.Postgres,
	docCtrl *controller.DocumentController,
	blockCtrl *controller.BlockController,
) {
	// Middleware
	r.Use(gin.Recovery())

	// Health Check Endpoint
	r.GET("/health", func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(c.Request.Context(), 2*time.Second)
		defer cancel()

		dbStatus := "healthy"
		if err := db.Pool.Ping(ctx); err != nil {
			dbStatus = "unhealthy"
		}

		c.JSON(http.StatusOK, gin.H{
			"status":   "ok",
			"env":      cfg.Env,
			"database": dbStatus,
			"message":  "Generative Document Engine API is live",
		})
	})

	// API v1 Group
	v1 := r.Group("/api/v1")
	{
		// Document Endpoints
		docs := v1.Group("/documents")
		
		{
			docs.POST("", docCtrl.Create)
			docs.GET("/:id", docCtrl.GetByID)
			docs.PUT("/:id", docCtrl.Update)
			docs.DELETE("/:id", docCtrl.Delete)
		}

		// Block Endpoints
		blocks := v1.Group("/blocks")
		{
			blocks.POST("", blockCtrl.Create)
			blocks.GET("/:id", blockCtrl.GetByID)
			blocks.PUT("/:id", blockCtrl.Update)
			blocks.DELETE("/:id", blockCtrl.Delete)
		}
	}
}