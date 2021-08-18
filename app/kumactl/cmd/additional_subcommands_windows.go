package cmd

import (
	"github.com/kumahq/kuma/app/kumactl/cmd/apply"
	"github.com/kumahq/kuma/app/kumactl/cmd/completion"
	"github.com/kumahq/kuma/app/kumactl/cmd/config"
	"github.com/kumahq/kuma/app/kumactl/cmd/delete"
	"github.com/kumahq/kuma/app/kumactl/cmd/generate"
	"github.com/kumahq/kuma/app/kumactl/cmd/get"
	"github.com/kumahq/kuma/app/kumactl/cmd/inspect"
	"github.com/kumahq/kuma/app/kumactl/cmd/install"
	"github.com/kumahq/kuma/pkg/cmd/version"
	"github.com/spf13/cobra"

	kumactl_cmd "github.com/kumahq/kuma/app/kumactl/pkg/cmd"
)

var additionalSubcommands []func(*kumactl_cmd.RootContext) *cobra.Command
