package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadReadsBackendEnvWhenStartedFromRepoRoot(t *testing.T) {
	tempDir := t.TempDir()
	backendDir := filepath.Join(tempDir, "backend")
	if err := os.MkdirAll(backendDir, 0o755); err != nil {
		t.Fatalf("mkdir backend: %v", err)
	}
	if err := os.WriteFile(filepath.Join(backendDir, ".env"), []byte("GEMINI_API_KEY=test-key\nGEMINI_MODEL=test-model\n"), 0o600); err != nil {
		t.Fatalf("write backend .env: %v", err)
	}

	oldWd, err := os.Getwd()
	if err != nil {
		t.Fatalf("get wd: %v", err)
	}
	t.Cleanup(func() {
		_ = os.Chdir(oldWd)
		os.Unsetenv("GEMINI_API_KEY")
		os.Unsetenv("GEMINI_MODEL")
	})
	os.Unsetenv("GEMINI_API_KEY")
	os.Unsetenv("GEMINI_MODEL")

	if err := os.Chdir(tempDir); err != nil {
		t.Fatalf("chdir: %v", err)
	}

	cfg := Load()
	if cfg.GeminiAPIKey != "test-key" {
		t.Fatalf("expected Gemini key from backend/.env, got %q", cfg.GeminiAPIKey)
	}
	if cfg.GeminiModel != "test-model" {
		t.Fatalf("expected Gemini model from backend/.env, got %q", cfg.GeminiModel)
	}
}
