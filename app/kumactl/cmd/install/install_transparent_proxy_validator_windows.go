package install

import (
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
)

func newInstallTransparentProxyValidator() *cobra.Command {
	return &cobra.Command{
		Use:   "transparent-proxy-validator",
		Short: "Validates if transparent proxy has been set up successfully",
		RunE: func(_ *cobra.Command, _ []string) error {
			return errors.New("This command is not supported on your operating system")
		},
	}
}
