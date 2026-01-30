package appclips

import (
	"context"
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/peterbourgon/ff/v3/ffcli"

	"github.com/rudrankriyam/App-Store-Connect-CLI/internal/asc"
)

// AppClipAdvancedExperienceImagesCommand returns the images command group.
func AppClipAdvancedExperienceImagesCommand() *ffcli.Command {
	fs := flag.NewFlagSet("images", flag.ExitOnError)

	return &ffcli.Command{
		Name:       "images",
		ShortUsage: "asc app-clips advanced-experiences images <subcommand> [flags]",
		ShortHelp:  "Manage App Clip advanced experience images.",
		LongHelp: `Manage App Clip advanced experience images.

Examples:
  asc app-clips advanced-experiences images create --experience-id "EXP_ID" --file path/to/image.png
  asc app-clips advanced-experiences images delete --id "IMAGE_ID" --confirm`,
		FlagSet:   fs,
		UsageFunc: DefaultUsageFunc,
		Subcommands: []*ffcli.Command{
			AppClipAdvancedExperienceImagesCreateCommand(),
			AppClipAdvancedExperienceImagesDeleteCommand(),
		},
		Exec: func(ctx context.Context, args []string) error {
			return flag.ErrHelp
		},
	}
}

// AppClipAdvancedExperienceImagesCreateCommand uploads an image.
func AppClipAdvancedExperienceImagesCreateCommand() *ffcli.Command {
	fs := flag.NewFlagSet("create", flag.ExitOnError)

	experienceID := fs.String("experience-id", "", "Advanced experience ID")
	filePath := fs.String("file", "", "Path to image file (PNG)")
	output := fs.String("output", "json", "Output format: json (default), table, markdown")
	pretty := fs.Bool("pretty", false, "Pretty-print JSON output")

	return &ffcli.Command{
		Name:       "create",
		ShortUsage: "asc app-clips advanced-experiences images create --experience-id \"EXP_ID\" --file path/to/image.png",
		ShortHelp:  "Upload an image for an advanced experience.",
		LongHelp: `Upload an image for an advanced experience.

The upload process reserves an upload slot, uploads the image, commits the upload,
and associates the image with the experience.

Examples:
  asc app-clips advanced-experiences images create --experience-id "EXP_ID" --file path/to/image.png`,
		FlagSet:   fs,
		UsageFunc: DefaultUsageFunc,
		Exec: func(ctx context.Context, args []string) error {
			experienceValue := strings.TrimSpace(*experienceID)
			if experienceValue == "" {
				fmt.Fprintln(os.Stderr, "Error: --experience-id is required")
				return flag.ErrHelp
			}

			fileValue := strings.TrimSpace(*filePath)
			if fileValue == "" {
				fmt.Fprintln(os.Stderr, "Error: --file is required")
				return flag.ErrHelp
			}

			client, err := getASCClient()
			if err != nil {
				return fmt.Errorf("app-clips advanced-experiences images create: %w", err)
			}

			requestCtx, cancel := contextWithUploadTimeout(ctx)
			defer cancel()

			result, err := client.UploadAppClipAdvancedExperienceImage(requestCtx, fileValue)
			if err != nil {
				return fmt.Errorf("app-clips advanced-experiences images create: %w", err)
			}

			if _, err := client.UpdateAppClipAdvancedExperience(requestCtx, experienceValue, nil, "", result.ID, nil); err != nil {
				return fmt.Errorf("app-clips advanced-experiences images create: failed to attach image: %w", err)
			}

			result.ExperienceID = experienceValue
			return printOutput(result, *output, *pretty)
		},
	}
}

// AppClipAdvancedExperienceImagesDeleteCommand deletes an image.
func AppClipAdvancedExperienceImagesDeleteCommand() *ffcli.Command {
	fs := flag.NewFlagSet("delete", flag.ExitOnError)

	imageID := fs.String("id", "", "Image ID")
	confirm := fs.Bool("confirm", false, "Confirm deletion")
	output := fs.String("output", "json", "Output format: json (default), table, markdown")
	pretty := fs.Bool("pretty", false, "Pretty-print JSON output")

	return &ffcli.Command{
		Name:       "delete",
		ShortUsage: "asc app-clips advanced-experiences images delete --id \"IMAGE_ID\" --confirm",
		ShortHelp:  "Delete an advanced experience image.",
		LongHelp: `Delete an advanced experience image.

Examples:
  asc app-clips advanced-experiences images delete --id "IMAGE_ID" --confirm`,
		FlagSet:   fs,
		UsageFunc: DefaultUsageFunc,
		Exec: func(ctx context.Context, args []string) error {
			idValue := strings.TrimSpace(*imageID)
			if idValue == "" {
				fmt.Fprintln(os.Stderr, "Error: --id is required")
				return flag.ErrHelp
			}
			if !*confirm {
				fmt.Fprintln(os.Stderr, "Error: --confirm is required to delete")
				return flag.ErrHelp
			}

			client, err := getASCClient()
			if err != nil {
				return fmt.Errorf("app-clips advanced-experiences images delete: %w", err)
			}

			requestCtx, cancel := contextWithTimeout(ctx)
			defer cancel()

			if err := client.DeleteAppClipAdvancedExperienceImage(requestCtx, idValue); err != nil {
				return fmt.Errorf("app-clips advanced-experiences images delete: failed to delete: %w", err)
			}

			result := &asc.AppClipAdvancedExperienceImageDeleteResult{
				ID:      idValue,
				Deleted: true,
			}

			return printOutput(result, *output, *pretty)
		},
	}
}
