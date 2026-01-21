package cmd

import (
	"context"
	"errors"
	"flag"
	"io"
	"strings"
	"testing"
)

func TestSandboxCreateValidationErrors(t *testing.T) {
	tests := []struct {
		name    string
		args    []string
		wantErr string
	}{
		{
			name:    "missing email",
			args:    []string{"sandbox", "create", "--first-name", "Test", "--last-name", "User", "--password", "Passwordtest1", "--confirm-password", "Passwordtest1", "--secret-question", "Question", "--secret-answer", "Answer", "--birth-date", "1980-03-01", "--territory", "USA"},
			wantErr: "--email is required",
		},
		{
			name:    "missing first name",
			args:    []string{"sandbox", "create", "--email", "tester@example.com", "--last-name", "User", "--password", "Passwordtest1", "--confirm-password", "Passwordtest1", "--secret-question", "Question", "--secret-answer", "Answer", "--birth-date", "1980-03-01", "--territory", "USA"},
			wantErr: "--first-name is required",
		},
		{
			name:    "missing last name",
			args:    []string{"sandbox", "create", "--email", "tester@example.com", "--first-name", "Test", "--password", "Passwordtest1", "--confirm-password", "Passwordtest1", "--secret-question", "Question", "--secret-answer", "Answer", "--birth-date", "1980-03-01", "--territory", "USA"},
			wantErr: "--last-name is required",
		},
		{
			name:    "missing password",
			args:    []string{"sandbox", "create", "--email", "tester@example.com", "--first-name", "Test", "--last-name", "User", "--confirm-password", "Passwordtest1", "--secret-question", "Question", "--secret-answer", "Answer", "--birth-date", "1980-03-01", "--territory", "USA"},
			wantErr: "--password is required",
		},
		{
			name:    "missing confirm password",
			args:    []string{"sandbox", "create", "--email", "tester@example.com", "--first-name", "Test", "--last-name", "User", "--password", "Passwordtest1", "--secret-question", "Question", "--secret-answer", "Answer", "--birth-date", "1980-03-01", "--territory", "USA"},
			wantErr: "--confirm-password is required",
		},
		{
			name:    "missing secret question",
			args:    []string{"sandbox", "create", "--email", "tester@example.com", "--first-name", "Test", "--last-name", "User", "--password", "Passwordtest1", "--confirm-password", "Passwordtest1", "--secret-answer", "Answer", "--birth-date", "1980-03-01", "--territory", "USA"},
			wantErr: "--secret-question is required",
		},
		{
			name:    "missing secret answer",
			args:    []string{"sandbox", "create", "--email", "tester@example.com", "--first-name", "Test", "--last-name", "User", "--password", "Passwordtest1", "--confirm-password", "Passwordtest1", "--secret-question", "Question", "--birth-date", "1980-03-01", "--territory", "USA"},
			wantErr: "--secret-answer is required",
		},
		{
			name:    "missing birth date",
			args:    []string{"sandbox", "create", "--email", "tester@example.com", "--first-name", "Test", "--last-name", "User", "--password", "Passwordtest1", "--confirm-password", "Passwordtest1", "--secret-question", "Question", "--secret-answer", "Answer", "--territory", "USA"},
			wantErr: "--birth-date is required",
		},
		{
			name:    "missing territory",
			args:    []string{"sandbox", "create", "--email", "tester@example.com", "--first-name", "Test", "--last-name", "User", "--password", "Passwordtest1", "--confirm-password", "Passwordtest1", "--secret-question", "Question", "--secret-answer", "Answer", "--birth-date", "1980-03-01"},
			wantErr: "--territory is required",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			root := RootCommand("1.2.3")
			root.FlagSet.SetOutput(io.Discard)

			stdout, stderr := captureOutput(t, func() {
				if err := root.Parse(test.args); err != nil {
					t.Fatalf("parse error: %v", err)
				}
				err := root.Run(context.Background())
				if !errors.Is(err, flag.ErrHelp) {
					t.Fatalf("expected ErrHelp, got %v", err)
				}
			})

			if stdout != "" {
				t.Fatalf("expected empty stdout, got %q", stdout)
			}
			if !strings.Contains(stderr, test.wantErr) {
				t.Fatalf("expected error %q, got %q", test.wantErr, stderr)
			}
		})
	}
}

func TestSandboxGetValidationErrors(t *testing.T) {
	root := RootCommand("1.2.3")
	root.FlagSet.SetOutput(io.Discard)

	stdout, stderr := captureOutput(t, func() {
		if err := root.Parse([]string{"sandbox", "get"}); err != nil {
			t.Fatalf("parse error: %v", err)
		}
		err := root.Run(context.Background())
		if !errors.Is(err, flag.ErrHelp) {
			t.Fatalf("expected ErrHelp, got %v", err)
		}
	})

	if stdout != "" {
		t.Fatalf("expected empty stdout, got %q", stdout)
	}
	if !strings.Contains(stderr, "--id or --email is required") {
		t.Fatalf("expected error, got %q", stderr)
	}
}

func TestSandboxDeleteValidationErrors(t *testing.T) {
	tests := []struct {
		name    string
		args    []string
		wantErr string
	}{
		{
			name:    "missing confirm",
			args:    []string{"sandbox", "delete", "--id", "tester-1"},
			wantErr: "--confirm is required",
		},
		{
			name:    "missing id and email",
			args:    []string{"sandbox", "delete", "--confirm"},
			wantErr: "--id or --email is required",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			root := RootCommand("1.2.3")
			root.FlagSet.SetOutput(io.Discard)

			stdout, stderr := captureOutput(t, func() {
				if err := root.Parse(test.args); err != nil {
					t.Fatalf("parse error: %v", err)
				}
				err := root.Run(context.Background())
				if !errors.Is(err, flag.ErrHelp) {
					t.Fatalf("expected ErrHelp, got %v", err)
				}
			})

			if stdout != "" {
				t.Fatalf("expected empty stdout, got %q", stdout)
			}
			if !strings.Contains(stderr, test.wantErr) {
				t.Fatalf("expected error %q, got %q", test.wantErr, stderr)
			}
		})
	}
}
