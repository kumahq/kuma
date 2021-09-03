package install

import (
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
)

func newInstallTransparentProxy() *cobra.Command {
	return &cobra.Command{
		Use:   "transparent-proxy",
		Short: "Install Transparent Proxy pre-requisites on the host",
		RunE: func(_ *cobra.Command, _ []string) error {
			return errors.New("This command is not supported on your operating system")
		},
	}
}
