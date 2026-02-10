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

## Phase 30 - `--next` Parity for App Events List and Media Surfaces

- [x] `internal/cli/app_events` (`app-events list` invalid/malformed `--next` validation + `--paginate --next` without `--app`)
- [x] `internal/cli/app_events` (`app-events localizations list` invalid/malformed `--next` validation + `--paginate --next` without `--event-id`)
- [x] `internal/cli/app_events` (`app-events screenshots list` invalid/malformed `--next` validation + `--paginate --next` without `--event-id/--localization-id`)
- [x] `internal/cli/app_events` (`app-events video-clips list` invalid/malformed `--next` validation + `--paginate --next` without `--event-id/--localization-id`)
- [x] Commit Phase 30

## Phase 31 - `--next` Parity for App Events Relationship and Localization Media Linkages

- [x] `internal/cli/app_events` (`app-events relationships` invalid/malformed `--next` validation + `--paginate --next` without `--event-id`)
- [x] `internal/cli/app_events` (`app-events screenshots relationships` invalid/malformed `--next` validation + `--paginate --next` without `--event-id/--localization-id`)
- [x] `internal/cli/app_events` (`app-events video-clips relationships` invalid/malformed `--next` validation + `--paginate --next` without `--event-id/--localization-id`)
- [x] `internal/cli/app_events` (`app-events localizations screenshots list` invalid/malformed `--next` validation + `--paginate --next` path)
- [x] `internal/cli/app_events` (`app-events localizations video-clips list` invalid/malformed `--next` validation + `--paginate --next` path)
- [x] Commit Phase 31

## Phase 32 - `--next` Parity for Xcode Cloud Build-Action and Build-Run Lists

- [x] `internal/cli/xcodecloud` (`xcode-cloud build-runs list` invalid/malformed `--next` validation + `--paginate --next` without `--workflow-id`)
- [x] `internal/cli/xcodecloud` (`xcode-cloud build-runs builds` invalid/malformed `--next` validation + `--paginate --next` without `--run-id`)
- [x] `internal/cli/xcodecloud` (`xcode-cloud issues list` invalid/malformed `--next` validation + `--paginate --next` without `--action-id`)
- [x] `internal/cli/xcodecloud` (`xcode-cloud test-results list` invalid/malformed `--next` validation + `--paginate --next` without `--action-id`)
- [x] `internal/cli/xcodecloud` (`xcode-cloud artifacts list` invalid/malformed `--next` validation + `--paginate --next` without `--action-id`)
- [x] Commit Phase 32

## Phase 33 - `--next` Parity for Xcode Cloud Products Lists and Repository Linkages

- [x] `internal/cli/xcodecloud` (`xcode-cloud products list` invalid/malformed `--next` validation + `--paginate --next` path)
- [x] `internal/cli/xcodecloud` (`xcode-cloud products build-runs` invalid/malformed `--next` validation + `--paginate --next` without `--id`)
- [x] `internal/cli/xcodecloud` (`xcode-cloud products workflows` invalid/malformed `--next` validation + `--paginate --next` without `--id`)
- [x] `internal/cli/xcodecloud` (`xcode-cloud products primary-repositories` invalid/malformed `--next` validation + `--paginate --next` without `--id`)
- [x] `internal/cli/xcodecloud` (`xcode-cloud products additional-repositories` invalid/malformed `--next` validation + `--paginate --next` without `--id`)
- [x] Commit Phase 33

## Phase 34 - `--next` Parity for Xcode Cloud macOS/Xcode Version Catalogs

- [x] `internal/cli/xcodecloud` (`xcode-cloud macos-versions list` invalid/malformed `--next` validation + `--paginate --next` path)
- [x] `internal/cli/xcodecloud` (`xcode-cloud xcode-versions list` invalid/malformed `--next` validation + `--paginate --next` path)
- [x] `internal/cli/xcodecloud` (`xcode-cloud macos-versions xcode-versions` invalid/malformed `--next` validation + `--paginate --next` without `--id`)
- [x] `internal/cli/xcodecloud` (`xcode-cloud xcode-versions macos-versions` invalid/malformed `--next` validation + `--paginate --next` without `--id`)
- [x] Commit Phase 34

## Phase 35 - `--next` Parity for Xcode Cloud SCM Provider/Repository Lists

- [x] `internal/cli/xcodecloud` (`xcode-cloud scm providers list` invalid/malformed `--next` validation + `--paginate --next` path)
- [x] `internal/cli/xcodecloud` (`xcode-cloud scm repositories list` invalid/malformed `--next` validation + `--paginate --next` path)
- [x] `internal/cli/xcodecloud` (`xcode-cloud scm providers repositories` invalid/malformed `--next` validation + `--paginate --next` without `--provider-id`)
- [x] `internal/cli/xcodecloud` (`xcode-cloud scm repositories git-references` invalid/malformed `--next` validation + `--paginate --next` without `--repo-id`)
- [x] `internal/cli/xcodecloud` (`xcode-cloud scm repositories pull-requests` invalid/malformed `--next` validation + `--paginate --next` without `--repo-id`)
- [x] Commit Phase 35

## Phase 36 - `--next` Parity for Xcode Cloud SCM Relationship Linkages

- [x] `internal/cli/xcodecloud` (`xcode-cloud scm repositories relationships git-references` invalid/malformed `--next` validation + `--paginate --next` without `--repo-id`)
- [x] `internal/cli/xcodecloud` (`xcode-cloud scm repositories relationships pull-requests` invalid/malformed `--next` validation + `--paginate --next` without `--repo-id`)
- [x] Commit Phase 36

## Phase 37 - `--next` Parity for Alternative Distribution Domain/Key/Package-Version Lists

- [x] `internal/cli/alternativedistribution` (`domains list` invalid/malformed `--next` validation + `--paginate --next` path)
- [x] `internal/cli/alternativedistribution` (`keys list` invalid/malformed `--next` validation + `--paginate --next` path)
- [x] `internal/cli/alternativedistribution` (`packages versions list` invalid/malformed `--next` validation + `--paginate --next` path)
- [x] `internal/cli/alternativedistribution` (`packages versions deltas` invalid/malformed `--next` validation + `--paginate --next` path)
- [x] `internal/cli/alternativedistribution` (`packages versions variants` invalid/malformed `--next` validation + `--paginate --next` path)
- [x] Commit Phase 37

## Phase 38 - `--next` Parity for Actors, Android Mapping, Feedback, and Crashes Lists

- [x] `internal/cli/actors` (`actors list` invalid/malformed `--next` validation + `--paginate --next` without `--id`)
- [x] `internal/cli/androidiosmapping` (`android-ios-mapping list` invalid/malformed `--next` validation + `--paginate --next` path)
- [x] `internal/cli/feedback` (`feedback` invalid/malformed `--next` validation + `--paginate --next` without `--app`)
- [x] `internal/cli/crashes` (`crashes` invalid/malformed `--next` validation + `--paginate --next` without `--app`)
- [x] Commit Phase 38

## Phase 39 - `--next` Parity for App Clips and App Event Localization Relationship Surfaces

- [x] `internal/cli/appclips` (`advanced-experiences-relationships` invalid/malformed `--next` validation + `--paginate --next` without `--app-clip-id`)
- [x] `internal/cli/appclips` (`default-experiences-relationships` invalid/malformed `--next` validation + `--paginate --next` without `--app-clip-id`)
- [x] `internal/cli/appclips` (`default-experiences localizations list` invalid/malformed `--next` validation + `--paginate --next` path)
- [x] `internal/cli/app_events` (`localizations screenshots-relationships` invalid/malformed `--next` validation + `--paginate --next` without `--localization-id`)
- [x] `internal/cli/app_events` (`localizations video-clips-relationships` invalid/malformed `--next` validation + `--paginate --next` without `--localization-id`)
- [x] Commit Phase 39

## Phase 40 - `--next` Parity for Analytics Request, Get, and Relationship Surfaces

- [x] `internal/cli/analytics` (`analytics requests` invalid/malformed `--next` validation + `--paginate --next` without `--app`)
- [x] `internal/cli/analytics` (`analytics get` invalid/malformed `--next` validation + `--paginate --next` without `--request-id`)
- [x] `internal/cli/analytics` (`analytics instances relationships` invalid/malformed `--next` validation + `--paginate --next` without `--instance-id`)
- [x] `internal/cli/analytics` (`analytics reports relationships` invalid/malformed `--next` validation + `--paginate --next` without `--report-id`)
- [x] Commit Phase 40

## Phase 41 - `--next` Parity for Bundle ID Lists, Capabilities, and Profiles

- [x] `internal/cli/bundleids` (`bundle-ids list` invalid/malformed `--next` validation + `--paginate --next` path)
- [x] `internal/cli/bundleids` (`bundle-ids capabilities list` invalid/malformed `--next` validation + `--paginate --next` without `--bundle`)
- [x] `internal/cli/bundleids` (`bundle-ids profiles list` invalid/malformed `--next` validation + `--paginate --next` without `--id`)
- [x] Commit Phase 41

## Phase 42 - `--next` Parity for Xcode Cloud Workflows and Actions Lists

- [x] `internal/cli/xcodecloud` (`xcode-cloud workflows list` invalid/malformed `--next` validation + `--paginate --next` without `--app`)
- [x] `internal/cli/xcodecloud` (`xcode-cloud actions list` invalid/malformed `--next` validation + `--paginate --next` without `--run-id`)
- [x] Commit Phase 42

## Phase 43 - `--next` Parity for Build Bundle File Size and App Clip Invocation Lists

- [x] `internal/cli/buildbundles` (`build-bundles file-sizes list` invalid/malformed `--next` validation + `--paginate --next` path)
- [x] `internal/cli/buildbundles` (`build-bundles app-clip invocations list` invalid/malformed `--next` validation + `--paginate --next` path)
- [x] Commit Phase 43

## Phase 44 - `--next` Parity for App Tag Relationship and Territory Surfaces

- [x] `internal/cli/apps` (`app-tags relationships` invalid/malformed `--next` validation + `--paginate --next` without `--app`)
- [x] `internal/cli/apps` (`app-tags territories` invalid/malformed `--next` validation + `--paginate --next` without `--id`)
- [x] `internal/cli/apps` (`app-tags territories-relationships` invalid/malformed `--next` validation + `--paginate --next` without `--id`)
- [x] Commit Phase 44

## Phase 45 - `--next` Parity for Core App Lists, Search Keywords, Encryption Declarations, and Territory Age Ratings

- [x] `internal/cli/apps` (`apps` invalid/malformed `--next` validation + `--paginate --next` path)
- [x] `internal/cli/apps` (`apps search-keywords list` invalid/malformed `--next` validation + `--paginate --next` without `--app`)
- [x] `internal/cli/apps` (`apps app-encryption-declarations list` invalid/malformed `--next` validation + `--paginate --next` without `--id`)
- [x] `internal/cli/apps` (`app-info territory-age-ratings list` invalid/malformed `--next` validation + `--paginate --next` without `--id`)
- [x] Commit Phase 45

## Phase 46 - `--next` Parity for Background Asset Lists, Versions, and Upload Files

- [x] `internal/cli/backgroundassets` (`background-assets list` invalid/malformed `--next` validation + `--paginate --next` without `--app`)
- [x] `internal/cli/backgroundassets` (`background-assets versions list` invalid/malformed `--next` validation + `--paginate --next` without `--background-asset-id`)
- [x] `internal/cli/backgroundassets` (`background-assets upload-files list` invalid/malformed `--next` validation + `--paginate --next` without `--version-id`)
- [x] Commit Phase 46

## Phase 47 - `--next` Parity for Beta App, Beta Build, and Build Localization Lists

- [x] `internal/cli/betaapplocalizations` (`beta-app-localizations list` invalid/malformed `--next` validation + `--paginate --next` without `--app`)
- [x] `internal/cli/betabuildlocalizations` (`beta-build-localizations list` invalid/malformed `--next` validation + `--paginate --next` without `--build`)
- [x] `internal/cli/buildlocalizations` (`build-localizations list` invalid/malformed `--next` validation + `--paginate --next` path with build-version resolution)
- [x] Commit Phase 47

## Phase 48 - `--next` Parity for Builds Icons, Testers, Metrics, Relationships, Test Notes, and Upload Lists

- [x] `internal/cli/builds` (`builds icons list` invalid/malformed `--next` validation + `--paginate --next` path)
- [x] `internal/cli/builds` (`builds individual-testers list` invalid/malformed `--next` validation + `--paginate --next` path)
- [x] `internal/cli/builds` (`builds metrics beta-usages` invalid/malformed `--next` validation + `--next` fetch path)
- [x] `internal/cli/builds` (`builds relationships get` invalid/malformed `--next` validation + to-many `--paginate --next` path)
- [x] `internal/cli/builds` (`builds test-notes list` invalid/malformed `--next` validation + `--paginate --next` path)
- [x] `internal/cli/builds` (`builds uploads files list` invalid/malformed `--next` validation + `--paginate --next` path)
- [x] `internal/cli/builds` (`builds uploads list` invalid/malformed `--next` validation + `--paginate --next` without `--app`)
- [x] Commit Phase 48

## Phase 49 - `--next` Parity for Category Subcategory and Encryption Declaration Lists

- [x] `internal/cli/categories` (`categories subcategories` invalid/malformed `--next` validation + `--paginate --next` without `--category-id`)
- [x] `internal/cli/encryption` (`encryption declarations list` invalid/malformed `--next` validation + `--paginate --next` without `--app`)
- [x] Commit Phase 49

## Phase 50 - `--next` Parity for Custom Product Page Lists, Localizations, and Media Sets

- [x] `internal/cli/productpages` (`custom-pages list` invalid/malformed `--next` validation + `--paginate --next` path)
- [x] `internal/cli/productpages` (`custom-pages versions list` invalid/malformed `--next` validation + `--paginate --next` path)
- [x] `internal/cli/productpages` (`custom-pages localizations list` invalid/malformed `--next` validation + `--paginate --next` path)
- [x] `internal/cli/productpages` (`custom-pages localizations preview-sets list` invalid/malformed `--next` validation + `--paginate --next` without `--localization-id`)
- [x] `internal/cli/productpages` (`custom-pages localizations screenshot-sets list` invalid/malformed `--next` validation + `--paginate --next` without `--localization-id`)
- [x] Commit Phase 50

## Phase 51 - `--next` Parity for Game Center Achievement Lists, Localizations, and Releases

- [x] `internal/cli/gamecenter` (`game-center achievements list` invalid/malformed `--next` validation + `--paginate --next` without `--app`)
- [x] `internal/cli/gamecenter` (`game-center achievements localizations list` invalid/malformed `--next` validation + `--paginate --next` without `--achievement-id`)
- [x] `internal/cli/gamecenter` (`game-center achievements releases list` invalid/malformed `--next` validation + `--paginate --next` without `--achievement-id`)
- [x] Commit Phase 51

## Phase 52 - `--next` Parity for Game Center Achievement V2 Lists, Versions, and Localizations

- [x] `internal/cli/gamecenter` (`game-center achievements v2 list` invalid/malformed `--next` validation + `--paginate --next` without `--app`/`--group-id`)
- [x] `internal/cli/gamecenter` (`game-center achievements v2 versions list` invalid/malformed `--next` validation + `--paginate --next` without `--achievement-id`)
- [x] `internal/cli/gamecenter` (`game-center achievements v2 localizations list` invalid/malformed `--next` validation + `--paginate --next` without `--version-id`)
- [x] Commit Phase 52

## Phase 53 - `--next` Parity for Game Center Activities Lists, Versions, Localizations, and Releases

- [x] `internal/cli/gamecenter` (`game-center activities list` invalid/malformed `--next` validation + `--paginate --next` without `--app`)
- [x] `internal/cli/gamecenter` (`game-center activities versions list` invalid/malformed `--next` validation + `--paginate --next` without `--activity-id`)
- [x] `internal/cli/gamecenter` (`game-center activities localizations list` invalid/malformed `--next` validation + `--paginate --next` without `--version-id`)
- [x] `internal/cli/gamecenter` (`game-center activities releases list` invalid/malformed `--next` validation + `--paginate --next` without `--app`)
- [x] Commit Phase 53

## Phase 54 - `--next` Parity for Game Center Challenges Lists, Versions, Localizations, and Releases

- [x] `internal/cli/gamecenter` (`game-center challenges list` invalid/malformed `--next` validation + `--paginate --next` without `--app`)
- [x] `internal/cli/gamecenter` (`game-center challenges versions list` invalid/malformed `--next` validation + `--paginate --next` without `--challenge-id`)
- [x] `internal/cli/gamecenter` (`game-center challenges localizations list` invalid/malformed `--next` validation + `--paginate --next` without `--version-id`)
- [x] `internal/cli/gamecenter` (`game-center challenges releases list` invalid/malformed `--next` validation + `--paginate --next` without `--app`)
- [x] Commit Phase 54

## Phase 55 - `--next` Parity for Game Center App Version and Enabled Version Lists

- [x] `internal/cli/gamecenter` (`game-center app-versions list` invalid/malformed `--next` validation + `--paginate --next` without `--app`)
- [x] `internal/cli/gamecenter` (`game-center app-versions compatibility list` invalid/malformed `--next` validation + `--paginate --next` without `--id`)
- [x] `internal/cli/gamecenter` (`game-center enabled-versions list` invalid/malformed `--next` validation + `--paginate --next` without `--app`)
- [x] `internal/cli/gamecenter` (`game-center enabled-versions compatible-versions` invalid/malformed `--next` validation + `--paginate --next` without `--id`)
- [x] Commit Phase 55

## Phase 56 - `--next` Parity for Game Center Details List Surfaces

- [x] `internal/cli/gamecenter` (`game-center details app-versions list` invalid/malformed `--next` validation + `--paginate --next` without `--id`)
- [x] `internal/cli/gamecenter` (`game-center details achievements-v2 list` invalid/malformed `--next` validation + `--paginate --next` without `--id`)
- [x] `internal/cli/gamecenter` (`game-center details leaderboards-v2 list` invalid/malformed `--next` validation + `--paginate --next` without `--id`)
- [x] `internal/cli/gamecenter` (`game-center details leaderboard-sets-v2 list` invalid/malformed `--next` validation + `--paginate --next` without `--id`)
- [x] `internal/cli/gamecenter` (`game-center details achievement-releases list` invalid/malformed `--next` validation + `--paginate --next` without `--id`)
- [x] `internal/cli/gamecenter` (`game-center details leaderboard-releases list` invalid/malformed `--next` validation + `--paginate --next` without `--id`)
- [x] `internal/cli/gamecenter` (`game-center details leaderboard-set-releases list` invalid/malformed `--next` validation + `--paginate --next` without `--id`)
- [x] Commit Phase 56

## Phase 57 - `--next` Parity for Game Center Group List and Relationship Surfaces

- [x] `internal/cli/gamecenter` (`game-center groups list` invalid/malformed `--next` validation + `--paginate --next` without `--app`)
- [x] `internal/cli/gamecenter` (`game-center groups achievements list` invalid/malformed `--next` validation + `--paginate --next` without `--group-id`)
- [x] `internal/cli/gamecenter` (`game-center groups leaderboards list` invalid/malformed `--next` validation + `--paginate --next` without `--group-id`)
- [x] `internal/cli/gamecenter` (`game-center groups leaderboard-sets list` invalid/malformed `--next` validation + `--paginate --next` without `--group-id`)
- [x] `internal/cli/gamecenter` (`game-center groups activities list` invalid/malformed `--next` validation + `--paginate --next` without `--group-id`)
- [x] `internal/cli/gamecenter` (`game-center groups challenges list` invalid/malformed `--next` validation + `--paginate --next` without `--group-id`)
- [x] `internal/cli/gamecenter` (`game-center groups details list` invalid/malformed `--next` validation + `--paginate --next` without `--group-id`)
- [x] Commit Phase 57

## Phase 58 - `--next` Parity for Game Center Leaderboard List Surfaces

- [x] `internal/cli/gamecenter` (`game-center leaderboards list` invalid/malformed `--next` validation + `--paginate --next` without `--app`)
- [x] `internal/cli/gamecenter` (`game-center leaderboards localizations list` invalid/malformed `--next` validation + `--paginate --next` without `--leaderboard-id`)
- [x] `internal/cli/gamecenter` (`game-center leaderboards releases list` invalid/malformed `--next` validation + `--paginate --next` without `--id`)
- [x] `internal/cli/gamecenter` (`game-center leaderboards v2 list` invalid/malformed `--next` validation + `--paginate --next` without `--app`/`--group-id`)
- [x] `internal/cli/gamecenter` (`game-center leaderboards v2 versions list` invalid/malformed `--next` validation + `--paginate --next` without `--leaderboard-id`)
- [x] `internal/cli/gamecenter` (`game-center leaderboards v2 localizations list` invalid/malformed `--next` validation + `--paginate --next` without `--version-id`)
- [x] Commit Phase 58

## Phase 59 - `--next` Parity for Game Center Leaderboard Set List Surfaces

- [x] `internal/cli/gamecenter` (`game-center leaderboard-sets list` invalid/malformed `--next` validation + `--paginate --next` without `--app`)
- [x] `internal/cli/gamecenter` (`game-center leaderboard-sets localizations list` invalid/malformed `--next` validation + `--paginate --next` without `--set-id`)
- [x] `internal/cli/gamecenter` (`game-center leaderboard-sets member-localizations list` invalid/malformed `--next` validation + `--paginate --next` without `--set-id`/`--leaderboard-id`)
- [x] `internal/cli/gamecenter` (`game-center leaderboard-sets members list` invalid/malformed `--next` validation + `--paginate --next` without `--set-id`)
- [x] `internal/cli/gamecenter` (`game-center leaderboard-sets releases list` invalid/malformed `--next` validation + `--paginate --next` without `--set-id`)
- [x] `internal/cli/gamecenter` (`game-center leaderboard-sets v2 list` invalid/malformed `--next` validation + `--paginate --next` without `--app`/`--group-id`)
- [x] `internal/cli/gamecenter` (`game-center leaderboard-sets v2 localizations list` invalid/malformed `--next` validation + `--paginate --next` without `--version-id`)
- [x] `internal/cli/gamecenter` (`game-center leaderboard-sets v2 members list` invalid/malformed `--next` validation + `--paginate --next` without `--set-id`)
- [x] `internal/cli/gamecenter` (`game-center leaderboard-sets v2 versions list` invalid/malformed `--next` validation + `--paginate --next` without `--set-id`)
- [x] Commit Phase 59

## Phase 60 - `--next` Parity for Game Center Matchmaking List Surfaces

- [x] `internal/cli/gamecenter` (`game-center matchmaking queues list` invalid/malformed `--next` validation + `--paginate --next` path)
- [x] `internal/cli/gamecenter` (`game-center matchmaking rule-sets list` invalid/malformed `--next` validation + `--paginate --next` path)
- [x] `internal/cli/gamecenter` (`game-center matchmaking rule-sets queues list` invalid/malformed `--next` validation + `--paginate --next` without `--rule-set-id`)
- [x] `internal/cli/gamecenter` (`game-center matchmaking rules list` invalid/malformed `--next` validation + `--paginate --next` without `--rule-set-id`)
- [x] `internal/cli/gamecenter` (`game-center matchmaking teams list` invalid/malformed `--next` validation + `--paginate --next` without `--rule-set-id`)
- [x] Commit Phase 60

## Phase 61 - `--next` Parity for IAP List Surfaces

- [x] `internal/cli/iap` (`iap list` invalid/malformed `--next` validation + `--paginate --next` without `--app`)
- [x] `internal/cli/iap` (`iap images list` invalid/malformed `--next` validation + `--paginate --next` without `--iap-id`)
- [x] `internal/cli/iap` (`iap localizations list` invalid/malformed `--next` validation + `--paginate --next` without `--iap-id`)
- [x] `internal/cli/iap` (`iap offer-codes custom-codes list` invalid/malformed `--next` validation + `--paginate --next` without `--offer-code-id`)
- [x] `internal/cli/iap` (`iap offer-codes one-time-codes list` invalid/malformed `--next` validation + `--paginate --next` without `--offer-code-id`)
- [x] `internal/cli/iap` (`iap offer-codes prices` invalid/malformed `--next` validation + `--paginate --next` without `--offer-code-id`)
- [x] `internal/cli/iap` (`iap price-points list` invalid/malformed `--next` validation + `--paginate --next` without `--iap-id`)
- [x] `internal/cli/iap` (`iap price-schedules automatic-prices` invalid/malformed `--next` validation + `--paginate --next` without `--schedule-id`)
- [x] `internal/cli/iap` (`iap price-schedules manual-prices` invalid/malformed `--next` validation + `--paginate --next` without `--schedule-id`)
- [x] `internal/cli/iap` (`iap availabilities available-territories` invalid/malformed `--next` validation + `--paginate --next` without `--id`)
- [x] Commit Phase 61

## Validation Gate (after each phase and at end)

- [x] `go test -short ./...`
- [x] `go test -short -coverprofile=coverage.out ./...`
- [x] verify previously-targeted packages are no longer `0.0%`
- [x] `make test`
- [x] `make lint`
