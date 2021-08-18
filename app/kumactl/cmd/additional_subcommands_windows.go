package cmd

import (
	"github.com/spf13/cobra"

	kumactl_cmd "github.com/kumahq/kuma/app/kumactl/pkg/cmd"
)

var additionalSubcommands []func(*kumactl_cmd.RootContext) *cobra.Command
