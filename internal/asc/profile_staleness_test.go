package asc

import (
	"testing"
	"time"
)

func TestProfileExpirationPassed(t *testing.T) {
	now := time.Date(2026, 9, 17, 12, 0, 0, 0, time.UTC)
	tests := []struct {
		name  string
		value string
		want  bool
	}{
		{name: "rfc3339 past", value: "2026-09-17T11:59:59Z", want: true},
		{name: "rfc3339 future", value: "2026-09-17T12:00:01Z", want: false},
		{name: "exactly now is not expired", value: "2026-09-17T12:00:00Z", want: false},
		{name: "date-only past", value: "2026-09-16", want: true},
		{name: "date-only future", value: "2026-09-18", want: false},
		{name: "empty", value: "  ", want: false},
		{name: "unparseable", value: "not-a-date", want: false},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if got := ProfileExpirationPassed(test.value, now); got != test.want {
				t.Fatalf("ProfileExpirationPassed(%q) = %v, want %v", test.value, got, test.want)
			}
		})
	}
}
