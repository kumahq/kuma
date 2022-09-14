// Deprecated: `kumactl install logging` is deprecated, use `kumactl install observability` instead

package install

import (
	"github.com/pkg/errors"
	"github.com/spf13/cobra"

	kumactl_data "github.com/kumahq/kuma/app/kumactl/data"
	"github.com/kumahq/kuma/app/kumactl/pkg/client/k8s"
	kumactl_cmd "github.com/kumahq/kuma/app/kumactl/pkg/cmd"
	"github.com/kumahq/kuma/app/kumactl/pkg/install/data"
)

type loggingTemplateArgs struct {
	Namespace string
}

func newInstallLogging(pctx *kumactl_cmd.RootContext) *cobra.Command {
	args := pctx.InstallLoggingContext.TemplateArgs
	cmd := &cobra.Command{
		Use:   "logging",
		Short: "Install Logging backend in Kubernetes cluster (Loki)",
		Long:  `Install Logging backend in Kubernetes cluster (Loki) in its own namespace.`,
		RunE: func(cmd *cobra.Command, _ []string) error {
			_, _ = cmd.ErrOrStderr().Write([]byte("# `logging` is deprecated, please use `install observability` to install logging, metrics and tracing"))
			templateArgs := loggingTemplateArgs{
				Namespace: args.Namespace,
			}

			templateFiles, err := data.ReadFiles(kumactl_data.InstallDeprecatedLoggingFS())
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
	cmd.Flags().StringVar(&args.Namespace, "namespace", args.Namespace, "namespace to install logging to")
	return cmd
}
