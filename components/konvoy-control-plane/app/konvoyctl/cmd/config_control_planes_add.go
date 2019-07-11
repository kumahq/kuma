package cmd

import (
	config_proto "github.com/Kong/konvoy/components/konvoy-control-plane/pkg/config/app/konvoyctl/v1alpha1"
	"github.com/spf13/cobra"
)

type configControlPlanesAddContext struct {
	pctx *rootContext
	args struct {
		name string
	}
}

func newConfigControlPlanesAddCmd(pctx *rootContext) *cobra.Command {
	ctx := &configControlPlanesAddContext{pctx: pctx}
	cmd := &cobra.Command{
		Use:   "add",
		Short: "Add a Control Plane",
		Long:  `Add a Control Plane.`,
	}
	// flags
	cmd.PersistentFlags().StringVar(&ctx.args.name, "name", "", "reference name for a Control Plane")
	// sub-commands
	cmd.AddCommand(newConfigControlPlanesAddKubernetesCmd(ctx))
	return cmd
}

func (c *configControlPlanesAddContext) AddControlPlane(cp *config_proto.ControlPlane) error {
	cfg := c.pctx.Config()
	cfg.AddControlPlane(cp)
	cfg.AddContext(&config_proto.Context{
		Name:         cp.Name,
		ControlPlane: cp.Name,
	})
	cfg.CurrentContext = cp.Name
	return c.pctx.SaveConfig()
}
