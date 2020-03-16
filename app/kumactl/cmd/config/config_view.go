package config

import (
	"github.com/pkg/errors"
	"github.com/spf13/cobra"

	kumactl_cmd "github.com/Kong/kuma/app/kumactl/pkg/cmd"
	util_proto "github.com/Kong/kuma/pkg/util/proto"
)

func newConfigViewCmd(pctx *kumactl_cmd.RootContext) *cobra.Command {
	return &cobra.Command{
		Use:   "view",
		Short: "Show kumactl config",
		Long:  `Show kumactl config.`,
		RunE: func(cmd *cobra.Command, _ []string) error {
			cfg := pctx.Config()
			contents, err := util_proto.ToYAML(cfg)
			if err != nil {
				return errors.Wrapf(err, "Cannot format configuration: %#v", cfg)
			}
			cmd.Println(string(contents))
			return nil
		},
	}
}
