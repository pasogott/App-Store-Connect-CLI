package update

import (
	"context"
	"crypto/sha256"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/schollz/progressbar/v3"
)

func downloadAndReplace(ctx context.Context, opts Options, execPath string) error {
	asset := assetName(opts.BinaryName, opts.OS, opts.Arch)
	if asset == "" {
		return fmt.Errorf("unsupported platform: %s/%s", opts.OS, opts.Arch)
	}
	base := strings.TrimSuffix(opts.DownloadBaseURL, "/")
	downloadURL := fmt.Sprintf("%s/%s/releases/latest/download/%s", base, opts.Repo, asset)
	checksumsURL := fmt.Sprintf("%s/%s/releases/latest/download/checksums.txt", base, opts.Repo)

	dir := filepath.Dir(execPath)
	tempFile, err := os.CreateTemp(dir, "."+opts.BinaryName+"-update-*")
	if err != nil {
		return fmt.Errorf("update failed: %w", err)
	}
	tempPath := tempFile.Name()
	removeTemp := true
	defer func() {
		if removeTemp {
			_ = os.Remove(tempPath)
		}
	}()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, downloadURL, nil)
	if err != nil {
		return err
	}
	resp, err := opts.Client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("download failed: %s", resp.Status)
	}

	hash := sha256.New()
	writer := io.MultiWriter(tempFile, hash)
	if opts.ShowProgress && resp.ContentLength > 0 {
		bar := progressbar.NewOptions64(
			resp.ContentLength,
			progressbar.OptionSetWriter(opts.Output),
			progressbar.OptionSetDescription("Downloading update"),
			progressbar.OptionShowBytes(true),
			progressbar.OptionClearOnFinish(),
			progressbar.OptionSetWidth(24),
			progressbar.OptionSetTheme(progressbar.Theme{
				Saucer:        "=",
				SaucerHead:    ">",
				SaucerPadding: "-",
				BarStart:      "[",
				BarEnd:        "]",
			}),
		)
		writer = io.MultiWriter(tempFile, hash, bar)
	}

	if _, err := io.Copy(writer, resp.Body); err != nil {
		return err
	}
	if err := tempFile.Sync(); err != nil {
		return err
	}
	if err := tempFile.Close(); err != nil {
		return err
	}

	actual := fmt.Sprintf("%x", hash.Sum(nil))
	expected, err := fetchChecksum(ctx, opts, checksumsURL, asset)
	if err != nil {
		fmt.Fprintf(opts.Output, "Warning: failed to verify checksum: %v\n", err)
	} else if expected == "" {
		fmt.Fprintln(opts.Output, "Warning: checksum not found; skipping verification.")
	} else if !strings.EqualFold(actual, expected) {
		return fmt.Errorf("checksum verification failed")
	}

	if runtime.GOOS != "windows" {
		if err := os.Chmod(tempPath, 0o755); err != nil && !errors.Is(err, os.ErrPermission) {
			return err
		}
	}

	if err := os.Rename(tempPath, execPath); err != nil {
		return err
	}
	removeTemp = false
	return nil
}

func assetName(binaryName, osName, arch string) string {
	if binaryName == "" || osName == "" || arch == "" {
		return ""
	}
	name := fmt.Sprintf("%s-%s-%s", binaryName, osName, arch)
	if osName == "windows" {
		return name + ".exe"
	}
	return name
}

func fetchChecksum(ctx context.Context, opts Options, url, asset string) (string, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return "", err
	}
	req.Header.Set("User-Agent", userAgent(opts.CurrentVersion))

	resp, err := opts.Client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("checksum request failed: %s", resp.Status)
	}
	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	return parseChecksum(string(data), asset), nil
}

func parseChecksum(data, asset string) string {
	for _, line := range strings.Split(data, "\n") {
		fields := strings.Fields(line)
		if len(fields) < 2 {
			continue
		}
		if fields[len(fields)-1] == asset {
			return fields[0]
		}
	}
	return ""
}
