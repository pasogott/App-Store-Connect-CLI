package notarization

import (
	"bytes"
	"context"
	"flag"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/peterbourgon/ff/v3/ffcli"

	"github.com/rudrankriyam/App-Store-Connect-CLI/internal/asc"
	"github.com/rudrankriyam/App-Store-Connect-CLI/internal/cli/shared"
)

// notarizationCommand returns the top-level notarization command.
func notarizationCommand() *ffcli.Command {
	return &ffcli.Command{
		Name:       "notarization",
		ShortUsage: "asc notarization <subcommand> [flags]",
		ShortHelp:  "Manage macOS notarization submissions.",
		LongHelp: `Manage macOS notarization submissions via the Apple Notary API.

Examples:
  asc notarization submit --file ./MyApp.zip
  asc notarization submit --file ./MyApp.zip --wait
  asc notarization status --id "SUBMISSION_ID"
  asc notarization log --id "SUBMISSION_ID"
  asc notarization history`,
		UsageFunc: DefaultUsageFunc,
		Subcommands: []*ffcli.Command{
			submitCommand(),
			statusCommand(),
			logCommand(),
			historyCommand(),
		},
		Exec: func(ctx context.Context, args []string) error {
			return flag.ErrHelp
		},
	}
}

// submitCommand returns the submit subcommand.
func submitCommand() *ffcli.Command {
	fs := flag.NewFlagSet("notarization submit", flag.ExitOnError)

	filePath := fs.String("file", "", "Path to the file to notarize (required, zip/dmg/pkg)")
	wait := fs.Bool("wait", false, "Wait for notarization to complete")
	pollInterval := fs.String("poll-interval", "15s", "Polling interval when using --wait")
	timeout := fs.String("timeout", "30m", "Timeout when using --wait")
	output := fs.String("output", "json", "Output format: json (default), table, markdown")
	pretty := fs.Bool("pretty", false, "Pretty-print JSON output")

	return &ffcli.Command{
		Name:       "submit",
		ShortUsage: "asc notarization submit --file <path> [flags]",
		ShortHelp:  "Submit software for notarization.",
		LongHelp: `Submit a file for macOS notarization via the Apple Notary API.

The file must be a zip, dmg, or pkg archive. The command computes the file's
SHA-256 hash, creates a submission, uploads the file to Apple's S3 bucket,
and optionally waits for the notarization to complete.

Examples:
  asc notarization submit --file ./MyApp.zip
  asc notarization submit --file ./MyApp.zip --wait
  asc notarization submit --file ./MyApp.zip --wait --poll-interval 30s --timeout 1h
  asc notarization submit --file ./MyApp.zip --output table`,
		FlagSet:   fs,
		UsageFunc: DefaultUsageFunc,
		Exec: func(ctx context.Context, args []string) error {
			pathValue := strings.TrimSpace(*filePath)
			if pathValue == "" {
				fmt.Fprintln(os.Stderr, "Error: --file is required")
				return flag.ErrHelp
			}

			interval, err := time.ParseDuration(strings.TrimSpace(*pollInterval))
			if err != nil || interval <= 0 {
				return fmt.Errorf("notarization submit: --poll-interval must be a valid positive duration (e.g. 15s, 1m)")
			}

			timeoutDuration, err := time.ParseDuration(strings.TrimSpace(*timeout))
			if err != nil || timeoutDuration <= 0 {
				return fmt.Errorf("notarization submit: --timeout must be a valid positive duration (e.g. 30m, 1h)")
			}

			// Validate file
			info, err := os.Lstat(pathValue)
			if err != nil {
				return fmt.Errorf("notarization submit: %w", err)
			}
			if info.Mode()&os.ModeSymlink != 0 {
				return fmt.Errorf("notarization submit: refusing to read symlink %q", pathValue)
			}
			if info.IsDir() {
				return fmt.Errorf("notarization submit: %q is a directory", pathValue)
			}
			if info.Size() <= 0 {
				return fmt.Errorf("notarization submit: file must not be empty")
			}

			// Compute SHA-256
			if shared.ProgressEnabled() {
				fmt.Fprintf(os.Stderr, "Computing SHA-256 hash of %s...\n", pathValue)
			}
			sha256Hash, err := asc.ComputeFileSHA256(pathValue)
			if err != nil {
				return fmt.Errorf("notarization submit: failed to compute SHA-256: %w", err)
			}

			client, err := getASCClient()
			if err != nil {
				return fmt.Errorf("notarization submit: %w", err)
			}

			requestCtx, cancel := contextWithTimeout(ctx)
			defer cancel()

			// Submit to Notary API
			submissionName := info.Name()
			if shared.ProgressEnabled() {
				fmt.Fprintf(os.Stderr, "Submitting %s for notarization...\n", submissionName)
			}

			submitResp, err := client.SubmitNotarization(requestCtx, sha256Hash, submissionName)
			if err != nil {
				return fmt.Errorf("notarization submit: %w", err)
			}

			submissionID := submitResp.Data.ID
			if shared.ProgressEnabled() {
				fmt.Fprintf(os.Stderr, "Submission created: %s\n", submissionID)
			}

			// Upload file to S3
			if shared.ProgressEnabled() {
				fmt.Fprintf(os.Stderr, "Uploading %s to Apple...\n", submissionName)
			}

			fileData, err := os.ReadFile(pathValue)
			if err != nil {
				return fmt.Errorf("notarization submit: failed to read file: %w", err)
			}

			uploadCtx, uploadCancel := contextWithUploadTimeout(ctx)
			defer uploadCancel()

			creds := asc.S3Credentials{
				AccessKeyID:     submitResp.Data.Attributes.AwsAccessKeyID,
				SecretAccessKey: submitResp.Data.Attributes.AwsSecretAccessKey,
				SessionToken:    submitResp.Data.Attributes.AwsSessionToken,
				Bucket:          submitResp.Data.Attributes.Bucket,
				Object:          submitResp.Data.Attributes.Object,
			}

			if err := asc.UploadToS3(uploadCtx, creds, bytes.NewReader(fileData)); err != nil {
				return fmt.Errorf("notarization submit: upload failed: %w", err)
			}

			if shared.ProgressEnabled() {
				fmt.Fprintln(os.Stderr, "Upload complete.")
			}

			// If not waiting, print the submission response and exit
			if !*wait {
				result := &notarizationSubmitResult{
					SubmissionID: submissionID,
					Status:       string(asc.NotaryStatusInProgress),
					Name:         submissionName,
					Message:      "Successfully submitted for notarization. Use 'asc notarization status --id " + submissionID + "' to check progress.",
				}
				return printOutput(result, *output, *pretty)
			}

			// Wait for notarization to complete
			if shared.ProgressEnabled() {
				fmt.Fprintf(os.Stderr, "Waiting for notarization (polling every %s, timeout %s)...\n", interval, timeoutDuration)
			}

			waitCtx, waitCancel := context.WithTimeout(ctx, timeoutDuration)
			defer waitCancel()

			statusResp, err := waitForNotarization(waitCtx, client, submissionID, interval)
			if err != nil {
				return fmt.Errorf("notarization submit: %w", err)
			}

			result := &notarizationSubmitResult{
				SubmissionID: submissionID,
				Status:       string(statusResp.Data.Attributes.Status),
				Name:         string(statusResp.Data.Attributes.Name),
				CreatedDate:  statusResp.Data.Attributes.CreatedDate,
			}

			if err := printOutput(result, *output, *pretty); err != nil {
				return err
			}

			switch statusResp.Data.Attributes.Status {
			case asc.NotaryStatusAccepted:
				if shared.ProgressEnabled() {
					fmt.Fprintln(os.Stderr, "Notarization complete! Status: Accepted")
				}
				return nil
			case asc.NotaryStatusInvalid, asc.NotaryStatusRejected:
				if shared.ProgressEnabled() {
					fmt.Fprintf(os.Stderr, "Notarization failed. Status: %s\n", statusResp.Data.Attributes.Status)
					fmt.Fprintf(os.Stderr, "Run 'asc notarization log --id %s' for details.\n", submissionID)
				}
				return shared.NewReportedError(fmt.Errorf("notarization %s: %s", submissionID, statusResp.Data.Attributes.Status))
			default:
				return nil
			}
		},
	}
}

// statusCommand returns the status subcommand.
func statusCommand() *ffcli.Command {
	fs := flag.NewFlagSet("notarization status", flag.ExitOnError)

	submissionID := fs.String("id", "", "Submission ID (required)")
	output := fs.String("output", "json", "Output format: json (default), table, markdown")
	pretty := fs.Bool("pretty", false, "Pretty-print JSON output")

	return &ffcli.Command{
		Name:       "status",
		ShortUsage: "asc notarization status --id \"SUBMISSION_ID\"",
		ShortHelp:  "Get the status of a notarization submission.",
		LongHelp: `Get the status of a notarization submission.

Status values: Accepted, In Progress, Invalid, Rejected.

Examples:
  asc notarization status --id "SUBMISSION_ID"
  asc notarization status --id "SUBMISSION_ID" --output table`,
		FlagSet:   fs,
		UsageFunc: DefaultUsageFunc,
		Exec: func(ctx context.Context, args []string) error {
			idValue := strings.TrimSpace(*submissionID)
			if idValue == "" {
				fmt.Fprintln(os.Stderr, "Error: --id is required")
				return flag.ErrHelp
			}

			client, err := getASCClient()
			if err != nil {
				return fmt.Errorf("notarization status: %w", err)
			}

			requestCtx, cancel := contextWithTimeout(ctx)
			defer cancel()

			resp, err := client.GetNotarizationStatus(requestCtx, idValue)
			if err != nil {
				return fmt.Errorf("notarization status: failed to fetch: %w", err)
			}

			return printOutput(resp, *output, *pretty)
		},
	}
}

// logCommand returns the log subcommand.
func logCommand() *ffcli.Command {
	fs := flag.NewFlagSet("notarization log", flag.ExitOnError)

	submissionID := fs.String("id", "", "Submission ID (required)")
	output := fs.String("output", "json", "Output format: json (default), table, markdown")
	pretty := fs.Bool("pretty", false, "Pretty-print JSON output")

	return &ffcli.Command{
		Name:       "log",
		ShortUsage: "asc notarization log --id \"SUBMISSION_ID\"",
		ShortHelp:  "Get the developer log URL for a notarization submission.",
		LongHelp: `Get the developer log URL for a notarization submission.

The log contains detailed information about the notarization result,
including any issues found during the scan.

Examples:
  asc notarization log --id "SUBMISSION_ID"
  asc notarization log --id "SUBMISSION_ID" --output table`,
		FlagSet:   fs,
		UsageFunc: DefaultUsageFunc,
		Exec: func(ctx context.Context, args []string) error {
			idValue := strings.TrimSpace(*submissionID)
			if idValue == "" {
				fmt.Fprintln(os.Stderr, "Error: --id is required")
				return flag.ErrHelp
			}

			client, err := getASCClient()
			if err != nil {
				return fmt.Errorf("notarization log: %w", err)
			}

			requestCtx, cancel := contextWithTimeout(ctx)
			defer cancel()

			resp, err := client.GetNotarizationLogs(requestCtx, idValue)
			if err != nil {
				return fmt.Errorf("notarization log: failed to fetch: %w", err)
			}

			return printOutput(resp, *output, *pretty)
		},
	}
}

// historyCommand returns the history subcommand.
func historyCommand() *ffcli.Command {
	fs := flag.NewFlagSet("notarization history", flag.ExitOnError)

	output := fs.String("output", "json", "Output format: json (default), table, markdown")
	pretty := fs.Bool("pretty", false, "Pretty-print JSON output")

	return &ffcli.Command{
		Name:       "history",
		ShortUsage: "asc notarization history [flags]",
		ShortHelp:  "List previous notarization submissions.",
		LongHelp: `List previous notarization submissions.

Examples:
  asc notarization history
  asc notarization history --output table`,
		FlagSet:   fs,
		UsageFunc: DefaultUsageFunc,
		Exec: func(ctx context.Context, args []string) error {
			client, err := getASCClient()
			if err != nil {
				return fmt.Errorf("notarization history: %w", err)
			}

			requestCtx, cancel := contextWithTimeout(ctx)
			defer cancel()

			resp, err := client.ListNotarizations(requestCtx)
			if err != nil {
				return fmt.Errorf("notarization history: failed to fetch: %w", err)
			}

			return printOutput(resp, *output, *pretty)
		},
	}
}

// notarizationSubmitResult is the CLI output for a notarization submission.
type notarizationSubmitResult struct {
	SubmissionID string `json:"submissionId"`
	Status       string `json:"status"`
	Name         string `json:"name"`
	CreatedDate  string `json:"createdDate,omitempty"`
	Message      string `json:"message,omitempty"`
}

// waitForNotarization polls the notarization status until it completes or the context is cancelled.
func waitForNotarization(ctx context.Context, client *asc.Client, submissionID string, pollInterval time.Duration) (*asc.NotarySubmissionStatusResponse, error) {
	ticker := time.NewTicker(pollInterval)
	defer ticker.Stop()

	for {
		requestCtx, cancel := shared.ContextWithTimeout(ctx)
		resp, err := client.GetNotarizationStatus(requestCtx, submissionID)
		cancel()

		if err != nil {
			return nil, fmt.Errorf("failed to check status: %w", err)
		}

		switch resp.Data.Attributes.Status {
		case asc.NotaryStatusAccepted, asc.NotaryStatusInvalid, asc.NotaryStatusRejected:
			return resp, nil
		default:
			// Treat unknown statuses (including InProgress) as non-terminal and continue polling
			if shared.ProgressEnabled() {
				fmt.Fprintf(os.Stderr, "Status: %s (checking again in %s)\n", resp.Data.Attributes.Status, pollInterval)
			}
		}

		select {
		case <-ctx.Done():
			return nil, fmt.Errorf("timed out waiting for notarization: %w", ctx.Err())
		case <-ticker.C:
		}
	}
}
