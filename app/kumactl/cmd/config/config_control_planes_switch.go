package config

import (
	"github.com/pkg/errors"
	"github.com/spf13/cobra"

	kumactl_cmd "github.com/kumahq/kuma/app/kumactl/pkg/cmd"
)

func newConfigControlPlanesSwitchCmd(pctx *kumactl_cmd.RootContext) *cobra.Command {
	args := struct {
		name string
	}{}
	cmd := &cobra.Command{
		Use:   "switch",
		Short: "Switch active Control Plane",
		Long:  `Switch active Control Plane.`,
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			cfg := pctx.Config()
			if !cfg.SwitchContext(args.name) {
				return errors.Errorf("there is no Control Plane with name %q", args.name)
			}
			if err := pctx.SaveConfig(); err != nil {
				return err
			}
			cmd.Printf("switched active Control Plane to %q\n", args.name)
			return nil
		},
	}
	// flags
	cmd.Flags().StringVar(&args.name, "name", "", "reference name for the Control Plane (required)")
	_ = cmd.MarkFlagRequired("name")
	return cmd
}
