package install

import (
	"github.com/pkg/errors"
	"github.com/spf13/cobra"

	"github.com/Kong/kuma/app/kumactl/pkg/install/data"
	"github.com/Kong/kuma/app/kumactl/pkg/install/k8s"
	"github.com/Kong/kuma/app/kumactl/pkg/install/k8s/tracing"
)

type tracingTemplateArgs struct {
	Namespace string
}

func newInstallTracing() *cobra.Command {
	args := struct {
		Namespace string
	}{
		Namespace: "kuma-tracing",
	}
	cmd := &cobra.Command{
		Use:   "tracing",
		Short: "Install Tracing backend in Kubernetes cluster (Jaeger)",
		Long:  `Install Tracing backend in Kubernetes cluster (Jaeger) in a 'kuma-tracing' namespace`,
		RunE: func(cmd *cobra.Command, _ []string) error {
			templateArgs := tracingTemplateArgs{
				Namespace: args.Namespace,
			}

			templateFiles, err := data.ReadFiles(tracing.Templates)
			if err != nil {
				return errors.Wrap(err, "Failed to read template files")
			}

			renderedFiles, err := renderFiles(templateFiles, templateArgs, simpleTemplateRenderer)
			if err != nil {
				return errors.Wrap(err, "Failed to render template files")
			}

			sortedResources := k8s.SortResourcesByKind(renderedFiles)

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
