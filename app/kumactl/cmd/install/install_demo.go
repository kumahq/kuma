package install

import (
	"github.com/pkg/errors"
	"github.com/spf13/cobra"

	install_context "github.com/kumahq/kuma/v3/app/kumactl/cmd/install/context"
	kumactl_data "github.com/kumahq/kuma/v3/app/kumactl/data"
	"github.com/kumahq/kuma/v3/app/kumactl/pkg/install/k8s"
	"github.com/kumahq/kuma/v3/pkg/util/data"
)

type demoTemplateArgs struct {
	Namespace       string
	SystemNamespace string
	Zone            string
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

			renderedFiles, err := renderFilesWithFilter(templateFiles, templateArgs, simpleTemplateRenderer, NoneFilter{})
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
	return cmd
}
