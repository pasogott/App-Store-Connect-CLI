package update

import (
	"context"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

type failRoundTripper struct {
	t *testing.T
}

func (f failRoundTripper) RoundTrip(*http.Request) (*http.Response, error) {
	f.t.Helper()
	f.t.Fatalf("unexpected HTTP request")
	return nil, fmt.Errorf("unexpected HTTP request")
}

func TestCheckAndUpdate_SkipsWhenDisabled(t *testing.T) {
	t.Setenv(noUpdateEnvVar, "1")

	client := &http.Client{Transport: failRoundTripper{t: t}}
	result, err := CheckAndUpdate(context.Background(), Options{
		CurrentVersion: "1.0.0",
		Client:         client,
	})
	if err != nil {
		t.Fatalf("CheckAndUpdate() error: %v", err)
	}
	if !result.Skipped {
		t.Fatal("expected update to be skipped when disabled")
	}
}

func TestCheckAndUpdate_UsesCacheWithoutNetwork(t *testing.T) {
	cachePath := filepath.Join(t.TempDir(), "update.json")
	cache := cacheFile{
		CheckedAt:     time.Now(),
		LatestVersion: "1.1.0",
	}
	data, err := json.Marshal(cache)
	if err != nil {
		t.Fatalf("json.Marshal error: %v", err)
	}
	if err := os.WriteFile(cachePath, data, 0o644); err != nil {
		t.Fatalf("os.WriteFile error: %v", err)
	}

	client := &http.Client{Transport: failRoundTripper{t: t}}
	result, err := CheckAndUpdate(context.Background(), Options{
		CurrentVersion: "1.0.0",
		AutoUpdate:     false,
		CachePath:      cachePath,
		CheckInterval:  24 * time.Hour,
		Client:         client,
	})
	if err != nil {
		t.Fatalf("CheckAndUpdate() error: %v", err)
	}
	if !result.UpdateAvailable {
		t.Fatal("expected update to be available from cache")
	}
}

func TestCheckAndUpdate_HomebrewSkipsDownload(t *testing.T) {
	downloaded := false
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case strings.HasPrefix(r.URL.Path, "/repos/"):
			_, _ = io.WriteString(w, `{"tag_name":"1.1.0"}`)
		case strings.Contains(r.URL.Path, "/releases/latest/download/"):
			downloaded = true
			w.WriteHeader(http.StatusBadRequest)
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	t.Cleanup(server.Close)

	result, err := CheckAndUpdate(context.Background(), Options{
		CurrentVersion:  "1.0.0",
		AutoUpdate:      true,
		APIBaseURL:      server.URL,
		DownloadBaseURL: server.URL,
		Output:          io.Discard,
		ExecutablePath:  "/opt/homebrew/bin/asc",
		EvalSymlinks: func(string) (string, error) {
			return "/opt/homebrew/Cellar/asc/1.0.0/bin/asc", nil
		},
		Client: server.Client(),
	})
	if err != nil {
		t.Fatalf("CheckAndUpdate() error: %v", err)
	}
	if !result.UpdateAvailable {
		t.Fatal("expected update to be available")
	}
	if result.Updated {
		t.Fatal("expected homebrew install to skip auto-update")
	}
	if downloaded {
		t.Fatal("expected download to be skipped for homebrew installs")
	}
}

func TestCheckAndUpdate_AutoUpdatesBinary(t *testing.T) {
	assetName := "asc-darwin-amd64"
	newBinary := []byte("new-binary")
	hash := sha256.Sum256(newBinary)
	checksums := fmt.Sprintf("%x  %s\n", hash, assetName)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/repos/rudrankriyam/App-Store-Connect-CLI/releases/latest":
			_, _ = io.WriteString(w, `{"tag_name":"1.1.0"}`)
		case "/rudrankriyam/App-Store-Connect-CLI/releases/latest/download/" + assetName:
			w.Header().Set("Content-Length", fmt.Sprintf("%d", len(newBinary)))
			_, _ = w.Write(newBinary)
		case "/rudrankriyam/App-Store-Connect-CLI/releases/latest/download/checksums.txt":
			_, _ = io.WriteString(w, checksums)
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	t.Cleanup(server.Close)

	execDir := t.TempDir()
	execPath := filepath.Join(execDir, "asc")
	if err := os.WriteFile(execPath, []byte("old-binary"), 0o755); err != nil {
		t.Fatalf("os.WriteFile error: %v", err)
	}

	result, err := CheckAndUpdate(context.Background(), Options{
		CurrentVersion:  "1.0.0",
		AutoUpdate:      true,
		APIBaseURL:      server.URL,
		DownloadBaseURL: server.URL,
		Output:          io.Discard,
		ShowProgress:    false,
		ExecutablePath:  execPath,
		EvalSymlinks: func(path string) (string, error) {
			return path, nil
		},
		Client:    server.Client(),
		OS:        "darwin",
		Arch:      "amd64",
		CachePath: filepath.Join(t.TempDir(), "update.json"),
	})
	if err != nil {
		t.Fatalf("CheckAndUpdate() error: %v", err)
	}
	if !result.Updated {
		t.Fatal("expected update to be applied")
	}

	updated, err := os.ReadFile(execPath)
	if err != nil {
		t.Fatalf("os.ReadFile error: %v", err)
	}
	if string(updated) != string(newBinary) {
		t.Fatalf("expected binary to be updated, got %q", string(updated))
	}
}
