package testflight

import (
	"context"
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/peterbourgon/ff/v3/ffcli"
)

// BetaFeedbackCommand returns the beta-feedback command group.
func BetaFeedbackCommand() *ffcli.Command {
	fs := flag.NewFlagSet("beta-feedback", flag.ExitOnError)

	return &ffcli.Command{
		Name:       "beta-feedback",
		ShortUsage: "asc testflight beta-feedback <subcommand> [flags]",
		ShortHelp:  "View TestFlight beta feedback submissions.",
		LongHelp: `View TestFlight beta feedback submissions.

Examples:
  asc testflight beta-feedback crash-submissions get --id "SUBMISSION_ID"
  asc testflight beta-feedback screenshot-submissions get --id "SUBMISSION_ID"
  asc testflight beta-feedback crash-log get --id "SUBMISSION_ID"`,
		FlagSet:   fs,
		UsageFunc: DefaultUsageFunc,
		Subcommands: []*ffcli.Command{
			BetaFeedbackCrashSubmissionsCommand(),
			BetaFeedbackScreenshotSubmissionsCommand(),
			BetaFeedbackCrashLogCommand(),
		},
		Exec: func(ctx context.Context, args []string) error {
			return flag.ErrHelp
		},
	}
}

// BetaFeedbackCrashSubmissionsCommand returns the crash-submissions command group.
func BetaFeedbackCrashSubmissionsCommand() *ffcli.Command {
	fs := flag.NewFlagSet("crash-submissions", flag.ExitOnError)

	return &ffcli.Command{
		Name:       "crash-submissions",
		ShortUsage: "asc testflight beta-feedback crash-submissions <subcommand> [flags]",
		ShortHelp:  "Fetch beta feedback crash submission details.",
		LongHelp: `Fetch beta feedback crash submission details.

Examples:
  asc testflight beta-feedback crash-submissions get --id "SUBMISSION_ID"`,
		FlagSet:   fs,
		UsageFunc: DefaultUsageFunc,
		Subcommands: []*ffcli.Command{
			BetaFeedbackCrashSubmissionsGetCommand(),
		},
		Exec: func(ctx context.Context, args []string) error {
			return flag.ErrHelp
		},
	}
}

// BetaFeedbackCrashSubmissionsGetCommand returns the crash-submissions get subcommand.
func BetaFeedbackCrashSubmissionsGetCommand() *ffcli.Command {
	fs := flag.NewFlagSet("crash-submissions get", flag.ExitOnError)

	id := fs.String("id", "", "Beta feedback crash submission ID")
	output := fs.String("output", "json", "Output format: json (default), table, markdown")
	pretty := fs.Bool("pretty", false, "Pretty-print JSON output")

	return &ffcli.Command{
		Name:       "get",
		ShortUsage: "asc testflight beta-feedback crash-submissions get --id \"SUBMISSION_ID\"",
		ShortHelp:  "Get a beta feedback crash submission by ID.",
		LongHelp: `Get a beta feedback crash submission by ID.

Examples:
  asc testflight beta-feedback crash-submissions get --id "SUBMISSION_ID"`,
		FlagSet:   fs,
		UsageFunc: DefaultUsageFunc,
		Exec: func(ctx context.Context, args []string) error {
			idValue := strings.TrimSpace(*id)
			if idValue == "" {
				fmt.Fprintln(os.Stderr, "Error: --id is required")
				return flag.ErrHelp
			}

			client, err := getASCClient()
			if err != nil {
				return fmt.Errorf("testflight beta-feedback crash-submissions get: %w", err)
			}

			requestCtx, cancel := contextWithTimeout(ctx)
			defer cancel()

			resp, err := client.GetBetaFeedbackCrashSubmission(requestCtx, idValue)
			if err != nil {
				return fmt.Errorf("testflight beta-feedback crash-submissions get: failed to fetch: %w", err)
			}

			return printOutput(resp, *output, *pretty)
		},
	}
}

// BetaFeedbackScreenshotSubmissionsCommand returns the screenshot-submissions command group.
func BetaFeedbackScreenshotSubmissionsCommand() *ffcli.Command {
	fs := flag.NewFlagSet("screenshot-submissions", flag.ExitOnError)

	return &ffcli.Command{
		Name:       "screenshot-submissions",
		ShortUsage: "asc testflight beta-feedback screenshot-submissions <subcommand> [flags]",
		ShortHelp:  "Fetch beta feedback screenshot submission details.",
		LongHelp: `Fetch beta feedback screenshot submission details.

Examples:
  asc testflight beta-feedback screenshot-submissions get --id "SUBMISSION_ID"`,
		FlagSet:   fs,
		UsageFunc: DefaultUsageFunc,
		Subcommands: []*ffcli.Command{
			BetaFeedbackScreenshotSubmissionsGetCommand(),
		},
		Exec: func(ctx context.Context, args []string) error {
			return flag.ErrHelp
		},
	}
}

// BetaFeedbackScreenshotSubmissionsGetCommand returns the screenshot-submissions get subcommand.
func BetaFeedbackScreenshotSubmissionsGetCommand() *ffcli.Command {
	fs := flag.NewFlagSet("screenshot-submissions get", flag.ExitOnError)

	id := fs.String("id", "", "Beta feedback screenshot submission ID")
	output := fs.String("output", "json", "Output format: json (default), table, markdown")
	pretty := fs.Bool("pretty", false, "Pretty-print JSON output")

	return &ffcli.Command{
		Name:       "get",
		ShortUsage: "asc testflight beta-feedback screenshot-submissions get --id \"SUBMISSION_ID\"",
		ShortHelp:  "Get a beta feedback screenshot submission by ID.",
		LongHelp: `Get a beta feedback screenshot submission by ID.

Examples:
  asc testflight beta-feedback screenshot-submissions get --id "SUBMISSION_ID"`,
		FlagSet:   fs,
		UsageFunc: DefaultUsageFunc,
		Exec: func(ctx context.Context, args []string) error {
			idValue := strings.TrimSpace(*id)
			if idValue == "" {
				fmt.Fprintln(os.Stderr, "Error: --id is required")
				return flag.ErrHelp
			}

			client, err := getASCClient()
			if err != nil {
				return fmt.Errorf("testflight beta-feedback screenshot-submissions get: %w", err)
			}

			requestCtx, cancel := contextWithTimeout(ctx)
			defer cancel()

			resp, err := client.GetBetaFeedbackScreenshotSubmission(requestCtx, idValue)
			if err != nil {
				return fmt.Errorf("testflight beta-feedback screenshot-submissions get: failed to fetch: %w", err)
			}

			return printOutput(resp, *output, *pretty)
		},
	}
}

// BetaFeedbackCrashLogCommand returns the crash-log command group.
func BetaFeedbackCrashLogCommand() *ffcli.Command {
	fs := flag.NewFlagSet("crash-log", flag.ExitOnError)

	return &ffcli.Command{
		Name:       "crash-log",
		ShortUsage: "asc testflight beta-feedback crash-log <subcommand> [flags]",
		ShortHelp:  "Fetch crash logs for beta feedback crash submissions.",
		LongHelp: `Fetch crash logs for beta feedback crash submissions.

Examples:
  asc testflight beta-feedback crash-log get --id "SUBMISSION_ID"`,
		FlagSet:   fs,
		UsageFunc: DefaultUsageFunc,
		Subcommands: []*ffcli.Command{
			BetaFeedbackCrashLogGetCommand(),
		},
		Exec: func(ctx context.Context, args []string) error {
			return flag.ErrHelp
		},
	}
}

// BetaFeedbackCrashLogGetCommand returns the crash-log get subcommand.
func BetaFeedbackCrashLogGetCommand() *ffcli.Command {
	fs := flag.NewFlagSet("crash-log get", flag.ExitOnError)

	id := fs.String("id", "", "Beta feedback crash submission ID")
	output := fs.String("output", "json", "Output format: json (default), table, markdown")
	pretty := fs.Bool("pretty", false, "Pretty-print JSON output")

	return &ffcli.Command{
		Name:       "get",
		ShortUsage: "asc testflight beta-feedback crash-log get --id \"SUBMISSION_ID\"",
		ShortHelp:  "Get the crash log for a beta feedback crash submission.",
		LongHelp: `Get the crash log for a beta feedback crash submission.

Examples:
  asc testflight beta-feedback crash-log get --id "SUBMISSION_ID"`,
		FlagSet:   fs,
		UsageFunc: DefaultUsageFunc,
		Exec: func(ctx context.Context, args []string) error {
			idValue := strings.TrimSpace(*id)
			if idValue == "" {
				fmt.Fprintln(os.Stderr, "Error: --id is required")
				return flag.ErrHelp
			}

			client, err := getASCClient()
			if err != nil {
				return fmt.Errorf("testflight beta-feedback crash-log get: %w", err)
			}

			requestCtx, cancel := contextWithTimeout(ctx)
			defer cancel()

			resp, err := client.GetBetaFeedbackCrashSubmissionCrashLog(requestCtx, idValue)
			if err != nil {
				return fmt.Errorf("testflight beta-feedback crash-log get: failed to fetch: %w", err)
			}

			return printOutput(resp, *output, *pretty)
		},
	}
}
