package cmd

import "github.com/spf13/cobra"

type RunnableWrapper = func(func(*cobra.Command, []string) error) func(*cobra.Command, []string) error

func WrapRunnables(cmd *cobra.Command, wrapper RunnableWrapper) {
	if cmd.PersistentPreRunE != nil {
		cmd.PersistentPreRunE = wrapper(cmd.PersistentPreRunE)
	}
	if cmd.PreRunE != nil {
		cmd.PreRunE = wrapper(cmd.PreRunE)
	}
	if cmd.RunE != nil {
		cmd.RunE = wrapper(cmd.RunE)
	}
	if cmd.PostRunE != nil {
		cmd.PostRunE = wrapper(cmd.PostRunE)
	}
	if cmd.PersistentPostRunE != nil {
		cmd.PersistentPostRunE = wrapper(cmd.PersistentPostRunE)
	}
	for _, command := range cmd.Commands() {
		WrapRunnables(command, wrapper)
	}
}
