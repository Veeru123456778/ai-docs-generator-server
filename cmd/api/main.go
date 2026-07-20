package main

import (
	"log/slog"
	"os"

	"ai-docs-generator/internal/config"
	"ai-docs-generator/internal/database"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

func main() {
	// Initialize structured logger
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	slog.SetDefault(logger)    

	// Load .env file if present
	if err := godotenv.Load(); err != nil {
		slog.Info("No .env file found, reading from environment variables")
	}

	cfg := config.LoadConfig()

	db,err := database.NewPostgres(cfg)

	if err!=nil{
		slog.Error("Failed to initialize database connection", "error", err)
		os.Exit(1)
	}
    
	defer db.Close()

	// Initialize Gin router
	r := gin.New()
	r.Use(gin.Recovery())

	// Health check endpoint
	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status":  "ok",
			"env":     cfg.Env,
			"message": "Generative Document Engine API is live",
		})
	})

	slog.Info("Starting server", "port", cfg.Port)
	if err := r.Run(":" + cfg.Port); err != nil {
		slog.Error("Server failed to start", "error", err)
		os.Exit(1)
	}
}

