package install

import (
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
)

func newInstallLogging() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "logging",
		Short: "Install Logging backend in Kubernetes cluster (Loki)",
		Long:  `Install Logging backend in Kubernetes cluster (Loki) in its own namespace.`,
		RunE: func(cmd *cobra.Command, _ []string) error {
			return errors.New("we're migrating to `observability`, please use `install observability`")
		},
	}
	return cmd
}
