package install

import (
	"github.com/pkg/errors"
	"github.com/spf13/cobra"

	install_context "github.com/kumahq/kuma/app/kumactl/cmd/install/context"
	kumactl_data "github.com/kumahq/kuma/app/kumactl/data"
	"github.com/kumahq/kuma/app/kumactl/pkg/install/data"
	"github.com/kumahq/kuma/app/kumactl/pkg/install/k8s"
)

type gatewayTemplateArgs struct {
	Namespace string
}

func newInstallGatewayCmd(ctx *install_context.InstallGatewayContext) *cobra.Command {
	args := ctx.Args
	cmd := &cobra.Command{
		Use:   "gateway",
		Short: "Install ingress gateway on Kubernetes",
		Long:  "Install ingress gateway on Kubernetes in a 'kuma-gateway' namespace.",
		RunE: func(cmd *cobra.Command, _ []string) error {
			if err := validateGWArgs(args); err != nil {
				return err
			}

			templateArgs := gatewayTemplateArgs{
				Namespace: args.Namespace,
			}

			templateFiles, err := data.ReadFiles(kumactl_data.InstallGatewayFS())
			if err != nil {
				return errors.Wrap(err, "Failed to read template files")
			}

			renderedFiles, err := renderFiles(templateFiles, templateArgs, simpleTemplateRenderer)

			if err != nil {
				return errors.Wrap(err, "Failed to render template files")
			}

			sortedResources, err := k8s.SortResourcesByKind(renderedFiles)
			if err != nil {
				return errors.Wrap(err, "Failed to sort resources by kind")
			}

			singleFile := data.JoinYAML(sortedResources)

			if _, err := cmd.OutOrStdout().Write(singleFile.Data); err != nil {
				return errors.Wrap(err, "Failed to output rendered resources")
			}
			return nil
		},
	}
	cmd.Flags().StringVar(&args.Type, "type", args.Type, "type of gateway to install. Available types: 'kong'")
	cmd.MarkFlagRequired("type")
	cmd.Flags().StringVar(&args.Namespace, "namespace", args.Namespace, "namespace to install gateway to")
	return cmd
}

func validateGWArgs(args install_context.InstallGatewayArgs) error {
	if args.Type != "kong" {
		return errors.New("Only gateway type 'kong' currently supported")
	}
	return nil
}
