package actors

import "github.com/peterbourgon/ff/v3/ffcli"

// Command returns the actors command group.
func Command() *ffcli.Command {
	return ActorsCommand()
}
