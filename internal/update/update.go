package update

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"
)

const (
	defaultRepo            = "rudrankriyam/App-Store-Connect-CLI"
	defaultBinaryName      = "asc"
	defaultAPIBaseURL      = "https://api.github.com"
	defaultDownloadBaseURL = "https://github.com"
	defaultCheckInterval   = 24 * time.Hour
	defaultTimeout         = 12 * time.Second
)

const (
	noUpdateEnvVar   = "ASC_NO_UPDATE"
	skipUpdateEnvVar = "ASC_SKIP_UPDATE"
)

// InstallMethod represents how the CLI was installed.
type InstallMethod string

const (
	InstallMethodUnknown  InstallMethod = "unknown"
	InstallMethodHomebrew InstallMethod = "homebrew"
	InstallMethodManual   InstallMethod = "manual"
)

// Options configures update checks and auto-updates.
type Options struct {
	Repo            string
	BinaryName      string
	CurrentVersion  string
	AutoUpdate      bool
	NoUpdate        bool
	CheckInterval   time.Duration
	CachePath       string
	APIBaseURL      string
	DownloadBaseURL string
	Client          *http.Client
	Output          io.Writer
	ShowProgress    bool
	Now             func() time.Time
	OS              string
	Arch            string
	ExecutablePath  string
	EvalSymlinks    func(string) (string, error)
}

// Result describes the outcome of an update check.
type Result struct {
	Skipped         bool
	SkipReason      string
	UpdateAvailable bool
	Updated         bool
	LatestVersion   string
	InstallMethod   InstallMethod
	ExecutablePath  string
}

// CheckAndUpdate checks for updates and optionally applies them.
func CheckAndUpdate(ctx context.Context, opts Options) (Result, error) {
	res := Result{}

	normalized, err := opts.withDefaults()
	if err != nil {
		return res, err
	}
	opts = normalized
	res.ExecutablePath = opts.ExecutablePath

	if opts.NoUpdate || envBool(noUpdateEnvVar) || envBool(skipUpdateEnvVar) {
		return Result{Skipped: true, SkipReason: "disabled", ExecutablePath: opts.ExecutablePath}, nil
	}

	currentDisplay, currentSemver, ok := normalizeVersion(opts.CurrentVersion)
	if !ok {
		return Result{Skipped: true, SkipReason: "non-release build", ExecutablePath: opts.ExecutablePath}, nil
	}

	latestVersion, usedCache, err := resolveLatestVersion(ctx, opts)
	if err != nil {
		return res, err
	}
	if latestVersion == "" {
		return Result{Skipped: true, SkipReason: "missing latest version", ExecutablePath: opts.ExecutablePath}, nil
	}

	latestDisplay, latestSemver, ok := normalizeVersion(latestVersion)
	if !ok {
		return Result{Skipped: true, SkipReason: "invalid latest version", ExecutablePath: opts.ExecutablePath}, nil
	}

	if compareVersions(currentSemver, latestSemver) >= 0 {
		if usedCache {
			res.LatestVersion = latestDisplay
		}
		return res, nil
	}

	res.UpdateAvailable = true
	res.LatestVersion = latestDisplay
	res.InstallMethod = detectInstallMethod(opts.ExecutablePath, opts.EvalSymlinks)

	if res.InstallMethod == InstallMethodHomebrew {
		fmt.Fprintf(opts.Output, "Update available (%s → %s). Run: brew upgrade rudrankriyam/tap/asc\n", currentDisplay, latestDisplay)
		return res, nil
	}

	if opts.OS == "windows" {
		fmt.Fprintf(opts.Output, "Update available (%s → %s). Download the latest release from GitHub.\n", currentDisplay, latestDisplay)
		return res, nil
	}

	if !opts.AutoUpdate {
		fmt.Fprintf(opts.Output, "Update available (%s → %s).\n", currentDisplay, latestDisplay)
		return res, nil
	}

	fmt.Fprintf(opts.Output, "Updating asc to %s...\n", latestDisplay)

	execPath := resolvedExecutable(opts.ExecutablePath, opts.EvalSymlinks)
	if execPath == "" {
		return res, fmt.Errorf("update failed: could not resolve executable path")
	}

	if err := downloadAndReplace(ctx, opts, execPath); err != nil {
		return res, err
	}

	fmt.Fprintf(opts.Output, "Updated asc to %s.\n", latestDisplay)
	res.Updated = true
	return res, nil
}

func (opts Options) withDefaults() (Options, error) {
	if opts.Repo == "" {
		opts.Repo = defaultRepo
	}
	if opts.BinaryName == "" {
		opts.BinaryName = defaultBinaryName
	}
	if opts.APIBaseURL == "" {
		opts.APIBaseURL = defaultAPIBaseURL
	}
	if opts.DownloadBaseURL == "" {
		opts.DownloadBaseURL = defaultDownloadBaseURL
	}
	if opts.CheckInterval == 0 {
		opts.CheckInterval = defaultCheckInterval
	}
	if opts.Client == nil {
		opts.Client = &http.Client{Timeout: defaultTimeout}
	}
	if opts.Output == nil {
		opts.Output = io.Discard
	}
	if opts.Now == nil {
		opts.Now = time.Now
	}
	if opts.OS == "" {
		opts.OS = runtime.GOOS
	}
	if opts.Arch == "" {
		opts.Arch = runtime.GOARCH
	}
	if opts.EvalSymlinks == nil {
		opts.EvalSymlinks = filepath.EvalSymlinks
	}
	if opts.ExecutablePath == "" {
		if path, err := os.Executable(); err == nil {
			opts.ExecutablePath = path
		}
	}
	if opts.CachePath == "" {
		if path, err := defaultCachePath(); err == nil {
			opts.CachePath = path
		}
	}
	return opts, nil
}

func envBool(name string) bool {
	value := strings.TrimSpace(os.Getenv(name))
	if value == "" {
		return false
	}
	parsed, err := parseBool(value)
	if err != nil {
		return true
	}
	return parsed
}
