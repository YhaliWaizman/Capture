package detector

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"
)

func TestHardcodedSecretDetector_DetectsCommonSecretTypes(t *testing.T) {
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "secrets.js")

	// Split fixtures in source so GitHub secret scanning does not flag the tests.
	content := fmt.Sprintf(`const stripe = "%s";
const aws = "%s";
const github = "%s";
const apiKey = "%s";
const jwt = "%s";
const password = "S3cur3Passw0rd!";
const privateKey = "-----BEGIN PRIVATE KEY-----";
`,
		"sk"+"_"+"live_"+"AbCdEf1234567890AbCdEf12",
		"AKIA"+"1234567890ABCDEF",
		"gh"+"p_"+"abcdefghijklmnopqrstuvwxyz1234567890AB",
		"AIza"+"SyD0Example1234567890abcdefghiJKLMN",
		"eyJ"+"hbGciOiJIUzI1NiJ9."+"eyJzdWIiOiIxMjM0NTY3ODkwIn0."+"c2lnbmF0dXJlX3Rva2Vu",
	)

	if err := os.WriteFile(testFile, []byte(content), 0o644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	detector := NewHardcodedSecretDetector()
	findings, err := detector.Detect(testFile)
	if err != nil {
		t.Fatalf("Detect failed: %v", err)
	}

	wantTypes := map[string]bool{
		"AWS Access Key":       false,
		"GitHub Token":         false,
		"Stripe Key":           false,
		"Google API Key":       false,
		"JWT Token":            false,
		"Private Key":          false,
		"Hardcoded credential": false,
	}
	for _, finding := range findings {
		if _, ok := wantTypes[finding.Type]; ok {
			wantTypes[finding.Type] = true
		}
		if finding.Location.FilePath != testFile {
			t.Errorf("finding path = %q, want %q", finding.Location.FilePath, testFile)
		}
		if finding.Location.LineNumber < 1 {
			t.Errorf("invalid line number: %d", finding.Location.LineNumber)
		}
		if finding.Suggestion == "" {
			t.Error("suggestion should not be empty")
		}
	}

	for typ, seen := range wantTypes {
		if !seen {
			t.Errorf("expected finding type %q", typ)
		}
	}
}

func TestHardcodedSecretDetector_SkipsPlaceholderValues(t *testing.T) {
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "safe.js")
	content := `const password = "password";
const secret = "${API_SECRET}";
const token = "PROCESS_ENV_TOKEN";
`

	if err := os.WriteFile(testFile, []byte(content), 0o644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	detector := NewHardcodedSecretDetector()
	findings, err := detector.Detect(testFile)
	if err != nil {
		t.Fatalf("Detect failed: %v", err)
	}

	if len(findings) != 0 {
		t.Fatalf("expected no findings for placeholders, got %d", len(findings))
	}
}
