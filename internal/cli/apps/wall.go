package apps

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/peterbourgon/ff/v3/ffcli"

	"github.com/rudrankriyam/App-Store-Connect-CLI/internal/asc"
	"github.com/rudrankriyam/App-Store-Connect-CLI/internal/cli/shared"
)

const (
	appStoreLookupURL     = "https://itunes.apple.com/lookup"
	appStoreURLTemplate   = "https://apps.apple.com/app/id%s"
	appStoreLookupBatch   = 100
	defaultAppsWallSort   = "name"
	defaultAppsWallOutput = "table"
)

type appStoreLookupResponse struct {
	Results []appStoreLookupResult `json:"results"`
}

type appStoreLookupResult struct {
	TrackID                   int64    `json:"trackId"`
	Kind                      string   `json:"kind"`
	SellerName                string   `json:"sellerName"`
	ArtistName                string   `json:"artistName"`
	TrackViewURL              string   `json:"trackViewUrl"`
	ArtworkURL60              string   `json:"artworkUrl60"`
	ArtworkURL100             string   `json:"artworkUrl100"`
	ArtworkURL512             string   `json:"artworkUrl512"`
	SupportedDevices          []string `json:"supportedDevices"`
	ReleaseDate               string   `json:"releaseDate"`
	CurrentVersionReleaseDate string   `json:"currentVersionReleaseDate"`
}

// AppsWallCommand returns the apps wall subcommand.
func AppsWallCommand() *ffcli.Command {
	fs := flag.NewFlagSet("apps wall", flag.ExitOnError)

	output := fs.String("output", defaultAppsWallOutput, "Output format: table (default), json, markdown")
	pretty := fs.Bool("pretty", false, "Pretty-print JSON output")
	sortBy := fs.String("sort", defaultAppsWallSort, "Sort by name, -name, releaseDate, or -releaseDate")
	limit := fs.Int("limit", 0, "Maximum number of apps to include (1-200)")
	includePlatforms := fs.String("include-platforms", "", "Filter by platform(s): IOS, MAC_OS, TV_OS, VISION_OS (comma-separated)")

	return &ffcli.Command{
		Name:       "wall",
		ShortUsage: "asc apps wall [flags]",
		ShortHelp:  "Generate a wall of your apps with creator and platform metadata.",
		LongHelp: `Generate a wall of your apps with creator and platform metadata.

Examples:
  asc apps wall
  asc apps wall --output markdown
  asc apps wall --include-platforms IOS,MAC_OS --limit 20
  asc apps wall --sort -releaseDate`,
		FlagSet:   fs,
		UsageFunc: shared.DefaultUsageFunc,
		Exec: func(ctx context.Context, args []string) error {
			return appsWall(ctx, *output, *pretty, *sortBy, *limit, *includePlatforms)
		},
	}
}

func appsWall(ctx context.Context, output string, pretty bool, sortBy string, limit int, includePlatforms string) error {
	if limit != 0 && (limit < 1 || limit > 200) {
		fmt.Fprintln(os.Stderr, "Error: --limit must be between 1 and 200")
		return flag.ErrHelp
	}
	if err := shared.ValidateSort(sortBy, "name", "-name", "releaseDate", "-releaseDate"); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %s\n", err.Error())
		return flag.ErrHelp
	}

	platforms, err := shared.NormalizeAppStoreVersionPlatforms(shared.SplitCSVUpper(includePlatforms))
	if err != nil {
		message := strings.Replace(err.Error(), "--platform", "--include-platforms", 1)
		fmt.Fprintf(os.Stderr, "Error: %s\n", message)
		return flag.ErrHelp
	}

	client, err := shared.GetASCClient()
	if err != nil {
		return fmt.Errorf("apps wall: %w", err)
	}

	requestCtx, cancel := shared.ContextWithTimeout(ctx)
	defer cancel()

	appsResp, err := fetchAllAppsForWall(requestCtx, client, sortBy)
	if err != nil {
		return err
	}

	result, err := buildAppsWallResult(requestCtx, appsResp, sortBy, limit, platforms)
	if err != nil {
		return err
	}

	return shared.PrintOutput(result, output, pretty)
}

func fetchAllAppsForWall(ctx context.Context, client *asc.Client, sortBy string) (*asc.AppsResponse, error) {
	ascSort := "name"
	if sortBy == "name" || sortBy == "-name" {
		ascSort = sortBy
	}

	firstPage, err := client.GetApps(ctx, asc.WithAppsLimit(200), asc.WithAppsSort(ascSort))
	if err != nil {
		return nil, fmt.Errorf("apps wall: failed to fetch apps: %w", err)
	}

	all, err := asc.PaginateAll(ctx, firstPage, func(ctx context.Context, nextURL string) (asc.PaginatedResponse, error) {
		return client.GetApps(ctx, asc.WithAppsNextURL(nextURL))
	})
	if err != nil {
		return nil, fmt.Errorf("apps wall: %w", err)
	}

	resp, ok := all.(*asc.AppsResponse)
	if !ok {
		return nil, fmt.Errorf("apps wall: unexpected pagination response type %T", all)
	}
	return resp, nil
}

func buildAppsWallResult(ctx context.Context, appsResp *asc.AppsResponse, sortBy string, limit int, includePlatforms []string) (*asc.AppsWallResult, error) {
	entries := make([]asc.AppWallEntry, 0, len(appsResp.Data))
	for _, item := range appsResp.Data {
		appID := strings.TrimSpace(item.ID)
		entries = append(entries, asc.AppWallEntry{
			ID:          appID,
			Name:        strings.TrimSpace(item.Attributes.Name),
			AppStoreURL: fallbackAppStoreURL(appID),
		})
	}

	metadataByID, err := fetchAppStoreLookupMetadata(ctx, appIDsFromEntries(entries))
	if err != nil {
		return nil, err
	}
	applyLookupMetadata(entries, metadataByID)

	entries = filterWallEntriesByPlatforms(entries, includePlatforms)
	sortWallEntries(entries, sortBy)

	if limit > 0 && len(entries) > limit {
		entries = entries[:limit]
	}

	return &asc.AppsWallResult{Data: entries}, nil
}

func appIDsFromEntries(entries []asc.AppWallEntry) []string {
	ids := make([]string, 0, len(entries))
	for _, entry := range entries {
		if entry.ID == "" {
			continue
		}
		ids = append(ids, entry.ID)
	}
	return ids
}

func fetchAppStoreLookupMetadata(ctx context.Context, appIDs []string) (map[string]appStoreLookupResult, error) {
	metadataByID := make(map[string]appStoreLookupResult, len(appIDs))

	numericIDs := make([]string, 0, len(appIDs))
	for _, appID := range appIDs {
		if isNumericAppID(appID) {
			numericIDs = append(numericIDs, appID)
		}
	}
	if len(numericIDs) == 0 {
		return metadataByID, nil
	}

	httpClient := &http.Client{Timeout: asc.ResolveTimeout()}
	for _, batch := range chunkIDs(numericIDs, appStoreLookupBatch) {
		values := url.Values{}
		values.Set("id", strings.Join(batch, ","))

		req, err := http.NewRequestWithContext(ctx, http.MethodGet, appStoreLookupURL+"?"+values.Encode(), nil)
		if err != nil {
			return nil, fmt.Errorf("apps wall: failed to create app store lookup request: %w", err)
		}

		resp, err := httpClient.Do(req)
		if err != nil {
			return nil, fmt.Errorf("apps wall: app store lookup request failed: %w", err)
		}

		if resp.StatusCode != http.StatusOK {
			body, _ := io.ReadAll(io.LimitReader(resp.Body, 1024))
			_ = resp.Body.Close()
			if strings.TrimSpace(string(body)) == "" {
				return nil, fmt.Errorf("apps wall: app store lookup failed with status %s", resp.Status)
			}
			return nil, fmt.Errorf("apps wall: app store lookup failed with status %s: %s", resp.Status, strings.TrimSpace(string(body)))
		}

		var payload appStoreLookupResponse
		if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
			_ = resp.Body.Close()
			return nil, fmt.Errorf("apps wall: failed to parse app store lookup response: %w", err)
		}
		_ = resp.Body.Close()

		for _, result := range payload.Results {
			if result.TrackID <= 0 {
				continue
			}
			metadataByID[strconv.FormatInt(result.TrackID, 10)] = result
		}
	}

	return metadataByID, nil
}

func applyLookupMetadata(entries []asc.AppWallEntry, metadataByID map[string]appStoreLookupResult) {
	for i := range entries {
		lookup, ok := metadataByID[entries[i].ID]
		if !ok {
			continue
		}

		if trackURL := strings.TrimSpace(lookup.TrackViewURL); trackURL != "" {
			entries[i].AppStoreURL = trackURL
		}
		entries[i].Creator = firstNonEmpty(lookup.SellerName, lookup.ArtistName)
		entries[i].IconURL = firstNonEmpty(lookup.ArtworkURL512, lookup.ArtworkURL100, lookup.ArtworkURL60)

		releaseDate := strings.TrimSpace(lookup.CurrentVersionReleaseDate)
		if releaseDate == "" {
			releaseDate = strings.TrimSpace(lookup.ReleaseDate)
		}
		entries[i].ReleaseDate = releaseDate
		entries[i].Platform = inferLookupPlatforms(lookup.SupportedDevices)
		if len(entries[i].Platform) == 0 {
			entries[i].Platform = inferLookupPlatformFromKind(lookup.Kind)
		}
	}
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if trimmed := strings.TrimSpace(value); trimmed != "" {
			return trimmed
		}
	}
	return ""
}

func inferLookupPlatforms(supportedDevices []string) []string {
	set := map[string]struct{}{}
	for _, device := range supportedDevices {
		value := strings.ToLower(strings.TrimSpace(device))
		switch {
		case strings.Contains(value, "iphone"), strings.Contains(value, "ipad"), strings.Contains(value, "ipod"):
			set["IOS"] = struct{}{}
		case strings.Contains(value, "mac"):
			set["MAC_OS"] = struct{}{}
		case strings.Contains(value, "appletv"), strings.Contains(value, "tvos"):
			set["TV_OS"] = struct{}{}
		case strings.Contains(value, "vision"):
			set["VISION_OS"] = struct{}{}
		}
	}

	ordered := []string{"IOS", "MAC_OS", "TV_OS", "VISION_OS"}
	platforms := make([]string, 0, len(ordered))
	for _, platform := range ordered {
		if _, ok := set[platform]; ok {
			platforms = append(platforms, platform)
		}
	}
	return platforms
}

func inferLookupPlatformFromKind(kind string) []string {
	switch strings.ToLower(strings.TrimSpace(kind)) {
	case "mac-software":
		return []string{"MAC_OS"}
	case "appletv-software":
		return []string{"TV_OS"}
	case "software":
		return []string{"IOS"}
	default:
		return nil
	}
}

func filterWallEntriesByPlatforms(entries []asc.AppWallEntry, includePlatforms []string) []asc.AppWallEntry {
	if len(includePlatforms) == 0 {
		return entries
	}

	allowed := make(map[string]struct{}, len(includePlatforms))
	for _, value := range includePlatforms {
		allowed[value] = struct{}{}
	}

	filtered := make([]asc.AppWallEntry, 0, len(entries))
	for _, entry := range entries {
		if hasAnyPlatform(entry.Platform, allowed) {
			filtered = append(filtered, entry)
		}
	}
	return filtered
}

func hasAnyPlatform(platforms []string, allowed map[string]struct{}) bool {
	for _, platform := range platforms {
		if _, ok := allowed[platform]; ok {
			return true
		}
	}
	return false
}

func sortWallEntries(entries []asc.AppWallEntry, sortBy string) {
	switch sortBy {
	case "", "name":
		sort.SliceStable(entries, func(i, j int) bool {
			return lessByName(entries[i], entries[j])
		})
	case "-name":
		sort.SliceStable(entries, func(i, j int) bool {
			return lessByName(entries[j], entries[i])
		})
	case "releaseDate":
		sort.SliceStable(entries, func(i, j int) bool {
			return lessByReleaseDate(entries[i], entries[j], false)
		})
	case "-releaseDate":
		sort.SliceStable(entries, func(i, j int) bool {
			return lessByReleaseDate(entries[i], entries[j], true)
		})
	default:
		sort.SliceStable(entries, func(i, j int) bool {
			return lessByName(entries[i], entries[j])
		})
	}
}

func lessByName(left, right asc.AppWallEntry) bool {
	leftName := strings.ToLower(strings.TrimSpace(left.Name))
	rightName := strings.ToLower(strings.TrimSpace(right.Name))
	if leftName != rightName {
		return leftName < rightName
	}
	return left.ID < right.ID
}

func lessByReleaseDate(left, right asc.AppWallEntry, descending bool) bool {
	leftDate, leftOK := parseReleaseDate(left.ReleaseDate)
	rightDate, rightOK := parseReleaseDate(right.ReleaseDate)

	switch {
	case leftOK && rightOK:
		if !leftDate.Equal(rightDate) {
			if descending {
				return leftDate.After(rightDate)
			}
			return leftDate.Before(rightDate)
		}
	case leftOK != rightOK:
		// Always put entries with known release dates first.
		return leftOK
	}

	return lessByName(left, right)
}

func parseReleaseDate(value string) (time.Time, bool) {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return time.Time{}, false
	}
	parsed, err := time.Parse(time.RFC3339, trimmed)
	if err != nil {
		return time.Time{}, false
	}
	return parsed, true
}

func fallbackAppStoreURL(appID string) string {
	return fmt.Sprintf(appStoreURLTemplate, strings.TrimSpace(appID))
}

func isNumericAppID(value string) bool {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return false
	}
	for _, char := range trimmed {
		if char < '0' || char > '9' {
			return false
		}
	}
	return true
}

func chunkIDs(values []string, batchSize int) [][]string {
	if batchSize <= 0 || len(values) == 0 {
		return nil
	}

	chunks := make([][]string, 0, (len(values)+batchSize-1)/batchSize)
	for start := 0; start < len(values); start += batchSize {
		end := start + batchSize
		if end > len(values) {
			end = len(values)
		}
		chunks = append(chunks, values[start:end])
	}
	return chunks
}
