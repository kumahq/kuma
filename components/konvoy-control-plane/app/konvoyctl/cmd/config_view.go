package cmd

import (
	util_proto "github.com/Kong/konvoy/components/konvoy-control-plane/pkg/util/proto"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
)

func newConfigViewCmd(pctx *rootContext) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "view",
		Short: "Show konvoyctl config",
		Long:  `Show konvoyctl config.`,
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
	return cmd
}
