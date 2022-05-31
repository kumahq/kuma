// Deprecated: `kumactl install tracing` is deprecated, use `kumactl install observability` instead

package install

import (
	"github.com/pkg/errors"
	"github.com/spf13/cobra"

	kumactl_data "github.com/kumahq/kuma/app/kumactl/data"
	kumactl_cmd "github.com/kumahq/kuma/app/kumactl/pkg/cmd"
	"github.com/kumahq/kuma/app/kumactl/pkg/install/data"
	"github.com/kumahq/kuma/app/kumactl/pkg/install/k8s"
)

type tracingTemplateArgs struct {
	Namespace string
}

func newInstallTracing(pctx *kumactl_cmd.RootContext) *cobra.Command {
	args := pctx.InstallTracingContext.TemplateArgs
	cmd := &cobra.Command{
		Use:   "tracing",
		Short: "Install Tracing backend in Kubernetes cluster (Jaeger)",
		Long:  `Install Tracing backend in Kubernetes cluster (Jaeger) in its own namespace.`,
		RunE: func(cmd *cobra.Command, _ []string) error {
			cmd.Deprecated = "we're migrating to `observability`, please use `install observability`"
			templateArgs := tracingTemplateArgs{
				Namespace: args.Namespace,
			}

			templateFiles, err := data.ReadFiles(kumactl_data.InstallDeprecatedTracingFS())
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
	cmd.Flags().StringVar(&args.Namespace, "namespace", args.Namespace, "namespace to install tracing to")
	return cmd
}
