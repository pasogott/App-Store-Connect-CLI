package metadata

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/peterbourgon/ff/v3/ffcli"

	"github.com/rudrankriyam/App-Store-Connect-CLI/internal/asc"
	"github.com/rudrankriyam/App-Store-Connect-CLI/internal/cli/shared"
)

var (
	appInfoPlanFields = []string{
		"name",
		"subtitle",
		"privacyPolicyUrl",
		"privacyChoicesUrl",
		"privacyPolicyText",
	}
	versionPlanFields = []string{
		"description",
		"keywords",
		"marketingUrl",
		"promotionalText",
		"supportUrl",
		"whatsNew",
	}
)

// PlanItem represents one deterministic metadata change entry.
type PlanItem struct {
	Key     string `json:"key"`
	Scope   string `json:"scope"`
	Locale  string `json:"locale"`
	Version string `json:"version,omitempty"`
	Field   string `json:"field"`
	Reason  string `json:"reason"`
	From    string `json:"from,omitempty"`
	To      string `json:"to,omitempty"`
}

// PlanAPICall is an estimated API call summary for the plan.
type PlanAPICall struct {
	Operation string `json:"operation"`
	Scope     string `json:"scope"`
	Count     int    `json:"count"`
}

// ApplyAction represents one executed mutation action.
type ApplyAction struct {
	Scope          string `json:"scope"`
	Locale         string `json:"locale"`
	Version        string `json:"version,omitempty"`
	Action         string `json:"action"`
	LocalizationID string `json:"localizationId,omitempty"`
}

// PushPlanResult is the push dry-run output artifact.
type PushPlanResult struct {
	AppID     string        `json:"appId"`
	AppInfoID string        `json:"appInfoId"`
	Version   string        `json:"version"`
	VersionID string        `json:"versionId"`
	Dir       string        `json:"dir"`
	DryRun    bool          `json:"dryRun"`
	Applied   bool          `json:"applied,omitempty"`
	Includes  []string      `json:"includes"`
	Adds      []PlanItem    `json:"adds"`
	Updates   []PlanItem    `json:"updates"`
	Deletes   []PlanItem    `json:"deletes"`
	APICalls  []PlanAPICall `json:"apiCalls,omitempty"`
	Actions   []ApplyAction `json:"actions,omitempty"`
}

type scopeCallCounts struct {
	create int
	update int
	delete int
}

type localMetadataBundle struct {
	appInfo        map[string]AppInfoLocalization
	version        map[string]VersionLocalization
	defaultAppInfo *AppInfoLocalization
	defaultVersion *VersionLocalization
}

// MetadataPushCommand returns the metadata push subcommand.
func MetadataPushCommand() *ffcli.Command {
	fs := flag.NewFlagSet("metadata push", flag.ExitOnError)

	appID := fs.String("app", "", "App Store Connect app ID (or ASC_APP_ID env)")
	version := fs.String("version", "", "App version string (for example 1.2.3)")
	platform := fs.String("platform", "", "Optional platform: IOS, MAC_OS, TV_OS, or VISION_OS")
	dir := fs.String("dir", "", "Metadata root directory (required)")
	include := fs.String("include", includeLocalizations, "Included metadata scopes (comma-separated)")
	dryRun := fs.Bool("dry-run", false, "Preview changes without mutating App Store Connect")
	allowDeletes := fs.Bool("allow-deletes", false, "Allow destructive delete operations when applying changes")
	confirm := fs.Bool("confirm", false, "Confirm destructive operations (required with --allow-deletes)")
	output := shared.BindOutputFlags(fs)

	return &ffcli.Command{
		Name:       "push",
		ShortUsage: "asc metadata push --app \"APP_ID\" --version \"1.2.3\" --dir \"./metadata\" [--dry-run]",
		ShortHelp:  "Push metadata changes from canonical files.",
		LongHelp: `Push metadata changes from canonical files.

Examples:
  asc metadata push --app "APP_ID" --version "1.2.3" --dir "./metadata" --dry-run
  asc metadata push --app "APP_ID" --version "1.2.3" --platform IOS --dir "./metadata" --dry-run
  asc metadata push --app "APP_ID" --version "1.2.3" --dir "./metadata"
  asc metadata push --app "APP_ID" --version "1.2.3" --dir "./metadata" --allow-deletes --confirm`,
		FlagSet:   fs,
		UsageFunc: shared.DefaultUsageFunc,
		Exec: func(ctx context.Context, args []string) error {
			if len(args) > 0 {
				return shared.UsageError("metadata push does not accept positional arguments")
			}
			resolvedAppID := shared.ResolveAppID(*appID)
			if resolvedAppID == "" {
				return shared.UsageError("--app is required (or set ASC_APP_ID)")
			}
			versionValue := strings.TrimSpace(*version)
			if versionValue == "" {
				return shared.UsageError("--version is required")
			}
			dirValue := strings.TrimSpace(*dir)
			if dirValue == "" {
				return shared.UsageError("--dir is required")
			}

			platformValue := strings.TrimSpace(*platform)
			if platformValue != "" {
				normalizedPlatform, err := shared.NormalizeAppStoreVersionPlatform(platformValue)
				if err != nil {
					return shared.UsageError(err.Error())
				}
				platformValue = normalizedPlatform
			}

			includes, err := parseIncludes(*include)
			if err != nil {
				return shared.UsageError(err.Error())
			}

			localBundle, err := loadLocalMetadata(dirValue, versionValue)
			if err != nil {
				return err
			}

			client, err := shared.GetASCClient()
			if err != nil {
				return fmt.Errorf("metadata push: %w", err)
			}

			requestCtx, cancel := shared.ContextWithTimeout(ctx)
			defer cancel()

			appInfoIDValue, err := shared.ResolveAppInfoID(requestCtx, client, resolvedAppID, "")
			if err != nil {
				return fmt.Errorf("metadata push: %w", err)
			}
			versionIDValue, err := resolveVersionID(requestCtx, client, resolvedAppID, versionValue, platformValue)
			if err != nil {
				if errors.Is(err, flag.ErrHelp) {
					return err
				}
				return fmt.Errorf("metadata push: %w", err)
			}

			remoteAppInfoItems, err := fetchAppInfoLocalizations(requestCtx, client, appInfoIDValue)
			if err != nil {
				return fmt.Errorf("metadata push: %w", err)
			}
			remoteVersionItems, err := fetchVersionLocalizations(requestCtx, client, versionIDValue)
			if err != nil {
				return fmt.Errorf("metadata push: %w", err)
			}

			remoteAppInfo := make(map[string]AppInfoLocalization, len(remoteAppInfoItems))
			for _, item := range remoteAppInfoItems {
				locale := strings.TrimSpace(item.Attributes.Locale)
				if locale == "" {
					continue
				}
				remoteAppInfo[locale] = NormalizeAppInfoLocalization(AppInfoLocalization{
					Name:              item.Attributes.Name,
					Subtitle:          item.Attributes.Subtitle,
					PrivacyPolicyURL:  item.Attributes.PrivacyPolicyURL,
					PrivacyChoicesURL: item.Attributes.PrivacyChoicesURL,
					PrivacyPolicyText: item.Attributes.PrivacyPolicyText,
				})
			}

			remoteVersion := make(map[string]VersionLocalization, len(remoteVersionItems))
			for _, item := range remoteVersionItems {
				locale := strings.TrimSpace(item.Attributes.Locale)
				if locale == "" {
					continue
				}
				remoteVersion[locale] = NormalizeVersionLocalization(VersionLocalization{
					Description:     item.Attributes.Description,
					Keywords:        item.Attributes.Keywords,
					MarketingURL:    item.Attributes.MarketingURL,
					PromotionalText: item.Attributes.PromotionalText,
					SupportURL:      item.Attributes.SupportURL,
					WhatsNew:        item.Attributes.WhatsNew,
				})
			}

			localAppInfo := applyDefaultAppInfoFallback(localBundle.appInfo, localBundle.defaultAppInfo, remoteAppInfo)
			localVersion := applyDefaultVersionFallback(localBundle.version, localBundle.defaultVersion, remoteVersion)

			adds, updates, deletes, appInfoCalls := buildScopePlan(
				appInfoDirName,
				"",
				appInfoPlanFields,
				appInfoToFieldMap(localAppInfo),
				appInfoToFieldMap(remoteAppInfo),
			)
			versionAdds, versionUpdates, versionDeletes, versionCalls := buildScopePlan(
				versionDirName,
				versionValue,
				versionPlanFields,
				versionToFieldMap(localVersion),
				versionToFieldMap(remoteVersion),
			)
			adds = append(adds, versionAdds...)
			updates = append(updates, versionUpdates...)
			deletes = append(deletes, versionDeletes...)

			sortPlanItems(adds)
			sortPlanItems(updates)
			sortPlanItems(deletes)

			apiCalls := buildAPICallSummary(appInfoCalls, versionCalls)

			result := PushPlanResult{
				AppID:     resolvedAppID,
				AppInfoID: appInfoIDValue,
				Version:   versionValue,
				VersionID: versionIDValue,
				Dir:       dirValue,
				DryRun:    *dryRun,
				Includes:  includes,
				Adds:      adds,
				Updates:   updates,
				Deletes:   deletes,
				APICalls:  apiCalls,
			}

			if !*dryRun {
				if len(result.Deletes) > 0 {
					if !*allowDeletes {
						return shared.UsageError("--allow-deletes is required to apply delete operations")
					}
					if !*confirm {
						return shared.UsageError("--confirm is required when applying delete operations")
					}
				}

				actions, applyErr := applyMetadataPlan(
					requestCtx,
					client,
					appInfoIDValue,
					versionIDValue,
					versionValue,
					localAppInfo,
					localVersion,
					remoteAppInfoItems,
					remoteVersionItems,
					*allowDeletes,
				)
				if applyErr != nil {
					return fmt.Errorf("metadata push: %w", applyErr)
				}
				result.Applied = true
				result.Actions = actions
			}

			return shared.PrintOutputWithRenderers(
				result,
				*output.Output,
				*output.Pretty,
				func() error { return printPushPlanTable(result) },
				func() error { return printPushPlanMarkdown(result) },
			)
		},
	}
}

func loadLocalMetadata(dir, version string) (localMetadataBundle, error) {
	localAppInfo := make(map[string]AppInfoLocalization)
	localVersion := make(map[string]VersionLocalization)
	var defaultAppInfo *AppInfoLocalization
	var defaultVersion *VersionLocalization
	filesSeen := 0

	appInfoDir := filepath.Join(dir, appInfoDirName)
	appInfoEntries, err := os.ReadDir(appInfoDir)
	if err != nil && !errors.Is(err, os.ErrNotExist) {
		return localMetadataBundle{}, fmt.Errorf("metadata push: failed to read %s: %w", appInfoDir, err)
	}
	if err == nil {
		for _, entry := range appInfoEntries {
			if entry.IsDir() || filepath.Ext(entry.Name()) != ".json" {
				continue
			}
			locale := strings.TrimSuffix(entry.Name(), ".json")
			resolvedLocale, localeErr := validateLocale(locale)
			if localeErr != nil {
				return localMetadataBundle{}, shared.UsageErrorf("invalid app-info localization file %q: %v", entry.Name(), localeErr)
			}
			filePath := filepath.Join(appInfoDir, entry.Name())
			loc, readErr := ReadAppInfoLocalizationFile(filePath)
			if readErr != nil {
				return localMetadataBundle{}, shared.UsageErrorf("invalid metadata schema in %s: %v", filePath, readErr)
			}
			issues := ValidateAppInfoLocalization(loc, ValidationOptions{})
			for _, issue := range issues {
				if issue.Field == "metadata" {
					return localMetadataBundle{}, shared.UsageErrorf("invalid metadata in %s: %s", filePath, issue.Message)
				}
			}
			if resolvedLocale == DefaultLocale {
				value := loc
				defaultAppInfo = &value
				filesSeen++
				continue
			}
			localAppInfo[resolvedLocale] = loc
			filesSeen++
		}
	}

	versionDir := filepath.Join(dir, versionDirName, version)
	versionEntries, err := os.ReadDir(versionDir)
	if err != nil && !errors.Is(err, os.ErrNotExist) {
		return localMetadataBundle{}, fmt.Errorf("metadata push: failed to read %s: %w", versionDir, err)
	}
	if err == nil {
		for _, entry := range versionEntries {
			if entry.IsDir() || filepath.Ext(entry.Name()) != ".json" {
				continue
			}
			locale := strings.TrimSuffix(entry.Name(), ".json")
			resolvedLocale, localeErr := validateLocale(locale)
			if localeErr != nil {
				return localMetadataBundle{}, shared.UsageErrorf("invalid version localization file %q: %v", entry.Name(), localeErr)
			}
			filePath := filepath.Join(versionDir, entry.Name())
			loc, readErr := ReadVersionLocalizationFile(filePath)
			if readErr != nil {
				return localMetadataBundle{}, shared.UsageErrorf("invalid metadata schema in %s: %v", filePath, readErr)
			}
			issues := ValidateVersionLocalization(loc)
			if len(issues) > 0 {
				return localMetadataBundle{}, shared.UsageErrorf("invalid metadata in %s: %s", filePath, issues[0].Message)
			}
			if resolvedLocale == DefaultLocale {
				value := loc
				defaultVersion = &value
				filesSeen++
				continue
			}
			localVersion[resolvedLocale] = loc
			filesSeen++
		}
	}

	if filesSeen == 0 {
		return localMetadataBundle{}, shared.UsageError("no metadata .json files found")
	}
	return localMetadataBundle{
		appInfo:        localAppInfo,
		version:        localVersion,
		defaultAppInfo: defaultAppInfo,
		defaultVersion: defaultVersion,
	}, nil
}

func applyDefaultAppInfoFallback(
	explicit map[string]AppInfoLocalization,
	defaultValue *AppInfoLocalization,
	remote map[string]AppInfoLocalization,
) map[string]AppInfoLocalization {
	result := make(map[string]AppInfoLocalization, len(explicit))
	for locale, value := range explicit {
		result[locale] = value
	}
	if defaultValue == nil {
		return result
	}
	for locale := range remote {
		if locale == DefaultLocale {
			continue
		}
		if _, ok := result[locale]; ok {
			continue
		}
		result[locale] = *defaultValue
	}
	return result
}

func applyDefaultVersionFallback(
	explicit map[string]VersionLocalization,
	defaultValue *VersionLocalization,
	remote map[string]VersionLocalization,
) map[string]VersionLocalization {
	result := make(map[string]VersionLocalization, len(explicit))
	for locale, value := range explicit {
		result[locale] = value
	}
	if defaultValue == nil {
		return result
	}
	for locale := range remote {
		if locale == DefaultLocale {
			continue
		}
		if _, ok := result[locale]; ok {
			continue
		}
		result[locale] = *defaultValue
	}
	return result
}

type remoteLocalizationState struct {
	id     string
	fields map[string]string
}

func applyMetadataPlan(
	ctx context.Context,
	client *asc.Client,
	appInfoID string,
	versionID string,
	version string,
	localAppInfo map[string]AppInfoLocalization,
	localVersion map[string]VersionLocalization,
	remoteAppInfoItems []asc.Resource[asc.AppInfoLocalizationAttributes],
	remoteVersionItems []asc.Resource[asc.AppStoreVersionLocalizationAttributes],
	allowDeletes bool,
) ([]ApplyAction, error) {
	actions := make([]ApplyAction, 0)

	appInfoActions, err := applyAppInfoChanges(ctx, client, appInfoID, localAppInfo, remoteAppInfoItems, allowDeletes)
	if err != nil {
		return nil, err
	}
	actions = append(actions, appInfoActions...)

	versionActions, err := applyVersionChanges(ctx, client, versionID, version, localVersion, remoteVersionItems, allowDeletes)
	if err != nil {
		return nil, err
	}
	actions = append(actions, versionActions...)

	return actions, nil
}

func applyAppInfoChanges(
	ctx context.Context,
	client *asc.Client,
	appInfoID string,
	local map[string]AppInfoLocalization,
	remoteItems []asc.Resource[asc.AppInfoLocalizationAttributes],
	allowDeletes bool,
) ([]ApplyAction, error) {
	remoteByLocale := make(map[string]remoteLocalizationState, len(remoteItems))
	for _, item := range remoteItems {
		locale := strings.TrimSpace(item.Attributes.Locale)
		if locale == "" {
			continue
		}
		remoteByLocale[locale] = remoteLocalizationState{
			id: item.ID,
			fields: appInfoFields(AppInfoLocalization{
				Name:              item.Attributes.Name,
				Subtitle:          item.Attributes.Subtitle,
				PrivacyPolicyURL:  item.Attributes.PrivacyPolicyURL,
				PrivacyChoicesURL: item.Attributes.PrivacyChoicesURL,
				PrivacyPolicyText: item.Attributes.PrivacyPolicyText,
			}),
		}
	}

	locales := make([]string, 0, len(local)+len(remoteByLocale))
	localeSet := make(map[string]struct{})
	for locale := range local {
		localeSet[locale] = struct{}{}
	}
	for locale := range remoteByLocale {
		localeSet[locale] = struct{}{}
	}
	for locale := range localeSet {
		locales = append(locales, locale)
	}
	sort.Strings(locales)

	actions := make([]ApplyAction, 0)
	for _, locale := range locales {
		localLoc, localExists := local[locale]
		localFields := appInfoFields(localLoc)

		remoteState, remoteExists := remoteByLocale[locale]
		remoteFields := remoteState.fields

		adds, updates, deletes := countFieldChanges(appInfoPlanFields, localFields, remoteFields)
		if adds == 0 && updates == 0 && deletes == 0 {
			continue
		}

		switch {
		case localExists && !remoteExists:
			if strings.TrimSpace(localLoc.Name) == "" {
				return nil, fmt.Errorf("cannot create app-info localization %q without name", locale)
			}
			resp, err := client.CreateAppInfoLocalization(ctx, appInfoID, appInfoAttributes(locale, localLoc, true))
			if err != nil {
				return nil, fmt.Errorf("create app-info localization %s: %w", locale, err)
			}
			actions = append(actions, ApplyAction{
				Scope:          appInfoDirName,
				Locale:         locale,
				Action:         "create",
				LocalizationID: resp.Data.ID,
			})
		case !localExists && remoteExists:
			if !allowDeletes {
				return nil, fmt.Errorf("delete operations require --allow-deletes")
			}
			if err := client.DeleteAppInfoLocalization(ctx, remoteState.id); err != nil {
				return nil, fmt.Errorf("delete app-info localization %s: %w", locale, err)
			}
			actions = append(actions, ApplyAction{
				Scope:          appInfoDirName,
				Locale:         locale,
				Action:         "delete",
				LocalizationID: remoteState.id,
			})
		case localExists && remoteExists:
			if deletes > 0 {
				if !allowDeletes {
					return nil, fmt.Errorf("delete operations require --allow-deletes")
				}
				if strings.TrimSpace(localLoc.Name) == "" {
					return nil, fmt.Errorf("cannot recreate app-info localization %q without name", locale)
				}
				if err := client.DeleteAppInfoLocalization(ctx, remoteState.id); err != nil {
					return nil, fmt.Errorf("delete app-info localization %s before recreate: %w", locale, err)
				}
				actions = append(actions, ApplyAction{
					Scope:          appInfoDirName,
					Locale:         locale,
					Action:         "delete",
					LocalizationID: remoteState.id,
				})
				resp, err := client.CreateAppInfoLocalization(ctx, appInfoID, appInfoAttributes(locale, localLoc, true))
				if err != nil {
					return nil, fmt.Errorf("recreate app-info localization %s: %w", locale, err)
				}
				actions = append(actions, ApplyAction{
					Scope:          appInfoDirName,
					Locale:         locale,
					Action:         "create",
					LocalizationID: resp.Data.ID,
				})
				continue
			}
			resp, err := client.UpdateAppInfoLocalization(ctx, remoteState.id, appInfoAttributes(locale, localLoc, false))
			if err != nil {
				return nil, fmt.Errorf("update app-info localization %s: %w", locale, err)
			}
			actions = append(actions, ApplyAction{
				Scope:          appInfoDirName,
				Locale:         locale,
				Action:         "update",
				LocalizationID: resp.Data.ID,
			})
		}
	}

	return actions, nil
}

func applyVersionChanges(
	ctx context.Context,
	client *asc.Client,
	versionID string,
	version string,
	local map[string]VersionLocalization,
	remoteItems []asc.Resource[asc.AppStoreVersionLocalizationAttributes],
	allowDeletes bool,
) ([]ApplyAction, error) {
	remoteByLocale := make(map[string]remoteLocalizationState, len(remoteItems))
	for _, item := range remoteItems {
		locale := strings.TrimSpace(item.Attributes.Locale)
		if locale == "" {
			continue
		}
		remoteByLocale[locale] = remoteLocalizationState{
			id: item.ID,
			fields: versionFields(VersionLocalization{
				Description:     item.Attributes.Description,
				Keywords:        item.Attributes.Keywords,
				MarketingURL:    item.Attributes.MarketingURL,
				PromotionalText: item.Attributes.PromotionalText,
				SupportURL:      item.Attributes.SupportURL,
				WhatsNew:        item.Attributes.WhatsNew,
			}),
		}
	}

	locales := make([]string, 0, len(local)+len(remoteByLocale))
	localeSet := make(map[string]struct{})
	for locale := range local {
		localeSet[locale] = struct{}{}
	}
	for locale := range remoteByLocale {
		localeSet[locale] = struct{}{}
	}
	for locale := range localeSet {
		locales = append(locales, locale)
	}
	sort.Strings(locales)

	actions := make([]ApplyAction, 0)
	for _, locale := range locales {
		localLoc, localExists := local[locale]
		localFields := versionFields(localLoc)
		remoteState, remoteExists := remoteByLocale[locale]
		remoteFields := remoteState.fields

		adds, updates, deletes := countFieldChanges(versionPlanFields, localFields, remoteFields)
		if adds == 0 && updates == 0 && deletes == 0 {
			continue
		}

		switch {
		case localExists && !remoteExists:
			resp, err := client.CreateAppStoreVersionLocalization(ctx, versionID, versionAttributes(locale, localLoc, true))
			if err != nil {
				return nil, fmt.Errorf("create version localization %s: %w", locale, err)
			}
			actions = append(actions, ApplyAction{
				Scope:          versionDirName,
				Locale:         locale,
				Version:        version,
				Action:         "create",
				LocalizationID: resp.Data.ID,
			})
		case !localExists && remoteExists:
			if !allowDeletes {
				return nil, fmt.Errorf("delete operations require --allow-deletes")
			}
			if err := client.DeleteAppStoreVersionLocalization(ctx, remoteState.id); err != nil {
				return nil, fmt.Errorf("delete version localization %s: %w", locale, err)
			}
			actions = append(actions, ApplyAction{
				Scope:          versionDirName,
				Locale:         locale,
				Version:        version,
				Action:         "delete",
				LocalizationID: remoteState.id,
			})
		case localExists && remoteExists:
			if deletes > 0 {
				if !allowDeletes {
					return nil, fmt.Errorf("delete operations require --allow-deletes")
				}
				if err := client.DeleteAppStoreVersionLocalization(ctx, remoteState.id); err != nil {
					return nil, fmt.Errorf("delete version localization %s before recreate: %w", locale, err)
				}
				actions = append(actions, ApplyAction{
					Scope:          versionDirName,
					Locale:         locale,
					Version:        version,
					Action:         "delete",
					LocalizationID: remoteState.id,
				})
				resp, err := client.CreateAppStoreVersionLocalization(ctx, versionID, versionAttributes(locale, localLoc, true))
				if err != nil {
					return nil, fmt.Errorf("recreate version localization %s: %w", locale, err)
				}
				actions = append(actions, ApplyAction{
					Scope:          versionDirName,
					Locale:         locale,
					Version:        version,
					Action:         "create",
					LocalizationID: resp.Data.ID,
				})
				continue
			}
			resp, err := client.UpdateAppStoreVersionLocalization(ctx, remoteState.id, versionAttributes(locale, localLoc, false))
			if err != nil {
				return nil, fmt.Errorf("update version localization %s: %w", locale, err)
			}
			actions = append(actions, ApplyAction{
				Scope:          versionDirName,
				Locale:         locale,
				Version:        version,
				Action:         "update",
				LocalizationID: resp.Data.ID,
			})
		}
	}

	return actions, nil
}

func appInfoAttributes(locale string, loc AppInfoLocalization, includeLocale bool) asc.AppInfoLocalizationAttributes {
	normalized := NormalizeAppInfoLocalization(loc)
	attrs := asc.AppInfoLocalizationAttributes{
		Name:              normalized.Name,
		Subtitle:          normalized.Subtitle,
		PrivacyPolicyURL:  normalized.PrivacyPolicyURL,
		PrivacyChoicesURL: normalized.PrivacyChoicesURL,
		PrivacyPolicyText: normalized.PrivacyPolicyText,
	}
	if includeLocale {
		attrs.Locale = locale
	}
	return attrs
}

func versionAttributes(locale string, loc VersionLocalization, includeLocale bool) asc.AppStoreVersionLocalizationAttributes {
	normalized := NormalizeVersionLocalization(loc)
	attrs := asc.AppStoreVersionLocalizationAttributes{
		Description:     normalized.Description,
		Keywords:        normalized.Keywords,
		MarketingURL:    normalized.MarketingURL,
		PromotionalText: normalized.PromotionalText,
		SupportURL:      normalized.SupportURL,
		WhatsNew:        normalized.WhatsNew,
	}
	if includeLocale {
		attrs.Locale = locale
	}
	return attrs
}

func countFieldChanges(fields []string, local map[string]string, remote map[string]string) (int, int, int) {
	adds := 0
	updates := 0
	deletes := 0
	for _, field := range fields {
		localValue, localHasField := local[field]
		remoteValue, remoteHasField := remote[field]
		switch {
		case !remoteHasField && localHasField:
			adds++
		case remoteHasField && !localHasField:
			deletes++
		case remoteHasField && localHasField && remoteValue != localValue:
			updates++
		}
	}
	return adds, updates, deletes
}

func appInfoToFieldMap(values map[string]AppInfoLocalization) map[string]map[string]string {
	result := make(map[string]map[string]string, len(values))
	for locale, value := range values {
		result[locale] = appInfoFields(value)
	}
	return result
}

func versionToFieldMap(values map[string]VersionLocalization) map[string]map[string]string {
	result := make(map[string]map[string]string, len(values))
	for locale, value := range values {
		result[locale] = versionFields(value)
	}
	return result
}

func appInfoFields(value AppInfoLocalization) map[string]string {
	fields := make(map[string]string)
	normalized := NormalizeAppInfoLocalization(value)
	if normalized.Name != "" {
		fields["name"] = normalized.Name
	}
	if normalized.Subtitle != "" {
		fields["subtitle"] = normalized.Subtitle
	}
	if normalized.PrivacyPolicyURL != "" {
		fields["privacyPolicyUrl"] = normalized.PrivacyPolicyURL
	}
	if normalized.PrivacyChoicesURL != "" {
		fields["privacyChoicesUrl"] = normalized.PrivacyChoicesURL
	}
	if normalized.PrivacyPolicyText != "" {
		fields["privacyPolicyText"] = normalized.PrivacyPolicyText
	}
	return fields
}

func versionFields(value VersionLocalization) map[string]string {
	fields := make(map[string]string)
	normalized := NormalizeVersionLocalization(value)
	if normalized.Description != "" {
		fields["description"] = normalized.Description
	}
	if normalized.Keywords != "" {
		fields["keywords"] = normalized.Keywords
	}
	if normalized.MarketingURL != "" {
		fields["marketingUrl"] = normalized.MarketingURL
	}
	if normalized.PromotionalText != "" {
		fields["promotionalText"] = normalized.PromotionalText
	}
	if normalized.SupportURL != "" {
		fields["supportUrl"] = normalized.SupportURL
	}
	if normalized.WhatsNew != "" {
		fields["whatsNew"] = normalized.WhatsNew
	}
	return fields
}

func buildScopePlan(
	scope string,
	version string,
	fields []string,
	local map[string]map[string]string,
	remote map[string]map[string]string,
) ([]PlanItem, []PlanItem, []PlanItem, scopeCallCounts) {
	localesMap := make(map[string]struct{})
	for locale := range local {
		localesMap[locale] = struct{}{}
	}
	for locale := range remote {
		localesMap[locale] = struct{}{}
	}

	locales := make([]string, 0, len(localesMap))
	for locale := range localesMap {
		locales = append(locales, locale)
	}
	sort.Strings(locales)

	adds := make([]PlanItem, 0)
	updates := make([]PlanItem, 0)
	deletes := make([]PlanItem, 0)
	callCounts := scopeCallCounts{}

	for _, locale := range locales {
		localValues, localExists := local[locale]
		remoteValues, remoteExists := remote[locale]
		localeChanged := false
		localeDeletes := 0

		for _, field := range fields {
			localValue, localHasField := localValues[field]
			remoteValue, remoteHasField := remoteValues[field]

			itemKey := buildPlanKey(scope, version, locale, field)
			switch {
			case !remoteHasField && localHasField:
				adds = append(adds, PlanItem{
					Key:     itemKey,
					Scope:   scope,
					Locale:  locale,
					Version: version,
					Field:   field,
					Reason:  "field exists locally but not remotely",
					To:      localValue,
				})
				localeChanged = true
			case remoteHasField && !localHasField:
				deletes = append(deletes, PlanItem{
					Key:     itemKey,
					Scope:   scope,
					Locale:  locale,
					Version: version,
					Field:   field,
					Reason:  "field exists remotely but not locally",
					From:    remoteValue,
				})
				localeDeletes++
				localeChanged = true
			case remoteHasField && localHasField && remoteValue != localValue:
				updates = append(updates, PlanItem{
					Key:     itemKey,
					Scope:   scope,
					Locale:  locale,
					Version: version,
					Field:   field,
					Reason:  "field value differs",
					From:    remoteValue,
					To:      localValue,
				})
				localeChanged = true
			}
		}

		if !localeChanged {
			continue
		}
		switch {
		case localExists && !remoteExists:
			callCounts.create++
		case !localExists && remoteExists:
			callCounts.delete++
		case localExists && remoteExists && localeDeletes > 0:
			callCounts.delete++
			callCounts.create++
		default:
			callCounts.update++
		}
	}

	return adds, updates, deletes, callCounts
}

func buildPlanKey(scope, version, locale, field string) string {
	if scope == appInfoDirName {
		return fmt.Sprintf("%s:%s:%s", scope, locale, field)
	}
	return fmt.Sprintf("%s:%s:%s:%s", scope, version, locale, field)
}

func buildAPICallSummary(appInfoCounts, versionCounts scopeCallCounts) []PlanAPICall {
	summary := make([]PlanAPICall, 0, 6)
	appendCalls := func(scope string, counts scopeCallCounts) {
		if counts.create > 0 {
			summary = append(summary, PlanAPICall{
				Operation: "create_localization",
				Scope:     scope,
				Count:     counts.create,
			})
		}
		if counts.update > 0 {
			summary = append(summary, PlanAPICall{
				Operation: "update_localization",
				Scope:     scope,
				Count:     counts.update,
			})
		}
		if counts.delete > 0 {
			summary = append(summary, PlanAPICall{
				Operation: "delete_localization",
				Scope:     scope,
				Count:     counts.delete,
			})
		}
	}
	appendCalls(appInfoDirName, appInfoCounts)
	appendCalls(versionDirName, versionCounts)

	sort.Slice(summary, func(i, j int) bool {
		if summary[i].Scope == summary[j].Scope {
			return summary[i].Operation < summary[j].Operation
		}
		return summary[i].Scope < summary[j].Scope
	})
	return summary
}

func sortPlanItems(items []PlanItem) {
	sort.Slice(items, func(i, j int) bool {
		return items[i].Key < items[j].Key
	})
}

func printPushPlanTable(result PushPlanResult) error {
	fmt.Printf("App ID: %s\n", result.AppID)
	fmt.Printf("Version: %s\n", result.Version)
	fmt.Printf("Dir: %s\n", result.Dir)
	fmt.Printf("Dry Run: %t\n\n", result.DryRun)
	if result.Applied {
		fmt.Printf("Applied: %t\n\n", result.Applied)
	}

	headers := []string{"change", "key", "scope", "locale", "version", "field", "reason", "from", "to"}
	rows := buildPlanRows(result)
	asc.RenderTable(headers, rows)

	if len(result.APICalls) > 0 {
		fmt.Println()
		asc.RenderTable([]string{"operation", "scope", "count"}, buildAPICallRows(result.APICalls))
	}
	if len(result.Actions) > 0 {
		fmt.Println()
		asc.RenderTable([]string{"scope", "locale", "version", "action", "localizationId"}, buildApplyActionRows(result.Actions))
	}
	return nil
}

func printPushPlanMarkdown(result PushPlanResult) error {
	fmt.Printf("**App ID:** %s\n\n", result.AppID)
	fmt.Printf("**Version:** %s\n\n", result.Version)
	fmt.Printf("**Dir:** %s\n\n", result.Dir)
	fmt.Printf("**Dry Run:** %t\n\n", result.DryRun)
	if result.Applied {
		fmt.Printf("**Applied:** %t\n\n", result.Applied)
	}

	headers := []string{"change", "key", "scope", "locale", "version", "field", "reason", "from", "to"}
	rows := buildPlanRows(result)
	asc.RenderMarkdown(headers, rows)

	if len(result.APICalls) > 0 {
		fmt.Println()
		asc.RenderMarkdown([]string{"operation", "scope", "count"}, buildAPICallRows(result.APICalls))
	}
	if len(result.Actions) > 0 {
		fmt.Println()
		asc.RenderMarkdown([]string{"scope", "locale", "version", "action", "localizationId"}, buildApplyActionRows(result.Actions))
	}
	return nil
}

func buildPlanRows(result PushPlanResult) [][]string {
	rows := make([][]string, 0, len(result.Adds)+len(result.Updates)+len(result.Deletes))
	for _, item := range result.Adds {
		rows = append(rows, []string{
			"add",
			item.Key,
			item.Scope,
			item.Locale,
			item.Version,
			item.Field,
			item.Reason,
			"",
			sanitizePlanCell(item.To),
		})
	}
	for _, item := range result.Updates {
		rows = append(rows, []string{
			"update",
			item.Key,
			item.Scope,
			item.Locale,
			item.Version,
			item.Field,
			item.Reason,
			sanitizePlanCell(item.From),
			sanitizePlanCell(item.To),
		})
	}
	for _, item := range result.Deletes {
		rows = append(rows, []string{
			"delete",
			item.Key,
			item.Scope,
			item.Locale,
			item.Version,
			item.Field,
			item.Reason,
			sanitizePlanCell(item.From),
			"",
		})
	}
	if len(rows) == 0 {
		rows = append(rows, []string{"none", "", "", "", "", "", "no changes", "", ""})
	}
	return rows
}

func buildAPICallRows(calls []PlanAPICall) [][]string {
	rows := make([][]string, 0, len(calls))
	for _, call := range calls {
		rows = append(rows, []string{call.Operation, call.Scope, fmt.Sprintf("%d", call.Count)})
	}
	return rows
}

func buildApplyActionRows(actions []ApplyAction) [][]string {
	rows := make([][]string, 0, len(actions))
	for _, action := range actions {
		rows = append(rows, []string{
			action.Scope,
			action.Locale,
			action.Version,
			action.Action,
			action.LocalizationID,
		})
	}
	return rows
}

func sanitizePlanCell(value string) string {
	normalized := strings.ReplaceAll(value, "\n", "\\n")
	const maxLen = 80
	if len([]rune(normalized)) <= maxLen {
		return normalized
	}
	runes := []rune(normalized)
	return string(runes[:77]) + "..."
}
