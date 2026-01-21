package cmd

import "testing"

func TestValidateSandboxEmail(t *testing.T) {
	if err := validateSandboxEmail("tester@example.com"); err != nil {
		t.Fatalf("expected valid email, got %v", err)
	}
	if err := validateSandboxEmail("not-an-email"); err == nil {
		t.Fatalf("expected invalid email error")
	}
}

func TestValidateSandboxPassword(t *testing.T) {
	tests := []struct {
		value   string
		wantErr bool
	}{
		{value: "Passwordtest1", wantErr: false},
		{value: "short1A", wantErr: true},
		{value: "alllowercase1", wantErr: true},
		{value: "ALLUPPERCASE1", wantErr: true},
		{value: "NoDigitsHere", wantErr: true},
	}

	for _, test := range tests {
		err := validateSandboxPassword(test.value)
		if test.wantErr && err == nil {
			t.Fatalf("expected error for %q", test.value)
		}
		if !test.wantErr && err != nil {
			t.Fatalf("unexpected error for %q: %v", test.value, err)
		}
	}
}

func TestValidateSandboxSecret(t *testing.T) {
	if err := validateSandboxSecret("--secret-question", "Question"); err != nil {
		t.Fatalf("expected valid secret, got %v", err)
	}
	if err := validateSandboxSecret("--secret-question", "short"); err == nil {
		t.Fatalf("expected error for short secret")
	}
}

func TestNormalizeSandboxBirthDate(t *testing.T) {
	if _, err := normalizeSandboxBirthDate("1980-03-01"); err != nil {
		t.Fatalf("expected valid birth date, got %v", err)
	}
	if _, err := normalizeSandboxBirthDate("03/01/1980"); err == nil {
		t.Fatalf("expected error for invalid birth date format")
	}
}

func TestNormalizeSandboxTerritory(t *testing.T) {
	if got, err := normalizeSandboxTerritory("usa"); err != nil || got != "USA" {
		t.Fatalf("expected USA, got %q err %v", got, err)
	}
	if _, err := normalizeSandboxTerritory("ZZZ"); err == nil {
		t.Fatalf("expected error for invalid territory")
	}
}

func TestNormalizeSandboxTerritoryFilter(t *testing.T) {
	if got, err := normalizeSandboxTerritoryFilter(""); err != nil || got != "" {
		t.Fatalf("expected empty territory, got %q err %v", got, err)
	}
}
