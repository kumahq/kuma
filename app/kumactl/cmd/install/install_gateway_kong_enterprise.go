package install

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	install_context "github.com/kumahq/kuma/app/kumactl/cmd/install/context"
	kumactl_data "github.com/kumahq/kuma/app/kumactl/data"
	"github.com/kumahq/kuma/app/kumactl/pkg/install/data"
	"github.com/kumahq/kuma/app/kumactl/pkg/install/k8s"
)

type templateArgs struct {
	Namespace   string
	LicenseText string
}

func newInstallGatewayKongEnterpriseCmd(ctx *install_context.InstallGatewayKongEnterpriseContext) *cobra.Command {
	args := ctx.Args
	cmd := &cobra.Command{
		Use:   "kong-enterprise",
		Short: "Install Kong ingress gateway on Kubernetes",
		Long:  "Install Kong ingress gateway on Kubernetes in its own namespace.",
		RunE: func(cmd *cobra.Command, _ []string) error {
			templateFiles, err := data.ReadFiles(kumactl_data.InstallGatewayKongEnterpriseFS())
			if err != nil {
				return fmt.Errorf("Failed to read template files: %w", err)
			}

			licenseBytes, err := os.ReadFile(args.LicensePath)
			if err != nil {
				return fmt.Errorf("Failed to read license file: %w", err)
			}

			templateArgs := templateArgs{
				Namespace:   args.Namespace,
				LicenseText: string(licenseBytes),
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
	cmd.Flags().StringVar(&args.Namespace, "namespace", args.Namespace, "namespace to install gateway to")
	cmd.Flags().StringVar(&args.LicensePath, "license-path", args.LicensePath, "path to license file")
	_ = cmd.MarkFlagRequired("license-path")
	return cmd
}
