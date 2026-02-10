package cmdtest

import (
	"context"
	"errors"
	"flag"
	"io"
	"net/http"
	"path/filepath"
	"strings"
	"testing"

	"github.com/peterbourgon/ff/v3/ffcli"
)

func findSubcommand(root *ffcli.Command, path ...string) *ffcli.Command {
	cmd := root
	for _, part := range path {
		var next *ffcli.Command
		for _, sub := range cmd.Subcommands {
			if sub.Name == part {
				next = sub
				break
			}
		}
		if next == nil {
			return nil
		}
		cmd = next
	}
	return cmd
}

func TestAppsWallFlagDefaults(t *testing.T) {
	root := RootCommand("1.2.3")
	cmd := findSubcommand(root, "apps", "wall")
	if cmd == nil {
		t.Fatal("expected apps wall command")
	}

	outputFlag := cmd.FlagSet.Lookup("output")
	if outputFlag == nil {
		t.Fatal("expected --output flag")
	}
	if got := outputFlag.DefValue; got != "table" {
		t.Fatalf("expected --output default table, got %q", got)
	}

	sortFlag := cmd.FlagSet.Lookup("sort")
	if sortFlag == nil {
		t.Fatal("expected --sort flag")
	}
	if got := sortFlag.DefValue; got != "name" {
		t.Fatalf("expected --sort default name, got %q", got)
	}
}

func TestAppsWallIncludePlatformsValidationError(t *testing.T) {
	root := RootCommand("1.2.3")
	root.FlagSet.SetOutput(io.Discard)

	stdout, stderr := captureOutput(t, func() {
		if err := root.Parse([]string{"apps", "wall", "--include-platforms", "ANDROID"}); err != nil {
			t.Fatalf("parse error: %v", err)
		}
		err := root.Run(context.Background())
		if !errors.Is(err, flag.ErrHelp) {
			t.Fatalf("expected ErrHelp, got %v", err)
		}
	})

	if stdout != "" {
		t.Fatalf("expected empty stdout, got %q", stdout)
	}
	if !strings.Contains(stderr, "Error: --include-platforms must be one of: IOS, MAC_OS, TV_OS, VISION_OS") {
		t.Fatalf("expected invalid platform error, got %q", stderr)
	}
}

func TestAppsWallDefaultTableOutputSortedByName(t *testing.T) {
	setupAuth(t)
	t.Setenv("ASC_CONFIG_PATH", filepath.Join(t.TempDir(), "nonexistent.json"))

	originalTransport := http.DefaultTransport
	t.Cleanup(func() {
		http.DefaultTransport = originalTransport
	})

	requestCount := 0
	http.DefaultTransport = roundTripFunc(func(req *http.Request) (*http.Response, error) {
		requestCount++
		switch req.Host {
		case "api.appstoreconnect.apple.com":
			if req.Method != http.MethodGet {
				t.Fatalf("expected GET for apps, got %s", req.Method)
			}
			if req.URL.Path != "/v1/apps" {
				t.Fatalf("expected /v1/apps, got %s", req.URL.Path)
			}
			if got := req.URL.Query().Get("limit"); got != "200" {
				t.Fatalf("expected limit=200, got %q", got)
			}
			if got := req.URL.Query().Get("sort"); got != "name" {
				t.Fatalf("expected sort=name, got %q", got)
			}
			body := `{
				"data":[
					{"type":"apps","id":"2","attributes":{"name":"Zulu App","bundleId":"com.example.zulu","sku":"SKU-Z"}},
					{"type":"apps","id":"1","attributes":{"name":"Alpha App","bundleId":"com.example.alpha","sku":"SKU-A"}}
				],
				"links":{"next":""}
			}`
			return &http.Response{
				StatusCode: http.StatusOK,
				Body:       io.NopCloser(strings.NewReader(body)),
				Header:     http.Header{"Content-Type": []string{"application/json"}},
			}, nil
		case "itunes.apple.com":
			if req.Method != http.MethodGet {
				t.Fatalf("expected GET for lookup, got %s", req.Method)
			}
			if req.URL.Path != "/lookup" {
				t.Fatalf("expected /lookup path, got %s", req.URL.Path)
			}
			body := `{
				"resultCount":2,
				"results":[
					{
						"trackId":1,
						"sellerName":"Creator Alpha",
						"trackViewUrl":"https://apps.apple.com/app/id1",
						"artworkUrl512":"https://example.com/alpha.png",
						"supportedDevices":["iPhone15,3"]
					},
					{
						"trackId":2,
						"sellerName":"Creator Zulu",
						"trackViewUrl":"https://apps.apple.com/app/id2",
						"artworkUrl512":"https://example.com/zulu.png",
						"supportedDevices":["AppleTV6,2"]
					}
				]
			}`
			return &http.Response{
				StatusCode: http.StatusOK,
				Body:       io.NopCloser(strings.NewReader(body)),
				Header:     http.Header{"Content-Type": []string{"application/json"}},
			}, nil
		default:
			t.Fatalf("unexpected host: %s", req.Host)
			return nil, nil
		}
	})

	root := RootCommand("1.2.3")
	root.FlagSet.SetOutput(io.Discard)

	stdout, stderr := captureOutput(t, func() {
		if err := root.Parse([]string{"apps", "wall"}); err != nil {
			t.Fatalf("parse error: %v", err)
		}
		if err := root.Run(context.Background()); err != nil {
			t.Fatalf("run error: %v", err)
		}
	})

	if stderr != "" {
		t.Fatalf("expected empty stderr, got %q", stderr)
	}
	if !strings.Contains(stdout, "App") || !strings.Contains(stdout, "Link") || !strings.Contains(stdout, "Creator") || !strings.Contains(stdout, "Platform") || !strings.Contains(stdout, "Icon") {
		t.Fatalf("expected table headers, got %q", stdout)
	}
	if !strings.Contains(stdout, "Creator Alpha") || !strings.Contains(stdout, "Creator Zulu") {
		t.Fatalf("expected creator names in output, got %q", stdout)
	}
	alphaIdx := strings.Index(stdout, "Alpha App")
	zuluIdx := strings.Index(stdout, "Zulu App")
	if alphaIdx == -1 || zuluIdx == -1 {
		t.Fatalf("expected both app names in output, got %q", stdout)
	}
	if alphaIdx > zuluIdx {
		t.Fatalf("expected Alpha App before Zulu App, got %q", stdout)
	}
}

func TestAppsWallMarkdownIncludePlatformsLimitAndSort(t *testing.T) {
	setupAuth(t)
	t.Setenv("ASC_CONFIG_PATH", filepath.Join(t.TempDir(), "nonexistent.json"))

	originalTransport := http.DefaultTransport
	t.Cleanup(func() {
		http.DefaultTransport = originalTransport
	})

	http.DefaultTransport = roundTripFunc(func(req *http.Request) (*http.Response, error) {
		switch req.Host {
		case "api.appstoreconnect.apple.com":
			body := `{
				"data":[
					{"type":"apps","id":"10","attributes":{"name":"Older Mac App","bundleId":"com.example.mac.old","sku":"SKU-M1"}},
					{"type":"apps","id":"11","attributes":{"name":"Newer Mac App","bundleId":"com.example.mac.new","sku":"SKU-M2"}},
					{"type":"apps","id":"12","attributes":{"name":"iOS App","bundleId":"com.example.ios","sku":"SKU-I"}}
				],
				"links":{"next":""}
			}`
			return &http.Response{
				StatusCode: http.StatusOK,
				Body:       io.NopCloser(strings.NewReader(body)),
				Header:     http.Header{"Content-Type": []string{"application/json"}},
			}, nil
		case "itunes.apple.com":
			body := `{
				"resultCount":3,
				"results":[
					{
						"trackId":10,
						"sellerName":"Creator Old Mac",
						"trackViewUrl":"https://apps.apple.com/app/id10",
						"artworkUrl512":"https://example.com/old-mac.png",
						"supportedDevices":["Mac14,7"],
						"currentVersionReleaseDate":"2024-01-02T00:00:00Z"
					},
					{
						"trackId":11,
						"sellerName":"Creator New Mac",
						"trackViewUrl":"https://apps.apple.com/app/id11",
						"artworkUrl512":"https://example.com/new-mac.png",
						"supportedDevices":["Mac14,7"],
						"currentVersionReleaseDate":"2025-01-02T00:00:00Z"
					},
					{
						"trackId":12,
						"sellerName":"Creator iOS",
						"trackViewUrl":"https://apps.apple.com/app/id12",
						"artworkUrl512":"https://example.com/ios.png",
						"supportedDevices":["iPhone15,3"],
						"currentVersionReleaseDate":"2026-01-02T00:00:00Z"
					}
				]
			}`
			return &http.Response{
				StatusCode: http.StatusOK,
				Body:       io.NopCloser(strings.NewReader(body)),
				Header:     http.Header{"Content-Type": []string{"application/json"}},
			}, nil
		default:
			t.Fatalf("unexpected host: %s", req.Host)
			return nil, nil
		}
	})

	root := RootCommand("1.2.3")
	root.FlagSet.SetOutput(io.Discard)

	stdout, stderr := captureOutput(t, func() {
		if err := root.Parse([]string{
			"apps", "wall",
			"--output", "markdown",
			"--include-platforms", "MAC_OS",
			"--sort", "-releaseDate",
			"--limit", "1",
		}); err != nil {
			t.Fatalf("parse error: %v", err)
		}
		if err := root.Run(context.Background()); err != nil {
			t.Fatalf("run error: %v", err)
		}
	})

	if stderr != "" {
		t.Fatalf("expected empty stderr, got %q", stderr)
	}
	if !strings.Contains(stdout, "| App") || !strings.Contains(stdout, "| Link") || !strings.Contains(stdout, "| Creator") || !strings.Contains(stdout, "| Platform") || !strings.Contains(stdout, "| Icon") {
		t.Fatalf("expected markdown header, got %q", stdout)
	}
	if strings.Contains(stdout, "Older Mac App") || strings.Contains(stdout, "iOS App") {
		t.Fatalf("expected filtered/limited output, got %q", stdout)
	}
	if !strings.Contains(stdout, "Newer Mac App") {
		t.Fatalf("expected newest mac app in output, got %q", stdout)
	}
	if !strings.Contains(stdout, "Creator New Mac") {
		t.Fatalf("expected creator in output, got %q", stdout)
	}
}
