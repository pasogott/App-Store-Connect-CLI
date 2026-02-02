//go:build integration

package update

import (
	"context"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"testing"
	"time"
)

func TestIntegrationAutoUpdate(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("auto-update is disabled on Windows")
	}
	if assetName(defaultBinaryName, runtime.GOOS, runtime.GOARCH) == "" {
		t.Skipf("unsupported platform: %s/%s", runtime.GOOS, runtime.GOARCH)
	}

	t.Setenv(noUpdateEnvVar, "")
	t.Setenv(skipUpdateEnvVar, "")

	tempDir := t.TempDir()
	executable := filepath.Join(tempDir, defaultBinaryName)
	if err := os.WriteFile(executable, []byte("placeholder"), 0o755); err != nil {
		t.Fatalf("failed to write placeholder binary: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	result, err := CheckAndUpdate(ctx, Options{
		CurrentVersion: "0.0.0",
		AutoUpdate:     true,
		Output:         io.Discard,
		ShowProgress:   false,
		ExecutablePath: executable,
		CachePath:      filepath.Join(tempDir, "update.json"),
		Client:         &http.Client{Timeout: 60 * time.Second},
		OS:             runtime.GOOS,
		Arch:           runtime.GOARCH,
	})
	if err != nil {
		t.Fatalf("CheckAndUpdate() error: %v", err)
	}
	if !result.Updated {
		t.Fatalf("expected update to be applied, got: %+v", result)
	}

	info, err := os.Stat(executable)
	if err != nil {
		t.Fatalf("failed to stat updated binary: %v", err)
	}
	if info.Size() <= int64(len("placeholder")) {
		t.Fatalf("updated binary size looks too small: %d", info.Size())
	}
}
