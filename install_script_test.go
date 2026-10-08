package main

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

const installStubBinaryContents = "fake-binary\n"

// installScriptAsset mirrors the OS/architecture mapping in install.sh for the
// platform the test runs on.
func installScriptAsset(t *testing.T, version, goos string) string {
	t.Helper()

	var osLabel string
	switch goos {
	case "darwin":
		osLabel = "macOS"
	case "linux":
		osLabel = "linux"
	default:
		t.Skipf("install.sh does not support GOOS %q", goos)
	}

	var archLabel string
	switch runtime.GOARCH {
	case "amd64":
		archLabel = "amd64"
	case "arm64":
		archLabel = "arm64"
	default:
		t.Skipf("install.sh does not support GOARCH %q", runtime.GOARCH)
	}

	return fmt.Sprintf("asc_%s_%s_%s", version, osLabel, archLabel)
}

// runInstallScript executes install.sh with a stubbed curl so no network
// access happens. checksumsMode controls what the stub serves for
// checksums.txt: valid, wrong, unlisted, or missing.
func runInstallScript(t *testing.T, checksumsMode string, extraEnv ...string) (output, installedBinary string, err error) {
	t.Helper()
	output, installedBinary, _, err = runInstallScriptPlatform(t, checksumsMode, runtime.GOOS, "13.0", "go 1.27.1", extraEnv...)
	return output, installedBinary, err
}

func runInstallScriptPlatform(t *testing.T, checksumsMode, goos, macOSVersion, goMod string, extraEnv ...string) (output, installedBinary, stderr string, err error) {
	t.Helper()

	repoRoot, wdErr := os.Getwd()
	if wdErr != nil {
		t.Fatalf("getwd: %v", wdErr)
	}

	version := "0.0.1"
	asset := installScriptAsset(t, version, goos)
	sum := sha256.Sum256([]byte(installStubBinaryContents))
	checksum := hex.EncodeToString(sum[:])

	workDir := t.TempDir()
	stubDir := filepath.Join(workDir, "stub-bin")
	installDir := filepath.Join(workDir, "install-bin")
	if mkErr := os.MkdirAll(stubDir, 0o755); mkErr != nil {
		t.Fatalf("mkdir stub dir: %v", mkErr)
	}

	curlStub := `#!/usr/bin/env bash
set -euo pipefail
printf '%s\n' "$*" >> "${STUB_STATE_DIR}/curl-requests"
out=""
url=""
prev=""
for arg in "$@"; do
  if [ "${prev}" = "-o" ]; then
    out="${arg}"
  fi
  case "${arg}" in
    http://*|https://*) url="${arg}" ;;
  esac
  prev="${arg}"
done

retry_target=""
case "${url}" in
  *_checksums.txt) retry_target="checksums" ;;
  */releases/download/*) retry_target="binary" ;;
esac
if [ -n "${retry_target}" ] && [ "${STUB_FAIL_ONCE:-}" = "${retry_target}" ]; then
  attempt_file="${STUB_STATE_DIR}/${retry_target}-attempted"
  if [ ! -e "${attempt_file}" ]; then
    touch "${attempt_file}"
    exit 22
  fi
fi

case "${url}" in
  */go.mod)
    if [ "${STUB_GO_MOD}" = "missing" ]; then exit 22; fi
    printf '%s\n' "${STUB_GO_MOD}"
    ;;
  */releases/latest)
    printf '%s' "https://github.com/rorkai/App-Store-Connect-CLI/releases/tag/${STUB_VERSION}"
    ;;
  *_checksums.txt)
    case "${STUB_CHECKSUMS_MODE}" in
      missing) exit 22 ;;
      valid) printf '%s  %s\n' "${STUB_SHA256}" "${STUB_ASSET}" > "${out}" ;;
      unlisted) printf '%s  %s\n' "${STUB_SHA256}" "${STUB_ASSET//./x}" > "${out}" ;;
      wrong) printf '%s  %s\n' "0000000000000000000000000000000000000000000000000000000000000000" "${STUB_ASSET}" > "${out}" ;;
      *) echo "unknown STUB_CHECKSUMS_MODE" >&2; exit 1 ;;
    esac
    ;;
  *)
    printf 'fake-binary\n' > "${out}"
    ;;
esac
`
	if writeErr := os.WriteFile(filepath.Join(stubDir, "curl"), []byte(curlStub), 0o755); writeErr != nil {
		t.Fatalf("write curl stub: %v", writeErr)
	}
	if writeErr := os.WriteFile(filepath.Join(stubDir, "sleep"), []byte("#!/usr/bin/env bash\nexit 0\n"), 0o755); writeErr != nil {
		t.Fatalf("write sleep stub: %v", writeErr)
	}

	unameOS := "Linux"
	if goos == "darwin" {
		unameOS = "Darwin"
	}
	unameStub := fmt.Sprintf("#!/usr/bin/env bash\ncase \"$1\" in\n-s) echo %q ;;\n-m) echo %q ;;\nesac\n", unameOS, runtime.GOARCH)
	if writeErr := os.WriteFile(filepath.Join(stubDir, "uname"), []byte(unameStub), 0o755); writeErr != nil {
		t.Fatalf("write uname stub: %v", writeErr)
	}
	swVersStub := "#!/usr/bin/env bash\nprintf '%s\\n' \"$STUB_MACOS_VERSION\"\n"
	if writeErr := os.WriteFile(filepath.Join(stubDir, "sw_vers"), []byte(swVersStub), 0o755); writeErr != nil {
		t.Fatalf("write sw_vers stub: %v", writeErr)
	}

	cmd := exec.Command("bash", filepath.Join(repoRoot, "install.sh"))
	cmd.Dir = workDir
	cmd.Env = append(
		os.Environ(),
		"PATH="+stubDir+string(os.PathListSeparator)+os.Getenv("PATH"),
		"INSTALL_DIR="+installDir,
		"STUB_VERSION="+version,
		"STUB_ASSET="+asset,
		"STUB_SHA256="+checksum,
		"STUB_CHECKSUMS_MODE="+checksumsMode,
		"STUB_STATE_DIR="+workDir,
		"STUB_MACOS_VERSION="+macOSVersion,
		"STUB_GO_MOD="+goMod,
	)
	cmd.Env = append(cmd.Env, extraEnv...)

	var stdout, stderrBuffer bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderrBuffer
	runErr := cmd.Run()
	return stdout.String() + stderrBuffer.String(), filepath.Join(installDir, "asc"), stderrBuffer.String(), runErr
}

func TestInstallScriptRetriesTransientBinaryDownloadFailure(t *testing.T) {
	output, installedBinary, err := runInstallScript(t, "valid", "STUB_FAIL_ONCE=binary")
	if err != nil {
		t.Fatalf("expected transient binary download failure to recover: %v\n%s", err, output)
	}
	if !strings.Contains(output, "Download failed; retrying (2/3)") {
		t.Fatalf("expected retry diagnostic, got:\n%s", output)
	}
	if _, statErr := os.Stat(installedBinary); statErr != nil {
		t.Fatalf("expected binary installed at %s: %v\n%s", installedBinary, statErr, output)
	}
}

func TestInstallScriptRetriesTransientChecksumDownloadFailure(t *testing.T) {
	output, installedBinary, err := runInstallScript(t, "valid", "STUB_FAIL_ONCE=checksums")
	if err != nil {
		t.Fatalf("expected transient checksum download failure to recover: %v\n%s", err, output)
	}
	if !strings.Contains(output, "Download failed; retrying (2/3)") {
		t.Fatalf("expected retry diagnostic, got:\n%s", output)
	}
	if _, statErr := os.Stat(installedBinary); statErr != nil {
		t.Fatalf("expected binary installed at %s: %v\n%s", installedBinary, statErr, output)
	}
}

func TestInstallScriptSyntaxIsValid(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("install.sh targets unix shells")
	}
	repoRoot, err := os.Getwd()
	if err != nil {
		t.Fatalf("getwd: %v", err)
	}
	cmd := exec.Command("bash", "-n", filepath.Join(repoRoot, "install.sh"))
	if output, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("bash -n install.sh failed: %v\n%s", err, output)
	}
}

func TestInstallScriptVerifiesChecksumAndInstalls(t *testing.T) {
	output, installedBinary, err := runInstallScript(t, "valid")
	if err != nil {
		t.Fatalf("install.sh failed: %v\n%s", err, output)
	}
	if !strings.Contains(output, "Checksum verified.") {
		t.Fatalf("expected checksum verification confirmation, got:\n%s", output)
	}
	contents, readErr := os.ReadFile(installedBinary)
	if readErr != nil {
		t.Fatalf("expected installed binary at %s: %v\n%s", installedBinary, readErr, output)
	}
	if string(contents) != installStubBinaryContents {
		t.Fatalf("installed binary contents mismatch: %q", contents)
	}
}

func TestInstallScriptFailsClosedWhenChecksumsMissing(t *testing.T) {
	output, installedBinary, err := runInstallScript(t, "missing")
	if err == nil {
		t.Fatalf("expected install.sh to fail without checksums.txt, got success:\n%s", output)
	}
	if !strings.Contains(output, "Refusing to install without SHA-256 checksum verification.") {
		t.Fatalf("expected fail-closed error message, got:\n%s", output)
	}
	if !strings.Contains(output, "ASC_INSTALL_INSECURE=1") {
		t.Fatalf("expected override hint in error output, got:\n%s", output)
	}
	if _, statErr := os.Stat(installedBinary); !os.IsNotExist(statErr) {
		t.Fatalf("expected no binary installed, stat err=%v\n%s", statErr, output)
	}
}

func TestInstallScriptFailsClosedWhenAssetUnlisted(t *testing.T) {
	output, installedBinary, err := runInstallScript(t, "unlisted")
	if err == nil {
		t.Fatalf("expected install.sh to fail when asset is missing from checksums.txt, got success:\n%s", output)
	}
	if !strings.Contains(output, "Refusing to install without SHA-256 checksum verification.") {
		t.Fatalf("expected fail-closed error message, got:\n%s", output)
	}
	if _, statErr := os.Stat(installedBinary); !os.IsNotExist(statErr) {
		t.Fatalf("expected no binary installed, stat err=%v\n%s", statErr, output)
	}
}

func TestInstallScriptInsecureOverrideAllowsMissingChecksums(t *testing.T) {
	output, installedBinary, err := runInstallScript(t, "missing", "ASC_INSTALL_INSECURE=1")
	if err != nil {
		t.Fatalf("expected ASC_INSTALL_INSECURE=1 to allow install, got: %v\n%s", err, output)
	}
	if !strings.Contains(output, "installing WITHOUT checksum verification") {
		t.Fatalf("expected loud insecure-install warning, got:\n%s", output)
	}
	if _, statErr := os.Stat(installedBinary); statErr != nil {
		t.Fatalf("expected binary installed at %s: %v\n%s", installedBinary, statErr, output)
	}
}

func TestInstallScriptChecksumMismatchFailsEvenWithInsecureOverride(t *testing.T) {
	output, installedBinary, err := runInstallScript(t, "wrong", "ASC_INSTALL_INSECURE=1")
	if err == nil {
		t.Fatalf("expected install.sh to fail on checksum mismatch, got success:\n%s", output)
	}
	if !strings.Contains(output, "Checksum verification failed") {
		t.Fatalf("expected checksum mismatch error, got:\n%s", output)
	}
	if _, statErr := os.Stat(installedBinary); !os.IsNotExist(statErr) {
		t.Fatalf("expected no binary installed, stat err=%v\n%s", statErr, output)
	}
}

func TestInstallScriptMacOSReleaseCompatibility(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("install.sh targets unix shells")
	}
	for _, tt := range []struct {
		name, goos, macOSVersion, goMod, wantError string
		wantMetadata                               bool
	}{
		{"monterey-compatible-release", "darwin", "12.7.6", "module example.com/asc\n\ngo 1.26.8\n", "", true},
		{"monterey-new-release", "darwin", "12.7.6", "go 1.27.1", "requires macOS 13", true},
		{"monterey-new-toolchain", "darwin", "12.7.6", "go 1.26.8\ntoolchain go1.27.1", "requires macOS 13", true},
		{"monterey-future-go", "darwin", "12.7.6", "go 1.28.0", "requires macOS 13", true},
		{"monterey-invalid-toolchain", "darwin", "12.7.6", "go 1.26.8\ntoolchain go1.banana", "Could not determine macOS compatibility", true},
		{"monterey-missing-metadata", "darwin", "12.7.6", "missing", "Could not determine macOS compatibility", true},
		{"monterey-invalid-metadata", "darwin", "12.7.6", "not a go.mod file", "Could not determine macOS compatibility", true},
		{"monterey-default-go", "darwin", "12.7.6", "go default", "Could not determine macOS compatibility", true},
		{"monterey-default-toolchain", "darwin", "12.7.6", "go 1.26.8\ntoolchain default", "", true},
		{"monterey-invalid-version", "darwin", "12.7.6", "go banana", "Could not determine macOS compatibility", true},
		{"ventura", "darwin", "13.0.1", "go 1.27.1", "", false},
		{"newer-macos", "darwin", "26.0", "go 1.27.1", "", false},
		{"linux", "linux", "", "go 1.27.1", "", false},
		{"invalid-macos-version", "darwin", "unknown", "go 1.27.1", "Could not determine macOS version", false},
	} {
		t.Run(tt.name, func(t *testing.T) {
			output, installedBinary, stderr, err := runInstallScriptPlatform(t, "valid", tt.goos, tt.macOSVersion, tt.goMod)
			if tt.wantError == "" {
				if err != nil {
					t.Fatalf("install failed: %v\n%s", err, output)
				}
				if contents, readErr := os.ReadFile(installedBinary); readErr != nil || string(contents) != installStubBinaryContents {
					t.Fatalf("expected installed binary, contents=%q err=%v\n%s", contents, readErr, output)
				}
			} else {
				if err == nil || !strings.Contains(stderr, tt.wantError) {
					t.Fatalf("expected %q, err=%v\n%s", tt.wantError, err, output)
				}
				if _, statErr := os.Stat(filepath.Dir(installedBinary)); !os.IsNotExist(statErr) {
					t.Fatalf("unsupported release created install directory: %v", statErr)
				}
			}
			requests, readErr := os.ReadFile(filepath.Join(filepath.Dir(filepath.Dir(installedBinary)), "curl-requests"))
			if readErr != nil && !os.IsNotExist(readErr) {
				t.Fatal(readErr)
			}
			metadataURL := "https://raw.githubusercontent.com/rorkai/App-Store-Connect-CLI/0.0.1/go.mod"
			if got := strings.Contains(string(requests), metadataURL); got != tt.wantMetadata {
				t.Fatalf("release metadata fetched=%v, want %v: %s", got, tt.wantMetadata, requests)
			}
			if tt.wantError != "" && strings.Contains(string(requests), "/releases/download/") {
				t.Fatalf("downloaded unsupported release: %s", requests)
			}
		})
	}
}
