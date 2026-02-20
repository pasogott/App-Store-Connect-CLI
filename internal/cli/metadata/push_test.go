package metadata

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadLocalMetadataTreatsDefaultLocaleCaseInsensitively(t *testing.T) {
	dir := t.TempDir()
	version := "1.2.3"

	if err := os.MkdirAll(filepath.Join(dir, appInfoDirName), 0o755); err != nil {
		t.Fatalf("mkdir app-info: %v", err)
	}
	if err := os.MkdirAll(filepath.Join(dir, versionDirName, version), 0o755); err != nil {
		t.Fatalf("mkdir version dir: %v", err)
	}
	if err := os.WriteFile(filepath.Join(dir, appInfoDirName, "Default.json"), []byte(`{"name":"Default App Name"}`), 0o644); err != nil {
		t.Fatalf("write app-info default file: %v", err)
	}
	if err := os.WriteFile(filepath.Join(dir, versionDirName, version, "DeFaUlT.json"), []byte(`{"description":"Default description"}`), 0o644); err != nil {
		t.Fatalf("write version default file: %v", err)
	}

	bundle, err := loadLocalMetadata(dir, version)
	if err != nil {
		t.Fatalf("loadLocalMetadata() error: %v", err)
	}
	if bundle.defaultAppInfo == nil {
		t.Fatal("expected default app-info localization")
	}
	if bundle.defaultVersion == nil {
		t.Fatal("expected default version localization")
	}
	if bundle.defaultAppInfo.Name != "Default App Name" {
		t.Fatalf("expected default app-info name, got %q", bundle.defaultAppInfo.Name)
	}
	if bundle.defaultVersion.Description != "Default description" {
		t.Fatalf("expected default version description, got %q", bundle.defaultVersion.Description)
	}
	if len(bundle.appInfo) != 0 {
		t.Fatalf("expected no explicit app-info locales, got %+v", bundle.appInfo)
	}
	if len(bundle.version) != 0 {
		t.Fatalf("expected no explicit version locales, got %+v", bundle.version)
	}
}

func TestBuildScopePlanCountsDeleteAndCreateForRecreate(t *testing.T) {
	local := map[string]map[string]string{
		"en-US": {
			"name": "Local Name",
		},
	}
	remote := map[string]map[string]string{
		"en-US": {
			"name":     "Remote Name",
			"subtitle": "Remote subtitle",
		},
	}

	adds, updates, deletes, calls := buildScopePlan(
		appInfoDirName,
		"",
		appInfoPlanFields,
		local,
		remote,
	)

	if len(adds) != 0 {
		t.Fatalf("expected no adds, got %+v", adds)
	}
	if len(updates) != 1 {
		t.Fatalf("expected one field update, got %+v", updates)
	}
	if len(deletes) != 1 {
		t.Fatalf("expected one field delete, got %+v", deletes)
	}
	if calls.create != 1 || calls.delete != 1 || calls.update != 0 {
		t.Fatalf("unexpected call counts: %+v", calls)
	}
}
