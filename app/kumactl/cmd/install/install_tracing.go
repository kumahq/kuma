package install

import (
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
)

func newInstallTracing() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "tracing",
		Short: "Install Tracing backend in Kubernetes cluster (Jaeger)",
		Long:  `Install Tracing backend in Kubernetes cluster (Jaeger) in its own namespace.`,
		RunE: func(cmd *cobra.Command, _ []string) error {
			return errors.New("we're migrating to `observability`, please use `install observability`")
		},
	}
	return cmd
}
