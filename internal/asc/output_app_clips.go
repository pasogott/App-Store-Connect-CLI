package asc

import (
	"fmt"
	"os"
	"strings"
	"text/tabwriter"
)

func printAppClipsTable(resp *AppClipsResponse) error {
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "ID\tBundle ID")
	for _, item := range resp.Data {
		fmt.Fprintf(w, "%s\t%s\n", item.ID, item.Attributes.BundleID)
	}
	return w.Flush()
}

func printAppClipsMarkdown(resp *AppClipsResponse) error {
	fmt.Fprintln(os.Stdout, "| ID | Bundle ID |")
	fmt.Fprintln(os.Stdout, "| --- | --- |")
	for _, item := range resp.Data {
		fmt.Fprintf(os.Stdout, "| %s | %s |\n",
			escapeMarkdown(item.ID),
			escapeMarkdown(item.Attributes.BundleID),
		)
	}
	return nil
}

func printAppClipDefaultExperiencesTable(resp *AppClipDefaultExperiencesResponse) error {
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "ID\tAction")
	for _, item := range resp.Data {
		fmt.Fprintf(w, "%s\t%s\n", item.ID, item.Attributes.Action)
	}
	return w.Flush()
}

func printAppClipDefaultExperiencesMarkdown(resp *AppClipDefaultExperiencesResponse) error {
	fmt.Fprintln(os.Stdout, "| ID | Action |")
	fmt.Fprintln(os.Stdout, "| --- | --- |")
	for _, item := range resp.Data {
		fmt.Fprintf(os.Stdout, "| %s | %s |\n",
			escapeMarkdown(item.ID),
			escapeMarkdown(string(item.Attributes.Action)),
		)
	}
	return nil
}

func printAppClipDefaultExperienceLocalizationsTable(resp *AppClipDefaultExperienceLocalizationsResponse) error {
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "ID\tLocale\tSubtitle")
	for _, item := range resp.Data {
		fmt.Fprintf(w, "%s\t%s\t%s\n",
			item.ID,
			item.Attributes.Locale,
			compactWhitespace(item.Attributes.Subtitle),
		)
	}
	return w.Flush()
}

func printAppClipDefaultExperienceLocalizationsMarkdown(resp *AppClipDefaultExperienceLocalizationsResponse) error {
	fmt.Fprintln(os.Stdout, "| ID | Locale | Subtitle |")
	fmt.Fprintln(os.Stdout, "| --- | --- | --- |")
	for _, item := range resp.Data {
		fmt.Fprintf(os.Stdout, "| %s | %s | %s |\n",
			escapeMarkdown(item.ID),
			escapeMarkdown(item.Attributes.Locale),
			escapeMarkdown(item.Attributes.Subtitle),
		)
	}
	return nil
}

func printAppClipAdvancedExperiencesTable(resp *AppClipAdvancedExperiencesResponse) error {
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "ID\tAction\tStatus\tBusiness Category\tDefault Language\tPowered By\tLink")
	for _, item := range resp.Data {
		fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\t%t\t%s\n",
			item.ID,
			item.Attributes.Action,
			item.Attributes.Status,
			item.Attributes.BusinessCategory,
			item.Attributes.DefaultLanguage,
			item.Attributes.IsPoweredBy,
			item.Attributes.Link,
		)
	}
	return w.Flush()
}

func printAppClipAdvancedExperiencesMarkdown(resp *AppClipAdvancedExperiencesResponse) error {
	fmt.Fprintln(os.Stdout, "| ID | Action | Status | Business Category | Default Language | Powered By | Link |")
	fmt.Fprintln(os.Stdout, "| --- | --- | --- | --- | --- | --- | --- |")
	for _, item := range resp.Data {
		fmt.Fprintf(os.Stdout, "| %s | %s | %s | %s | %s | %t | %s |\n",
			escapeMarkdown(item.ID),
			escapeMarkdown(string(item.Attributes.Action)),
			escapeMarkdown(item.Attributes.Status),
			escapeMarkdown(string(item.Attributes.BusinessCategory)),
			escapeMarkdown(string(item.Attributes.DefaultLanguage)),
			item.Attributes.IsPoweredBy,
			escapeMarkdown(item.Attributes.Link),
		)
	}
	return nil
}

func printBetaAppClipInvocationLocalizationsTable(resp *BetaAppClipInvocationLocalizationsResponse) error {
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "ID\tLocale\tTitle")
	for _, item := range resp.Data {
		fmt.Fprintf(w, "%s\t%s\t%s\n",
			item.ID,
			item.Attributes.Locale,
			compactWhitespace(item.Attributes.Title),
		)
	}
	return w.Flush()
}

func printBetaAppClipInvocationLocalizationsMarkdown(resp *BetaAppClipInvocationLocalizationsResponse) error {
	fmt.Fprintln(os.Stdout, "| ID | Locale | Title |")
	fmt.Fprintln(os.Stdout, "| --- | --- | --- |")
	for _, item := range resp.Data {
		fmt.Fprintf(os.Stdout, "| %s | %s | %s |\n",
			escapeMarkdown(item.ID),
			escapeMarkdown(item.Attributes.Locale),
			escapeMarkdown(item.Attributes.Title),
		)
	}
	return nil
}

func printAppClipAdvancedExperienceImageUploadResultTable(result *AppClipAdvancedExperienceImageUploadResult) error {
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "ID\tExperience ID\tFile Name\tFile Size\tState\tUploaded")
	fmt.Fprintf(w, "%s\t%s\t%s\t%d\t%s\t%t\n",
		result.ID,
		result.ExperienceID,
		result.FileName,
		result.FileSize,
		result.AssetDeliveryState,
		result.Uploaded,
	)
	return w.Flush()
}

func printAppClipAdvancedExperienceImageUploadResultMarkdown(result *AppClipAdvancedExperienceImageUploadResult) error {
	fmt.Fprintln(os.Stdout, "| ID | Experience ID | File Name | File Size | State | Uploaded |")
	fmt.Fprintln(os.Stdout, "| --- | --- | --- | --- | --- | --- |")
	fmt.Fprintf(os.Stdout, "| %s | %s | %s | %d | %s | %t |\n",
		escapeMarkdown(result.ID),
		escapeMarkdown(result.ExperienceID),
		escapeMarkdown(result.FileName),
		result.FileSize,
		escapeMarkdown(result.AssetDeliveryState),
		result.Uploaded,
	)
	return nil
}

func printAppClipHeaderImageUploadResultTable(result *AppClipHeaderImageUploadResult) error {
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "ID\tLocalization ID\tFile Name\tFile Size\tState\tUploaded")
	fmt.Fprintf(w, "%s\t%s\t%s\t%d\t%s\t%t\n",
		result.ID,
		result.LocalizationID,
		result.FileName,
		result.FileSize,
		result.AssetDeliveryState,
		result.Uploaded,
	)
	return w.Flush()
}

func printAppClipHeaderImageUploadResultMarkdown(result *AppClipHeaderImageUploadResult) error {
	fmt.Fprintln(os.Stdout, "| ID | Localization ID | File Name | File Size | State | Uploaded |")
	fmt.Fprintln(os.Stdout, "| --- | --- | --- | --- | --- | --- |")
	fmt.Fprintf(os.Stdout, "| %s | %s | %s | %d | %s | %t |\n",
		escapeMarkdown(result.ID),
		escapeMarkdown(result.LocalizationID),
		escapeMarkdown(result.FileName),
		result.FileSize,
		escapeMarkdown(result.AssetDeliveryState),
		result.Uploaded,
	)
	return nil
}

func printAppClipDefaultExperienceDeleteResultTable(result *AppClipDefaultExperienceDeleteResult) error {
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "ID\tDeleted")
	fmt.Fprintf(w, "%s\t%t\n", result.ID, result.Deleted)
	return w.Flush()
}

func printAppClipDefaultExperienceDeleteResultMarkdown(result *AppClipDefaultExperienceDeleteResult) error {
	fmt.Fprintln(os.Stdout, "| ID | Deleted |")
	fmt.Fprintln(os.Stdout, "| --- | --- |")
	fmt.Fprintf(os.Stdout, "| %s | %t |\n",
		escapeMarkdown(result.ID),
		result.Deleted,
	)
	return nil
}

func printAppClipDefaultExperienceLocalizationDeleteResultTable(result *AppClipDefaultExperienceLocalizationDeleteResult) error {
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "ID\tDeleted")
	fmt.Fprintf(w, "%s\t%t\n", result.ID, result.Deleted)
	return w.Flush()
}

func printAppClipDefaultExperienceLocalizationDeleteResultMarkdown(result *AppClipDefaultExperienceLocalizationDeleteResult) error {
	fmt.Fprintln(os.Stdout, "| ID | Deleted |")
	fmt.Fprintln(os.Stdout, "| --- | --- |")
	fmt.Fprintf(os.Stdout, "| %s | %t |\n",
		escapeMarkdown(result.ID),
		result.Deleted,
	)
	return nil
}

func printAppClipAdvancedExperienceDeleteResultTable(result *AppClipAdvancedExperienceDeleteResult) error {
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "ID\tDeleted")
	fmt.Fprintf(w, "%s\t%t\n", result.ID, result.Deleted)
	return w.Flush()
}

func printAppClipAdvancedExperienceDeleteResultMarkdown(result *AppClipAdvancedExperienceDeleteResult) error {
	fmt.Fprintln(os.Stdout, "| ID | Deleted |")
	fmt.Fprintln(os.Stdout, "| --- | --- |")
	fmt.Fprintf(os.Stdout, "| %s | %t |\n",
		escapeMarkdown(result.ID),
		result.Deleted,
	)
	return nil
}

func printAppClipAdvancedExperienceImageDeleteResultTable(result *AppClipAdvancedExperienceImageDeleteResult) error {
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "ID\tDeleted")
	fmt.Fprintf(w, "%s\t%t\n", result.ID, result.Deleted)
	return w.Flush()
}

func printAppClipAdvancedExperienceImageDeleteResultMarkdown(result *AppClipAdvancedExperienceImageDeleteResult) error {
	fmt.Fprintln(os.Stdout, "| ID | Deleted |")
	fmt.Fprintln(os.Stdout, "| --- | --- |")
	fmt.Fprintf(os.Stdout, "| %s | %t |\n",
		escapeMarkdown(result.ID),
		result.Deleted,
	)
	return nil
}

func printAppClipHeaderImageDeleteResultTable(result *AppClipHeaderImageDeleteResult) error {
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "ID\tDeleted")
	fmt.Fprintf(w, "%s\t%t\n", result.ID, result.Deleted)
	return w.Flush()
}

func printAppClipHeaderImageDeleteResultMarkdown(result *AppClipHeaderImageDeleteResult) error {
	fmt.Fprintln(os.Stdout, "| ID | Deleted |")
	fmt.Fprintln(os.Stdout, "| --- | --- |")
	fmt.Fprintf(os.Stdout, "| %s | %t |\n",
		escapeMarkdown(result.ID),
		result.Deleted,
	)
	return nil
}

func printBetaAppClipInvocationDeleteResultTable(result *BetaAppClipInvocationDeleteResult) error {
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "ID\tDeleted")
	fmt.Fprintf(w, "%s\t%t\n", result.ID, result.Deleted)
	return w.Flush()
}

func printBetaAppClipInvocationDeleteResultMarkdown(result *BetaAppClipInvocationDeleteResult) error {
	fmt.Fprintln(os.Stdout, "| ID | Deleted |")
	fmt.Fprintln(os.Stdout, "| --- | --- |")
	fmt.Fprintf(os.Stdout, "| %s | %t |\n",
		escapeMarkdown(result.ID),
		result.Deleted,
	)
	return nil
}

func printBetaAppClipInvocationLocalizationDeleteResultTable(result *BetaAppClipInvocationLocalizationDeleteResult) error {
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "ID\tDeleted")
	fmt.Fprintf(w, "%s\t%t\n", result.ID, result.Deleted)
	return w.Flush()
}

func printBetaAppClipInvocationLocalizationDeleteResultMarkdown(result *BetaAppClipInvocationLocalizationDeleteResult) error {
	fmt.Fprintln(os.Stdout, "| ID | Deleted |")
	fmt.Fprintln(os.Stdout, "| --- | --- |")
	fmt.Fprintf(os.Stdout, "| %s | %t |\n",
		escapeMarkdown(result.ID),
		result.Deleted,
	)
	return nil
}

func printAppClipAppStoreReviewDetailTable(resp *AppClipAppStoreReviewDetailResponse) error {
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "ID\tInvocation URLs")
	urls := strings.Join(resp.Data.Attributes.InvocationURLs, ", ")
	fmt.Fprintf(w, "%s\t%s\n", resp.Data.ID, compactWhitespace(urls))
	return w.Flush()
}

func printAppClipAppStoreReviewDetailMarkdown(resp *AppClipAppStoreReviewDetailResponse) error {
	fmt.Fprintln(os.Stdout, "| ID | Invocation URLs |")
	fmt.Fprintln(os.Stdout, "| --- | --- |")
	urls := strings.Join(resp.Data.Attributes.InvocationURLs, ", ")
	fmt.Fprintf(os.Stdout, "| %s | %s |\n",
		escapeMarkdown(resp.Data.ID),
		escapeMarkdown(urls),
	)
	return nil
}
