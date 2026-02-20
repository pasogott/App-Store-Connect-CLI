package cmdtest

import (
	"bytes"
	"compress/gzip"
	"context"
	"encoding/json"
	"errors"
	"flag"
	"io"
	"net/http"
	"path/filepath"
	"strings"
	"testing"
)

func TestInsightsWeeklyValidationErrors(t *testing.T) {
	t.Setenv("ASC_APP_ID", "")
	t.Setenv("ASC_VENDOR_NUMBER", "")
	t.Setenv("ASC_ANALYTICS_VENDOR_NUMBER", "")

	tests := []struct {
		name    string
		args    []string
		wantErr string
	}{
		{
			name:    "missing app",
			args:    []string{"insights", "weekly", "--source", "analytics", "--week", "2026-02-16"},
			wantErr: "--app is required",
		},
		{
			name:    "missing source",
			args:    []string{"insights", "weekly", "--app", "app-1", "--week", "2026-02-16"},
			wantErr: "--source is required",
		},
		{
			name:    "missing week",
			args:    []string{"insights", "weekly", "--app", "app-1", "--source", "analytics"},
			wantErr: "--week is required",
		},
		{
			name:    "sales missing vendor",
			args:    []string{"insights", "weekly", "--app", "app-1", "--source", "sales", "--week", "2026-02-16"},
			wantErr: "--vendor is required for --source sales",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			root := RootCommand("1.2.3")
			root.FlagSet.SetOutput(io.Discard)

			stdout, stderr := captureOutput(t, func() {
				if err := root.Parse(test.args); err != nil {
					t.Fatalf("parse error: %v", err)
				}
				err := root.Run(context.Background())
				if !errors.Is(err, flag.ErrHelp) {
					t.Fatalf("expected ErrHelp, got %v", err)
				}
			})

			if stdout != "" {
				t.Fatalf("expected empty stdout, got %q", stdout)
			}
			if !strings.Contains(stderr, test.wantErr) {
				t.Fatalf("expected error %q, got %q", test.wantErr, stderr)
			}
		})
	}
}

func TestInsightsWeeklySalesJSON(t *testing.T) {
	setupAuth(t)
	t.Setenv("ASC_CONFIG_PATH", filepath.Join(t.TempDir(), "nonexistent.json"))
	t.Setenv("ASC_APP_ID", "")

	originalTransport := http.DefaultTransport
	t.Cleanup(func() {
		http.DefaultTransport = originalTransport
	})

	http.DefaultTransport = roundTripFunc(func(req *http.Request) (*http.Response, error) {
		if req.URL.Path != "/v1/salesReports" {
			t.Fatalf("unexpected request: %s %s", req.Method, req.URL.String())
		}
		query := req.URL.Query()
		if query.Get("filter[vendorNumber]") != "12345678" {
			t.Fatalf("expected filter[vendorNumber]=12345678, got %q", query.Get("filter[vendorNumber]"))
		}
		if query.Get("filter[reportType]") != "SALES" {
			t.Fatalf("expected filter[reportType]=SALES, got %q", query.Get("filter[reportType]"))
		}
		if query.Get("filter[frequency]") != "WEEKLY" {
			t.Fatalf("expected filter[frequency]=WEEKLY, got %q", query.Get("filter[frequency]"))
		}
		switch query.Get("filter[reportDate]") {
		case "2026-02-22":
			return insightsGzipResponse(strings.Join([]string{
				"Provider\tUnits\tDeveloper Proceeds\tCustomer Price",
				"foo\t10\t2.50\t4.00",
				"foo\t5\t1.50\t3.00",
				"",
			}, "\n")), nil
		case "2026-02-15":
			return insightsGzipResponse(strings.Join([]string{
				"Provider\tUnits\tDeveloper Proceeds\tCustomer Price",
				"foo\t8\t1.25\t2.00",
				"",
			}, "\n")), nil
		default:
			t.Fatalf("unexpected report date filter %q", query.Get("filter[reportDate]"))
			return nil, nil
		}
	})

	root := RootCommand("1.2.3")
	root.FlagSet.SetOutput(io.Discard)

	stdout, stderr := captureOutput(t, func() {
		if err := root.Parse([]string{"insights", "weekly", "--app", "app-1", "--source", "sales", "--vendor", "12345678", "--week", "2026-02-16"}); err != nil {
			t.Fatalf("parse error: %v", err)
		}
		if err := root.Run(context.Background()); err != nil {
			t.Fatalf("run error: %v", err)
		}
	})

	if stderr != "" {
		t.Fatalf("expected empty stderr, got %q", stderr)
	}

	var payload map[string]any
	if err := json.Unmarshal([]byte(stdout), &payload); err != nil {
		t.Fatalf("unmarshal output: %v\nstdout=%s", err, stdout)
	}

	source, ok := payload["source"].(map[string]any)
	if !ok {
		t.Fatalf("expected source object, got %T", payload["source"])
	}
	if source["name"] != "sales" {
		t.Fatalf("expected source.name=sales, got %v", source["name"])
	}
	if source["vendorNumber"] != "12345678" {
		t.Fatalf("expected source.vendorNumber=12345678, got %v", source["vendorNumber"])
	}

	week, ok := payload["week"].(map[string]any)
	if !ok {
		t.Fatalf("expected week object, got %T", payload["week"])
	}
	if week["start"] != "2026-02-16" || week["end"] != "2026-02-22" {
		t.Fatalf("unexpected week range: %v", week)
	}

	metrics, ok := payload["metrics"].([]any)
	if !ok {
		t.Fatalf("expected metrics array, got %T", payload["metrics"])
	}
	unitsMetric := findMetric(t, metrics, "units")
	if unitsMetric["status"] != "ok" {
		t.Fatalf("expected units metric status ok, got %v", unitsMetric["status"])
	}
	if unitsMetric["thisWeek"] != 15.0 || unitsMetric["lastWeek"] != 8.0 {
		t.Fatalf("unexpected units totals: %v", unitsMetric)
	}
	unavailableMetric := findMetric(t, metrics, "active_devices")
	if unavailableMetric["status"] != "unavailable" {
		t.Fatalf("expected active_devices metric unavailable, got %v", unavailableMetric["status"])
	}
}

func TestInsightsWeeklySalesReportRowsUnavailableOnFetchError(t *testing.T) {
	setupAuth(t)
	t.Setenv("ASC_CONFIG_PATH", filepath.Join(t.TempDir(), "nonexistent.json"))
	t.Setenv("ASC_APP_ID", "")

	originalTransport := http.DefaultTransport
	t.Cleanup(func() {
		http.DefaultTransport = originalTransport
	})

	http.DefaultTransport = roundTripFunc(func(req *http.Request) (*http.Response, error) {
		if req.URL.Path != "/v1/salesReports" {
			t.Fatalf("unexpected request: %s %s", req.Method, req.URL.String())
		}

		switch req.URL.Query().Get("filter[reportDate]") {
		case "2026-02-22":
			return &http.Response{
				StatusCode: http.StatusInternalServerError,
				Body: io.NopCloser(strings.NewReader(`{
					"errors":[{"status":"500","code":"INTERNAL_ERROR","title":"Internal Server Error","detail":"temporary failure"}]
				}`)),
				Header: http.Header{"Content-Type": []string{"application/json"}},
			}, nil
		case "2026-02-15":
			return insightsGzipResponse(strings.Join([]string{
				"Provider\tUnits\tDeveloper Proceeds\tCustomer Price",
				"foo\t8\t1.25\t2.00",
				"",
			}, "\n")), nil
		default:
			t.Fatalf("unexpected report date filter %q", req.URL.Query().Get("filter[reportDate]"))
			return nil, nil
		}
	})

	root := RootCommand("1.2.3")
	root.FlagSet.SetOutput(io.Discard)

	stdout, stderr := captureOutput(t, func() {
		if err := root.Parse([]string{"insights", "weekly", "--app", "app-1", "--source", "sales", "--vendor", "12345678", "--week", "2026-02-16"}); err != nil {
			t.Fatalf("parse error: %v", err)
		}
		if err := root.Run(context.Background()); err != nil {
			t.Fatalf("run error: %v", err)
		}
	})

	if stderr != "" {
		t.Fatalf("expected empty stderr, got %q", stderr)
	}

	var payload map[string]any
	if err := json.Unmarshal([]byte(stdout), &payload); err != nil {
		t.Fatalf("unmarshal output: %v\nstdout=%s", err, stdout)
	}

	metrics, ok := payload["metrics"].([]any)
	if !ok {
		t.Fatalf("expected metrics array, got %T", payload["metrics"])
	}

	reportRowsMetric := findMetric(t, metrics, "report_rows")
	if reportRowsMetric["status"] != "unavailable" {
		t.Fatalf("expected report_rows metric unavailable, got %v", reportRowsMetric["status"])
	}
}

func TestInsightsWeeklyAnalyticsJSON(t *testing.T) {
	setupAuth(t)
	t.Setenv("ASC_CONFIG_PATH", filepath.Join(t.TempDir(), "nonexistent.json"))
	t.Setenv("ASC_APP_ID", "")

	originalTransport := http.DefaultTransport
	t.Cleanup(func() {
		http.DefaultTransport = originalTransport
	})

	http.DefaultTransport = roundTripFunc(func(req *http.Request) (*http.Response, error) {
		switch req.URL.Path {
		case "/v1/apps/app-1/analyticsReportRequests":
			return insightsJSONResponse(`{
				"data":[
					{
						"type":"analyticsReportRequests",
						"id":"req-1",
						"attributes":{"createdDate":"2026-02-16T10:00:00Z","state":"COMPLETED"}
					}
				],
				"links":{"next":""}
			}`), nil
		case "/v1/analyticsReportRequests/req-1/reports":
			return insightsJSONResponse(`{
				"data":[
					{"type":"analyticsReports","id":"report-1","attributes":{"name":"App Usage"}}
				],
				"links":{"next":""}
			}`), nil
		case "/v1/analyticsReports/report-1/instances":
			return insightsJSONResponse(`{
				"data":[
					{"type":"analyticsReportInstances","id":"inst-1","attributes":{"reportDate":"2026-02-18","processingDate":"2026-02-19T00:00:00Z"}},
					{"type":"analyticsReportInstances","id":"inst-2","attributes":{"reportDate":"2026-02-10","processingDate":"2026-02-11T00:00:00Z"}}
				],
				"links":{"next":""}
			}`), nil
		default:
			t.Fatalf("unexpected request: %s %s", req.Method, req.URL.String())
			return nil, nil
		}
	})

	root := RootCommand("1.2.3")
	root.FlagSet.SetOutput(io.Discard)

	stdout, stderr := captureOutput(t, func() {
		if err := root.Parse([]string{"insights", "weekly", "--app", "app-1", "--source", "analytics", "--week", "2026-02-16"}); err != nil {
			t.Fatalf("parse error: %v", err)
		}
		if err := root.Run(context.Background()); err != nil {
			t.Fatalf("run error: %v", err)
		}
	})

	if stderr != "" {
		t.Fatalf("expected empty stderr, got %q", stderr)
	}

	var payload map[string]any
	if err := json.Unmarshal([]byte(stdout), &payload); err != nil {
		t.Fatalf("unmarshal output: %v\nstdout=%s", err, stdout)
	}

	source, ok := payload["source"].(map[string]any)
	if !ok {
		t.Fatalf("expected source object, got %T", payload["source"])
	}
	if source["name"] != "analytics" {
		t.Fatalf("expected source.name=analytics, got %v", source["name"])
	}
	if source["requestsScanned"] != 1.0 {
		t.Fatalf("expected source.requestsScanned=1, got %v", source["requestsScanned"])
	}

	metrics, ok := payload["metrics"].([]any)
	if !ok {
		t.Fatalf("expected metrics array, got %T", payload["metrics"])
	}
	instancesMetric := findMetric(t, metrics, "instances_available")
	if instancesMetric["status"] != "ok" {
		t.Fatalf("expected instances_available status ok, got %v", instancesMetric["status"])
	}
	if instancesMetric["thisWeek"] != 1.0 || instancesMetric["lastWeek"] != 1.0 {
		t.Fatalf("unexpected instances metrics %v", instancesMetric)
	}
	unavailableMetric := findMetric(t, metrics, "business_conversion_rate")
	if unavailableMetric["status"] != "unavailable" {
		t.Fatalf("expected business_conversion_rate unavailable, got %v", unavailableMetric["status"])
	}
}

func TestInsightsWeeklyTableOutput(t *testing.T) {
	setupAuth(t)
	t.Setenv("ASC_CONFIG_PATH", filepath.Join(t.TempDir(), "nonexistent.json"))
	t.Setenv("ASC_APP_ID", "")

	originalTransport := http.DefaultTransport
	t.Cleanup(func() {
		http.DefaultTransport = originalTransport
	})

	http.DefaultTransport = roundTripFunc(func(req *http.Request) (*http.Response, error) {
		if req.URL.Path != "/v1/salesReports" {
			t.Fatalf("unexpected request: %s %s", req.Method, req.URL.String())
		}
		return insightsGzipResponse(strings.Join([]string{
			"Provider\tUnits\tDeveloper Proceeds\tCustomer Price",
			"foo\t1\t0.10\t0.20",
		}, "\n")), nil
	})

	root := RootCommand("1.2.3")
	root.FlagSet.SetOutput(io.Discard)

	stdout, stderr := captureOutput(t, func() {
		if err := root.Parse([]string{"insights", "weekly", "--app", "app-1", "--source", "sales", "--vendor", "12345678", "--week", "2026-02-16", "--output", "table"}); err != nil {
			t.Fatalf("parse error: %v", err)
		}
		if err := root.Run(context.Background()); err != nil {
			t.Fatalf("run error: %v", err)
		}
	})

	if stderr != "" {
		t.Fatalf("expected empty stderr, got %q", stderr)
	}
	if !strings.Contains(stdout, "CONTEXT") || !strings.Contains(stdout, "METRICS") {
		t.Fatalf("expected context and metrics sections in table output, got %q", stdout)
	}
}

func TestInsightsWeeklyAnalyticsForbiddenReturnsUnavailable(t *testing.T) {
	setupAuth(t)
	t.Setenv("ASC_CONFIG_PATH", filepath.Join(t.TempDir(), "nonexistent.json"))
	t.Setenv("ASC_APP_ID", "")

	originalTransport := http.DefaultTransport
	t.Cleanup(func() {
		http.DefaultTransport = originalTransport
	})

	http.DefaultTransport = roundTripFunc(func(req *http.Request) (*http.Response, error) {
		if req.URL.Path != "/v1/apps/app-1/analyticsReportRequests" {
			t.Fatalf("unexpected request: %s %s", req.Method, req.URL.String())
		}
		return &http.Response{
			StatusCode: http.StatusForbidden,
			Body: io.NopCloser(strings.NewReader(`{
				"errors":[{"status":"403","code":"FORBIDDEN","title":"Forbidden","detail":"security restriction"}]
			}`)),
			Header: http.Header{"Content-Type": []string{"application/json"}},
		}, nil
	})

	root := RootCommand("1.2.3")
	root.FlagSet.SetOutput(io.Discard)

	stdout, stderr := captureOutput(t, func() {
		if err := root.Parse([]string{"insights", "weekly", "--app", "app-1", "--source", "analytics", "--week", "2026-02-16"}); err != nil {
			t.Fatalf("parse error: %v", err)
		}
		if err := root.Run(context.Background()); err != nil {
			t.Fatalf("run error: %v", err)
		}
	})

	if stderr != "" {
		t.Fatalf("expected empty stderr, got %q", stderr)
	}

	var payload map[string]any
	if err := json.Unmarshal([]byte(stdout), &payload); err != nil {
		t.Fatalf("unmarshal output: %v\nstdout=%s", err, stdout)
	}

	metrics, ok := payload["metrics"].([]any)
	if !ok || len(metrics) == 0 {
		t.Fatalf("expected metrics array, got %T %v", payload["metrics"], payload["metrics"])
	}
	completedMetric := findMetric(t, metrics, "completed_requests")
	if completedMetric["status"] != "unavailable" {
		t.Fatalf("expected completed_requests unavailable, got %v", completedMetric["status"])
	}
}

func TestInsightsWeeklyAnalyticsNestedForbiddenReturnsUnavailable(t *testing.T) {
	setupAuth(t)
	t.Setenv("ASC_CONFIG_PATH", filepath.Join(t.TempDir(), "nonexistent.json"))
	t.Setenv("ASC_APP_ID", "")

	originalTransport := http.DefaultTransport
	t.Cleanup(func() {
		http.DefaultTransport = originalTransport
	})

	http.DefaultTransport = roundTripFunc(func(req *http.Request) (*http.Response, error) {
		switch req.URL.Path {
		case "/v1/apps/app-1/analyticsReportRequests":
			return insightsJSONResponse(`{
				"data":[
					{"type":"analyticsReportRequests","id":"req-1","attributes":{"state":"COMPLETED","createdDate":"2026-02-16T10:00:00Z"}}
				],
				"links":{"next":""}
			}`), nil
		case "/v1/analyticsReportRequests/req-1/reports":
			return &http.Response{
				StatusCode: http.StatusForbidden,
				Body: io.NopCloser(strings.NewReader(`{
					"errors":[{"status":"403","code":"FORBIDDEN","title":"Forbidden","detail":"security restriction"}]
				}`)),
				Header: http.Header{"Content-Type": []string{"application/json"}},
			}, nil
		default:
			t.Fatalf("unexpected request: %s %s", req.Method, req.URL.String())
			return nil, nil
		}
	})

	root := RootCommand("1.2.3")
	root.FlagSet.SetOutput(io.Discard)

	stdout, stderr := captureOutput(t, func() {
		if err := root.Parse([]string{"insights", "weekly", "--app", "app-1", "--source", "analytics", "--week", "2026-02-16"}); err != nil {
			t.Fatalf("parse error: %v", err)
		}
		if err := root.Run(context.Background()); err != nil {
			t.Fatalf("run error: %v", err)
		}
	})

	if stderr != "" {
		t.Fatalf("expected empty stderr, got %q", stderr)
	}

	var payload map[string]any
	if err := json.Unmarshal([]byte(stdout), &payload); err != nil {
		t.Fatalf("unmarshal output: %v\nstdout=%s", err, stdout)
	}

	metrics, ok := payload["metrics"].([]any)
	if !ok || len(metrics) == 0 {
		t.Fatalf("expected metrics array, got %T %v", payload["metrics"], payload["metrics"])
	}
	reportsMetric := findMetric(t, metrics, "reports_available")
	if reportsMetric["status"] != "unavailable" {
		t.Fatalf("expected reports_available unavailable, got %v", reportsMetric["status"])
	}
}

func findMetric(t *testing.T, metrics []any, name string) map[string]any {
	t.Helper()
	for _, item := range metrics {
		metric, ok := item.(map[string]any)
		if !ok {
			continue
		}
		if metric["name"] == name {
			return metric
		}
	}
	t.Fatalf("metric %q not found in %v", name, metrics)
	return nil
}

func insightsJSONResponse(body string) *http.Response {
	return &http.Response{
		StatusCode: http.StatusOK,
		Body:       io.NopCloser(strings.NewReader(body)),
		Header:     http.Header{"Content-Type": []string{"application/json"}},
	}
}

func insightsGzipResponse(report string) *http.Response {
	var compressed bytes.Buffer
	zw := gzip.NewWriter(&compressed)
	_, _ = zw.Write([]byte(report))
	_ = zw.Close()
	return &http.Response{
		StatusCode: http.StatusOK,
		Body:       io.NopCloser(bytes.NewReader(compressed.Bytes())),
		Header:     http.Header{"Content-Type": []string{"application/a-gzip"}},
	}
}
