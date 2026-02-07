package iap

import (
	"testing"
	"time"
)

func TestParseIAPPriceScheduleIncluded_DecodesDatesFromResourceID(t *testing.T) {
	raw := []byte(`[
		{
			"type":"inAppPurchasePrices",
			"id":"eyJzIjoiNjc0OTI3MzQ1NyIsInQiOiJNVVMiLCJwIjoiMTAzNTciLCJzZCI6MC4wLCJlZCI6MTc3MTIyODgwMC4wMDAwMDAwMDB9",
			"attributes":{}
		},
		{
			"type":"inAppPurchasePrices",
			"id":"eyJzIjoiNjc0OTI3MzQ1NyIsInQiOiJNVVMiLCJwIjoiMTAzODciLCJzZCI6MTc3MTIyODgwMC4wMDAwMDAwMDAsImVkIjowLjB9",
			"attributes":{}
		}
	]`)

	entries, _, err := parseIAPPriceScheduleIncluded(raw)
	if err != nil {
		t.Fatalf("parseIAPPriceScheduleIncluded returned error: %v", err)
	}
	if len(entries) != 2 {
		t.Fatalf("expected 2 entries, got %d", len(entries))
	}

	now := time.Date(2026, time.February, 7, 0, 0, 0, 0, time.UTC)
	changes := buildScheduledChanges(entries, now, "")
	if len(changes) != 1 {
		t.Fatalf("expected 1 scheduled change, got %d", len(changes))
	}

	change := changes[0]
	if change.Territory != "MUS" {
		t.Fatalf("expected territory MUS, got %q", change.Territory)
	}
	if change.FromPricePoint != "10357" {
		t.Fatalf("expected from price point 10357, got %q", change.FromPricePoint)
	}
	if change.ToPricePoint != "10387" {
		t.Fatalf("expected to price point 10387, got %q", change.ToPricePoint)
	}
	if change.EffectiveDate != "2026-02-16" {
		t.Fatalf("expected effective date 2026-02-16, got %q", change.EffectiveDate)
	}
}
