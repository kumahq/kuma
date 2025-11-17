//go:build !windows
// +build !windows

package install

import (
	"os"

	"github.com/spf13/cobra"

	"github.com/kumahq/kuma/v2/pkg/core"
	kuma_log "github.com/kumahq/kuma/v2/pkg/log"
	"github.com/kumahq/kuma/v2/pkg/transparentproxy"
	"github.com/kumahq/kuma/v2/pkg/transparentproxy/validate"
)

const defaultLogName = "validator"

type transparentProxyValidatorArgs struct {
	IpFamilyMode         string
	ValidationServerPort uint16
}

func newInstallTransparentProxyValidator() *cobra.Command {
	args := transparentProxyValidatorArgs{
		IpFamilyMode:         "dualstack",
		ValidationServerPort: validate.ValidationServerPort,
	}
	cmd := &cobra.Command{
		Use:   "transparent-proxy-validator",
		Short: "Validates if transparent proxy has been set up successfully",
		Long: `Validates the transparent proxy setup by testing if the applied 
iptables rules are working correctly onto the pod.

Follow the following steps to validate:
 1) install the transparent proxy using 'kumactl install transparent-proxy'
 2) run this command

The result will be shown as text in stdout as well as the exit code.
`,
		RunE: func(cmd *cobra.Command, _ []string) error {
			log := core.NewLoggerTo(os.Stdout, kuma_log.InfoLevel).WithName(defaultLogName)

			ipv6Supported, _ := transparentproxy.HasLocalIPv6()
			useIPv6 := ipv6Supported && args.IpFamilyMode == "ipv6"

			sExit := make(chan struct{})
			validator := validate.NewValidator(useIPv6, args.ValidationServerPort, log)
			_, err := validator.RunServer(sExit)
			if err != nil {
				return err
			}

			// by using 0, we make the client to generate a random port to connect verifying the iptables rules are working
			err = validator.RunClient(uint16(0), sExit)
			return err
		},
	}

	cmd.Flags().StringVar(&args.IpFamilyMode, "ip-family-mode", args.IpFamilyMode, "The IP family mode that has enabled traffic redirection for when setting up transparent proxy. Can be 'dualstack' or 'ipv4'")
	cmd.Flags().Uint16Var(&args.ValidationServerPort, "validation-server-port", args.ValidationServerPort, "The port that the validation server will listen")
	return cmd
}
