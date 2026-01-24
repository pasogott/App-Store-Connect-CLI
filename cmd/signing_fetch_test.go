package cmd

import (
	"context"
	"encoding/base64"
	"errors"
	"flag"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestSigningFetchValidationErrors(t *testing.T) {
	tests := []struct {
		name    string
		args    []string
		wantErr string
	}{
		{
			name:    "missing bundle-id",
			args:    []string{"signing", "fetch", "--profile-type", "IOS_APP_STORE"},
			wantErr: "Error: --bundle-id is required",
		},
		{
			name:    "missing profile-type",
			args:    []string{"signing", "fetch", "--bundle-id", "com.example.app"},
			wantErr: "Error: --profile-type is required",
		},
		{
			name:    "missing device for development profile",
			args:    []string{"signing", "fetch", "--bundle-id", "com.example.app", "--profile-type", "IOS_APP_DEVELOPMENT", "--create-missing"},
			wantErr: "Error: --device is required for development profiles",
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

func TestSigningFetchWriteFiles_NoOverwrite(t *testing.T) {
	dir := t.TempDir()
	profilePath := filepath.Join(dir, "profile.mobileprovision")
	certPath := filepath.Join(dir, "cert.cer")

	profileContent := base64.StdEncoding.EncodeToString([]byte("profile"))
	certContent := base64.StdEncoding.EncodeToString([]byte("certificate"))

	if err := writeProfileFile(profilePath, profileContent); err != nil {
		t.Fatalf("writeProfileFile error: %v", err)
	}
	if err := writeCertificateFile(certPath, certContent); err != nil {
		t.Fatalf("writeCertificateFile error: %v", err)
	}

	if data, err := os.ReadFile(profilePath); err != nil {
		t.Fatalf("read profile error: %v", err)
	} else if string(data) != "profile" {
		t.Fatalf("unexpected profile content: %q", string(data))
	}

	if data, err := os.ReadFile(certPath); err != nil {
		t.Fatalf("read certificate error: %v", err)
	} else if string(data) != "certificate" {
		t.Fatalf("unexpected certificate content: %q", string(data))
	}

	if err := writeProfileFile(profilePath, profileContent); err == nil {
		t.Fatal("expected error when overwriting profile file")
	} else if !strings.Contains(err.Error(), "output file already exists") {
		t.Fatalf("expected overwrite error, got %v", err)
	}

	if err := writeCertificateFile(certPath, certContent); err == nil {
		t.Fatal("expected error when overwriting certificate file")
	} else if !strings.Contains(err.Error(), "output file already exists") {
		t.Fatalf("expected overwrite error, got %v", err)
	}
}
