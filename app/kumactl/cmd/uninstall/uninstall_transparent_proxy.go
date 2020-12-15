package uninstall

import (
	"fmt"
	"runtime"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"

	"github.com/kumahq/kuma/pkg/transparentproxy"
)

type transparenProxyArgs struct {
	DryRun bool
}

func newUninstallTransparentProxy() *cobra.Command {
	args := transparenProxyArgs{
		DryRun: false,
	}
	cmd := &cobra.Command{
		Use:   "transparent-proxy",
		Short: "Uninstall Transparent Proxy pre-requisites on the host",
		Long:  `Uninstall Transparent Proxy by restoring the hosts iptables and /etc/resolv.conf`,
		RunE: func(cmd *cobra.Command, _ []string) error {
			if !args.DryRun && runtime.GOOS != "linux" {
				return errors.Errorf("transparent proxy will work only on Linux OSes")
			}

			tp := transparentproxy.GetDefaultTransparentProxy()

			output, err := tp.Cleanup(args.DryRun)
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
	return cmd
}
