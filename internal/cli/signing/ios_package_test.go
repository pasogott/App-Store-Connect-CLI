package signing

import (
	"archive/zip"
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
	"errors"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"howett.net/plist"
)

func newIOSPackageFixture(t *testing.T) (string, string) {
	t.Helper()
	directory := t.TempDir()
	app := filepath.Join(directory, "App.app")
	if err := os.Mkdir(app, 0o755); err != nil {
		t.Fatal(err)
	}
	data, err := plist.Marshal(map[string]any{"CFBundleIdentifier": "com.example.app", "CFBundleExecutable": "App", "CFBundleShortVersionString": "1.0", "CFBundleVersion": "1", "CFBundleSupportedPlatforms": []string{"iPhoneOS"}}, plist.XMLFormat)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(app, "Info.plist"), data, 0o644); err != nil {
		t.Fatal(err)
	}
	var executable bytes.Buffer
	for _, v := range []uint32{0xfeedfacf, 0x100000c, 0, 2, 1, 24, 0, 0, 0x32, 24, 2, 0x120000, 0x1b0000, 0} {
		if err := binary.Write(&executable, binary.LittleEndian, v); err != nil {
			t.Fatal(err)
		}
	}
	if err := os.WriteFile(filepath.Join(app, "App"), executable.Bytes(), 0o755); err != nil {
		t.Fatal(err)
	}
	return directory, app
}

func TestIOSPackageAdHocStagesSourceAndPublishesCreateOnly(t *testing.T) {
	if os.PathSeparator == '\\' {
		t.Skip("POSIX signer fixture")
	}
	directory, app := newIOSPackageFixture(t)
	tools := filepath.Join(directory, "bin")
	if err := os.Mkdir(tools, 0o755); err != nil {
		t.Fatal(err)
	}
	// This is the unavailable third-party signer boundary. Live E2B exercises rcodesign.
	script := `#!/bin/sh
test "$1" = --config-file && test "$2" = /dev/null || exit 12
test "$3" = sign && test "$4" = --timestamp-url && test "$5" = none || exit 13
test -z "$ASC_KEY_ID" && test -z "$RCODESIGN_SIGN_PEM_FILE" && test -z "$AWS_SECRET_ACCESS_KEY" && test -z "$LD_PRELOAD" || exit 14
for app; do test "$app" != --entitlements-xml-file || exit 15; done
printf 'signer ran in private stage' > "$app/signer-proof"
`
	if err := os.WriteFile(filepath.Join(tools, "rcodesign"), []byte(script), 0o755); err != nil {
		t.Fatal(err)
	}
	t.Setenv("PATH", tools)
	t.Setenv("ASC_KEY_ID", "not-a-real-key")
	t.Setenv("RCODESIGN_SIGN_PEM_FILE", "must-be-ignored")
	t.Setenv("AWS_SECRET_ACCESS_KEY", "must-be-ignored")
	t.Setenv("LD_PRELOAD", "must-be-ignored")
	destination := filepath.Join(directory, "App.ipa")
	opts := IOSPackageOptions{AppPath: app, IPAPath: destination, AdHoc: true, LogWriter: io.Discard}
	result, err := PackageIOSApp(context.Background(), opts)
	if err != nil {
		t.Fatal(err)
	}
	if !result.Success || result.SigningType != "adHoc" || result.ProfileValidated || result.SignatureVerified || result.AppleAcceptance != "notVerified" {
		t.Fatalf("misleading receipt: %+v", result)
	}
	data, err := os.ReadFile(destination)
	if err != nil {
		t.Fatal(err)
	}
	digest := sha256.Sum256(data)
	if result.SHA256 != hex.EncodeToString(digest[:]) {
		t.Fatal("wrong artifact hash")
	}
	archive, err := zip.NewReader(bytes.NewReader(data), int64(len(data)))
	if err != nil {
		t.Fatal(err)
	}
	marker := false
	for _, entry := range archive.File {
		if !strings.HasPrefix(entry.Name, "Payload/App.app/") {
			t.Fatalf("unexpected entry: %s", entry.Name)
		}
		if entry.Name == "Payload/App.app/signer-proof" {
			marker = true
		}
	}
	if !marker {
		t.Fatal("signed stage was not packaged")
	}
	if _, err := os.Stat(filepath.Join(app, "signer-proof")); !errors.Is(err, os.ErrNotExist) {
		t.Fatal("source app was modified")
	}
	if _, err := PackageIOSApp(context.Background(), opts); err == nil {
		t.Fatal("existing output replaced")
	}
	after, err := os.ReadFile(destination)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(data, after) {
		t.Fatal("existing output changed")
	}
}

func TestIOSPackageRejectsUntrustedProfileBeforeSigning(t *testing.T) {
	directory, app := newIOSPackageFixture(t)
	fixture := newSigningRunFixture(t, signingRunFixtureOptions{})
	original := signingResignNowFn
	signingResignNowFn = func() time.Time { return fixture.now }
	t.Cleanup(func() { signingResignNowFn = original })
	profile := filepath.Join(directory, "untrusted.mobileprovision")
	if err := os.WriteFile(profile, fixture.profile, 0o600); err != nil {
		t.Fatal(err)
	}
	destination := filepath.Join(directory, "App.ipa")
	_, err := PackageIOSApp(context.Background(), IOSPackageOptions{AppPath: app, IPAPath: destination, IdentityPath: "not-read-before-trust", ProfilePath: profile, LogWriter: io.Discard})
	if err == nil {
		t.Fatal("non-Apple profile accepted")
	}
	if _, err := os.Stat(destination); !errors.Is(err, os.ErrNotExist) {
		t.Fatal("untrusted profile produced an artifact")
	}
}

func TestIOSPackageReportsPrivateStageCleanupFailure(t *testing.T) {
	directory, app := newIOSPackageFixture(t)
	original := removeIOSPackageStage
	var retained string
	removeIOSPackageStage = func(path string) error { retained = path; return errors.New("injected cleanup failure") }
	t.Cleanup(func() {
		removeIOSPackageStage = original
		if retained != "" {
			_ = os.RemoveAll(retained)
		}
	})
	t.Setenv("PATH", t.TempDir())
	result, err := PackageIOSApp(context.Background(), IOSPackageOptions{AppPath: app, IPAPath: filepath.Join(directory, "App.ipa"), AdHoc: true, LogWriter: io.Discard})
	if err == nil || !strings.Contains(err.Error(), "private signing stage cleanup failed") || result.Success {
		t.Fatalf("cleanup failure hidden: %+v %v", result, err)
	}
}
