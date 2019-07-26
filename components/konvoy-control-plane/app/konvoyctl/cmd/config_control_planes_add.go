package cmd

import (
	"github.com/pkg/errors"
	"github.com/spf13/cobra"

	config_proto "github.com/Kong/konvoy/components/konvoy-control-plane/pkg/config/app/konvoyctl/v1alpha1"
)

type configControlPlanesAddContext struct {
	*rootContext

	args struct {
		name string
	}
}

func newConfigControlPlanesAddCmd(pctx *rootContext) *cobra.Command {
	ctx := &configControlPlanesAddContext{rootContext: pctx}
	cmd := &cobra.Command{
		Use:   "add",
		Short: "Add a Control Plane",
		Long:  `Add a Control Plane.`,
	}
	// flags
	withCommonFlags := func(cmd *cobra.Command) *cobra.Command {
		cmd.Flags().StringVar(&ctx.args.name, "name", "", "reference name for the Control Plane (required)")
		cmd.MarkFlagRequired("name")
		return cmd
	}
	// sub-commands
	cmd.AddCommand(withCommonFlags(newConfigControlPlanesAddKubernetesCmd(ctx)))
	cmd.AddCommand(withCommonFlags(newConfigControlPlanesAddOtherCmd(ctx)))
	return cmd
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
