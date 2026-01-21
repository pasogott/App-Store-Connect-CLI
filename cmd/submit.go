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

func SubmitCommand() *ffcli.Command {
	return &ffcli.Command{
		Name:       "submit",
		ShortUsage: "asc submit <subcommand> [flags]",
		ShortHelp:  "Submit builds for App Store review.",
		LongHelp: `Submit builds for App Store review.

Subcommands:
  create  Submit a build for review.
  status  Check submission status.
  cancel  Cancel a submission.`,
		UsageFunc: DefaultUsageFunc,
		Subcommands: []*ffcli.Command{
			SubmitCreateCommand(),
			SubmitStatusCommand(),
			SubmitCancelCommand(),
		},
		Exec: func(ctx context.Context, args []string) error {
			return flag.ErrHelp
		},
	}
}

func SubmitCreateCommand() *ffcli.Command {
	fs := flag.NewFlagSet("submit create", flag.ExitOnError)

	appID := fs.String("app", "", "App Store Connect app ID (or ASC_APP_ID)")
	version := fs.String("version", "", "App Store version string")
	versionID := fs.String("version-id", "", "App Store version ID")
	buildID := fs.String("build", "", "Build ID to attach")
	platform := fs.String("platform", "IOS", "Platform: IOS, MAC_OS, TV_OS, VISION_OS")
	confirm := fs.Bool("confirm", false, "Confirm submission (required)")
	output := fs.String("output", "json", "Output format: json (default), table, markdown")
	pretty := fs.Bool("pretty", false, "Pretty-print JSON output")

	return &ffcli.Command{
		Name:       "create",
		ShortUsage: "asc submit create [flags]",
		ShortHelp:  "Submit a build for App Store review.",
		LongHelp: `Submit a build for App Store review.

Examples:
  asc submit create --app "123456789" --version "1.0.0" --build "BUILD_ID" --confirm
  asc submit create --app "123456789" --version-id "VERSION_ID" --build "BUILD_ID" --confirm`,
		FlagSet:   fs,
		UsageFunc: DefaultUsageFunc,
		Exec: func(ctx context.Context, args []string) error {
			if !*confirm {
				fmt.Fprintln(os.Stderr, "Error: --confirm is required to submit for review")
				return flag.ErrHelp
			}
			if strings.TrimSpace(*buildID) == "" {
				fmt.Fprintln(os.Stderr, "Error: --build is required")
				return flag.ErrHelp
			}
			if strings.TrimSpace(*version) == "" && strings.TrimSpace(*versionID) == "" {
				fmt.Fprintln(os.Stderr, "Error: --version or --version-id is required")
				return flag.ErrHelp
			}
			if strings.TrimSpace(*version) != "" && strings.TrimSpace(*versionID) != "" {
				return fmt.Errorf("submit create: --version and --version-id are mutually exclusive")
			}

			resolvedAppID := resolveAppID(*appID)
			if resolvedAppID == "" {
				fmt.Fprintln(os.Stderr, "Error: --app is required (or set ASC_APP_ID)")
				return flag.ErrHelp
			}

			normalizedPlatform, err := normalizeSubmitPlatform(*platform)
			if err != nil {
				return fmt.Errorf("submit create: %w", err)
			}

			client, err := getASCClient()
			if err != nil {
				return fmt.Errorf("submit create: %w", err)
			}

			requestCtx, cancel := contextWithTimeout(ctx)
			defer cancel()

			resolvedVersionID := strings.TrimSpace(*versionID)
			if resolvedVersionID == "" {
				resolvedVersionID, err = resolveAppStoreVersionID(requestCtx, client, resolvedAppID, strings.TrimSpace(*version), normalizedPlatform)
				if err != nil {
					return fmt.Errorf("submit create: %w", err)
				}
			}

			if err := client.AttachBuildToVersion(requestCtx, resolvedVersionID, strings.TrimSpace(*buildID)); err != nil {
				return fmt.Errorf("submit create: failed to attach build: %w", err)
			}

			submitReq := asc.AppStoreVersionSubmissionCreateRequest{
				Data: asc.AppStoreVersionSubmissionCreateData{
					Type: asc.ResourceTypeAppStoreVersionSubmissions,
					Relationships: &asc.AppStoreVersionSubmissionRelationships{
						AppStoreVersion: &asc.Relationship{
							Data: asc.ResourceData{Type: asc.ResourceTypeAppStoreVersions, ID: resolvedVersionID},
						},
					},
				},
			}

			submitResp, err := client.CreateAppStoreVersionSubmission(requestCtx, submitReq)
			if err != nil {
				return fmt.Errorf("submit create: failed to create submission: %w", err)
			}

			result := &asc.AppStoreVersionSubmissionCreateResult{
				SubmissionID: submitResp.Data.ID,
				VersionID:    resolvedVersionID,
				BuildID:      strings.TrimSpace(*buildID),
				CreatedDate:  submitResp.Data.Attributes.CreatedDate,
			}

			return printOutput(result, *output, *pretty)
		},
	}
}

func SubmitStatusCommand() *ffcli.Command {
	fs := flag.NewFlagSet("submit status", flag.ExitOnError)

	submissionID := fs.String("id", "", "Submission ID")
	versionID := fs.String("version-id", "", "App Store version ID")
	output := fs.String("output", "json", "Output format: json (default), table, markdown")
	pretty := fs.Bool("pretty", false, "Pretty-print JSON output")

	return &ffcli.Command{
		Name:       "status",
		ShortUsage: "asc submit status [flags]",
		ShortHelp:  "Check submission status.",
		LongHelp: `Check submission status.

Examples:
  asc submit status --id "SUBMISSION_ID"
  asc submit status --version-id "VERSION_ID"`,
		FlagSet:   fs,
		UsageFunc: DefaultUsageFunc,
		Exec: func(ctx context.Context, args []string) error {
			if strings.TrimSpace(*submissionID) == "" && strings.TrimSpace(*versionID) == "" {
				fmt.Fprintln(os.Stderr, "Error: --id or --version-id is required")
				return flag.ErrHelp
			}
			if strings.TrimSpace(*submissionID) != "" && strings.TrimSpace(*versionID) != "" {
				return fmt.Errorf("submit status: --id and --version-id are mutually exclusive")
			}

			client, err := getASCClient()
			if err != nil {
				return fmt.Errorf("submit status: %w", err)
			}

			requestCtx, cancel := contextWithTimeout(ctx)
			defer cancel()

			var submissionResp *asc.AppStoreVersionSubmissionResourceResponse
			resolvedVersionID := strings.TrimSpace(*versionID)
			if strings.TrimSpace(*submissionID) != "" {
				submissionResp, err = client.GetAppStoreVersionSubmissionResource(requestCtx, strings.TrimSpace(*submissionID))
				if err != nil && asc.IsNotFound(err) {
					return fmt.Errorf("submit status: no submission found for ID %q", strings.TrimSpace(*submissionID))
				}
			} else {
				submissionResp, err = client.GetAppStoreVersionSubmissionForVersion(requestCtx, resolvedVersionID)
				if err != nil && asc.IsNotFound(err) {
					return fmt.Errorf("submit status: no submission found for version %q", resolvedVersionID)
				}
			}
			if err != nil {
				return fmt.Errorf("submit status: %w", err)
			}

			resolvedSubmissionID := submissionResp.Data.ID
			if submissionResp.Data.Relationships.AppStoreVersion != nil && submissionResp.Data.Relationships.AppStoreVersion.Data.ID != "" {
				resolvedVersionID = submissionResp.Data.Relationships.AppStoreVersion.Data.ID
			}

			result := &asc.AppStoreVersionSubmissionStatusResult{
				ID:          resolvedSubmissionID,
				VersionID:   resolvedVersionID,
				CreatedDate: submissionResp.Data.Attributes.CreatedDate,
			}

			if resolvedVersionID != "" {
				versionResp, err := client.GetAppStoreVersion(requestCtx, resolvedVersionID)
				if err != nil {
					return fmt.Errorf("submit status: %w", err)
				}
				result.VersionString = versionResp.Data.Attributes.VersionString
				result.Platform = string(versionResp.Data.Attributes.Platform)
				result.State = resolveAppStoreVersionState(versionResp.Data.Attributes)
			}

			return printOutput(result, *output, *pretty)
		},
	}
}

func SubmitCancelCommand() *ffcli.Command {
	fs := flag.NewFlagSet("submit cancel", flag.ExitOnError)

	submissionID := fs.String("id", "", "Submission ID")
	versionID := fs.String("version-id", "", "App Store version ID")
	confirm := fs.Bool("confirm", false, "Confirm cancellation (required)")
	output := fs.String("output", "json", "Output format: json (default), table, markdown")
	pretty := fs.Bool("pretty", false, "Pretty-print JSON output")

	return &ffcli.Command{
		Name:       "cancel",
		ShortUsage: "asc submit cancel [flags]",
		ShortHelp:  "Cancel a submission.",
		LongHelp: `Cancel a submission.

Examples:
  asc submit cancel --id "SUBMISSION_ID" --confirm
  asc submit cancel --version-id "VERSION_ID" --confirm`,
		FlagSet:   fs,
		UsageFunc: DefaultUsageFunc,
		Exec: func(ctx context.Context, args []string) error {
			if !*confirm {
				fmt.Fprintln(os.Stderr, "Error: --confirm is required to cancel a submission")
				return flag.ErrHelp
			}
			if strings.TrimSpace(*submissionID) == "" && strings.TrimSpace(*versionID) == "" {
				fmt.Fprintln(os.Stderr, "Error: --id or --version-id is required")
				return flag.ErrHelp
			}
			if strings.TrimSpace(*submissionID) != "" && strings.TrimSpace(*versionID) != "" {
				return fmt.Errorf("submit cancel: --id and --version-id are mutually exclusive")
			}

			client, err := getASCClient()
			if err != nil {
				return fmt.Errorf("submit cancel: %w", err)
			}

			requestCtx, cancel := contextWithTimeout(ctx)
			defer cancel()

			resolvedSubmissionID := strings.TrimSpace(*submissionID)
			if resolvedSubmissionID == "" {
				submissionResp, err := client.GetAppStoreVersionSubmissionForVersion(requestCtx, strings.TrimSpace(*versionID))
				if err != nil {
					if asc.IsNotFound(err) {
						return fmt.Errorf("submit cancel: no submission found for version %q", strings.TrimSpace(*versionID))
					}
					return fmt.Errorf("submit cancel: %w", err)
				}
				resolvedSubmissionID = submissionResp.Data.ID
			}

			if err := client.DeleteAppStoreVersionSubmission(requestCtx, resolvedSubmissionID); err != nil {
				if asc.IsNotFound(err) {
					return fmt.Errorf("submit cancel: no submission found for ID %q", resolvedSubmissionID)
				}
				return fmt.Errorf("submit cancel: %w", err)
			}

			result := &asc.AppStoreVersionSubmissionCancelResult{
				ID:        resolvedSubmissionID,
				Cancelled: true,
			}

			return printOutput(result, *output, *pretty)
		},
	}
}

func normalizeSubmitPlatform(value string) (string, error) {
	normalized := strings.ToUpper(strings.TrimSpace(value))
	if normalized == "" {
		return "", fmt.Errorf("--platform is required")
	}
	if _, ok := appStoreVersionPlatforms[normalized]; !ok {
		return "", fmt.Errorf("--platform must be one of: %s", strings.Join(appStoreVersionPlatformList(), ", "))
	}
	return normalized, nil
}

func resolveAppStoreVersionID(ctx context.Context, client *asc.Client, appID, version, platform string) (string, error) {
	opts := []asc.AppStoreVersionsOption{
		asc.WithAppStoreVersionsVersionStrings([]string{version}),
		asc.WithAppStoreVersionsPlatforms([]string{platform}),
		asc.WithAppStoreVersionsLimit(10),
	}
	resp, err := client.GetAppStoreVersions(ctx, appID, opts...)
	if err != nil {
		return "", err
	}
	if resp == nil || len(resp.Data) == 0 {
		return "", fmt.Errorf("app store version not found for version %q and platform %q", version, platform)
	}
	if len(resp.Data) > 1 {
		return "", fmt.Errorf("multiple app store versions found for version %q and platform %q (use --version-id)", version, platform)
	}
	return resp.Data[0].ID, nil
}
