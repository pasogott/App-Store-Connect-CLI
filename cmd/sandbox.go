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

// SandboxCommand returns the sandbox testers command with subcommands.
func SandboxCommand() *ffcli.Command {
	fs := flag.NewFlagSet("sandbox", flag.ExitOnError)

	return &ffcli.Command{
		Name:       "sandbox",
		ShortUsage: "asc sandbox <subcommand> [flags]",
		ShortHelp:  "Manage App Store Connect sandbox testers.",
		LongHelp: `Manage sandbox testers for in-app purchase testing.

Examples:
  asc sandbox list
  asc sandbox list --email "tester@example.com"
  asc sandbox create --email "tester@example.com" --first-name "Test" --last-name "User" --password "Passwordtest1" --confirm-password "Passwordtest1" --secret-question "Question" --secret-answer "Answer" --birth-date "1980-03-01" --territory "USA"
  asc sandbox get --id "SANDBOX_TESTER_ID"
  asc sandbox delete --email "tester@example.com" --confirm`,
		FlagSet:   fs,
		UsageFunc: DefaultUsageFunc,
		Subcommands: []*ffcli.Command{
			SandboxListCommand(),
			SandboxCreateCommand(),
			SandboxGetCommand(),
			SandboxDeleteCommand(),
		},
		Exec: func(ctx context.Context, args []string) error {
			return flag.ErrHelp
		},
	}
}

// SandboxListCommand returns the sandbox list subcommand.
func SandboxListCommand() *ffcli.Command {
	fs := flag.NewFlagSet("list", flag.ExitOnError)

	email := fs.String("email", "", "Filter by tester email")
	territory := fs.String("territory", "", "Filter by territory (e.g., USA, JPN)")
	limit := fs.Int("limit", 0, "Maximum results per page (1-200)")
	next := fs.String("next", "", "Fetch next page using a links.next URL")
	output := fs.String("output", "json", "Output format: json (default), table, markdown")
	pretty := fs.Bool("pretty", false, "Pretty-print JSON output")

	return &ffcli.Command{
		Name:       "list",
		ShortUsage: "asc sandbox list [flags]",
		ShortHelp:  "List sandbox testers.",
		LongHelp: `List sandbox testers for the App Store Connect team.

Examples:
  asc sandbox list
  asc sandbox list --email "tester@example.com"
  asc sandbox list --territory "USA"
  asc sandbox list --limit 50`,
		FlagSet:   fs,
		UsageFunc: DefaultUsageFunc,
		Exec: func(ctx context.Context, args []string) error {
			if *limit != 0 && (*limit < 1 || *limit > 200) {
				return fmt.Errorf("sandbox list: --limit must be between 1 and 200")
			}
			if err := validateNextURL(*next); err != nil {
				return fmt.Errorf("sandbox list: %w", err)
			}
			if strings.TrimSpace(*email) != "" {
				if err := validateSandboxEmail(*email); err != nil {
					return fmt.Errorf("sandbox list: %w", err)
				}
			}
			normalizedTerritory, err := normalizeSandboxTerritoryFilter(*territory)
			if err != nil {
				return fmt.Errorf("sandbox list: %w", err)
			}

			client, err := getASCClient()
			if err != nil {
				return err
			}

			requestCtx, cancel := contextWithTimeout(ctx)
			defer cancel()

			opts := []asc.SandboxTestersOption{
				asc.WithSandboxTestersLimit(*limit),
				asc.WithSandboxTestersNextURL(*next),
			}
			if strings.TrimSpace(*email) != "" {
				opts = append(opts, asc.WithSandboxTestersEmail(*email))
			}
			if normalizedTerritory != "" {
				opts = append(opts, asc.WithSandboxTestersTerritory(normalizedTerritory))
			}

			resp, err := client.GetSandboxTesters(requestCtx, opts...)
			if err != nil {
				return fmt.Errorf("sandbox list: %w", err)
			}

			return printOutput(resp, *output, *pretty)
		},
	}
}

// SandboxCreateCommand returns the sandbox create subcommand.
func SandboxCreateCommand() *ffcli.Command {
	fs := flag.NewFlagSet("create", flag.ExitOnError)

	email := fs.String("email", "", "Tester email address")
	firstName := fs.String("first-name", "", "Tester first name")
	lastName := fs.String("last-name", "", "Tester last name")
	password := fs.String("password", "", "Tester password (8+ chars, uppercase, lowercase, number)")
	confirmPassword := fs.String("confirm-password", "", "Confirm password (must match --password)")
	secretQuestion := fs.String("secret-question", "", "Secret question (6+ chars)")
	secretAnswer := fs.String("secret-answer", "", "Secret answer (6+ chars)")
	birthDate := fs.String("birth-date", "", "Birth date (YYYY-MM-DD)")
	territory := fs.String("territory", "", "App Store territory code (e.g., USA, JPN)")
	output := fs.String("output", "json", "Output format: json (default), table, markdown")
	pretty := fs.Bool("pretty", false, "Pretty-print JSON output")

	return &ffcli.Command{
		Name:       "create",
		ShortUsage: "asc sandbox create [flags]",
		ShortHelp:  "Create a sandbox tester.",
		LongHelp: `Create a new sandbox tester account for in-app purchase testing.

Examples:
  asc sandbox create --email "tester@example.com" --first-name "Test" --last-name "User" --password "Passwordtest1" --confirm-password "Passwordtest1" --secret-question "Question" --secret-answer "Answer" --birth-date "1980-03-01" --territory "USA"`,
		FlagSet:   fs,
		UsageFunc: DefaultUsageFunc,
		Exec: func(ctx context.Context, args []string) error {
			if strings.TrimSpace(*email) == "" {
				fmt.Fprintln(os.Stderr, "Error: --email is required")
				return flag.ErrHelp
			}
			if strings.TrimSpace(*firstName) == "" {
				fmt.Fprintln(os.Stderr, "Error: --first-name is required")
				return flag.ErrHelp
			}
			if strings.TrimSpace(*lastName) == "" {
				fmt.Fprintln(os.Stderr, "Error: --last-name is required")
				return flag.ErrHelp
			}
			if strings.TrimSpace(*password) == "" {
				fmt.Fprintln(os.Stderr, "Error: --password is required")
				return flag.ErrHelp
			}
			if strings.TrimSpace(*confirmPassword) == "" {
				fmt.Fprintln(os.Stderr, "Error: --confirm-password is required")
				return flag.ErrHelp
			}
			if strings.TrimSpace(*secretQuestion) == "" {
				fmt.Fprintln(os.Stderr, "Error: --secret-question is required")
				return flag.ErrHelp
			}
			if strings.TrimSpace(*secretAnswer) == "" {
				fmt.Fprintln(os.Stderr, "Error: --secret-answer is required")
				return flag.ErrHelp
			}
			if strings.TrimSpace(*birthDate) == "" {
				fmt.Fprintln(os.Stderr, "Error: --birth-date is required")
				return flag.ErrHelp
			}
			if strings.TrimSpace(*territory) == "" {
				fmt.Fprintln(os.Stderr, "Error: --territory is required")
				return flag.ErrHelp
			}

			if err := validateSandboxEmail(*email); err != nil {
				return fmt.Errorf("sandbox create: %w", err)
			}
			if err := validateSandboxPassword(*password); err != nil {
				return fmt.Errorf("sandbox create: %w", err)
			}
			if strings.TrimSpace(*confirmPassword) != strings.TrimSpace(*password) {
				return fmt.Errorf("sandbox create: --confirm-password must match --password")
			}
			if err := validateSandboxSecret("--secret-question", *secretQuestion); err != nil {
				return fmt.Errorf("sandbox create: %w", err)
			}
			if err := validateSandboxSecret("--secret-answer", *secretAnswer); err != nil {
				return fmt.Errorf("sandbox create: %w", err)
			}

			normalizedBirthDate, err := normalizeSandboxBirthDate(*birthDate)
			if err != nil {
				return fmt.Errorf("sandbox create: %w", err)
			}
			normalizedTerritory, err := normalizeSandboxTerritory(*territory)
			if err != nil {
				return fmt.Errorf("sandbox create: %w", err)
			}

			client, err := getASCClient()
			if err != nil {
				return err
			}

			requestCtx, cancel := contextWithTimeout(ctx)
			defer cancel()

			attrs := asc.SandboxTesterCreateAttributes{
				FirstName:         strings.TrimSpace(*firstName),
				LastName:          strings.TrimSpace(*lastName),
				Email:             strings.TrimSpace(*email),
				Password:          strings.TrimSpace(*password),
				ConfirmPassword:   strings.TrimSpace(*confirmPassword),
				SecretQuestion:    strings.TrimSpace(*secretQuestion),
				SecretAnswer:      strings.TrimSpace(*secretAnswer),
				BirthDate:         normalizedBirthDate,
				AppStoreTerritory: normalizedTerritory,
			}

			resp, err := client.CreateSandboxTester(requestCtx, attrs)
			if err != nil {
				if asc.IsNotFound(err) {
					return fmt.Errorf("sandbox create: sandbox tester creation is not available via the App Store Connect API for this account")
				}
				return fmt.Errorf("sandbox create: %w", err)
			}

			return printOutput(resp, *output, *pretty)
		},
	}
}

// SandboxGetCommand returns the sandbox get subcommand.
func SandboxGetCommand() *ffcli.Command {
	fs := flag.NewFlagSet("get", flag.ExitOnError)

	testerID := fs.String("id", "", "Sandbox tester ID")
	email := fs.String("email", "", "Tester email address")
	output := fs.String("output", "json", "Output format: json (default), table, markdown")
	pretty := fs.Bool("pretty", false, "Pretty-print JSON output")

	return &ffcli.Command{
		Name:       "get",
		ShortUsage: "asc sandbox get [flags]",
		ShortHelp:  "Get sandbox tester details.",
		LongHelp: `Get sandbox tester details by ID or email.

Examples:
  asc sandbox get --id "SANDBOX_TESTER_ID"
  asc sandbox get --email "tester@example.com"`,
		FlagSet:   fs,
		UsageFunc: DefaultUsageFunc,
		Exec: func(ctx context.Context, args []string) error {
			if strings.TrimSpace(*testerID) == "" && strings.TrimSpace(*email) == "" {
				fmt.Fprintln(os.Stderr, "Error: --id or --email is required")
				return flag.ErrHelp
			}
			if strings.TrimSpace(*email) != "" {
				if err := validateSandboxEmail(*email); err != nil {
					return fmt.Errorf("sandbox get: %w", err)
				}
			}

			client, err := getASCClient()
			if err != nil {
				return err
			}

			requestCtx, cancel := contextWithTimeout(ctx)
			defer cancel()

			var response *asc.SandboxTesterResponse
			if strings.TrimSpace(*testerID) != "" {
				response, err = client.GetSandboxTester(requestCtx, strings.TrimSpace(*testerID))
			} else {
				response, err = findSandboxTesterByEmail(requestCtx, client, strings.TrimSpace(*email))
			}
			if err != nil {
				return fmt.Errorf("sandbox get: %w", err)
			}

			return printOutput(response, *output, *pretty)
		},
	}
}

// SandboxDeleteCommand returns the sandbox delete subcommand.
func SandboxDeleteCommand() *ffcli.Command {
	fs := flag.NewFlagSet("delete", flag.ExitOnError)

	testerID := fs.String("id", "", "Sandbox tester ID")
	email := fs.String("email", "", "Tester email address")
	confirm := fs.Bool("confirm", false, "Confirm deletion")
	output := fs.String("output", "json", "Output format: json (default), table, markdown")
	pretty := fs.Bool("pretty", false, "Pretty-print JSON output")

	return &ffcli.Command{
		Name:       "delete",
		ShortUsage: "asc sandbox delete [flags]",
		ShortHelp:  "Delete a sandbox tester.",
		LongHelp: `Delete a sandbox tester by ID or email.

Examples:
  asc sandbox delete --id "SANDBOX_TESTER_ID" --confirm
  asc sandbox delete --email "tester@example.com" --confirm`,
		FlagSet:   fs,
		UsageFunc: DefaultUsageFunc,
		Exec: func(ctx context.Context, args []string) error {
			if !*confirm {
				fmt.Fprintln(os.Stderr, "Error: --confirm is required")
				return flag.ErrHelp
			}
			if strings.TrimSpace(*testerID) == "" && strings.TrimSpace(*email) == "" {
				fmt.Fprintln(os.Stderr, "Error: --id or --email is required")
				return flag.ErrHelp
			}
			if strings.TrimSpace(*email) != "" {
				if err := validateSandboxEmail(*email); err != nil {
					return fmt.Errorf("sandbox delete: %w", err)
				}
			}

			client, err := getASCClient()
			if err != nil {
				return err
			}

			requestCtx, cancel := contextWithTimeout(ctx)
			defer cancel()

			resolvedID := strings.TrimSpace(*testerID)
			resolvedEmail := strings.TrimSpace(*email)
			if resolvedID == "" {
				resolvedID, err = findSandboxTesterIDByEmail(requestCtx, client, resolvedEmail)
				if err != nil {
					return fmt.Errorf("sandbox delete: %w", err)
				}
			}

			if err := client.DeleteSandboxTester(requestCtx, resolvedID); err != nil {
				if asc.IsNotFound(err) {
					return fmt.Errorf("sandbox delete: sandbox tester deletion is not available via the App Store Connect API for this account")
				}
				return fmt.Errorf("sandbox delete: %w", err)
			}

			result := &asc.SandboxTesterDeleteResult{
				ID:      resolvedID,
				Email:   resolvedEmail,
				Deleted: true,
			}

			return printOutput(result, *output, *pretty)
		},
	}
}
