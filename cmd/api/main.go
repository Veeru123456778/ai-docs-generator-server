package main

import (
	"log/slog"
	"os"

	"ai-docs-generator/internal/config"
	"ai-docs-generator/internal/controller"
	"ai-docs-generator/internal/database"
	"ai-docs-generator/internal/repository"
	"ai-docs-generator/internal/routes"
	"ai-docs-generator/internal/service"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

func main() {
	// 1. Structured Logging Setup
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	slog.SetDefault(logger)

	// 2. Load Environment Variables
	if err := godotenv.Load(); err != nil {
		slog.Info("No .env file found, relying on system environment variables")
	}

	cfg := config.LoadConfig()

	// 3. Database Pool Connection
	db, err := database.NewPostgres(cfg)
	if err != nil {
		slog.Error("Failed to initialize database connection", "error", err)
		os.Exit(1)
	}
	defer db.Close()

	// 4. Repositories
	docRepo := repository.NewPostgresDocumentRepository(db.Pool)
	blockRepo := repository.NewPostgresBlockRepository(db.Pool)

	// 5. Services
	docService := service.NewDocumentService(docRepo, blockRepo)
	blockService := service.NewBlockService(blockRepo, docRepo)

	// 6. Controllers
	docController := controller.NewDocumentController(docService)
	blockController := controller.NewBlockController(blockService)

	// 7. Router Setup & Route Registration
	r := gin.New()
	routes.RegisterRoutes(r, cfg, db, docController, blockController)

	// 8. Start HTTP Server
	slog.Info("Starting HTTP server", "port", cfg.Port)
	if err := r.Run(":" + cfg.Port); err != nil {
		slog.Error("Server execution failed", "error", err)
		os.Exit(1)
	}
}