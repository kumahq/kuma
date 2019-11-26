package config

import (
	kumactl_cmd "github.com/Kong/kuma/app/kumactl/pkg/cmd"
	util_proto "github.com/Kong/kuma/pkg/util/proto"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
)

func newConfigViewCmd(pctx *kumactl_cmd.RootContext) *cobra.Command {
	return &cobra.Command{
		Use:   "view",
		Short: "Show kumactl config",
		Long:  `Show kumactl config.`,
		Example: `:$ kumactl config view
control_planes:
    - name: my-first-cp
        coordinates:
            api_server:
            url: https://cp1.internal:5681
    - name: my-second-cp
        coordinates:
            api_server:
            url: https://cp1.internal:5681
		
contexts:
    - name: stage1
        control_plane: my-first-cp
        defaults:
            mesh: pilot
    - name: stage2
        control_plane: test2
        defaults:
            mesh: default

current_context: stage1
		`,
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
