package config

import (
	"github.com/pkg/errors"
	"github.com/spf13/cobra"

	kumactl_cmd "github.com/Kong/kuma/app/kumactl/pkg/cmd"
)

func newConfigControlPlanesSwitchCmd(pctx *kumactl_cmd.RootContext) *cobra.Command {
	args := struct {
		name string
	}{}
	cmd := &cobra.Command{
		Use:   "switch",
		Short: "Switch active Control Plane",
		Long:  `Switch active Control Plane.`,
		Example: `If you have in your deployment configuration several contexts, for example:
contexts:
    - name: ctx1
        control_plane: cp1
        defaults:
           mesh: pilot
    - name: ctx2
        control_plane: cp2
        defaults:
        mesh: default

:$ kumactl config control-planes switch ctx2
switched active Control Plate to "ctx2"
`,
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
