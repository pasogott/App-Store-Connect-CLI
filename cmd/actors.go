package cmd

import (
	"github.com/peterbourgon/ff/v3/ffcli"

	actorscli "github.com/rudrankriyam/App-Store-Connect-CLI/internal/cli/actors"
)

// ActorsCommand returns the actors command group.
func ActorsCommand() *ffcli.Command {
	return actorscli.ActorsCommand()
}
