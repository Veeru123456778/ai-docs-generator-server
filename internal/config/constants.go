package config

import "time"

const (
	DefaultPort = "8080"
	DefaultEnv = "development"
	DefaultDatabaseURL = "postgres://postgres:postgres@localhost:5432/docengine?sslmode=disable"
)

// Define numerical connection pool settings for the database
const (
	// Maximum number of open connections allowed in the pool
	DBMaxConns = 25
	// Minimum number of idle connections kept ready in the pool
	DBMinConns = 5
	// Maximum lifetime duration for a single database connection
	DBMaxConnLifetime = 1 * time.Hour
	// Maximum idle time before an unused connection is closed
	DBMaxConnIdleTime = 30 * time.Minute
)