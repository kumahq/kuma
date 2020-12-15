package install

import (
	"fmt"
	"runtime"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"

	"github.com/kumahq/kuma/pkg/transparentproxy"
)

type transparenProxyArgs struct {
	DryRun              bool
	ExcludeInboundPorts string
}

func newInstallTransparentProxy() *cobra.Command {
	args := transparenProxyArgs{
		DryRun:              false,
		ExcludeInboundPorts: "",
	}
	cmd := &cobra.Command{
		Use:   "transparent-proxy",
		Short: "Install Transparent Proxy pre-requisites on the host",
		Long:  `Install Transparent Proxy by modifying the hosts iptables, /etc/resolv.conf and creates \'envoy\' user`,
		RunE: func(cmd *cobra.Command, _ []string) error {
			if !args.DryRun && runtime.GOOS != "linux" {
				return errors.Errorf("transparent proxy will work only on Linux OSes")
			}

			tp := transparentproxy.GetDefaultTransparentProxy()

			output, err := tp.Setup(args.DryRun, args.ExcludeInboundPorts)
			if err != nil {
				return errors.Wrap(err, "Failed to setup transparent proxy")
			}

			if args.DryRun {
				fmt.Println(output)
			} else {
				fmt.Println("Transparent proxy set up successfully")
			}

			return nil
		},
	}
	cmd.Flags().BoolVar(&args.DryRun, "dry-run", args.DryRun, "dry run")
	cmd.Flags().StringVar(&args.ExcludeInboundPorts, "exclude-inbound-ports", args.ExcludeInboundPorts, "inbound ports to exclude from redirect to Envoy")
	return cmd
}
