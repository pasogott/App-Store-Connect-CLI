package install

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/peterbourgon/ff/v3/ffcli"

	"github.com/rudrankriyam/App-Store-Connect-CLI/internal/cli/shared"
)

const defaultSkillsPackage = "rudrankriyam/asc-skills"

var (
	lookupNpx  = exec.LookPath
	runCommand = defaultRunCommand
)

// InstallCommand returns the install command factory.
func InstallCommand() *ffcli.Command {
	fs := flag.NewFlagSet("install", flag.ExitOnError)

	return &ffcli.Command{
		Name:       "install",
		ShortUsage: "asc install <subcommand> [flags]",
		ShortHelp:  "Install optional ASC components.",
		LongHelp: `Install optional ASC components.

Subcommands:
  skills    Install the App Store Connect CLI skill pack.`,
		FlagSet:   fs,
		UsageFunc: shared.DefaultUsageFunc,
		Subcommands: []*ffcli.Command{
			InstallSkillsCommand(),
		},
		Exec: func(ctx context.Context, args []string) error {
			if len(args) == 0 {
				return flag.ErrHelp
			}
			fmt.Fprintf(os.Stderr, "Unknown subcommand: %s\n\n", args[0])
			return flag.ErrHelp
		},
	}
}

// InstallSkillsCommand returns the install skills subcommand.
func InstallSkillsCommand() *ffcli.Command {
	fs := flag.NewFlagSet("install skills", flag.ExitOnError)
	packageName := fs.String("package", defaultSkillsPackage, "NPM package name or repo for the skill pack")

	return &ffcli.Command{
		Name:       "skills",
		ShortUsage: "asc install skills [flags]",
		ShortHelp:  "Install the App Store Connect CLI skill pack.",
		LongHelp: `Install the App Store Connect CLI skill pack.

Examples:
  asc install skills
  asc install skills --package "rudrankriyam/asc-skills"`,
		FlagSet:   fs,
		UsageFunc: shared.DefaultUsageFunc,
		Exec: func(ctx context.Context, args []string) error {
			if err := installSkills(ctx, *packageName); err != nil {
				return fmt.Errorf("install skills: %w", err)
			}
			return nil
		},
	}
}

func installSkills(ctx context.Context, pkg string) error {
	pkg = strings.TrimSpace(pkg)
	if pkg == "" {
		return fmt.Errorf("--package is required")
	}

	path, err := lookupNpx("npx")
	if err != nil {
		return fmt.Errorf("npx not found; install Node.js to continue")
	}

	return runCommand(ctx, path, "--yes", "add-skill", pkg)
}

func defaultRunCommand(ctx context.Context, name string, args ...string) error {
	cmd := exec.CommandContext(ctx, name, args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin
	return cmd.Run()
}
