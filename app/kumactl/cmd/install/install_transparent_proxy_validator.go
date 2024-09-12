//go:build !windows
// +build !windows

package install

import (
	std_errors "errors"
	"fmt"
	"os"
	"strings"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"

	"github.com/kumahq/kuma/pkg/core"
	kuma_log "github.com/kumahq/kuma/pkg/log"
	tproxy_config "github.com/kumahq/kuma/pkg/transparentproxy/config"
	tproxy_consts "github.com/kumahq/kuma/pkg/transparentproxy/consts"
	tproxy_validate "github.com/kumahq/kuma/pkg/transparentproxy/validate"
)

const defaultLogName = "transparentproxy.validator"

func newInstallTransparentProxyValidator() *cobra.Command {
	ipFamilyMode := tproxy_config.IPFamilyModeDualStack
	serverPort := tproxy_validate.ServerPort

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
			logger := core.NewLoggerTo(os.Stdout, kuma_log.InfoLevel).WithName(defaultLogName)
			hasLocalIPv6Addr, _ := tproxy_config.HasLocalIPv6()
			validateOnlyIPv4 := ipFamilyMode == tproxy_config.IPFamilyModeIPv4

			validate := func(ipv6 bool) error {
				if ipv6 && !hasLocalIPv6Addr || validateOnlyIPv4 {
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
		},
	}

	cmd.Flags().Var(
		&ipFamilyMode,
		"ip-family-mode",
		fmt.Sprintf(
			"specify the IP family mode for traffic redirection when setting up the transparent proxy; accepted values: %s",
			tproxy_config.AllowedIPFamilyModes(),
		),
	)

	cmd.Flags().Uint16Var(
		&serverPort,
		"validation-server-port",
		serverPort,
		"port number for the validation server to listen on",
	)

	return cmd
}
