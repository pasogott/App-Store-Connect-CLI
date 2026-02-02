package update

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
)

type latestReleaseResponse struct {
	TagName string `json:"tag_name"`
}

func resolveLatestVersion(ctx context.Context, opts Options) (string, bool, error) {
	cache, cacheErr := readCache(opts.CachePath)
	if cacheErr == nil && cache.LatestVersion != "" && opts.CheckInterval > 0 {
		if opts.Now().Sub(cache.CheckedAt) < opts.CheckInterval {
			return cache.LatestVersion, true, nil
		}
	}

	latest, err := fetchLatestVersion(ctx, opts)
	if err != nil {
		if cache.LatestVersion != "" {
			return cache.LatestVersion, true, nil
		}
		return "", false, err
	}

	_ = writeCache(opts.CachePath, cacheFile{
		CheckedAt:     opts.Now(),
		LatestVersion: latest,
	})

	return latest, false, nil
}

func fetchLatestVersion(ctx context.Context, opts Options) (string, error) {
	url := strings.TrimSuffix(opts.APIBaseURL, "/") + "/repos/" + opts.Repo + "/releases/latest"
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return "", err
	}
	req.Header.Set("Accept", "application/vnd.github+json")
	req.Header.Set("User-Agent", userAgent(opts.CurrentVersion))

	resp, err := opts.Client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("latest release request failed: %s", resp.Status)
	}

	var payload latestReleaseResponse
	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		return "", err
	}
	return strings.TrimSpace(payload.TagName), nil
}

func userAgent(versionInfo string) string {
	if display, _, ok := normalizeVersion(versionInfo); ok {
		return "asc/" + display
	}
	return "asc"
}
