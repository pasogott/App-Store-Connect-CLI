package cmd

import (
	"context"
	"fmt"
	"net/mail"
	"strings"
	"time"
	"unicode"

	"github.com/rudrankriyam/App-Store-Connect-CLI/internal/asc"
)

func validateSandboxEmail(value string) error {
	address := strings.TrimSpace(value)
	if address == "" {
		return fmt.Errorf("--email is required")
	}
	if _, err := mail.ParseAddress(address); err != nil {
		return fmt.Errorf("--email must be a valid email address")
	}
	return nil
}

func validateSandboxPassword(value string) error {
	if len(value) < 8 {
		return fmt.Errorf("--password must be at least 8 characters")
	}
	var hasUpper, hasLower, hasDigit bool
	for _, r := range value {
		switch {
		case unicode.IsUpper(r):
			hasUpper = true
		case unicode.IsLower(r):
			hasLower = true
		case unicode.IsDigit(r):
			hasDigit = true
		}
	}
	if !hasUpper || !hasLower || !hasDigit {
		return fmt.Errorf("--password must include uppercase, lowercase, and a number")
	}
	return nil
}

func validateSandboxSecret(flagName, value string) error {
	if len(strings.TrimSpace(value)) < 6 {
		return fmt.Errorf("%s must be at least 6 characters", flagName)
	}
	return nil
}

func normalizeSandboxBirthDate(value string) (string, error) {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return "", fmt.Errorf("--birth-date is required")
	}
	parsed, err := time.Parse("2006-01-02", trimmed)
	if err != nil {
		return "", fmt.Errorf("--birth-date must be in YYYY-MM-DD format")
	}
	return parsed.Format("2006-01-02"), nil
}

func normalizeSandboxTerritory(value string) (string, error) {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return "", fmt.Errorf("--territory is required")
	}
	upper := strings.ToUpper(trimmed)
	if _, ok := sandboxTerritoryCodes[upper]; !ok {
		return "", fmt.Errorf("--territory must be a valid App Store territory code")
	}
	return upper, nil
}

func normalizeSandboxTerritoryFilter(value string) (string, error) {
	if strings.TrimSpace(value) == "" {
		return "", nil
	}
	return normalizeSandboxTerritory(value)
}

func findSandboxTesterByEmail(ctx context.Context, client *asc.Client, email string) (*asc.SandboxTesterResponse, error) {
	next := ""
	for {
		resp, err := client.GetSandboxTesters(ctx,
			asc.WithSandboxTestersEmail(email),
			asc.WithSandboxTestersLimit(200),
			asc.WithSandboxTestersNextURL(next),
		)
		if err != nil {
			return nil, err
		}
		if len(resp.Data) > 1 {
			return nil, fmt.Errorf("multiple sandbox testers found for %q", strings.TrimSpace(email))
		}
		if len(resp.Data) == 1 {
			return &asc.SandboxTesterResponse{Data: resp.Data[0], Links: resp.Links}, nil
		}
		if strings.TrimSpace(resp.Links.Next) == "" {
			break
		}
		next = resp.Links.Next
	}
	return nil, fmt.Errorf("no sandbox tester found for %q", strings.TrimSpace(email))
}

func findSandboxTesterIDByEmail(ctx context.Context, client *asc.Client, email string) (string, error) {
	response, err := findSandboxTesterByEmail(ctx, client, email)
	if err != nil {
		return "", err
	}
	return response.Data.ID, nil
}

var sandboxTerritoryCodes = map[string]struct{}{
	"ABW": {},
	"AFG": {},
	"AGO": {},
	"AIA": {},
	"ALB": {},
	"AND": {},
	"ANT": {},
	"ARE": {},
	"ARG": {},
	"ARM": {},
	"ASM": {},
	"ATG": {},
	"AUS": {},
	"AUT": {},
	"AZE": {},
	"BDI": {},
	"BEL": {},
	"BEN": {},
	"BES": {},
	"BFA": {},
	"BGD": {},
	"BGR": {},
	"BHR": {},
	"BHS": {},
	"BIH": {},
	"BLR": {},
	"BLZ": {},
	"BMU": {},
	"BOL": {},
	"BRA": {},
	"BRB": {},
	"BRN": {},
	"BTN": {},
	"BWA": {},
	"CAF": {},
	"CAN": {},
	"CHE": {},
	"CHL": {},
	"CHN": {},
	"CIV": {},
	"CMR": {},
	"COD": {},
	"COG": {},
	"COK": {},
	"COL": {},
	"COM": {},
	"CPV": {},
	"CRI": {},
	"CUB": {},
	"CUW": {},
	"CXR": {},
	"CYM": {},
	"CYP": {},
	"CZE": {},
	"DEU": {},
	"DJI": {},
	"DMA": {},
	"DNK": {},
	"DOM": {},
	"DZA": {},
	"ECU": {},
	"EGY": {},
	"ERI": {},
	"ESP": {},
	"EST": {},
	"ETH": {},
	"FIN": {},
	"FJI": {},
	"FLK": {},
	"FRA": {},
	"FRO": {},
	"FSM": {},
	"GAB": {},
	"GBR": {},
	"GEO": {},
	"GGY": {},
	"GHA": {},
	"GIB": {},
	"GIN": {},
	"GLP": {},
	"GMB": {},
	"GNB": {},
	"GNQ": {},
	"GRC": {},
	"GRD": {},
	"GRL": {},
	"GTM": {},
	"GUF": {},
	"GUM": {},
	"GUY": {},
	"HKG": {},
	"HND": {},
	"HRV": {},
	"HTI": {},
	"HUN": {},
	"IDN": {},
	"IMN": {},
	"IND": {},
	"IRL": {},
	"IRQ": {},
	"ISL": {},
	"ISR": {},
	"ITA": {},
	"JAM": {},
	"JEY": {},
	"JOR": {},
	"JPN": {},
	"KAZ": {},
	"KEN": {},
	"KGZ": {},
	"KHM": {},
	"KIR": {},
	"KNA": {},
	"KOR": {},
	"KWT": {},
	"LAO": {},
	"LBN": {},
	"LBR": {},
	"LBY": {},
	"LCA": {},
	"LIE": {},
	"LKA": {},
	"LSO": {},
	"LTU": {},
	"LUX": {},
	"LVA": {},
	"MAC": {},
	"MAR": {},
	"MCO": {},
	"MDA": {},
	"MDG": {},
	"MDV": {},
	"MEX": {},
	"MHL": {},
	"MKD": {},
	"MLI": {},
	"MLT": {},
	"MMR": {},
	"MNE": {},
	"MNG": {},
	"MNP": {},
	"MOZ": {},
	"MRT": {},
	"MSR": {},
	"MTQ": {},
	"MUS": {},
	"MWI": {},
	"MYS": {},
	"MYT": {},
	"NAM": {},
	"NCL": {},
	"NER": {},
	"NFK": {},
	"NGA": {},
	"NIC": {},
	"NIU": {},
	"NLD": {},
	"NOR": {},
	"NPL": {},
	"NRU": {},
	"NZL": {},
	"OMN": {},
	"PAK": {},
	"PAN": {},
	"PER": {},
	"PHL": {},
	"PLW": {},
	"PNG": {},
	"POL": {},
	"PRI": {},
	"PRT": {},
	"PRY": {},
	"PSE": {},
	"PYF": {},
	"QAT": {},
	"REU": {},
	"ROU": {},
	"RUS": {},
	"RWA": {},
	"SAU": {},
	"SEN": {},
	"SGP": {},
	"SHN": {},
	"SLB": {},
	"SLE": {},
	"SLV": {},
	"SMR": {},
	"SOM": {},
	"SPM": {},
	"SRB": {},
	"SSD": {},
	"STP": {},
	"SUR": {},
	"SVK": {},
	"SVN": {},
	"SWE": {},
	"SWZ": {},
	"SXM": {},
	"SYC": {},
	"TCA": {},
	"TCD": {},
	"TGO": {},
	"THA": {},
	"TJK": {},
	"TKM": {},
	"TLS": {},
	"TON": {},
	"TTO": {},
	"TUN": {},
	"TUR": {},
	"TUV": {},
	"TWN": {},
	"TZA": {},
	"UGA": {},
	"UKR": {},
	"UMI": {},
	"URY": {},
	"USA": {},
	"UZB": {},
	"VAT": {},
	"VCT": {},
	"VEN": {},
	"VGB": {},
	"VIR": {},
	"VNM": {},
	"VUT": {},
	"WLF": {},
	"WSM": {},
	"XKS": {},
	"YEM": {},
	"ZAF": {},
	"ZMB": {},
	"ZWE": {},
}
