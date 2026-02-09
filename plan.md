# Coverage Expansion Plan (audit/fix-security-and-test-gaps)

Goal: eliminate current `0.0%` coverage packages by adding package-local tests in phases, then re-run short/full suites.

## Baseline (captured)

- [x] Run `go test -short -coverprofile=coverage.out ./...`
- [x] Capture current zero-coverage package list
- [x] Re-check baseline list after each phase

## Phase 0 - Planning and Tracking

- [x] Create `plan.md`
- [x] Keep this file updated after every phase commit

## Phase 1 - Core Utility and Small Command Packages

- [x] `internal/asc/types`
- [x] `internal/cli/shared/suggest`
- [x] `internal/cli/completion`
- [x] `internal/cli/registry`
- [x] `internal/cli/crashes`
- [x] `internal/cli/feedback`
- [x] `internal/cli/submit`
- [x] `internal/cli/routingcoverage`
- [x] Commit Phase 1

## Phase 2 - Identity/Metadata Command Families

- [x] `internal/cli/accessibility`
- [x] `internal/cli/actors`
- [x] `internal/cli/agerating`
- [x] `internal/cli/agreements`
- [x] `internal/cli/categories`
- [x] `internal/cli/eula`
- [x] `internal/cli/nominations`
- [x] `internal/cli/merchantids`
- [x] `internal/cli/passtypeids`
- [x] Commit Phase 2

## Phase 3 - App Distribution/Release Command Families

- [x] `internal/cli/androidiosmapping`
- [x] `internal/cli/app_events`
- [x] `internal/cli/appclips`
- [x] `internal/cli/assets`
- [x] `internal/cli/betaapplocalizations`
- [x] `internal/cli/betabuildlocalizations`
- [x] `internal/cli/buildbundles`
- [x] `internal/cli/buildlocalizations`
- [x] `internal/cli/localizations`
- [x] `internal/cli/prerelease`
- [x] `internal/cli/productpages`
- [x] Commit Phase 3

## Phase 4 - Commerce/Security/Operational Command Families

- [x] `internal/cli/encryption`
- [x] `internal/cli/offercodes`
- [x] `internal/cli/performance`
- [x] `internal/cli/promotedpurchases`
- [x] `internal/cli/reviews`
- [x] `internal/cli/winbackoffers`
- [x] `internal/cli/notarization`
- [x] Commit Phase 4

## Phase 5 - Large Surface Command Families

- [x] `internal/cli/gamecenter`
- [x] `internal/cli/xcodecloud`
- [x] Commit Phase 5

## Phase 6 - Remaining Non-CLI Root Packages

- [x] `github.com/rudrankriyam/App-Store-Connect-CLI` (main package)
- [x] Commit Phase 6

## Phase 7 - High-Risk Package Behavior Depth

- [x] `internal/cli/subscriptions` (normalization + parsing behavior)
- [x] `internal/cli/builds` (expire-all time parsing/threshold behavior)
- [x] `internal/cli/apps` (include/field normalization behavior)
- [x] `internal/cli/testflight` (recruitment/metrics helper behavior)
- [x] `internal/cli/iap` (offer/schedule parsing behavior)
- [x] Commit Phase 7

## Phase 8 - High-Risk Package API Interaction Coverage

- [x] `internal/cli/builds` (`builds latest` multi-preReleaseVersion selection via API)
- [x] `internal/cli/apps` (`app-tags list` query + response behavior)
- [x] `internal/cli/testflight` (`metrics public-link` and `metrics testers` API output)
- [x] `internal/cli/iap` (`offer-codes create` default eligibilities + payload behavior)
- [x] `internal/cli/subscriptions` (`offer-codes create` normalization + payload behavior)
- [x] Commit Phase 8

## Phase 9 - High-Risk Package Negative API Paths

- [x] `internal/cli/builds` (`builds latest` pre-release lookup API failure propagation)
- [x] `internal/cli/apps` (`app-tags list` fetch failure propagation)
- [x] `internal/cli/testflight` (`metrics public-link` fetch failure propagation)
- [x] `internal/cli/iap` (`offer-codes create` API create failure propagation)
- [x] `internal/cli/subscriptions` (`offer-codes create` API create failure propagation)
- [x] Commit Phase 9

## Phase 10 - Pagination and `--next` Edge Cases

- [x] `internal/cli/builds` (`builds latest` repeated pre-release pagination URL detection)
- [x] `internal/cli/apps` (`app-tags list --paginate` repeated next URL detection)
- [x] `internal/cli/testflight` (`metrics beta-tester-usages --paginate` repeated next URL detection)
- [x] `internal/cli/iap` (`offer-codes list` invalid `--next` host rejection)
- [x] `internal/cli/subscriptions` (`offer-codes list --paginate` second-page API failure propagation)
- [x] Commit Phase 10

## Phase 11 - Streaming/NDJSON and Mixed Output-Mode Errors

- [x] `internal/cli/subscriptions` (`price-points list --paginate --stream` repeated `next` URL detection)
- [x] `internal/cli/subscriptions` (`price-points list --paginate --stream` second-page API failure propagation)
- [x] `internal/cli/builds` (`builds latest` unsupported `--output` and `--pretty` non-JSON failures)
- [x] `internal/cli/apps` (`app-tags list` unsupported `--output` and `--pretty` non-JSON failures)
- [x] `internal/cli/testflight` (`metrics public-link` unsupported `--output` and `--pretty` non-JSON failures)
- [x] `internal/cli/iap` (`offer-codes list` unsupported `--output` and `--pretty` non-JSON failures)
- [x] `internal/cli/subscriptions` (`offer-codes list` unsupported `--output` and `--pretty` non-JSON failures)
- [x] Commit Phase 11

## Phase 12 - `--next`/`--paginate` Combinatorics and Human Output Paths

- [x] `internal/cli/builds` (`builds latest --output table` success rendering path)
- [x] `internal/cli/apps` (`app-tags list --paginate --next` without `--app` + markdown rendering path)
- [x] `internal/cli/testflight` (`metrics testers --output table` success rendering path)
- [x] `internal/cli/iap` (`offer-codes list --paginate --next` without `--iap-id` + table rendering path)
- [x] `internal/cli/subscriptions` (`offer-codes list --paginate --next` without `--subscription-id` + markdown rendering path)
- [x] Commit Phase 12

## Validation Gate (after each phase and at end)

- [x] `go test -short ./...`
- [x] `go test -short -coverprofile=coverage.out ./...`
- [x] verify previously-targeted packages are no longer `0.0%`
- [x] `make test`
- [x] `make lint`
