package update

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
)

// Restart runs the updated binary and returns its exit code.
func Restart(executable string, args []string, env []string) (int, error) {
	if executable == "" {
		return 1, errors.New("missing executable path for restart")
	}
	if len(args) == 0 {
		return 1, fmt.Errorf("missing args for restart")
	}
	cmd := exec.Command(executable, args[1:]...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin
	cmd.Env = append(env, skipUpdateEnvVar+"=1")

	if err := cmd.Run(); err != nil {
		var exitErr *exec.ExitError
		if errors.As(err, &exitErr) {
			return exitErr.ExitCode(), nil
		}
		return 1, err
	}
	return 0, nil
}
