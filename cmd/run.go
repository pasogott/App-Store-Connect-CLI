package cmd

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"os"
	"time"

	"github.com/rudrankriyam/App-Store-Connect-CLI/internal/cli/shared"
	"github.com/rudrankriyam/App-Store-Connect-CLI/internal/cli/shared/errfmt"
	"github.com/rudrankriyam/App-Store-Connect-CLI/internal/update"
)

// Run executes the CLI using the provided args (not including argv[0]) and version string.
// It returns the intended process exit code.
func Run(args []string, versionInfo string) int {
	root := RootCommand(versionInfo)
	defer CleanupTempPrivateKeys()

	if err := root.Parse(args); err != nil {
		if err == flag.ErrHelp {
			return 0
		}
		fmt.Fprint(os.Stderr, errfmt.FormatStderr(err))
		return 1
	}

	if versionRequested {
		if err := root.Run(context.Background()); err != nil {
			if errors.Is(err, flag.ErrHelp) {
				return 1
			}
			fmt.Fprint(os.Stderr, errfmt.FormatStderr(err))
			return 1
		}
		return 0
	}

	updateResult, err := update.CheckAndUpdate(context.Background(), update.Options{
		CurrentVersion: versionInfo,
		AutoUpdate:     true,
		NoUpdate:       shared.NoUpdate(),
		Output:         os.Stderr,
		ShowProgress:   shared.ProgressEnabled(),
		CheckInterval:  24 * time.Hour,
	})
	if err != nil {
		fmt.Fprintf(os.Stderr, "Update check failed: %v\n", err)
	}
	if updateResult.Updated {
		exitCode, restartErr := update.Restart(updateResult.ExecutablePath, os.Args, os.Environ())
		if restartErr != nil {
			fmt.Fprintf(os.Stderr, "Restart failed after update: %v\n", restartErr)
		} else {
			return exitCode
		}
	}

	if err := root.Run(context.Background()); err != nil {
		var reported ReportedError
		if errors.As(err, &reported) {
			return 1
		}
		if errors.Is(err, flag.ErrHelp) {
			return 1
		}
		fmt.Fprint(os.Stderr, errfmt.FormatStderr(err))
		return 1
	}

	return 0
}
