package config

import (
	"errors"
	"os"
)

type Config struct {
	DatabaseURL string
	Port        string
}

func Load() (Config, error) {
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		return Config{}, errors.New("DATABASE_URL is required")
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	return Config{DatabaseURL: dbURL, Port: port}, nil
}
