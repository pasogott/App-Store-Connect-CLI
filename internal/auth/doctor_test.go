package auth

import (
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/rudrankriyam/App-Store-Connect-CLI/internal/config"
)

func TestDoctorConfigPermissionsWarning(t *testing.T) {
	t.Setenv("ASC_BYPASS_KEYCHAIN", "1")

	configPath := filepath.Join(t.TempDir(), "config.json")
	if err := os.WriteFile(configPath, []byte("{}"), 0o644); err != nil {
		t.Fatalf("write config error: %v", err)
	}
	t.Setenv("ASC_CONFIG_PATH", configPath)

	report := Doctor(DoctorOptions{})
	section := findDoctorSection(t, report, "Storage")
	if !sectionHasStatus(section, DoctorWarn, "Config file permissions") {
		t.Fatalf("expected config permissions warning, got %#v", section.Checks)
	}

	Doctor(DoctorOptions{Fix: true})
	info, err := os.Stat(configPath)
	if err != nil {
		t.Fatalf("stat config error: %v", err)
	}
	if info.Mode().Perm()&0o077 != 0 {
		t.Fatalf("expected config permissions fixed to 0600, got %#o", info.Mode().Perm())
	}
}

func TestDoctorTempFilesWarns(t *testing.T) {
	t.Setenv("ASC_BYPASS_KEYCHAIN", "1")
	t.Setenv("ASC_CONFIG_PATH", filepath.Join(t.TempDir(), "config.json"))

	tempFile, err := os.CreateTemp(os.TempDir(), "asc-key-*.p8")
	if err != nil {
		t.Fatalf("CreateTemp() error: %v", err)
	}
	tempFile.Close()
	t.Cleanup(func() {
		_ = os.Remove(tempFile.Name())
	})

	report := Doctor(DoctorOptions{})
	section := findDoctorSection(t, report, "Temp Files")
	if !sectionHasStatus(section, DoctorWarn, "orphaned temp key file") {
		t.Fatalf("expected temp file warning, got %#v", section.Checks)
	}
}

func TestDoctorPrivateKeyPermissionsFix(t *testing.T) {
	t.Setenv("ASC_BYPASS_KEYCHAIN", "1")

	tempDir := t.TempDir()
	keyPath := filepath.Join(tempDir, "AuthKey.p8")
	writeECDSAPEM(t, keyPath, 0o600, true)
	if err := os.Chmod(keyPath, 0o644); err != nil {
		t.Fatalf("chmod key error: %v", err)
	}

	cfg := &config.Config{
		DefaultKeyName: "test",
		Keys: []config.Credential{
			{
				Name:           "test",
				KeyID:          "KEY123",
				IssuerID:       "ISS456",
				PrivateKeyPath: keyPath,
			},
		},
	}
	configPath := filepath.Join(tempDir, "config.json")
	if err := config.SaveAt(configPath, cfg); err != nil {
		t.Fatalf("save config error: %v", err)
	}
	t.Setenv("ASC_CONFIG_PATH", configPath)

	report := Doctor(DoctorOptions{Fix: true})
	section := findDoctorSection(t, report, "Private Keys")
	if !sectionHasStatus(section, DoctorOK, "permissions fixed to 0600") {
		t.Fatalf("expected private key permissions fix, got %#v", section.Checks)
	}
}

func TestDoctorMigrationHintsDetected(t *testing.T) {
	repo := t.TempDir()
	if err := os.Mkdir(filepath.Join(repo, ".git"), 0o755); err != nil {
		t.Fatalf("create .git error: %v", err)
	}
	fastlaneDir := filepath.Join(repo, "fastlane")
	if err := os.MkdirAll(fastlaneDir, 0o755); err != nil {
		t.Fatalf("mkdir fastlane error: %v", err)
	}

	secretValue := "SECRET_TOKEN_123"
	appfile := `app_identifier "com.example.app"
apple_id "user@example.com"
team_id "TEAM123"
`
	if err := os.WriteFile(filepath.Join(fastlaneDir, "Appfile"), []byte(appfile), 0o644); err != nil {
		t.Fatalf("write Appfile error: %v", err)
	}
	fastfile := `lane :beta do
  app_store_connect_api_key(
    key_content: "` + secretValue + `"
  )
  deliver
  upload_to_testflight
  app_store_build_number
end
`
	if err := os.WriteFile(filepath.Join(fastlaneDir, "Fastfile"), []byte(fastfile), 0o644); err != nil {
		t.Fatalf("write Fastfile error: %v", err)
	}

	previousDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("Getwd() error: %v", err)
	}
	if err := os.Chdir(repo); err != nil {
		t.Fatalf("Chdir() error: %v", err)
	}
	t.Cleanup(func() {
		_ = os.Chdir(previousDir)
	})

	t.Setenv("ASC_BYPASS_KEYCHAIN", "1")
	t.Setenv("ASC_CONFIG_PATH", filepath.Join(repo, "config.json"))

	report := Doctor(DoctorOptions{})
	section := findDoctorSection(t, report, "Migration Hints")
	if !sectionHasStatus(section, DoctorInfo, "Detected Appfile") {
		t.Fatalf("expected Appfile detection, got %#v", section.Checks)
	}
	if !sectionHasStatus(section, DoctorInfo, "Detected Fastfile") {
		t.Fatalf("expected Fastfile detection, got %#v", section.Checks)
	}
	if !sectionHasStatus(section, DoctorInfo, "keys: app_identifier") {
		t.Fatalf("expected Appfile keys in output, got %#v", section.Checks)
	}
	if !sectionHasStatus(section, DoctorInfo, "actions: app_store_connect_api_key") {
		t.Fatalf("expected Fastfile actions in output, got %#v", section.Checks)
	}

	if report.Migration == nil {
		t.Fatal("expected migration hints in report")
	}
	expectedActions := []string{
		"app_store_connect_api_key",
		"deliver",
		"upload_to_testflight",
		"app_store_build_number",
	}
	if !reflect.DeepEqual(report.Migration.DetectedActions, expectedActions) {
		t.Fatalf("DetectedActions = %#v, want %#v", report.Migration.DetectedActions, expectedActions)
	}

	expectedCommands := []string{
		`asc auth login --name "MyKey" --key-id "KEY_ID" --issuer-id "ISSUER_ID" --private-key /path/to/AuthKey.p8`,
		"asc migrate validate --fastlane-dir ./fastlane",
		`asc migrate import --app "APP_ID" --version-id "VERSION_ID" --fastlane-dir ./fastlane`,
		`asc builds latest --app "APP_ID"`,
		`asc publish testflight --app "APP_ID" --ipa app.ipa --group "GROUP_ID"`,
	}
	if !reflect.DeepEqual(report.Migration.SuggestedCommands, expectedCommands) {
		t.Fatalf("SuggestedCommands = %#v, want %#v", report.Migration.SuggestedCommands, expectedCommands)
	}

	assertNoSecretInDoctorReport(t, report, secretValue)
}

func TestDoctorMigrationHintsMissingFilesInfoOnly(t *testing.T) {
	repo := t.TempDir()
	if err := os.Mkdir(filepath.Join(repo, ".git"), 0o755); err != nil {
		t.Fatalf("create .git error: %v", err)
	}

	previousDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("Getwd() error: %v", err)
	}
	if err := os.Chdir(repo); err != nil {
		t.Fatalf("Chdir() error: %v", err)
	}
	t.Cleanup(func() {
		_ = os.Chdir(previousDir)
	})

	t.Setenv("ASC_BYPASS_KEYCHAIN", "1")
	t.Setenv("ASC_CONFIG_PATH", filepath.Join(repo, "config.json"))

	report := Doctor(DoctorOptions{})
	section := findDoctorSection(t, report, "Migration Hints")
	if len(section.Checks) == 0 {
		t.Fatal("expected migration hints checks")
	}
	for _, check := range section.Checks {
		if check.Status != DoctorInfo {
			t.Fatalf("expected info-only checks, got %#v", section.Checks)
		}
	}
	if report.Migration == nil {
		t.Fatal("expected migration hints in report")
	}
	if len(report.Migration.DetectedFiles) != 0 {
		t.Fatalf("expected no detected files, got %#v", report.Migration.DetectedFiles)
	}
	if len(report.Migration.DetectedActions) != 0 {
		t.Fatalf("expected no detected actions, got %#v", report.Migration.DetectedActions)
	}
	if len(report.Migration.SuggestedCommands) != 0 {
		t.Fatalf("expected no suggested commands, got %#v", report.Migration.SuggestedCommands)
	}
}

func assertNoSecretInDoctorReport(t *testing.T, report DoctorReport, secret string) {
	t.Helper()
	for _, section := range report.Sections {
		for _, check := range section.Checks {
			if strings.Contains(check.Message, secret) {
				t.Fatalf("secret leaked in message: %q", check.Message)
			}
			if strings.Contains(check.Recommendation, secret) {
				t.Fatalf("secret leaked in recommendation: %q", check.Recommendation)
			}
		}
	}
	if report.Migration != nil {
		for _, cmd := range report.Migration.SuggestedCommands {
			if strings.Contains(cmd, secret) {
				t.Fatalf("secret leaked in suggested command: %q", cmd)
			}
		}
		for _, file := range report.Migration.DetectedFiles {
			if strings.Contains(file, secret) {
				t.Fatalf("secret leaked in detected file: %q", file)
			}
		}
	}
}

func findDoctorSection(t *testing.T, report DoctorReport, title string) DoctorSection {
	t.Helper()
	for _, section := range report.Sections {
		if section.Title == title {
			return section
		}
	}
	t.Fatalf("expected section %q, got %#v", title, report.Sections)
	return DoctorSection{}
}

func sectionHasStatus(section DoctorSection, status DoctorStatus, contains string) bool {
	for _, check := range section.Checks {
		if check.Status == status && strings.Contains(check.Message, contains) {
			return true
		}
	}
	return false
}
