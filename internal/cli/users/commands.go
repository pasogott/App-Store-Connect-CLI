package users

import "github.com/peterbourgon/ff/v3/ffcli"

// Command returns the users command group.
func Command() *ffcli.Command {
	return UsersCommand()
}
