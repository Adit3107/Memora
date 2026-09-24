package config

import (
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/joho/godotenv"
)

type Config struct {
	Port                    string
	DatabaseURL             string
	AIServiceURL            string
	EmbeddingDimension      int
	EmbeddingMaxConcurrency int
	GeminiAPIKey            string
	GeminiModel             string
	RAG                     RAGConfig
	Storage                 StorageConfig
}

type RAGConfig struct {
	TopK               int
	HistoryMessages    int
	ContextMaxChars    int
	ChunkMaxChars      int
	AllowLocalFallback bool
}

type StorageConfig struct {
	AWSRegion string
	S3Bucket  string
}

func Load() Config {
	// Load both common local paths so backend/.env works whether the server is
	// started from the backend directory or the repository root.
	// In production, environment variables usually come from the hosting platform.
	_ = godotenv.Load(".env")
	_ = godotenv.Load("backend/.env")

	port := getEnv("PORT", "8080")
	aiURL := getEnv("AI_SERVICE_URL", "http://127.0.0.1:8000")
	aiURL = detectAIServiceURL(aiURL)

	return Config{
		Port:                    port,
		DatabaseURL:             os.Getenv("DATABASE_URL"),
		AIServiceURL:            aiURL,
		EmbeddingDimension:      getEnvInt("EMBEDDING_DIMENSION", 384),
		EmbeddingMaxConcurrency: getEnvInt("EMBEDDING_MAX_CONCURRENCY", 2),
		GeminiAPIKey:            os.Getenv("GEMINI_API_KEY"),
		GeminiModel:             getEnv("GEMINI_MODEL", "gemini-3.8-flash"),
		RAG: RAGConfig{
			TopK:               getEnvInt("RAG_TOP_K", 8),
			HistoryMessages:    getEnvInt("RAG_HISTORY_MESSAGES", 6),
			ContextMaxChars:    getEnvInt("RAG_CONTEXT_MAX_CHARS", 12000),
			ChunkMaxChars:      getEnvInt("RAG_CHUNK_MAX_CHARS", 1600),
			AllowLocalFallback: getEnvBool("RAG_ALLOW_LOCAL_FALLBACK", true),
		},
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

func getEnvInt(key string, fallback int) int {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}

	parsed, err := strconv.Atoi(value)
	if err != nil || parsed <= 0 {
		return fallback
	}

	return parsed
}

func getEnvBool(key string, fallback bool) bool {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}

	parsed, err := strconv.ParseBool(value)
	if err != nil {
		return fallback
	}

	return parsed
}

func detectAIServiceURL(preferred string) string {
	cleaned := strings.TrimRight(strings.TrimSpace(preferred), "/")
	cleaned = strings.Replace(cleaned, "localhost", "127.0.0.1", -1)

	client := &http.Client{Timeout: 500 * time.Millisecond}
	resp, err := client.Get(cleaned + "/health")
	if err == nil {
		resp.Body.Close()
		if resp.StatusCode == http.StatusOK {
			return cleaned
		}
	}

	alt := ""
	if strings.Contains(cleaned, ":8001") {
		alt = strings.Replace(cleaned, ":8001", ":8000", 1)
	} else if strings.Contains(cleaned, ":8000") {
		alt = strings.Replace(cleaned, ":8000", ":8001", 1)
	}

	if alt != "" {
		if respAlt, errAlt := client.Get(alt + "/health"); errAlt == nil {
			respAlt.Body.Close()
			if respAlt.StatusCode == http.StatusOK {
				return alt
			}
		}
	}

	return cleaned
}

// Why this file exists:
// This file centralizes environment configuration for the backend.
// Code outside config should use Config fields instead of repeatedly calling os.Getenv.
// DATABASE_URL stays in .env because Neon credentials are secrets and must not be committed.
// AWS credentials also stay in .env or the deployment platform; they must never go to frontend code.
