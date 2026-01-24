package cmd

import (
	"context"
	"flag"
	"fmt"
	"strings"

	"github.com/peterbourgon/ff/v3/ffcli"

	"github.com/rudrankriyam/App-Store-Connect-CLI/internal/asc"
)

// AppsCommand returns the apps command factory.
func AppsCommand() *ffcli.Command {
	fs := flag.NewFlagSet("apps", flag.ExitOnError)

	output := fs.String("output", "json", "Output format: json (default), table, markdown")
	pretty := fs.Bool("pretty", false, "Pretty-print JSON output")
	bundleID := fs.String("bundle-id", "", "Filter by bundle ID(s), comma-separated")
	name := fs.String("name", "", "Filter by app name(s), comma-separated")
	sku := fs.String("sku", "", "Filter by SKU(s), comma-separated")
	sort := fs.String("sort", "", "Sort by name, -name, bundleId, or -bundleId")
	limit := fs.Int("limit", 0, "Maximum results per page (1-200)")
	next := fs.String("next", "", "Fetch next page using a links.next URL")
	paginate := fs.Bool("paginate", false, "Automatically fetch all pages (aggregate results)")

	return &ffcli.Command{
		Name:       "apps",
		ShortUsage: "asc apps [flags]",
		ShortHelp:  "List apps from App Store Connect.",
		LongHelp: `List apps from App Store Connect.

Examples:
  asc apps
  asc apps --bundle-id "com.example.app"
  asc apps --name "My App"
  asc apps --limit 10
  asc apps --sort name
  asc apps --output table
  asc apps --next "<links.next>"
  asc apps --paginate`,
		FlagSet:   fs,
		UsageFunc: DefaultUsageFunc,
		Exec: func(ctx context.Context, args []string) error {
			if *limit != 0 && (*limit < 1 || *limit > 200) {
				return fmt.Errorf("apps: --limit must be between 1 and 200")
			}
			if err := validateNextURL(*next); err != nil {
				return fmt.Errorf("apps: %w", err)
			}
			if err := validateSort(*sort, "name", "-name", "bundleId", "-bundleId"); err != nil {
				return fmt.Errorf("apps: %w", err)
			}

			client, err := getASCClient()
			if err != nil {
				return fmt.Errorf("apps: %w", err)
			}

			requestCtx, cancel := contextWithTimeout(ctx)
			defer cancel()

			opts := []asc.AppsOption{
			asc.WithAppsBundleIDs(splitCSV(*bundleID)),
			asc.WithAppsNames(splitCSV(*name)),
			asc.WithAppsSKUs(splitCSV(*sku)),
				asc.WithAppsLimit(*limit),
				asc.WithAppsNextURL(*next),
			}
			if strings.TrimSpace(*sort) != "" {
				opts = append(opts, asc.WithAppsSort(*sort))
			}

			if *paginate {
				// Fetch first page with limit set for consistent pagination
				paginateOpts := append(opts, asc.WithAppsLimit(200))
				firstPage, err := client.GetApps(requestCtx, paginateOpts...)
				if err != nil {
					return fmt.Errorf("apps: failed to fetch: %w", err)
				}

				// Fetch all remaining pages
				apps, err := asc.PaginateAll(requestCtx, firstPage, func(ctx context.Context, nextURL string) (asc.PaginatedResponse, error) {
					return client.GetApps(ctx, asc.WithAppsNextURL(nextURL))
				})
				if err != nil {
					return fmt.Errorf("apps: %w", err)
				}

				format := *output
				return printOutput(apps, format, *pretty)
			}

			apps, err := client.GetApps(requestCtx, opts...)
			if err != nil {
				return fmt.Errorf("apps: failed to fetch: %w", err)
			}

			format := *output

			return printOutput(apps, format, *pretty)
		},
	}
}
