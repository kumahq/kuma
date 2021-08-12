package cmd

import "github.com/spf13/cobra"

func FindSubCommand(cmd *cobra.Command, name ...string) *cobra.Command {
	if len(name) == 0 {
		return cmd
	}

	for _, command := range cmd.Commands() {
		if command.Name() == name[0] {
			return FindSubCommand(command, name[1:]...)
		}
	}

	return nil
}
