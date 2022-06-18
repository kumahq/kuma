// Deprecated: `kumactl install tracing` is deprecated, use `kumactl install observability` instead

package install

import (
	"fmt"

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
			_, _ = cmd.ErrOrStderr().Write([]byte("# `tracing` is deprecated, please use `install observability` to install logging, metrics and tracing"))
			templateArgs := tracingTemplateArgs{
				Namespace: args.Namespace,
			}

			templateFiles, err := data.ReadFiles(kumactl_data.InstallDeprecatedTracingFS())
			if err != nil {
				return fmt.Errorf("Failed to read template files: %w", err)
			}

			renderedFiles, err := renderFiles(templateFiles, templateArgs, simpleTemplateRenderer)
			if err != nil {
				return fmt.Errorf("Failed to render template files: %w", err)
			}

			sortedResources, err := k8s.SortResourcesByKind(renderedFiles)
			if err != nil {
				return fmt.Errorf("Failed to sort resources by kind: %w", err)
			}

			singleFile := data.JoinYAML(sortedResources)

			if _, err := cmd.OutOrStdout().Write(singleFile.Data); err != nil {
				return fmt.Errorf("Failed to output rendered resources: %w", err)
			}
			return nil
		},
	}
	cmd.Flags().StringVar(&args.Namespace, "namespace", args.Namespace, "namespace to install tracing to")
	return cmd
}
