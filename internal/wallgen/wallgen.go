package wallgen

import (
	"encoding/json"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

const (
	startMarker = "<!-- WALL-OF-APPS:START -->"
	endMarker   = "<!-- WALL-OF-APPS:END -->"
)

var platformDisplayNames = map[string]string{
	"IOS":       "iOS",
	"MAC_OS":    "macOS",
	"TV_OS":     "tvOS",
	"VISION_OS": "visionOS",
}

var platformAliases = map[string]string{
	"ios":       "IOS",
	"macos":     "MAC_OS",
	"mac_os":    "MAC_OS",
	"tvos":      "TV_OS",
	"tv_os":     "TV_OS",
	"visionos":  "VISION_OS",
	"vision_os": "VISION_OS",
}

type wallEntry struct {
	App      string   `json:"app"`
	Link     string   `json:"link"`
	Creator  string   `json:"creator"`
	Platform []string `json:"platform"`
}

// Result contains generated output paths.
type Result struct {
	GeneratedPath string
	ReadmePath    string
}

// Generate reads docs/wall-of-apps.json and updates docs/readme wall snippets.
func Generate(repoRoot string) (Result, error) {
	sourcePath := filepath.Join(repoRoot, "docs", "wall-of-apps.json")
	generatedPath := filepath.Join(repoRoot, "docs", "generated", "app-wall.md")
	readmePath := filepath.Join(repoRoot, "README.md")

	entries, err := readEntries(sourcePath)
	if err != nil {
		return Result{}, err
	}
	snippet := buildSnippet(entries)

	if err := writeGenerated(snippet, generatedPath); err != nil {
		return Result{}, err
	}
	if err := syncReadme(snippet, readmePath); err != nil {
		return Result{}, err
	}

	return Result{GeneratedPath: generatedPath, ReadmePath: readmePath}, nil
}

func readEntries(sourcePath string) ([]wallEntry, error) {
	raw, err := os.ReadFile(sourcePath)
	if err != nil {
		return nil, fmt.Errorf("missing source file: %s", sourcePath)
	}
	if strings.TrimSpace(string(raw)) == "" {
		return nil, fmt.Errorf("source file is empty: %s", sourcePath)
	}

	var parsed []wallEntry
	if err := json.Unmarshal(raw, &parsed); err != nil {
		return nil, fmt.Errorf("invalid JSON in %s: %w", sourcePath, err)
	}
	if len(parsed) == 0 {
		return nil, fmt.Errorf("source file has no entries: %s", sourcePath)
	}

	normalized := make([]wallEntry, 0, len(parsed))
	for idx, entry := range parsed {
		item, err := normalizeEntry(entry, idx+1)
		if err != nil {
			return nil, err
		}
		normalized = append(normalized, item)
	}

	sort.SliceStable(normalized, func(i, j int) bool {
		leftApp := strings.ToLower(normalized[i].App)
		rightApp := strings.ToLower(normalized[j].App)
		if leftApp != rightApp {
			return leftApp < rightApp
		}
		return strings.ToLower(normalized[i].Link) < strings.ToLower(normalized[j].Link)
	})

	return normalized, nil
}

func normalizeEntry(entry wallEntry, index int) (wallEntry, error) {
	entry.App = strings.TrimSpace(entry.App)
	entry.Link = strings.TrimSpace(entry.Link)
	entry.Creator = strings.TrimSpace(entry.Creator)
	if entry.App == "" {
		return wallEntry{}, fmt.Errorf("entry #%d: 'app' is required", index)
	}
	if entry.Link == "" {
		return wallEntry{}, fmt.Errorf("entry #%d: 'link' is required", index)
	}
	if entry.Creator == "" {
		return wallEntry{}, fmt.Errorf("entry #%d: 'creator' is required", index)
	}
	parsedURL, err := url.Parse(entry.Link)
	if err != nil || parsedURL.Scheme == "" || parsedURL.Host == "" {
		return wallEntry{}, fmt.Errorf("entry #%d: 'link' must be a valid http/https URL", index)
	}
	if parsedURL.Scheme != "http" && parsedURL.Scheme != "https" {
		return wallEntry{}, fmt.Errorf("entry #%d: 'link' must be a valid http/https URL", index)
	}
	if len(entry.Platform) == 0 {
		return wallEntry{}, fmt.Errorf("entry #%d: 'platform' must be a non-empty array", index)
	}

	platforms := make([]string, 0, len(entry.Platform))
	for _, value := range entry.Platform {
		token := strings.TrimSpace(value)
		normalized, ok := normalizePlatform(token)
		if !ok {
			allowed := strings.Join(allowedPlatformDisplayValues(), ", ")
			return wallEntry{}, fmt.Errorf(
				"entry #%d: invalid platform %q (allowed: %s)",
				index,
				token,
				allowed,
			)
		}
		if !contains(platforms, normalized) {
			platforms = append(platforms, normalized)
		}
	}
	entry.Platform = platforms
	return entry, nil
}

func normalizePlatform(value string) (string, bool) {
	key := strings.ToLower(value)
	key = strings.ReplaceAll(key, "-", "_")
	key = strings.ReplaceAll(key, " ", "")
	normalized, ok := platformAliases[key]
	return normalized, ok
}

func allowedPlatformDisplayValues() []string {
	return []string{"iOS", "macOS", "tvOS", "visionOS"}
}

func contains(values []string, needle string) bool {
	for _, value := range values {
		if value == needle {
			return true
		}
	}
	return false
}

func buildSnippet(entries []wallEntry) string {
	lines := []string{
		"## Wall of Apps",
		"",
		"Apps shipping with asc-cli. [Add yours via PR](https://github.com/rudrankriyam/App-Store-Connect-CLI/pulls)!",
		"",
		"| App | Link | Creator | Platform |",
		"|:----|:-----|:--------|:---------|",
	}

	for _, entry := range entries {
		platforms := make([]string, 0, len(entry.Platform))
		for _, platform := range entry.Platform {
			platforms = append(platforms, displayPlatform(platform))
		}
		lines = append(lines, fmt.Sprintf(
			"| %s | [Open](%s) | %s | %s |",
			escapeCell(entry.App),
			entry.Link,
			escapeCell(entry.Creator),
			escapeCell(strings.Join(platforms, ", ")),
		))
	}

	return strings.Join(lines, "\n") + "\n"
}

func displayPlatform(value string) string {
	if name, ok := platformDisplayNames[value]; ok {
		return name
	}
	return value
}

func escapeCell(value string) string {
	escaped := strings.ReplaceAll(value, "|", "\\|")
	return strings.TrimSpace(strings.ReplaceAll(escaped, "\n", " "))
}

func writeGenerated(snippet string, generatedPath string) error {
	if err := os.MkdirAll(filepath.Dir(generatedPath), 0o755); err != nil {
		return fmt.Errorf("create generated docs directory: %w", err)
	}
	header := "<!-- Generated from docs/wall-of-apps.json by tools/update-wall-of-apps. -->\n\n"
	if err := os.WriteFile(generatedPath, []byte(header+snippet), 0o644); err != nil {
		return fmt.Errorf("write generated doc: %w", err)
	}
	return nil
}

func syncReadme(snippet string, readmePath string) error {
	contentBytes, err := os.ReadFile(readmePath)
	if err != nil {
		return fmt.Errorf("missing README file: %s", readmePath)
	}
	content := string(contentBytes)
	start := strings.Index(content, startMarker)
	end := strings.Index(content, endMarker)
	if start == -1 || end == -1 || end < start {
		return fmt.Errorf("README markers not found. Expected WALL-OF-APPS markers in README.md")
	}

	before := content[:start]
	after := content[end+len(endMarker):]
	updated := before + startMarker + "\n" + snippet + endMarker + after

	if err := os.WriteFile(readmePath, []byte(updated), 0o644); err != nil {
		return fmt.Errorf("write README: %w", err)
	}
	return nil
}
