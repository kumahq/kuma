package config

import (
	net_url "net/url"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"

	kumactl_cmd "github.com/kumahq/kuma/app/kumactl/pkg/cmd"
	"github.com/kumahq/kuma/app/kumactl/pkg/config"
	config_proto "github.com/kumahq/kuma/pkg/config/app/kumactl/v1alpha1"
)

type controlPlaneAddArgs struct {
	name           string
	apiServerURL   string
	overwrite      bool
	clientCertFile string
	clientKeyFile  string
	caCertFile     string
	skipVerify     bool
}

func newConfigControlPlanesAddCmd(pctx *kumactl_cmd.RootContext) *cobra.Command {
	var args controlPlaneAddArgs
	cmd := &cobra.Command{
		Use:   "add",
		Short: "Add a Control Plane",
		Long:  `Add a Control Plane.`,
		RunE: func(cmd *cobra.Command, _ []string) error {
			if err := validateArgs(args); err != nil {
				return err
			}

			cp := &config_proto.ControlPlane{
				Name: args.name,
				Coordinates: &config_proto.ControlPlaneCoordinates{
					ApiServer: &config_proto.ControlPlaneCoordinates_ApiServer{
						Url:            args.apiServerURL,
						CaCertFile:     args.caCertFile,
						ClientCertFile: args.clientCertFile,
						ClientKeyFile:  args.clientKeyFile,
					},
				},
			}
			cfg := pctx.Config()
			if err := config.ValidateCpCoordinates(cp); err != nil {
				return err
			}
			if !cfg.AddControlPlane(cp, args.overwrite) {
				return errors.Errorf("Control Plane with name %q already exists. Use --overwrite to replace an existing one.", cp.Name)
			}
			ctx := &config_proto.Context{
				Name:         cp.Name,
				ControlPlane: cp.Name,
			}
			if !cfg.AddContext(ctx, args.overwrite) {
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
	cmd.Flags().StringVar(&args.apiServerURL, "address", "", "URL of the Control Plane API Server (required). Example: http://localhost:5681 or https://localhost:5682)")
	_ = cmd.MarkFlagRequired("address")
	cmd.Flags().BoolVar(&args.overwrite, "overwrite", false, "overwrite existing Control Plane with the same reference name")
	cmd.Flags().StringVar(&args.clientCertFile, "client-cert-file", "", "path to the certificate of a client that is authorized to use the Admin operations of the Control Plane (kumactl stores only a reference to this file)")
	cmd.Flags().StringVar(&args.clientKeyFile, "client-key-file", "", "path to the certificate key of a client that is authorized to use the Admin operations of the Control Plane (kumactl stores only a reference to this file)")
	cmd.Flags().StringVar(&args.caCertFile, "ca-cert-file", "", "path to the certificate authority which will be used to verify the Control Plane certificate (kumactl stores only a reference to this file)")
	cmd.Flags().BoolVar(&args.skipVerify, "skip-verify", false, "skip CA verification")
	return cmd
}

func validateArgs(args controlPlaneAddArgs) error {
	url, err := net_url.ParseRequestURI(args.apiServerURL)
	if err != nil {
		return errors.Wrap(err, "API Server URL is invalid")
	}
	if url.Scheme == "https" {
		if args.caCertFile == "" && !args.skipVerify {
			return errors.New("HTTPS is used. You need to specify either --ca-cert-file so kumactl can verify authenticity of the Control Plane or --skip-verify to skip verification")
		}
	}
	if (args.clientKeyFile != "" && args.clientCertFile == "") || (args.clientKeyFile == "" && args.clientCertFile != "") {
		return errors.New("Both --client-cert-file and --client-key-file needs to be specified")
	}
	return nil
}
