//go:build !windows
// +build !windows

package uninstall

import (
	"fmt"
	"os"
	"runtime"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"

	"github.com/kumahq/kuma/pkg/transparentproxy"
	"github.com/kumahq/kuma/pkg/transparentproxy/config"
)

func newUninstallTransparentProxy() *cobra.Command {
	cfg := config.DefaultConfig()

	cmd := &cobra.Command{
		Use:   "transparent-proxy",
		Short: "Uninstall Transparent Proxy pre-requisites on the host",
		Long: "Uninstall Transparent Proxy by restoring the hosts iptables " +
			"and /etc/resolv.conf or removing leftover ebpf objects",
		PreRun: func(cmd *cobra.Command, _ []string) {
			cfg.RuntimeStdout = cmd.OutOrStdout()
			cfg.RuntimeStderr = cmd.ErrOrStderr()
		},
		RunE: func(cmd *cobra.Command, _ []string) error {
			if !cfg.DryRun && runtime.GOOS != "linux" {
				return errors.Errorf("transparent proxy will work only on Linux OSes")
			}

			initializedConfig, err := cfg.Initialize()
			if err != nil {
				return errors.Wrap(err, "failed to initialize config")
			}

			output, err := transparentproxy.Cleanup(initializedConfig)
			if err != nil {
				return errors.Wrap(err, "transparent proxy cleanup failed")
			}

			if cfg.Ebpf.Enabled {
				return nil
			}

			if cfg.DryRun {
				fmt.Fprintln(cfg.RuntimeStdout, output)
			}

			if _, err := os.Stat("/etc/resolv.conf.kuma-backup"); !os.IsNotExist(err) {
				content, err := os.ReadFile("/etc/resolv.conf.kuma-backup")
				if err != nil {
					return errors.Wrap(err, "unable to open /etc/resolv.conf.kuma-backup")
				}

				if !cfg.DryRun {
					err = os.WriteFile("/etc/resolv.conf", content, 0o600)
					if err != nil {
						return errors.Wrap(err, "unable to write /etc/resolv.conf")
					}
				}

				fmt.Fprintln(cfg.RuntimeStdout, string(content))
			}

			fmt.Fprintln(cfg.RuntimeStdout, "Transparent proxy cleaned up successfully")

			return nil
		},
	}

	// ebpf
	cmd.Flags().BoolVar(&cfg.Ebpf.Enabled, "ebpf-enabled", cfg.Ebpf.Enabled, "uninstall transparent proxy with ebpf mode")
	cmd.Flags().StringVar(&cfg.Ebpf.BPFFSPath, "ebpf-bpffs-path", cfg.Ebpf.BPFFSPath, "the path of the BPF filesystem")

	cmd.Flags().BoolVar(&cfg.DryRun, "dry-run", cfg.DryRun, "dry run")
	cmd.Flags().BoolVar(&cfg.Verbose, "verbose", cfg.Verbose, "verbose")

	return cmd
}
