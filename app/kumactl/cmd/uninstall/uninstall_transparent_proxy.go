//go:build !windows
// +build !windows

package uninstall

import (
	"os"
	"runtime"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"

	"github.com/kumahq/kuma/pkg/transparentproxy"
)

type transparentProxyArgs struct {
	DryRun  bool
	Verbose bool
}

func newUninstallTransparentProxy() *cobra.Command {
	args := transparentProxyArgs{
		DryRun:  false,
		Verbose: false,
	}
	cmd := &cobra.Command{
		Use:   "transparent-proxy",
		Short: "Uninstall Transparent Proxy pre-requisites on the host",
		Long:  `Uninstall Transparent Proxy by restoring the hosts iptables and /etc/resolv.conf`,
		RunE: func(cmd *cobra.Command, _ []string) error {
			if !args.DryRun && runtime.GOOS != "linux" {
				return errors.Errorf("transparent proxy will work only on Linux OSes")
			}

			tp := transparentproxy.DefaultTransparentProxy()

			output, err := tp.Cleanup(args.DryRun, args.Verbose)
			if err != nil {
				return errors.Wrap(err, "Failed to cleanup transparent proxy")
			}

			if args.DryRun {
				_, _ = cmd.OutOrStdout().Write([]byte(output))
				_, _ = cmd.OutOrStdout().Write([]byte("\n"))
			}

			if _, err := os.Stat("/etc/resolv.conf.kuma-backup"); !os.IsNotExist(err) {
				content, err := os.ReadFile("/etc/resolv.conf.kuma-backup")
				if err != nil {
					return errors.Wrap(err, "unable to open /etc/resolv.conf.kuma-backup")
				}

				if !args.DryRun {
					err = os.WriteFile("/etc/resolv.conf", content, 0644)
					if err != nil {
						return errors.Wrap(err, "unable to write /etc/resolv.conf")
					}
				}

				_, _ = cmd.OutOrStdout().Write(content)
			}
			_, _ = cmd.OutOrStdout().Write([]byte("\nTransparent proxy cleaned up successfully\n"))

			return nil
		},
	}

	cmd.Flags().BoolVar(&args.DryRun, "dry-run", args.DryRun, "dry run")
	cmd.Flags().BoolVar(&args.Verbose, "verbose", args.Verbose, "verbose")
	return cmd
}
