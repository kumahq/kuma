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
	Namespace       string
	SystemNamespace string
	Zone            string
}

type GatewayFilter struct{}

func (GatewayFilter) Filter(name string) bool {
	return !strings.HasSuffix(name, "gateway.yaml")
}

func newInstallDemoCmd(ctx *install_context.InstallDemoContext) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "demo",
		Short: "Install Kuma demo on Kubernetes",
		Long:  "Install Kuma demo on Kubernetes in its own namespace.",
		RunE: func(cmd *cobra.Command, _ []string) error {
			templateArgs := demoTemplateArgs{
				Namespace:       ctx.Args.Namespace,
				Zone:            ctx.Args.Zone,
				SystemNamespace: ctx.Args.SystemNamespace,
			}

			templateFiles, err := data.ReadFiles(kumactl_data.InstallDemoFS())
			if err != nil {
				return errors.Wrap(err, "Failed to read template files")
			}

			var filter templateFilter = NoneFilter{}
			if ctx.Args.WithoutGateway {
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
	cmd.Flags().StringVar(&ctx.Args.Zone, "zone", ctx.Args.Zone, "Zone in which to install demo")
	cmd.Flags().StringVar(&ctx.Args.Namespace, "namespace", ctx.Args.Namespace, "Namespace to install demo to")
	cmd.Flags().StringVar(&ctx.Args.SystemNamespace, "system-namespace", ctx.Args.SystemNamespace, "System namespace of the control plane")
	cmd.Flags().BoolVar(&ctx.Args.WithoutGateway, "without-gateway", ctx.Args.WithoutGateway, "Skip MeshGateway resources")
	return cmd
}
