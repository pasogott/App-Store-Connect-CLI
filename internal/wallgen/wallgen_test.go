package wallgen

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func writeFile(t *testing.T, path string, content string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatalf("mkdir parent dir for %s: %v", path, err)
	}
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("write file %s: %v", path, err)
	}
}

func TestGenerateWritesDocsAndReadmeSnippet(t *testing.T) {
	tmpRepo := t.TempDir()

	writeFile(t, filepath.Join(tmpRepo, "docs", "wall-of-apps.json"), `[
  {
    "app": "Zulu App",
    "link": "https://example.com/zulu",
    "creator": "Zulu Creator",
    "platform": ["iOS"]
  },
  {
    "app": "Alpha App",
    "link": "https://example.com/alpha",
    "creator": "Alpha Creator",
    "platform": ["macOS", "iOS"]
  }
]`)

	writeFile(t, filepath.Join(tmpRepo, "README.md"), `# Demo

Before.
<!-- WALL-OF-APPS:START -->
Old content.
<!-- WALL-OF-APPS:END -->
After.
`)

	result, err := Generate(tmpRepo)
	if err != nil {
		t.Fatalf("generate failed: %v", err)
	}

	generatedContentBytes, err := os.ReadFile(result.GeneratedPath)
	if err != nil {
		t.Fatalf("read generated app wall: %v", err)
	}
	generatedContent := string(generatedContentBytes)

	if !strings.Contains(generatedContent, "Generated from docs/wall-of-apps.json") {
		t.Fatalf("expected generated header, got:\n%s", generatedContent)
	}
	if !strings.Contains(generatedContent, "| App | Link | Creator | Platform |") {
		t.Fatalf("expected markdown table header, got:\n%s", generatedContent)
	}

	alphaRow := "| Alpha App | [Open](https://example.com/alpha) | Alpha Creator | macOS, iOS |"
	zuluRow := "| Zulu App | [Open](https://example.com/zulu) | Zulu Creator | iOS |"

	alphaIdx := strings.Index(generatedContent, alphaRow)
	zuluIdx := strings.Index(generatedContent, zuluRow)
	if alphaIdx == -1 || zuluIdx == -1 {
		t.Fatalf("expected both generated rows, got:\n%s", generatedContent)
	}
	if alphaIdx > zuluIdx {
		t.Fatalf("expected deterministic app sorting, got:\n%s", generatedContent)
	}

	readmeBytes, err := os.ReadFile(result.ReadmePath)
	if err != nil {
		t.Fatalf("read generated README: %v", err)
	}
	readme := string(readmeBytes)
	if strings.Contains(readme, "Old content.") {
		t.Fatalf("expected README snippet to be replaced, got:\n%s", readme)
	}
	if !strings.Contains(readme, alphaRow) || !strings.Contains(readme, zuluRow) {
		t.Fatalf("expected README to include generated rows, got:\n%s", readme)
	}
}

func TestGenerateFailsWhenCreatorMissing(t *testing.T) {
	tmpRepo := t.TempDir()

	writeFile(t, filepath.Join(tmpRepo, "docs", "wall-of-apps.json"), `[
  {
    "app": "No Creator App",
    "link": "https://example.com/no-creator",
    "platform": ["iOS"]
  }
]`)

	writeFile(t, filepath.Join(tmpRepo, "README.md"), `# Demo
<!-- WALL-OF-APPS:START -->
Old content.
<!-- WALL-OF-APPS:END -->
`)

	_, err := Generate(tmpRepo)
	if err == nil {
		t.Fatal("expected generate to fail for missing creator")
	}
	if !strings.Contains(err.Error(), "'creator' is required") {
		t.Fatalf("expected missing creator error, got %v", err)
	}
}
