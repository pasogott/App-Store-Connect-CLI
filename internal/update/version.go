package update

import (
	"strings"

	"golang.org/x/mod/semver"
)

func normalizeVersion(value string) (string, string, bool) {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return "", "", false
	}
	fields := strings.Fields(trimmed)
	if len(fields) == 0 {
		return "", "", false
	}
	raw := strings.TrimPrefix(fields[0], "v")
	if raw == "" || strings.EqualFold(raw, "dev") {
		return "", "", false
	}
	normalized := "v" + raw
	if !semver.IsValid(normalized) {
		return "", "", false
	}
	return raw, normalized, true
}

func compareVersions(current, latest string) int {
	return semver.Compare(current, latest)
}
