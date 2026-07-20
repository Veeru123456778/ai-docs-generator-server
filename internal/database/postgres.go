// Declare that this file belongs to the 'database' package
package database

import (
	// Standard context package for managing timeouts and deadlines
	"context"
	// Standard fmt package for formatting text strings and errors
	"fmt"
	// Standard log/slog package for structured JSON logging
	"log/slog"
	// Standard time package for defining timeout durations
	"time"

	// Import your project config package using 'ai-docs-generator' root
	"ai-docs-generator/internal/config"

	// Import pgxpool driver for PostgreSQL connection pooling
	"github.com/jackc/pgx/v5/pgxpool"
)

// Define a wrapper struct that holds our PostgreSQL connection pool
type Postgres struct {
	// Pointer to pgxpool's connection pool manager
	Pool *pgxpool.Pool
}

// NewPostgres creates and verifies a connection pool to PostgreSQL
func NewPostgres(cfg *config.Config) (*Postgres, error) {
	// Create a context with a 10-second deadline for the initial DB connection
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	// Guarantee that context resources are cleaned up when function ends
	defer cancel()

	// Parse the raw database connection string into pgx configuration
	poolCfg, err := pgxpool.ParseConfig(cfg.DatabaseURL)
	// Check if the connection string format was invalid
	if err != nil {
		// Return nil and wrap the error with helpful context
		return nil, fmt.Errorf("unable to parse database url: %w", err)
	}

	// Set maximum open database connections from config constants
	poolCfg.MaxConns = config.DBMaxConns
	// Set minimum idle database connections from config constants
	poolCfg.MinConns = config.DBMinConns
	// Set maximum connection lifetime duration from config constants
	poolCfg.MaxConnLifetime = config.DBMaxConnLifetime
	// Set maximum connection idle time duration from config constants
	poolCfg.MaxConnIdleTime = config.DBMaxConnIdleTime

	// Create the actual connection pool using parsed configuration
	pool, err := pgxpool.NewWithConfig(ctx, poolCfg)
	// Check if creating the connection pool failed
	if err != nil {
		// Return nil and return error explaining connection failure
		return nil, fmt.Errorf("unable to connect to database: %w", err)
	}

	// Send a test ping query to verify database is reachable right now
	if err := pool.Ping(ctx); err != nil {
		// Return nil and error if database did not respond to ping
		return nil, fmt.Errorf("database ping failed: %w", err)
	}

	// Log success message to server output
	slog.Info("Successfully connected to PostgreSQL database")
	// Return the initialized Postgres struct wrapper
	return &Postgres{Pool: pool}, nil
}

// Close gracefully closes all active database connections in the pool
func (p *Postgres) Close() {
	// Check if the pool instance actually exists to prevent panic
	if p.Pool != nil {
		// Close all open network connections in the pool
		p.Pool.Close()
		// Log message confirming database connections are closed
		slog.Info("PostgreSQL connection pool closed")
	}
}