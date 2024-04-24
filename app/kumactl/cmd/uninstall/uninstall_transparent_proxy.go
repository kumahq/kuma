//go:build !windows
// +build !windows

package uninstall

import (
	"os"
	"runtime"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"

	"github.com/kumahq/kuma/pkg/transparentproxy"
	"github.com/kumahq/kuma/pkg/transparentproxy/config"
)

func newUninstallTransparentProxy() *cobra.Command {
	cfg := config.Config{
		Ebpf: config.Ebpf{
			Enabled:   false,
			BPFFSPath: "/sys/fs/bpf",
		},
		Verbose: false,
		DryRun:  false,
	}

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

			output, err := transparentproxy.Cleanup(cfg)
			if err != nil {
				return errors.Wrap(err, "transparent proxy cleanup failed")
			}

			if cfg.Ebpf.Enabled {
				return nil
			}

			if cfg.DryRun {
				_, _ = cmd.OutOrStdout().Write([]byte(output))
				_, _ = cmd.OutOrStdout().Write([]byte("\n"))
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

				_, _ = cmd.OutOrStdout().Write(content)
			}
			_, _ = cmd.OutOrStdout().Write([]byte("\nTransparent proxy cleaned up successfully\n"))

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
