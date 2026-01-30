package appclips

import "github.com/peterbourgon/ff/v3/ffcli"

// Command returns the app-clips command group.
func Command() *ffcli.Command {
	return AppClipsCommand()
}
