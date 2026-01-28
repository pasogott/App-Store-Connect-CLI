package cmd

import (
	"github.com/peterbourgon/ff/v3/ffcli"

	userscli "github.com/rudrankriyam/App-Store-Connect-CLI/internal/cli/users"
)

// UsersCommand returns the users command group.
func UsersCommand() *ffcli.Command {
	return userscli.UsersCommand()
}
