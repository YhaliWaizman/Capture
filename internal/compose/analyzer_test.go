package compose

import (
	"os"
	"path/filepath"
	"testing"
)

func TestComposeAnalyzer_Analyze(t *testing.T) {
	tmpDir := t.TempDir()

	extraEnv := filepath.Join(tmpDir, ".env.extra")
	if err := os.WriteFile(extraEnv, []byte("EXTRA_VAR=1\n"), 0644); err != nil {
		t.Fatalf("Failed to create env file: %v", err)
	}

	composePath := filepath.Join(tmpDir, "docker-compose.yml")
	content := `services:
  app:
    image: myapp:${VERSION:-latest}
    environment:
      - NODE_ENV=production
      - API_KEY=${API_KEY}
      - DATABASE_URL
    env_file:
      - .env.extra
      - .env.missing
  worker:
    environment:
      REDIS_URL: redis://localhost
      SECRET_TOKEN: ${SECRET_TOKEN}
`
	if err := os.WriteFile(composePath, []byte(content), 0644); err != nil {
		t.Fatalf("Failed to create compose file: %v", err)
	}

	analyzer := NewComposeAnalyzer()
	result, err := analyzer.Analyze(composePath)
	if err != nil {
		t.Fatalf("Analyze failed: %v", err)
	}

	expectedDeclared := []string{"NODE_ENV", "API_KEY", "DATABASE_URL", "REDIS_URL", "SECRET_TOKEN"}
	for _, varName := range expectedDeclared {
		if _, ok := result.Declared[varName]; !ok {
			t.Errorf("Expected %s to be declared", varName)
		}
	}

	expectedUsed := []string{"VERSION", "API_KEY", "SECRET_TOKEN"}
	for _, varName := range expectedUsed {
		if _, ok := result.Used[varName]; !ok {
			t.Errorf("Expected %s to be used", varName)
		}
	}

	if len(result.EnvFiles) != 2 {
		t.Errorf("Expected 2 env_file entries, got %d", len(result.EnvFiles))
	}

	missingPath := filepath.Clean(filepath.Join(tmpDir, ".env.missing"))
	if _, ok := result.MissingEnvFiles[missingPath]; !ok {
		t.Errorf("Expected missing env file %s to be reported", missingPath)
	}
}
