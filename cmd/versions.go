package cmd

import (
	"context"
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/peterbourgon/ff/v3/ffcli"

	"github.com/rudrankriyam/App-Store-Connect-CLI/internal/asc"
)

var appStoreVersionPlatforms = map[string]struct{}{
	"IOS":       {},
	"MAC_OS":    {},
	"TV_OS":     {},
	"VISION_OS": {},
}

var appStoreVersionStates = map[string]struct{}{
	"ACCEPTED":                      {},
	"DEVELOPER_REMOVED_FROM_SALE":   {},
	"DEVELOPER_REJECTED":            {},
	"IN_REVIEW":                     {},
	"INVALID_BINARY":                {},
	"METADATA_REJECTED":             {},
	"PENDING_APPLE_RELEASE":         {},
	"PENDING_CONTRACT":              {},
	"PENDING_DEVELOPER_RELEASE":     {},
	"PREPARE_FOR_SUBMISSION":        {},
	"PREORDER_READY_FOR_SALE":       {},
	"PROCESSING_FOR_APP_STORE":      {},
	"READY_FOR_REVIEW":              {},
	"READY_FOR_SALE":                {},
	"REJECTED":                      {},
	"REMOVED_FROM_SALE":             {},
	"WAITING_FOR_EXPORT_COMPLIANCE": {},
	"WAITING_FOR_REVIEW":            {},
	"REPLACED_WITH_NEW_VERSION":     {},
	"NOT_APPLICABLE":                {},
}

func VersionsCommand() *ffcli.Command {
	return &ffcli.Command{
		Name:       "versions",
		ShortUsage: "asc versions <subcommand> [flags]",
		ShortHelp:  "Manage App Store versions.",
		LongHelp: `Manage App Store versions.

Subcommands:
  list          List app store versions for an app.
  get           Get details for an app store version.
  attach-build  Attach a build to an app store version.`,
		UsageFunc: DefaultUsageFunc,
		Subcommands: []*ffcli.Command{
			VersionsListCommand(),
			VersionsGetCommand(),
			VersionsAttachBuildCommand(),
		},
		Exec: func(ctx context.Context, args []string) error {
			return flag.ErrHelp
		},
	}
}

func VersionsListCommand() *ffcli.Command {
	fs := flag.NewFlagSet("versions list", flag.ExitOnError)

	appID := fs.String("app", "", "App Store Connect app ID (or ASC_APP_ID)")
	version := fs.String("version", "", "Filter by version string (comma-separated)")
	platform := fs.String("platform", "", "Filter by platform: IOS, MAC_OS, TV_OS, VISION_OS (comma-separated)")
	state := fs.String("state", "", "Filter by state (comma-separated)")
	limit := fs.Int("limit", 0, "Maximum results per page (1-200)")
	next := fs.String("next", "", "Next page URL from a previous response")
	output := fs.String("output", "json", "Output format: json (default), table, markdown")
	pretty := fs.Bool("pretty", false, "Pretty-print JSON output")

	return &ffcli.Command{
		Name:       "list",
		ShortUsage: "asc versions list [flags]",
		ShortHelp:  "List app store versions for an app.",
		LongHelp: `List app store versions for an app.

Examples:
  asc versions list --app "123456789"
  asc versions list --app "123456789" --version "1.0.0"
  asc versions list --app "123456789" --platform IOS --state READY_FOR_REVIEW`,
		FlagSet:   fs,
		UsageFunc: DefaultUsageFunc,
		Exec: func(ctx context.Context, args []string) error {
			if *limit != 0 && (*limit < 1 || *limit > 200) {
				return fmt.Errorf("versions list: --limit must be between 1 and 200")
			}
			if err := validateNextURL(*next); err != nil {
				return fmt.Errorf("versions list: %w", err)
			}

			platforms, err := normalizeAppStoreVersionPlatforms(splitCSVUpper(*platform))
			if err != nil {
				return fmt.Errorf("versions list: %w", err)
			}
			states, err := normalizeAppStoreVersionStates(splitCSVUpper(*state))
			if err != nil {
				return fmt.Errorf("versions list: %w", err)
			}

			resolvedAppID := resolveAppID(*appID)
			if resolvedAppID == "" && strings.TrimSpace(*next) == "" {
				fmt.Fprintln(os.Stderr, "Error: --app is required (or set ASC_APP_ID)")
				return flag.ErrHelp
			}

			client, err := getASCClient()
			if err != nil {
				return fmt.Errorf("versions list: %w", err)
			}

			requestCtx, cancel := contextWithTimeout(ctx)
			defer cancel()

			opts := []asc.AppStoreVersionsOption{
				asc.WithAppStoreVersionsLimit(*limit),
				asc.WithAppStoreVersionsPlatforms(platforms),
				asc.WithAppStoreVersionsVersionStrings(splitCSV(*version)),
				asc.WithAppStoreVersionsStates(states),
				asc.WithAppStoreVersionsNextURL(*next),
			}

			versions, err := client.GetAppStoreVersions(requestCtx, resolvedAppID, opts...)
			if err != nil {
				return fmt.Errorf("versions list: %w", err)
			}

			return printOutput(versions, *output, *pretty)
		},
	}
}

func VersionsGetCommand() *ffcli.Command {
	fs := flag.NewFlagSet("versions get", flag.ExitOnError)

	versionID := fs.String("version-id", "", "App Store version ID (required)")
	includeBuild := fs.Bool("include-build", false, "Include attached build information")
	includeSubmission := fs.Bool("include-submission", false, "Include submission information")
	output := fs.String("output", "json", "Output format: json (default), table, markdown")
	pretty := fs.Bool("pretty", false, "Pretty-print JSON output")

	return &ffcli.Command{
		Name:       "get",
		ShortUsage: "asc versions get [flags]",
		ShortHelp:  "Get details for an app store version.",
		LongHelp: `Get details for an app store version.

Examples:
  asc versions get --version-id "VERSION_ID"
  asc versions get --version-id "VERSION_ID" --include-build --include-submission`,
		FlagSet:   fs,
		UsageFunc: DefaultUsageFunc,
		Exec: func(ctx context.Context, args []string) error {
			if strings.TrimSpace(*versionID) == "" {
				fmt.Fprintln(os.Stderr, "Error: --version-id is required")
				return flag.ErrHelp
			}

			client, err := getASCClient()
			if err != nil {
				return fmt.Errorf("versions get: %w", err)
			}

			requestCtx, cancel := contextWithTimeout(ctx)
			defer cancel()

			versionResp, err := client.GetAppStoreVersion(requestCtx, strings.TrimSpace(*versionID))
			if err != nil {
				return fmt.Errorf("versions get: %w", err)
			}

			result := &asc.AppStoreVersionDetailResult{
				ID:            versionResp.Data.ID,
				VersionString: versionResp.Data.Attributes.VersionString,
				Platform:      string(versionResp.Data.Attributes.Platform),
				State:         resolveAppStoreVersionState(versionResp.Data.Attributes),
			}

			if *includeBuild {
				buildResp, err := fetchOptionalBuild(requestCtx, strings.TrimSpace(*versionID), client.GetAppStoreVersionBuild)
				if err != nil {
					return fmt.Errorf("versions get: %w", err)
				}
				if buildResp != nil {
					result.BuildID = buildResp.Data.ID
					result.BuildVersion = buildResp.Data.Attributes.Version
				}
			}

			if *includeSubmission {
				submissionResp, err := fetchOptionalSubmission(requestCtx, strings.TrimSpace(*versionID), client.GetAppStoreVersionSubmissionForVersion)
				if err != nil {
					return fmt.Errorf("versions get: %w", err)
				}
				if submissionResp != nil {
					result.SubmissionID = submissionResp.Data.ID
				}
			}

			return printOutput(result, *output, *pretty)
		},
	}
}

func VersionsAttachBuildCommand() *ffcli.Command {
	fs := flag.NewFlagSet("versions attach-build", flag.ExitOnError)

	versionID := fs.String("version-id", "", "App Store version ID (required)")
	buildID := fs.String("build", "", "Build ID to attach (required)")
	output := fs.String("output", "json", "Output format: json (default), table, markdown")
	pretty := fs.Bool("pretty", false, "Pretty-print JSON output")

	return &ffcli.Command{
		Name:       "attach-build",
		ShortUsage: "asc versions attach-build [flags]",
		ShortHelp:  "Attach a build to an app store version.",
		LongHelp: `Attach a build to an app store version.

Examples:
  asc versions attach-build --version-id "VERSION_ID" --build "BUILD_ID"`,
		FlagSet:   fs,
		UsageFunc: DefaultUsageFunc,
		Exec: func(ctx context.Context, args []string) error {
			if strings.TrimSpace(*versionID) == "" {
				fmt.Fprintln(os.Stderr, "Error: --version-id is required")
				return flag.ErrHelp
			}
			if strings.TrimSpace(*buildID) == "" {
				fmt.Fprintln(os.Stderr, "Error: --build is required")
				return flag.ErrHelp
			}

			client, err := getASCClient()
			if err != nil {
				return fmt.Errorf("versions attach-build: %w", err)
			}

			requestCtx, cancel := contextWithTimeout(ctx)
			defer cancel()

			if err := client.AttachBuildToVersion(requestCtx, strings.TrimSpace(*versionID), strings.TrimSpace(*buildID)); err != nil {
				return fmt.Errorf("versions attach-build: %w", err)
			}

			result := &asc.AppStoreVersionAttachBuildResult{
				VersionID: strings.TrimSpace(*versionID),
				BuildID:   strings.TrimSpace(*buildID),
				Attached:  true,
			}

			return printOutput(result, *output, *pretty)
		},
	}
}

func normalizeAppStoreVersionPlatforms(values []string) ([]string, error) {
	if len(values) == 0 {
		return nil, nil
	}
	for _, value := range values {
		if _, ok := appStoreVersionPlatforms[value]; !ok {
			return nil, fmt.Errorf("--platform must be one of: %s", strings.Join(appStoreVersionPlatformList(), ", "))
		}
	}
	return values, nil
}

func normalizeAppStoreVersionStates(values []string) ([]string, error) {
	if len(values) == 0 {
		return nil, nil
	}
	for _, value := range values {
		if _, ok := appStoreVersionStates[value]; !ok {
			return nil, fmt.Errorf("--state must be one of: %s", strings.Join(appStoreVersionStateList(), ", "))
		}
	}
	return values, nil
}

func appStoreVersionPlatformList() []string {
	return []string{"IOS", "MAC_OS", "TV_OS", "VISION_OS"}
}

func appStoreVersionStateList() []string {
	return []string{
		"ACCEPTED",
		"DEVELOPER_REMOVED_FROM_SALE",
		"DEVELOPER_REJECTED",
		"IN_REVIEW",
		"INVALID_BINARY",
		"METADATA_REJECTED",
		"PENDING_APPLE_RELEASE",
		"PENDING_CONTRACT",
		"PENDING_DEVELOPER_RELEASE",
		"PREPARE_FOR_SUBMISSION",
		"PREORDER_READY_FOR_SALE",
		"PROCESSING_FOR_APP_STORE",
		"READY_FOR_REVIEW",
		"READY_FOR_SALE",
		"REJECTED",
		"REMOVED_FROM_SALE",
		"WAITING_FOR_EXPORT_COMPLIANCE",
		"WAITING_FOR_REVIEW",
		"REPLACED_WITH_NEW_VERSION",
		"NOT_APPLICABLE",
	}
}

func resolveAppStoreVersionState(attrs asc.AppStoreVersionAttributes) string {
	if attrs.AppVersionState != "" {
		return attrs.AppVersionState
	}
	return attrs.AppStoreState
}

func fetchOptionalBuild(ctx context.Context, versionID string, fetch func(context.Context, string) (*asc.BuildResponse, error)) (*asc.BuildResponse, error) {
	resp, err := fetch(ctx, versionID)
	if err != nil {
		if asc.IsNotFound(err) {
			return nil, nil
		}
		return nil, err
	}
	return resp, nil
}

func fetchOptionalSubmission(ctx context.Context, versionID string, fetch func(context.Context, string) (*asc.AppStoreVersionSubmissionResourceResponse, error)) (*asc.AppStoreVersionSubmissionResourceResponse, error) {
	resp, err := fetch(ctx, versionID)
	if err != nil {
		if asc.IsNotFound(err) {
			return nil, nil
		}
		return nil, err
	}
	return resp, nil
}
