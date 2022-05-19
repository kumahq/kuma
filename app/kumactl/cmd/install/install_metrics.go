package install

import (
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
)

func newInstallMetrics() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "metrics",
		Short: "Install Metrics backend in Kubernetes cluster (Prometheus + Grafana)",
		Long:  `Install Metrics backend in Kubernetes cluster (Prometheus + Grafana) in its own namespace.`,
		RunE: func(cmd *cobra.Command, _ []string) error {
			return errors.New("we're migrating to `observability`, please use `install observability`")
		},
	}
	return cmd
}
