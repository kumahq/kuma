package config

import (
	"github.com/spf13/cobra"

	kumactl_cmd "github.com/kumahq/kuma/app/kumactl/pkg/cmd"
	"github.com/kumahq/kuma/app/kumactl/pkg/config"
	util_proto "github.com/kumahq/kuma/pkg/util/proto"
)

func newConfigDefaultCmd(pctx *kumactl_cmd.RootContext) *cobra.Command {
	return &cobra.Command{
		Use:   "default",
		Short: "print the default configuration file",
		Long:  "Print the default configuration file.",
		RunE: func(cmd *cobra.Command, _ []string) error {
			cfg := config.DefaultConfiguration()

			contents, err := util_proto.ToYAML(&cfg)
			if err != nil {
				return err
			}

			cmd.Println(string(contents))
			return nil
		},
	}
}
