package storage

import (
	"crypto/rand"
	"encoding/hex"
	"path/filepath"
	"regexp"
	"strings"
)

func NewObjectKey(ownerName string, originalName string) (string, error) {
	randomPart, err := randomHex(8)
	if err != nil {
		return "", err
	}

	fileName := cleanFileName(originalName)

	return "users/" + cleanPathPart(ownerName) + "/docs/" + randomPart + "-" + fileName, nil
}

func randomHex(byteCount int) (string, error) {
	bytes := make([]byte, byteCount)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}

	return hex.EncodeToString(bytes), nil
}

func cleanPathPart(value string) string {
	value = strings.ToLower(strings.TrimSpace(value))
	value = strings.ReplaceAll(value, "\\", "-")
	value = strings.ReplaceAll(value, "/", "-")
	value = strings.ReplaceAll(value, " ", "-")
	value = regexp.MustCompile(`[^a-z0-9._-]+`).ReplaceAllString(value, "-")
	value = strings.Trim(value, ".-_")

	if value == "" {
		return "unknown"
	}

	return value
}

func cleanFileName(value string) string {
	value = filepath.Base(strings.TrimSpace(value))
	value = strings.ReplaceAll(value, "\\", "-")
	value = regexp.MustCompile(`[^A-Za-z0-9._ -]+`).ReplaceAllString(value, "-")
	value = strings.Join(strings.Fields(value), " ")
	value = strings.Trim(value, ".-_ ")
	if value == "" || value == "." {
		return "upload.bin"
	}

	return value
}

// Why this file exists:
// S3 object keys are like paths, but they are not real folders.
// We generate keys instead of trusting original filenames so users cannot control storage paths.
// The random part avoids collisions when two files have the same name.
