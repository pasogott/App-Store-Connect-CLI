package update

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"time"

	"github.com/rudrankriyam/App-Store-Connect-CLI/internal/config"
)

type cacheFile struct {
	CheckedAt     time.Time `json:"checked_at"`
	LatestVersion string    `json:"latest_version"`
}

func defaultCachePath() (string, error) {
	path, err := config.GlobalPath()
	if err != nil {
		return "", err
	}
	return filepath.Join(filepath.Dir(path), "update.json"), nil
}

func readCache(path string) (cacheFile, error) {
	if path == "" {
		return cacheFile{}, nil
	}
	data, err := os.ReadFile(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return cacheFile{}, nil
		}
		return cacheFile{}, err
	}
	if len(data) == 0 {
		return cacheFile{}, nil
	}
	var cache cacheFile
	if err := json.Unmarshal(data, &cache); err != nil {
		return cacheFile{}, err
	}
	return cache, nil
}

func writeCache(path string, cache cacheFile) error {
	if path == "" {
		return nil
	}
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}
	data, err := json.Marshal(cache)
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0o644)
}
