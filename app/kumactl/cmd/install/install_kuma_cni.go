package install

import (
	"github.com/Kong/kuma/app/kumactl/pkg/install/data"
	"github.com/Kong/kuma/app/kumactl/pkg/install/k8s"
	kumacni "github.com/Kong/kuma/app/kumactl/pkg/install/k8s/kuma-cni"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
)

func newInstallKumaCNICmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "kuma-cni",
		Short: "Install Kuma CNI plugin on Kubernetes",
		Long:  "Install Kuma CNI plugin on Kubernetes, in a 'kube-system' namespace.",
		RunE: func(cmd *cobra.Command, _ []string) error {
			templateFiles, err := data.ReadFiles(kumacni.Templates)
			if err != nil {
				return errors.Wrap(err, "Failed to read template files")
			}

			renderedFiles, err := renderFiles(templateFiles, nil, simpleTemplateRenderer)
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
	return cmd
}
