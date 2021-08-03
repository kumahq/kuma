package cmd

import "github.com/spf13/cobra"

type Plugin interface {
	// CustomizeContext is a hook that allows a plugin to customize
	// the kumactl root context before it is used to build the root
	// command.
	CustomizeContext(*RootContext)

	// CustomizeCommand is a hook that allows a plugin to customize
	// the kumactl root command after it has been created.
	CustomizeCommand(*cobra.Command)
}

var Plugins []Plugin
