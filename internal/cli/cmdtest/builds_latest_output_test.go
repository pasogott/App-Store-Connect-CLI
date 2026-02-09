package cmdtest

import (
	"context"
	"encoding/json"
	"errors"
	"flag"
	"io"
	"net/http"
	"path/filepath"
	"strings"
	"testing"
)

func TestBuildsLatestSelectsNewestAcrossPlatformPreReleaseVersions(t *testing.T) {
	setupAuth(t)
	t.Setenv("ASC_CONFIG_PATH", filepath.Join(t.TempDir(), "nonexistent.json"))

	const nextPreReleaseURL = "https://api.appstoreconnect.apple.com/v1/preReleaseVersions?page=2"

	originalTransport := http.DefaultTransport
	t.Cleanup(func() {
		http.DefaultTransport = originalTransport
	})

	http.DefaultTransport = roundTripFunc(func(req *http.Request) (*http.Response, error) {
		switch {
		case req.Method == http.MethodGet && req.URL.Path == "/v1/preReleaseVersions" && req.URL.Query().Get("page") == "":
			query := req.URL.Query()
			if query.Get("filter[app]") != "app-1" {
				t.Fatalf("expected filter[app]=app-1, got %q", query.Get("filter[app]"))
			}
			if query.Get("filter[platform]") != "IOS" {
				t.Fatalf("expected filter[platform]=IOS, got %q", query.Get("filter[platform]"))
			}
			if query.Get("limit") != "200" {
				t.Fatalf("expected limit=200, got %q", query.Get("limit"))
			}
			body := `{
				"data":[{"type":"preReleaseVersions","id":"prv-old"}],
				"links":{"next":"` + nextPreReleaseURL + `"}
			}`
			return &http.Response{
				StatusCode: http.StatusOK,
				Body:       io.NopCloser(strings.NewReader(body)),
				Header:     http.Header{"Content-Type": []string{"application/json"}},
			}, nil

		case req.Method == http.MethodGet && req.URL.String() == nextPreReleaseURL:
			body := `{
				"data":[{"type":"preReleaseVersions","id":"prv-new"}],
				"links":{"next":""}
			}`
			return &http.Response{
				StatusCode: http.StatusOK,
				Body:       io.NopCloser(strings.NewReader(body)),
				Header:     http.Header{"Content-Type": []string{"application/json"}},
			}, nil

		case req.Method == http.MethodGet && req.URL.Path == "/v1/builds":
			query := req.URL.Query()
			if query.Get("filter[app]") != "app-1" {
				t.Fatalf("expected filter[app]=app-1, got %q", query.Get("filter[app]"))
			}
			if query.Get("sort") != "-uploadedDate" {
				t.Fatalf("expected sort=-uploadedDate, got %q", query.Get("sort"))
			}
			if query.Get("limit") != "1" {
				t.Fatalf("expected limit=1, got %q", query.Get("limit"))
			}

			switch query.Get("filter[preReleaseVersion]") {
			case "prv-old":
				body := `{
					"data":[{"type":"builds","id":"build-old","attributes":{"uploadedDate":"2026-01-01T00:00:00Z"}}]
				}`
				return &http.Response{
					StatusCode: http.StatusOK,
					Body:       io.NopCloser(strings.NewReader(body)),
					Header:     http.Header{"Content-Type": []string{"application/json"}},
				}, nil
			case "prv-new":
				body := `{
					"data":[{"type":"builds","id":"build-new","attributes":{"uploadedDate":"2026-02-01T00:00:00Z"}}]
				}`
				return &http.Response{
					StatusCode: http.StatusOK,
					Body:       io.NopCloser(strings.NewReader(body)),
					Header:     http.Header{"Content-Type": []string{"application/json"}},
				}, nil
			default:
				t.Fatalf("unexpected filter[preReleaseVersion]=%q", query.Get("filter[preReleaseVersion]"))
				return nil, nil
			}

		default:
			t.Fatalf("unexpected request: %s %s", req.Method, req.URL.String())
			return nil, nil
		}
	})

	root := RootCommand("1.2.3")
	root.FlagSet.SetOutput(io.Discard)

	stdout, stderr := captureOutput(t, func() {
		if err := root.Parse([]string{"builds", "latest", "--app", "app-1", "--platform", "ios"}); err != nil {
			t.Fatalf("parse error: %v", err)
		}
		if err := root.Run(context.Background()); err != nil {
			t.Fatalf("run error: %v", err)
		}
	})

	if stderr != "" {
		t.Fatalf("expected empty stderr, got %q", stderr)
	}

	var out struct {
		Data struct {
			ID string `json:"id"`
		} `json:"data"`
	}
	if err := json.Unmarshal([]byte(stdout), &out); err != nil {
		t.Fatalf("unmarshal output: %v\nstdout: %s", err, stdout)
	}
	if out.Data.ID != "build-new" {
		t.Fatalf("expected latest build id build-new, got %q", out.Data.ID)
	}
}

func TestBuildsLatestReturnsPreReleaseLookupFailure(t *testing.T) {
	setupAuth(t)
	t.Setenv("ASC_CONFIG_PATH", filepath.Join(t.TempDir(), "nonexistent.json"))

	originalTransport := http.DefaultTransport
	t.Cleanup(func() {
		http.DefaultTransport = originalTransport
	})

	http.DefaultTransport = roundTripFunc(func(req *http.Request) (*http.Response, error) {
		if req.Method != http.MethodGet {
			t.Fatalf("expected GET, got %s", req.Method)
		}
		if req.URL.Path != "/v1/preReleaseVersions" {
			t.Fatalf("expected pre-release versions path, got %s", req.URL.Path)
		}
		body := `{"errors":[{"status":"500","title":"Server Error","detail":"pre-release lookup failed"}]}`
		return &http.Response{
			StatusCode: http.StatusInternalServerError,
			Body:       io.NopCloser(strings.NewReader(body)),
			Header:     http.Header{"Content-Type": []string{"application/json"}},
		}, nil
	})

	root := RootCommand("1.2.3")
	root.FlagSet.SetOutput(io.Discard)

	var runErr error
	stdout, _ := captureOutput(t, func() {
		if err := root.Parse([]string{"builds", "latest", "--app", "app-1", "--platform", "ios"}); err != nil {
			t.Fatalf("parse error: %v", err)
		}
		runErr = root.Run(context.Background())
	})

	if runErr == nil {
		t.Fatal("expected error, got nil")
	}
	if !strings.Contains(runErr.Error(), "builds latest: failed to lookup pre-release versions") {
		t.Fatalf("expected pre-release lookup error, got %v", runErr)
	}
	if stdout != "" {
		t.Fatalf("expected empty stdout, got %q", stdout)
	}
}

func TestBuildsLatestRejectsRepeatedPreReleasePaginationURL(t *testing.T) {
	setupAuth(t)
	t.Setenv("ASC_CONFIG_PATH", filepath.Join(t.TempDir(), "nonexistent.json"))

	const repeatedNextURL = "https://api.appstoreconnect.apple.com/v1/preReleaseVersions?cursor=AQ"

	originalTransport := http.DefaultTransport
	t.Cleanup(func() {
		http.DefaultTransport = originalTransport
	})

	requestCount := 0
	http.DefaultTransport = roundTripFunc(func(req *http.Request) (*http.Response, error) {
		requestCount++
		switch requestCount {
		case 1:
			if req.Method != http.MethodGet || req.URL.Path != "/v1/preReleaseVersions" {
				t.Fatalf("unexpected first request: %s %s", req.Method, req.URL.String())
			}
			body := `{
				"data":[{"type":"preReleaseVersions","id":"prv-1"}],
				"links":{"next":"` + repeatedNextURL + `"}
			}`
			return &http.Response{
				StatusCode: http.StatusOK,
				Body:       io.NopCloser(strings.NewReader(body)),
				Header:     http.Header{"Content-Type": []string{"application/json"}},
			}, nil
		case 2:
			if req.Method != http.MethodGet || req.URL.String() != repeatedNextURL {
				t.Fatalf("unexpected second request: %s %s", req.Method, req.URL.String())
			}
			body := `{
				"data":[{"type":"preReleaseVersions","id":"prv-2"}],
				"links":{"next":"` + repeatedNextURL + `"}
			}`
			return &http.Response{
				StatusCode: http.StatusOK,
				Body:       io.NopCloser(strings.NewReader(body)),
				Header:     http.Header{"Content-Type": []string{"application/json"}},
			}, nil
		default:
			t.Fatalf("unexpected extra request: %s %s", req.Method, req.URL.String())
			return nil, nil
		}
	})

	root := RootCommand("1.2.3")
	root.FlagSet.SetOutput(io.Discard)

	var runErr error
	stdout, _ := captureOutput(t, func() {
		if err := root.Parse([]string{"builds", "latest", "--app", "app-1", "--platform", "IOS"}); err != nil {
			t.Fatalf("parse error: %v", err)
		}
		runErr = root.Run(context.Background())
	})

	if runErr == nil {
		t.Fatal("expected error, got nil")
	}
	if !strings.Contains(runErr.Error(), "detected repeated pagination URL") {
		t.Fatalf("expected repeated pagination URL error, got %v", runErr)
	}
	if !strings.Contains(runErr.Error(), "failed to paginate pre-release versions") {
		t.Fatalf("expected pre-release pagination context, got %v", runErr)
	}
	if stdout != "" {
		t.Fatalf("expected empty stdout, got %q", stdout)
	}
}

func TestBuildsLatestOutputErrors(t *testing.T) {
	setupAuth(t)
	t.Setenv("ASC_CONFIG_PATH", filepath.Join(t.TempDir(), "nonexistent.json"))

	originalTransport := http.DefaultTransport
	t.Cleanup(func() {
		http.DefaultTransport = originalTransport
	})

	http.DefaultTransport = roundTripFunc(func(req *http.Request) (*http.Response, error) {
		if req.Method != http.MethodGet {
			t.Fatalf("expected GET, got %s", req.Method)
		}
		if req.URL.Path != "/v1/builds" {
			t.Fatalf("expected path /v1/builds, got %s", req.URL.Path)
		}
		query := req.URL.Query()
		if query.Get("filter[app]") != "app-1" {
			t.Fatalf("expected filter[app]=app-1, got %q", query.Get("filter[app]"))
		}
		if query.Get("sort") != "-uploadedDate" {
			t.Fatalf("expected sort=-uploadedDate, got %q", query.Get("sort"))
		}
		if query.Get("limit") != "1" {
			t.Fatalf("expected limit=1, got %q", query.Get("limit"))
		}
		body := `{
			"data":[{"type":"builds","id":"build-1","attributes":{"uploadedDate":"2026-02-01T00:00:00Z"}}]
		}`
		return &http.Response{
			StatusCode: http.StatusOK,
			Body:       io.NopCloser(strings.NewReader(body)),
			Header:     http.Header{"Content-Type": []string{"application/json"}},
		}, nil
	})

	tests := []struct {
		name    string
		args    []string
		wantErr string
	}{
		{
			name:    "unsupported output",
			args:    []string{"builds", "latest", "--app", "app-1", "--output", "yaml"},
			wantErr: "unsupported format: yaml",
		},
		{
			name:    "pretty with table",
			args:    []string{"builds", "latest", "--app", "app-1", "--output", "table", "--pretty"},
			wantErr: "--pretty is only valid with JSON output",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			root := RootCommand("1.2.3")
			root.FlagSet.SetOutput(io.Discard)

			var runErr error
			stdout, stderr := captureOutput(t, func() {
				if err := root.Parse(test.args); err != nil {
					t.Fatalf("parse error: %v", err)
				}
				runErr = root.Run(context.Background())
			})

			if runErr == nil {
				t.Fatal("expected error, got nil")
			}
			if errors.Is(runErr, flag.ErrHelp) {
				t.Fatalf("expected non-help error, got %v", runErr)
			}
			if !strings.Contains(runErr.Error(), test.wantErr) {
				t.Fatalf("expected error %q, got %v", test.wantErr, runErr)
			}
			if stdout != "" {
				t.Fatalf("expected empty stdout, got %q", stdout)
			}
			if stderr != "" {
				t.Fatalf("expected empty stderr, got %q", stderr)
			}
		})
	}
}

func TestBuildsLatestTableOutput(t *testing.T) {
	setupAuth(t)
	t.Setenv("ASC_CONFIG_PATH", filepath.Join(t.TempDir(), "nonexistent.json"))

	originalTransport := http.DefaultTransport
	t.Cleanup(func() {
		http.DefaultTransport = originalTransport
	})

	http.DefaultTransport = roundTripFunc(func(req *http.Request) (*http.Response, error) {
		if req.Method != http.MethodGet {
			t.Fatalf("expected GET, got %s", req.Method)
		}
		if req.URL.Path != "/v1/builds" {
			t.Fatalf("expected path /v1/builds, got %s", req.URL.Path)
		}
		query := req.URL.Query()
		if query.Get("filter[app]") != "app-1" {
			t.Fatalf("expected filter[app]=app-1, got %q", query.Get("filter[app]"))
		}
		if query.Get("sort") != "-uploadedDate" {
			t.Fatalf("expected sort=-uploadedDate, got %q", query.Get("sort"))
		}
		if query.Get("limit") != "1" {
			t.Fatalf("expected limit=1, got %q", query.Get("limit"))
		}
		body := `{
			"data":[{"type":"builds","id":"build-table","attributes":{"uploadedDate":"2026-03-01T00:00:00Z"}}]
		}`
		return &http.Response{
			StatusCode: http.StatusOK,
			Body:       io.NopCloser(strings.NewReader(body)),
			Header:     http.Header{"Content-Type": []string{"application/json"}},
		}, nil
	})

	root := RootCommand("1.2.3")
	root.FlagSet.SetOutput(io.Discard)

	stdout, stderr := captureOutput(t, func() {
		if err := root.Parse([]string{"builds", "latest", "--app", "app-1", "--output", "table"}); err != nil {
			t.Fatalf("parse error: %v", err)
		}
		if err := root.Run(context.Background()); err != nil {
			t.Fatalf("run error: %v", err)
		}
	})

	if stderr != "" {
		t.Fatalf("expected empty stderr, got %q", stderr)
	}
	if !strings.Contains(stdout, "2026-03-01T00:00:00Z") {
		t.Fatalf("expected table output to contain uploaded timestamp, got %q", stdout)
	}
}
