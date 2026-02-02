package update

import (
	"path/filepath"
	"strings"
)

func detectInstallMethod(executablePath string, evalSymlinks func(string) (string, error)) InstallMethod {
	resolved := resolvedExecutable(executablePath, evalSymlinks)
	if resolved == "" {
		return InstallMethodUnknown
	}
	if isHomebrewPath(resolved) {
		return InstallMethodHomebrew
	}
	return InstallMethodManual
}

func resolvedExecutable(executablePath string, evalSymlinks func(string) (string, error)) string {
	if executablePath == "" {
		return ""
	}
	if evalSymlinks == nil {
		return executablePath
	}
	resolved, err := evalSymlinks(executablePath)
	if err != nil || resolved == "" {
		return executablePath
	}
	return resolved
}

func isHomebrewPath(path string) bool {
	cleaned := filepath.Clean(path)
	if strings.Contains(cleaned, string(filepath.Separator)+"Cellar"+string(filepath.Separator)) {
		return true
	}
	prefix := strings.TrimSpace(strings.TrimSuffix(getEnv("HOMEBREW_PREFIX"), string(filepath.Separator)))
	if prefix == "" {
		return false
	}
	cellar := filepath.Join(prefix, "Cellar") + string(filepath.Separator)
	return strings.HasPrefix(cleaned, cellar)
}
