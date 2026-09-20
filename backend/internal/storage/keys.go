package storage

import (
	"crypto/rand"
	"encoding/hex"
	"path/filepath"
	"strings"
	"time"
)

func NewObjectKey(userID string, originalName string) (string, error) {
	randomPart, err := randomHex(8)
	if err != nil {
		return "", err
	}

	extension := strings.ToLower(filepath.Ext(originalName))
	datePath := time.Now().UTC().Format("2006/01/02")

	return "users/" + cleanPathPart(userID) + "/" + datePath + "/" + randomPart + extension, nil
}

func randomHex(byteCount int) (string, error) {
	bytes := make([]byte, byteCount)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}

	return hex.EncodeToString(bytes), nil
}

func cleanPathPart(value string) string {
	value = strings.TrimSpace(value)
	value = strings.ReplaceAll(value, "\\", "-")
	value = strings.ReplaceAll(value, "/", "-")

	if value == "" {
		return "unknown"
	}

	return value
}

// Why this file exists:
// S3 object keys are like paths, but they are not real folders.
// We generate keys instead of trusting original filenames so users cannot control storage paths.
// The random part avoids collisions when two files have the same name.
