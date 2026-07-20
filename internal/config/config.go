package config

import "os"

type Config struct {
	// Holds the server port number (e.g., "8080")
	Port string
	Env string
	DatabaseURL string
}

// LoadConfig reads environment variables and falls back to constants if missing
func LoadConfig() *Config {
	// Try to get the PORT environment variable from the operating system
	port := os.Getenv("PORT")
	if port == "" {
		port = DefaultPort
	}

	env := os.Getenv("ENV")
	if env == "" {
		env = DefaultEnv
	}

	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		// Assign the default database URL constant if missing
		dbURL = DefaultDatabaseURL
	}

	return &Config{
		// Set the Port field in the Config struct
		Port: port,
		// Set the Env field in the Config struct
		Env: env,
		// Set the DatabaseURL field in the Config struct
		DatabaseURL: dbURL,
	}
}