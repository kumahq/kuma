//build: gateway

package install

import (
	"github.com/pkg/errors"
	"github.com/spf13/cobra"

	kumactl_cmd "github.com/kumahq/kuma/app/kumactl/pkg/cmd"
	"github.com/kumahq/kuma/app/kumactl/pkg/install/data"
	gateway_api_types "github.com/kumahq/kuma/deployments/gateway-api/types"
)

func init() {
	optionalInstallCmds = append(optionalInstallCmds,
		func(*kumactl_cmd.RootContext) *cobra.Command {
			return newInstallGatewayControllerCmd()
		})
}

func newInstallGatewayControllerCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "gateway-controller",
		Short: "Install the Kuma Gateway controller on Kubernetes",
		RunE: func(cmd *cobra.Command, _ []string) error {
			gatewayFiles, err := gateway_api_types.Generate()
			if err != nil {
				return err
			}

			singleFile := data.JoinYAML(gatewayFiles)
			if _, err := cmd.OutOrStdout().Write(singleFile.Data); err != nil {
				return errors.Wrap(err, "Failed to output rendered resources")
			}

			return nil
		},
	}

	return cmd
}
