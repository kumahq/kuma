package cmd

import (
	"github.com/spf13/cobra"

	config_proto "github.com/Kong/konvoy/components/konvoy-control-plane/pkg/config/app/konvoyctl/v1alpha1"
)

type configControlPlanesAddOtherContext struct {
	*configControlPlanesAddContext

	args struct {
		apiServerURL string
	}
}

func newConfigControlPlanesAddOtherCmd(pctx *configControlPlanesAddContext) *cobra.Command {
	ctx := &configControlPlanesAddOtherContext{configControlPlanesAddContext: pctx}
	cmd := &cobra.Command{
		Use:   "other",
		Short: "Add a Control Plane installed elsewhere",
		Long:  `Add a Control Plane installed elsewhere.`,
		RunE: func(cmd *cobra.Command, _ []string) error {
			name := pctx.args.name
			url := ctx.args.apiServerURL

			cp := &config_proto.ControlPlane{
				Name: name,
				Coordinates: &config_proto.ControlPlaneCoordinates{
					Type: &config_proto.ControlPlaneCoordinates_ApiServer_{
						ApiServer: &config_proto.ControlPlaneCoordinates_ApiServer{
							Url: url,
						},
					},
				},
			}

			return pctx.AddControlPlane(cp)
		},
	}
	// flags
	cmd.Flags().StringVar(&ctx.args.apiServerURL, "api-server-url", "", "URL of the Control Plane API Server (required)")
	cmd.MarkFlagRequired("api-server-url")
	return cmd
}
