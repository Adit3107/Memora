package config

import (
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	Port        string
	DatabaseURL string
	Storage     StorageConfig
}

type StorageConfig struct {
	AWSRegion string
	S3Bucket  string
}

func Load() Config {
	// godotenv loads backend/.env during local development.
	// In production, environment variables usually come from the hosting platform.
	_ = godotenv.Load()

	port := getEnv("PORT", "8080")

	return Config{
		Port:        port,
		DatabaseURL: os.Getenv("DATABASE_URL"),
		Storage: StorageConfig{
			AWSRegion: os.Getenv("AWS_REGION"),
			S3Bucket:  os.Getenv("AWS_S3_BUCKET"),
		},
	}
}

func getEnv(key string, fallback string) string {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}

	return value
}

// Why this file exists:
// This file centralizes environment configuration for the backend.
// Code outside config should use Config fields instead of repeatedly calling os.Getenv.
// DATABASE_URL stays in .env because Neon credentials are secrets and must not be committed.
// AWS credentials also stay in .env or the deployment platform; they must never go to frontend code.
