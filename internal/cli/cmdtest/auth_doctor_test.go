package cmdtest

import (
	"context"
	"encoding/json"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

type doctorMigrationJSON struct {
	DetectedFiles     []string `json:"detectedFiles"`
	DetectedActions   []string `json:"detectedActions"`
	SuggestedCommands []string `json:"suggestedCommands"`
}

type doctorReportJSON struct {
	Migration doctorMigrationJSON `json:"migration"`
}

func TestAuthDoctorJSONIncludesMigrationHints(t *testing.T) {
	withTempRepo(t, func(repo string) {
		fastlaneDir := filepath.Join(repo, "fastlane")
		if err := os.MkdirAll(fastlaneDir, 0o755); err != nil {
			t.Fatalf("mkdir fastlane error: %v", err)
		}
		if err := os.WriteFile(filepath.Join(fastlaneDir, "Appfile"), []byte(`app_identifier "com.example.app"`), 0o644); err != nil {
			t.Fatalf("write Appfile error: %v", err)
		}
		if err := os.WriteFile(filepath.Join(fastlaneDir, "Fastfile"), []byte("deliver\nupload_to_app_store\n"), 0o644); err != nil {
			t.Fatalf("write Fastfile error: %v", err)
		}

		t.Setenv("ASC_BYPASS_KEYCHAIN", "1")
		t.Setenv("ASC_CONFIG_PATH", filepath.Join(repo, "config.json"))

		root := RootCommand("1.2.3")
		root.FlagSet.SetOutput(io.Discard)

		stdout, _ := captureOutput(t, func() {
			if err := root.Parse([]string{"auth", "doctor", "--output", "json"}); err != nil {
				t.Fatalf("parse error: %v", err)
			}
			if err := root.Run(context.Background()); err != nil {
				t.Fatalf("run error: %v", err)
			}
		})

		var report doctorReportJSON
		if err := json.Unmarshal([]byte(stdout), &report); err != nil {
			t.Fatalf("unmarshal error: %v", err)
		}
		if !sliceContains(report.Migration.DetectedFiles, "fastlane/Appfile") {
			t.Fatalf("expected Appfile in detected files, got %#v", report.Migration.DetectedFiles)
		}
		if !sliceContains(report.Migration.DetectedActions, "deliver") {
			t.Fatalf("expected deliver action, got %#v", report.Migration.DetectedActions)
		}
		if len(report.Migration.SuggestedCommands) == 0 {
			t.Fatalf("expected suggested commands, got %#v", report.Migration.SuggestedCommands)
		}
	})
}

func TestAuthDoctorTextIncludesMigrationHints(t *testing.T) {
	withTempRepo(t, func(repo string) {
		fastlaneDir := filepath.Join(repo, "fastlane")
		if err := os.MkdirAll(fastlaneDir, 0o755); err != nil {
			t.Fatalf("mkdir fastlane error: %v", err)
		}
		if err := os.WriteFile(filepath.Join(fastlaneDir, "Deliverfile"), []byte("app_identifier \"com.example.app\""), 0o644); err != nil {
			t.Fatalf("write Deliverfile error: %v", err)
		}

		t.Setenv("ASC_BYPASS_KEYCHAIN", "1")
		t.Setenv("ASC_CONFIG_PATH", filepath.Join(repo, "config.json"))

		root := RootCommand("1.2.3")
		root.FlagSet.SetOutput(io.Discard)

		stdout, _ := captureOutput(t, func() {
			if err := root.Parse([]string{"auth", "doctor"}); err != nil {
				t.Fatalf("parse error: %v", err)
			}
			if err := root.Run(context.Background()); err != nil {
				t.Fatalf("run error: %v", err)
			}
		})

		if !strings.Contains(stdout, "Migration Hints:") {
			t.Fatalf("expected migration section heading, got %q", stdout)
		}
		if !strings.Contains(stdout, "Detected Deliverfile") {
			t.Fatalf("expected deliverfile detection, got %q", stdout)
		}
		if !strings.Contains(stdout, "Suggested:") {
			t.Fatalf("expected suggested commands, got %q", stdout)
		}
	})
}

func withTempRepo(t *testing.T, fn func(repo string)) {
	t.Helper()

	repo := t.TempDir()
	if err := os.Mkdir(filepath.Join(repo, ".git"), 0o755); err != nil {
		t.Fatalf("create .git error: %v", err)
	}
	previousDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("Getwd() error: %v", err)
	}
	if err := os.Chdir(repo); err != nil {
		t.Fatalf("Chdir() error: %v", err)
	}
	t.Cleanup(func() {
		_ = os.Chdir(previousDir)
	})

	fn(repo)
}

func sliceContains(values []string, target string) bool {
	for _, value := range values {
		if value == target {
			return true
		}
	}
	return false
}
