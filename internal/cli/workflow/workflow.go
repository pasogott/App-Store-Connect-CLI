package workflow

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/peterbourgon/ff/v3/ffcli"

	"github.com/rudrankriyam/App-Store-Connect-CLI/internal/cli/shared"
	wf "github.com/rudrankriyam/App-Store-Connect-CLI/internal/workflow"
)

// WorkflowCommand returns the top-level workflow command group.
func WorkflowCommand() *ffcli.Command {
	fs := flag.NewFlagSet("workflow", flag.ExitOnError)

	return &ffcli.Command{
		Name:       "workflow",
		ShortUsage: "asc workflow <subcommand> [flags]",
		ShortHelp:  "Run multi-step automation workflows.",
		LongHelp: `Define named, multi-step automation sequences in .asc/workflow.json.
Each workflow composes existing asc commands and shell commands.

Examples:
  asc workflow list
  asc workflow validate
  asc workflow run beta
  asc workflow run beta SUBMIT_BETA:true
  asc workflow run release VERSION:2.1.0
  asc workflow run beta --dry-run`,
		FlagSet:   fs,
		UsageFunc: shared.DefaultUsageFunc,
		Subcommands: []*ffcli.Command{
			workflowRunCommand(),
			workflowValidateCommand(),
			workflowListCommand(),
		},
		Exec: func(_ context.Context, _ []string) error {
			return flag.ErrHelp
		},
	}
}

func workflowRunCommand() *ffcli.Command {
	fs := flag.NewFlagSet("workflow run", flag.ExitOnError)
	filePath := fs.String("file", wf.DefaultPath, "Path to workflow.json")
	dryRun := fs.Bool("dry-run", false, "Preview steps without executing")
	pretty := fs.Bool("pretty", false, "Pretty-print JSON output")

	return &ffcli.Command{
		Name:       "run",
		ShortUsage: "asc workflow run <name> [KEY:VALUE ...] [flags]",
		ShortHelp:  "Run a named workflow.",
		FlagSet:    fs,
		UsageFunc:  shared.DefaultUsageFunc,
		Exec: func(ctx context.Context, args []string) error {
			if len(args) == 0 {
				return shared.UsageError("workflow name is required")
			}

			workflowName := args[0]
			paramArgs := args[1:]

			absPath, err := filepath.Abs(strings.TrimSpace(*filePath))
			if err != nil {
				return fmt.Errorf("workflow run: resolve path: %w", err)
			}

			def, err := wf.Load(absPath)
			if err != nil {
				return fmt.Errorf("workflow run: %w", err)
			}

			params, err := wf.ParseParams(paramArgs)
			if err != nil {
				return shared.UsageErrorf("%s", err)
			}

			result, err := wf.Run(ctx, def, wf.RunOptions{
				WorkflowName: workflowName,
				Params:       params,
				DryRun:       *dryRun,
				// Keep stdout machine-parseable JSON; stream step output to stderr.
				Stdout: os.Stderr,
				Stderr: os.Stderr,
			})
			if err != nil {
				if result != nil {
					_ = printJSON(os.Stdout, result, *pretty)
				}
				return shared.NewReportedError(err)
			}

			return printJSON(os.Stdout, result, *pretty)
		},
	}
}

func workflowValidateCommand() *ffcli.Command {
	fs := flag.NewFlagSet("workflow validate", flag.ExitOnError)
	filePath := fs.String("file", wf.DefaultPath, "Path to workflow.json")
	pretty := fs.Bool("pretty", false, "Pretty-print JSON output")

	return &ffcli.Command{
		Name:       "validate",
		ShortUsage: "asc workflow validate [flags]",
		ShortHelp:  "Validate workflow.json for errors and cycles.",
		FlagSet:    fs,
		UsageFunc:  shared.DefaultUsageFunc,
		Exec: func(_ context.Context, _ []string) error {
			absPath, err := filepath.Abs(strings.TrimSpace(*filePath))
			if err != nil {
				return fmt.Errorf("workflow validate: resolve path: %w", err)
			}

			def, err := wf.LoadUnvalidated(absPath)
			if err != nil {
				return fmt.Errorf("workflow validate: %w", err)
			}

			errs := wf.Validate(def)

			type validationResult struct {
				Valid  bool                  `json:"valid"`
				Errors []*wf.ValidationError `json:"errors,omitempty"`
			}
			result := validationResult{
				Valid:  len(errs) == 0,
				Errors: errs,
			}

			if printErr := printJSON(os.Stdout, result, *pretty); printErr != nil {
				return printErr
			}

			if !result.Valid {
				return shared.NewReportedError(
					fmt.Errorf("workflow validate: found %d error(s)", len(errs)),
				)
			}
			return nil
		},
	}
}

func workflowListCommand() *ffcli.Command {
	fs := flag.NewFlagSet("workflow list", flag.ExitOnError)
	filePath := fs.String("file", wf.DefaultPath, "Path to workflow.json")
	pretty := fs.Bool("pretty", false, "Pretty-print JSON output")

	return &ffcli.Command{
		Name:       "list",
		ShortUsage: "asc workflow list [flags]",
		ShortHelp:  "List available workflows.",
		FlagSet:    fs,
		UsageFunc:  shared.DefaultUsageFunc,
		Exec: func(_ context.Context, _ []string) error {
			absPath, err := filepath.Abs(strings.TrimSpace(*filePath))
			if err != nil {
				return fmt.Errorf("workflow list: resolve path: %w", err)
			}

			def, err := wf.LoadUnvalidated(absPath)
			if err != nil {
				return fmt.Errorf("workflow list: %w", err)
			}

			type workflowInfo struct {
				Name        string `json:"name"`
				Description string `json:"description,omitempty"`
				Private     bool   `json:"private,omitempty"`
				StepCount   int    `json:"step_count"`
			}

			workflows := make([]workflowInfo, 0, len(def.Workflows))
			for name, w := range def.Workflows {
				workflows = append(workflows, workflowInfo{
					Name:        name,
					Description: w.Description,
					Private:     w.Private,
					StepCount:   len(w.Steps),
				})
			}

			sort.Slice(workflows, func(i, j int) bool {
				return workflows[i].Name < workflows[j].Name
			})

			return printJSON(os.Stdout, workflows, *pretty)
		},
	}
}

// printJSON encodes data as JSON to the writer.
func printJSON(w io.Writer, data any, pretty bool) error {
	enc := json.NewEncoder(w)
	if pretty {
		enc.SetIndent("", "  ")
	}
	return enc.Encode(data)
}
