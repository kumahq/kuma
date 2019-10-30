package config

import (
	"github.com/Kong/kuma/app/kumactl/pkg/config"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"

	kumactl_cmd "github.com/Kong/kuma/app/kumactl/pkg/cmd"
	config_proto "github.com/Kong/kuma/pkg/config/app/kumactl/v1alpha1"
)

func newConfigControlPlanesAddCmd(pctx *kumactl_cmd.RootContext) *cobra.Command {
	args := struct {
		name         string
		apiServerURL string
	}{}
	cmd := &cobra.Command{
		Use:   "add",
		Short: "Add a Control Plane",
		Long:  `Add a Control Plane.`,
		RunE: func(cmd *cobra.Command, _ []string) error {
			cp := &config_proto.ControlPlane{
				Name: args.name,
				Coordinates: &config_proto.ControlPlaneCoordinates{
					ApiServer: &config_proto.ControlPlaneCoordinates_ApiServer{
						Url: args.apiServerURL,
					},
				},
			}

			cfg := pctx.Config()
			if err := cp.Validate(); err != nil {
				return errors.Wrapf(err, "Control Plane configuration is not valid")
			}
			if err := config.ValidateCpCoordinates(cp); err != nil {
				return err
			}
			if !cfg.AddControlPlane(cp) {
				return errors.Errorf("Control Plane with name %q already exists", cp.Name)
			}
			ctx := &config_proto.Context{
				Name:         cp.Name,
				ControlPlane: cp.Name,
			}
			if err := ctx.Validate(); err != nil {
				return errors.Wrapf(err, "Context configuration is not valid")
			}
			if !cfg.AddContext(ctx) {
				return errors.Errorf("Context with name %q already exists", ctx.Name)
			}
			cfg.CurrentContext = ctx.Name
			if err := pctx.SaveConfig(); err != nil {
				return err
			}
			cmd.Printf("added Control Plane %q\n", ctx.Name)
			cmd.Printf("switched active Control Plane to %q\n", ctx.Name)
			return nil
		},
	}
	// flags
	cmd.Flags().StringVar(&args.name, "name", "", "reference name for the Control Plane (required)")
	_ = cmd.MarkFlagRequired("name")
	cmd.Flags().StringVar(&args.apiServerURL, "address", "", "URL of the Control Plane API Server (required)")
	_ = cmd.MarkFlagRequired("address")
	return cmd
}
