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

// shouldSkipValidation returns true when the given IP version should not be
// validated. IPv6 is skipped when the node has no local IPv6 address or the
// mode is IPv4-only. IPv4 is never skipped.
func shouldSkipValidation(ipv6, hasLocalIPv6Addr, validateOnlyIPv4 bool) bool {
	return ipv6 && (!hasLocalIPv6Addr || validateOnlyIPv4)
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

<<<<<<< HEAD
			// by using 0, we make the client to generate a random port to connect verifying the iptables rules are working
			err = validator.RunClient(uint16(0), sExit)
			return err
=======
			hasLocalIPv6Addr, _ := tproxy_config.HasLocalIPv6()
			validateOnlyIPv4 := ipFamilyMode == tproxy_config.IPFamilyModeIPv4

			validate := func(ipv6 bool) error {
				if shouldSkipValidation(ipv6, hasLocalIPv6Addr, validateOnlyIPv4) {
					return nil
				}

				logger := logger.WithName(strings.ToLower(tproxy_consts.IPTypeMap[ipv6]))
				validator := tproxy_validate.NewValidator(ipv6, serverPort, logger)
				exitC := make(chan struct{})

				if _, err := validator.RunServer(cmd.Context(), exitC); err != nil {
					return err
				}

				// by using 0, we make the client to generate a random port to connect verifying
				// the iptables rules are working
				return validator.RunClient(cmd.Context(), 0, exitC)
			}

			return errors.Wrap(
				std_errors.Join(validate(false), validate(true)),
				"validation failed",
			)
>>>>>>> 6f4783c3f5 (fix(kuma-init): properly validate ip family condition (#16810))
		},
	}

	cmd.Flags().StringVar(&args.IpFamilyMode, "ip-family-mode", args.IpFamilyMode, "The IP family mode that has enabled traffic redirection for when setting up transparent proxy. Can be 'dualstack' or 'ipv4'")
	cmd.Flags().Uint16Var(&args.ValidationServerPort, "validation-server-port", args.ValidationServerPort, "The port that the validation server will listen")
	return cmd
}
