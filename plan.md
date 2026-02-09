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

## Phase 13 - `--next` Validation Matrix and Paginated Filter Validation

- [x] `internal/cli/apps` (`app-tags list` invalid `--next` + paginated filter/include query coverage)
- [x] `internal/cli/apps` (`app-tags list` invalid `--fields` and `--territory-fields` include requirement)
- [x] `internal/cli/testflight` (`metrics beta-tester-usages` invalid/malformed `--next` validation)
- [x] `internal/cli/iap` (`offer-codes list` malformed `--next` validation)
- [x] `internal/cli/subscriptions` (`offer-codes list` invalid/malformed `--next` validation)
- [x] Commit Phase 13

## Phase 14 - `--next` Parity for TestFlight Lists and Price Points

- [x] `internal/cli/testflight` (`apps list` invalid/malformed `--next` validation + `--paginate --next` path)
- [x] `internal/cli/testflight` (`beta-groups list` invalid/malformed `--next` validation + `--paginate --next` without `--app`)
- [x] `internal/cli/testflight` (`beta-testers list` invalid/malformed `--next` validation + `--paginate --next` without `--app`)
- [x] `internal/cli/subscriptions` (`price-points list` invalid/malformed `--next` validation + `--paginate --next` without `--subscription-id`)
- [x] Commit Phase 14

## Phase 15 - `--next` Parity for Devices, Users, and Versions Lists

- [x] `internal/cli/devices` (`devices list` invalid/malformed `--next` validation + `--paginate --next` path)
- [x] `internal/cli/users` (`users list` invalid/malformed `--next` validation + `--paginate --next` path)
- [x] `internal/cli/versions` (`versions list` invalid/malformed `--next` validation + `--paginate --next` without `--app`)
- [x] Commit Phase 15

## Phase 16 - `--next` Parity for Users Nested Lists and Version Relationship Surfaces

- [x] `internal/cli/users` (`users invites list` invalid/malformed `--next` validation + `--paginate --next` path)
- [x] `internal/cli/users` (`users visible-apps list/get` invalid-path `--next` handling + `--paginate --next` without `--id`)
- [x] `internal/cli/users` (`users invites visible-apps list --paginate --next` without `--id`)
- [x] `internal/cli/versions` (`versions relationships --type appStoreVersionExperiments` invalid/malformed `--next` validation + `--paginate --next` without `--version-id`)
- [x] `internal/cli/versions` (`versions experiments-v2 list` invalid/malformed `--next` validation + `--paginate --next` without `--version-id`)
- [x] `internal/cli/versions` (`versions customer-reviews list` invalid/malformed `--next` validation + `--paginate --next` without `--version-id`)
- [x] Commit Phase 16

## Phase 17 - `--next` Parity for Webhooks List and Delivery Surfaces

- [x] `internal/cli/webhooks` (`webhooks list` invalid/malformed `--next` validation + `--paginate --next` path)
- [x] `internal/cli/webhooks` (`webhooks deliveries` invalid/malformed `--next` validation + `--paginate --next` path)
- [x] `internal/cli/webhooks` (`webhooks deliveries relationships` invalid/malformed `--next` and invalid extraction path + `--paginate --next` without `--webhook-id`)
- [x] Commit Phase 17

## Phase 18 - `--next` Parity for TestFlight Review and Recruitment Surfaces

- [x] `internal/cli/testflight` (`review get` invalid/malformed `--next` validation + `--next` path without `--app`)
- [x] `internal/cli/testflight` (`beta-details get` invalid/malformed `--next` validation + `--next` path without `--build`)
- [x] `internal/cli/testflight` (`recruitment options` invalid/malformed `--next` validation + `--next` path)
- [x] `internal/cli/testflight` (`review submissions list` invalid/malformed `--next` validation + `--paginate --next` path)
- [x] Commit Phase 18

## Phase 19 - `--next` Parity for TestFlight Relationship Linkages

- [x] `internal/cli/testflight` (`beta-groups relationships get` invalid/malformed `--next` validation + `--paginate --next` without `--group-id`)
- [x] `internal/cli/testflight` (`beta-testers relationships get` invalid/malformed `--next` validation + `--paginate --next` without `--tester-id`)
- [x] Commit Phase 19

## Phase 20 - `--next` Parity for TestFlight Beta Tester Related Lists

- [x] `internal/cli/testflight` (`beta-testers apps list` invalid/malformed `--next` validation + `--paginate --next` path)
- [x] `internal/cli/testflight` (`beta-testers beta-groups list` invalid/malformed `--next` validation + `--paginate --next` path)
- [x] `internal/cli/testflight` (`beta-testers builds list` invalid/malformed `--next` validation + `--paginate --next` path)
- [x] Commit Phase 20

## Phase 21 - `--next` Parity for TestFlight License and Tester Metrics Lists

- [x] `internal/cli/testflight` (`beta-license-agreements list` invalid/malformed `--next` validation + `--paginate --next` path)
- [x] `internal/cli/testflight` (`beta-testers metrics` invalid/malformed `--next` validation + `--next` path without required `--tester-id/--app`)
- [x] Commit Phase 21

## Phase 22 - `--next` Parity for Sandbox, Reviews, and Promoted Purchases Lists

- [x] `internal/cli/sandbox` (`sandbox list` invalid/malformed `--next` validation + `--paginate --next` path)
- [x] `internal/cli/promotedpurchases` (`promoted-purchases list` invalid/malformed `--next` validation + `--paginate --next` without `--app`)
- [x] `internal/cli/reviews` (`reviews list` invalid/malformed `--next` validation + `--paginate --next` without `--app`)
- [x] `internal/cli/reviews` (`reviews summarizations` invalid/malformed `--next` validation + `--paginate --next` without `--app`)
- [x] Commit Phase 22

## Phase 23 - `--next` Parity for Profiles Lists and Relationship Linkages

- [x] `internal/cli/profiles` (`profiles list` invalid/malformed `--next` validation + `--paginate --next` path)
- [x] `internal/cli/profiles` (`profiles relationships certificates` invalid/malformed extraction `--next` + `--paginate --next` without `--id`)
- [x] `internal/cli/profiles` (`profiles relationships devices` invalid/malformed extraction `--next` + `--paginate --next` without `--id`)
- [x] Commit Phase 23

## Phase 24 - `--next` Parity for Certificates, Agreements, Nominations, and Accessibility Lists

- [x] `internal/cli/certificates` (`certificates list` invalid/malformed `--next` validation + `--paginate --next` path)
- [x] `internal/cli/agreements` (`agreements territories list` invalid/malformed extraction `--next` + `--paginate --next` without `--id`)
- [x] `internal/cli/nominations` (`nominations list` invalid/malformed `--next` validation + `--paginate --next` path)
- [x] `internal/cli/accessibility` (`accessibility list` invalid/malformed `--next` validation + `--paginate --next` without `--app`)
- [x] Commit Phase 24

## Phase 25 - `--next` Parity for Pass-Type IDs, Merchant IDs, and Offer Codes Lists

- [x] `internal/cli/passtypeids` (`pass-type-ids list` invalid/malformed `--next` validation + `--paginate --next` path)
- [x] `internal/cli/merchantids` (`merchant-ids list` invalid/malformed `--next` validation + `--paginate --next` path)
- [x] `internal/cli/offercodes` (`offer-codes list` invalid/malformed `--next` validation + `--paginate --next` without `--offer-code`)
- [x] `internal/cli/offercodes` (`offer-codes custom-codes list` invalid/malformed `--next` validation + `--paginate --next` without `--offer-code-id`)
- [x] `internal/cli/offercodes` (`offer-codes prices list` invalid/malformed `--next` validation + `--paginate --next` without `--offer-code-id`)
- [x] Commit Phase 25

## Phase 26 - `--next` Parity for Pass-Type and Merchant Certificate Surfaces

- [x] `internal/cli/passtypeids` (`pass-type-ids certificates list` invalid/malformed `--next` validation + `--paginate --next` path)
- [x] `internal/cli/passtypeids` (`pass-type-ids certificates get` invalid/malformed `--next` validation + `--paginate --next` path)
- [x] `internal/cli/merchantids` (`merchant-ids certificates list` invalid/malformed `--next` validation + `--paginate --next` without `--merchant-id`)
- [x] `internal/cli/merchantids` (`merchant-ids certificates get` invalid/malformed `--next` validation + `--paginate --next` without `--merchant-id`)
- [x] Commit Phase 26

## Phase 27 - `--next` Parity for Pre-Release Versions Lists and Relationships

- [x] `internal/cli/prerelease` (`pre-release-versions list` invalid/malformed `--next` validation + `--paginate --next` without `--app`)
- [x] `internal/cli/prerelease` (`pre-release-versions builds list` invalid/malformed `--next` validation + `--paginate --next` path)
- [x] `internal/cli/prerelease` (`pre-release-versions relationships get --type builds` invalid/malformed `--next` validation + `--paginate --next` without `--id`)
- [x] Commit Phase 27

## Phase 28 - `--next` Parity for Localizations Lists and Media-Set Relationships

- [x] `internal/cli/localizations` (`localizations list` invalid/malformed `--next` validation + `--paginate --next` path)
- [x] `internal/cli/localizations` (`localizations preview-sets list` invalid/malformed `--next` validation + `--paginate --next` without `--localization-id`)
- [x] `internal/cli/localizations` (`localizations preview-sets relationships` invalid/malformed `--next` validation + `--paginate --next` without `--localization-id`)
- [x] `internal/cli/localizations` (`localizations screenshot-sets list` invalid/malformed `--next` validation + `--paginate --next` without `--localization-id`)
- [x] `internal/cli/localizations` (`localizations screenshot-sets relationships` invalid/malformed `--next` validation + `--paginate --next` without `--localization-id`)
- [x] Commit Phase 28

## Phase 29 - `--next` Parity for App Clips Core List Surfaces

- [x] `internal/cli/appclips` (`app-clips list` invalid/malformed `--next` validation + `--paginate --next` path)
- [x] `internal/cli/appclips` (`app-clips advanced-experiences list` invalid/malformed `--next` validation + `--paginate --next` path)
- [x] `internal/cli/appclips` (`app-clips default-experiences list` invalid/malformed `--next` validation + `--paginate --next` path)
- [x] `internal/cli/appclips` (`app-clips invocations list` invalid/malformed `--next` validation + `--paginate --next` path)
- [x] Commit Phase 29

## Validation Gate (after each phase and at end)

- [x] `go test -short ./...`
- [x] `go test -short -coverprofile=coverage.out ./...`
- [x] verify previously-targeted packages are no longer `0.0%`
- [x] `make test`
- [x] `make lint`
