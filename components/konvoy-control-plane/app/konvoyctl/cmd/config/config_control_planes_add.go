package config

import (
	"github.com/pkg/errors"
	"github.com/spf13/cobra"

	"github.com/Kong/konvoy/components/konvoy-control-plane/app/konvoyctl/pkg/cmd"
	config_proto "github.com/Kong/konvoy/components/konvoy-control-plane/pkg/config/app/konvoyctl/v1alpha1"
)

type configControlPlanesAddContext struct {
	*cmd.RootContext

	args struct {
		name         string
		apiServerURL string
	}
}

func newConfigControlPlanesAddCmd(pctx *cmd.RootContext) *cobra.Command {
	ctx := &configControlPlanesAddContext{RootContext: pctx}
	command := &cobra.Command{
		Use:   "add",
		Short: "Add a Control Plane",
		Long:  `Add a Control Plane.`,
		RunE: func(cmd *cobra.Command, _ []string) error {
			cp := &config_proto.ControlPlane{
				Name: ctx.args.name,
				Coordinates: &config_proto.ControlPlaneCoordinates{
					ApiServer: &config_proto.ControlPlaneCoordinates_ApiServer{
						Url: ctx.args.apiServerURL,
					},
				},
			}

			return ctx.AddControlPlane(cp)
		},
	}

	command.Flags().StringVar(&ctx.args.name, "name", "", "reference name for the Control Plane (required)")
	command.MarkFlagRequired("name")
	command.Flags().StringVar(&ctx.args.apiServerURL, "api-server-url", "", "URL of the Control Plane API Server (required)")
	command.MarkFlagRequired("api-server-url")
	return command
}

func (c *configControlPlanesAddContext) AddControlPlane(cp *config_proto.ControlPlane) error {
	cfg := c.Config()
	if err := cp.Validate(); err != nil {
		return errors.Wrapf(err, "Control Plane configuration is not valid")
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
	cfg.CurrentContext = cp.Name
	return c.SaveConfig()
}
