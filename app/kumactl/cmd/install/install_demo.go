package install

import (
	"fmt"

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
	cmd.Flags().StringVar(&args.Zone, "zone", args.Zone, "Zone in which to install demo")
	cmd.Flags().StringVar(&args.Namespace, "namespace", args.Namespace, "Namespace to install demo to")
	return cmd
}

func validateDemoArgs(args install_context.InstallDemoArgs) error {
	return nil
}
