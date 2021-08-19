package uninstall

import (
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
)

func newUninstallTransparentProxy() *cobra.Command {
	return &cobra.Command{
		Use:   "transparent-proxy",
		Short: "Uninstall Transparent Proxy pre-requisites on the host",
		RunE: func(_ *cobra.Command, _ []string) error {
			return errors.New("This command is not supported on your operating system")
		},
	}
}
