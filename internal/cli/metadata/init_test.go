package metadata

import (
	"strings"
	"testing"
)

func TestMetadataInitCommandDefaultsLocale(t *testing.T) {
	cmd := MetadataInitCommand()
	f := cmd.FlagSet.Lookup("locale")
	if f == nil {
		t.Fatal("expected --locale flag to be defined")
		return
	}
	if f.DefValue != "en-US" {
		t.Fatalf("expected --locale default en-US, got %q", f.DefValue)
	}
}

func TestMetadataInitCommandUsageMentionsVersionAndLocale(t *testing.T) {
	cmd := MetadataInitCommand()
	for _, want := range []string{"metadata init", "--version", "--locale"} {
		if !strings.Contains(cmd.ShortUsage, want) {
			t.Fatalf("expected ShortUsage to mention %q, got %q", want, cmd.ShortUsage)
		}
	}
}
