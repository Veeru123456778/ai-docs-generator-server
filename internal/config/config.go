package config

import "os"

type Config struct {
	Port        string
	Env         string
	DatabaseURL string
}

func LoadConfig() *Config {
	port := os.Getenv("PORT")
	if port == "" {
		port = DEFAULT_PORT
	}

	env := os.Getenv("ENV")
	if env == "" {
		env = DEFAULT_ENVIRONMENT
	}

	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		dbURL = DEFAULT_DATABASE_URL
	}

	return &Config{
		Port:        port,
		Env:         env,
		DatabaseURL: dbURL,
	}
};
