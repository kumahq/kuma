package install

import (
	"strings"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"

	install_context "github.com/kumahq/kuma/app/kumactl/cmd/install/context"
	kumactl_data "github.com/kumahq/kuma/app/kumactl/data"
	"github.com/kumahq/kuma/app/kumactl/pkg/install/data"
	"github.com/kumahq/kuma/app/kumactl/pkg/install/k8s"
)

type demoTemplateArgs struct {
	Namespace string
	Zone      string
}

type GatewayFilter struct{}

func (GatewayFilter) Filter(name string) bool {
	return !strings.HasSuffix(name, "gateway.yaml")
}

func newInstallDemoCmd(ctx *install_context.InstallDemoContext) *cobra.Command {
	args := ctx.Args
	cmd := &cobra.Command{
		Use:   "demo",
		Short: "Install Kuma demo on Kubernetes",
		Long:  "Install Kuma demo on Kubernetes in its own namespace.",
		RunE: func(cmd *cobra.Command, _ []string) error {
			if err := validateDemoArgs(args); err != nil {
				return err
			}

			templateArgs := demoTemplateArgs{
				Namespace: args.Namespace,
				Zone:      args.Zone,
			}

			templateFiles, err := data.ReadFiles(kumactl_data.InstallDemoFS())
			if err != nil {
				return errors.Wrap(err, "Failed to read template files")
			}

			var filter templateFilter = NoneFilter{}
			if args.WithoutGateway {
				filter = GatewayFilter{}
			}

			renderedFiles, err := renderFilesWithFilter(templateFiles, templateArgs, simpleTemplateRenderer, filter)
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
	cmd.Flags().StringVar(&args.Zone, "zone", args.Zone, "Zone in which to install demo")
	cmd.Flags().StringVar(&args.Namespace, "namespace", args.Namespace, "Namespace to install demo to")
	cmd.Flags().BoolVar(&args.WithoutGateway, "without-gateway", args.WithoutGateway, "Skip MeshGateway resources")
	return cmd
}

func validateDemoArgs(args install_context.InstallDemoArgs) error {
	return nil
}
