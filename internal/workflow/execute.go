package workflow

import (
	"context"
	"fmt"
	"io"
	"os"
	"strings"
	"time"
)

// MaxCallDepth is the maximum nesting depth for sub-workflow calls.
const MaxCallDepth = 16

// RunOptions configures a workflow execution.
type RunOptions struct {
	WorkflowName string
	Params       map[string]string
	DryRun       bool
	Stdout       io.Writer
	Stderr       io.Writer
}

// StepResult records one executed step.
type StepResult struct {
	Index          int    `json:"index"`
	Name           string `json:"name,omitempty"`
	Command        string `json:"command,omitempty"`
	Workflow       string `json:"workflow,omitempty"`
	ParentWorkflow string `json:"parent_workflow,omitempty"`
	Status         string `json:"status"`
	DurationMS     int64  `json:"duration_ms"`
	Error          string `json:"error,omitempty"`
}

// RunResult is the structured output of a workflow execution.
type RunResult struct {
	Workflow   string       `json:"workflow"`
	Status     string       `json:"status"`
	Steps      []StepResult `json:"steps"`
	DurationMS int64        `json:"duration_ms"`
}

// Run executes a named workflow from the Definition.
func Run(ctx context.Context, def *Definition, opts RunOptions) (*RunResult, error) {
	if opts.Stdout == nil {
		opts.Stdout = os.Stdout
	}
	if opts.Stderr == nil {
		opts.Stderr = os.Stderr
	}

	wf, ok := def.Workflows[opts.WorkflowName]
	if !ok {
		return nil, fmt.Errorf("workflow: unknown workflow %q", opts.WorkflowName)
	}
	if wf.Private {
		return nil, fmt.Errorf("workflow: %q is private and cannot be run directly", opts.WorkflowName)
	}

	env := mergeEnv(def.Env, wf.Env, opts.Params)

	result := &RunResult{
		Workflow: opts.WorkflowName,
		Steps:    make([]StepResult, 0),
	}
	start := time.Now()

	// before_all hook
	if err := runHook(ctx, def.BeforeAll, env, opts.DryRun, opts.Stdout, opts.Stderr); err != nil {
		result.Status = "error"
		result.DurationMS = time.Since(start).Milliseconds()
		_ = runHook(ctx, def.Error, env, opts.DryRun, opts.Stdout, opts.Stderr)
		return result, fmt.Errorf("workflow: before_all hook failed: %w", err)
	}

	// Execute steps
	err := executeSteps(ctx, def, opts.WorkflowName, wf.Steps, env, 0, opts, result)
	result.DurationMS = time.Since(start).Milliseconds()

	if err != nil {
		result.Status = "error"
		if hook := strings.TrimSpace(def.Error); hook != "" {
			_ = runHook(ctx, hook, env, opts.DryRun, opts.Stdout, opts.Stderr)
		}
		return result, err
	}

	// after_all hook
	if hookErr := runHook(ctx, def.AfterAll, env, opts.DryRun, opts.Stdout, opts.Stderr); hookErr != nil {
		result.Status = "error"
		result.DurationMS = time.Since(start).Milliseconds()
		return result, fmt.Errorf("workflow: after_all hook failed: %w", hookErr)
	}

	result.Status = "ok"
	return result, nil
}

func executeSteps(ctx context.Context, def *Definition, workflowName string, steps []Step, env map[string]string, depth int, opts RunOptions, result *RunResult) error {
	for i, step := range steps {
		idx := i + 1
		stepStart := time.Now()

		sr := StepResult{
			Index:   idx,
			Name:    step.Name,
			Command: step.Run,
		}
		if workflowName != opts.WorkflowName {
			sr.ParentWorkflow = workflowName
		}

		// Check conditional
		if ifVar := strings.TrimSpace(step.If); ifVar != "" {
			val, ok := env[ifVar]
			if !ok {
				val = os.Getenv(ifVar)
			}
			if !isTruthy(val) {
				sr.Status = "skipped"
				sr.DurationMS = time.Since(stepStart).Milliseconds()
				result.Steps = append(result.Steps, sr)
				continue
			}
		}

		if ref := strings.TrimSpace(step.Workflow); ref != "" {
			sr.Workflow = ref

			if depth+1 > MaxCallDepth {
				sr.Status = "error"
				sr.Error = fmt.Sprintf("max call depth %d exceeded", MaxCallDepth)
				sr.DurationMS = time.Since(stepStart).Milliseconds()
				result.Steps = append(result.Steps, sr)
				return fmt.Errorf("workflow: %s step %d: max call depth %d exceeded", workflowName, idx, MaxCallDepth)
			}

			subWf, ok := def.Workflows[ref]
			if !ok {
				sr.Status = "error"
				sr.Error = fmt.Sprintf("unknown workflow %q", ref)
				sr.DurationMS = time.Since(stepStart).Milliseconds()
				result.Steps = append(result.Steps, sr)
				return fmt.Errorf("workflow: %s step %d: unknown workflow %q", workflowName, idx, ref)
			}

			// Sub-workflow env provides defaults; call-site "with" should override.
			subEnv := mergeEnv(env, subWf.Env, step.With)

			if opts.DryRun {
				fmt.Fprintf(opts.Stderr, "[dry-run] step %d: workflow %s\n", idx, ref)
			}

			if err := executeSteps(ctx, def, ref, subWf.Steps, subEnv, depth+1, opts, result); err != nil {
				return err
			}
			continue
		}

		// run: step
		if opts.DryRun {
			expanded := expandEnv(step.Run, env)
			fmt.Fprintf(opts.Stderr, "[dry-run] step %d: %s\n", idx, expanded)
			sr.Status = "dry-run"
			sr.DurationMS = time.Since(stepStart).Milliseconds()
			result.Steps = append(result.Steps, sr)
			continue
		}

		if err := runShellCommand(ctx, step.Run, env, opts.Stdout, opts.Stderr); err != nil {
			sr.Status = "error"
			sr.Error = err.Error()
			sr.DurationMS = time.Since(stepStart).Milliseconds()
			result.Steps = append(result.Steps, sr)
			return fmt.Errorf("workflow: %s step %d: %w", workflowName, idx, err)
		}

		sr.Status = "ok"
		sr.DurationMS = time.Since(stepStart).Milliseconds()
		result.Steps = append(result.Steps, sr)
	}
	return nil
}
