package config

import (
	"fmt"

	"github.com/spf13/cobra"

	kumactl_cmd "github.com/kumahq/kuma/app/kumactl/pkg/cmd"
	util_proto "github.com/kumahq/kuma/pkg/util/proto"
)

func newConfigViewCmd(pctx *kumactl_cmd.RootContext) *cobra.Command {
	return &cobra.Command{
		Use:   "view",
		Short: "Show kumactl config",
		Long:  `Show kumactl config.`,
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			cfg := pctx.Config()
			contents, err := util_proto.ToYAML(cfg)
			if err != nil {
				return fmt.Errorf("Cannot format configuration: %#v: %w", cfg, err)
			}
			cmd.Println(string(contents))
			return nil
		},
	}
}
