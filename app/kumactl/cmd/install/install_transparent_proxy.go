//go:build !windows
// +build !windows

package install

import (
	"fmt"
	"runtime"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"

	core_config "github.com/kumahq/kuma/pkg/config"
	"github.com/kumahq/kuma/pkg/transparentproxy"
	"github.com/kumahq/kuma/pkg/transparentproxy/config"
	"github.com/kumahq/kuma/pkg/transparentproxy/firewalld"
	"github.com/kumahq/kuma/pkg/transparentproxy/iptables/consts"
)

func newInstallTransparentProxy() *cobra.Command {
	var configFile string

	cfg := config.DefaultConfig()

	cmd := &cobra.Command{
		Use:   "transparent-proxy",
		Short: "Install Transparent Proxy pre-requisites on the host",
		Long: `Install Transparent Proxy by modifying the hosts iptables.

Follow the following steps to use the Kuma data plane proxy in Transparent Proxy mode:

 1) create a dedicated user for the Kuma data plane proxy, e.g. 'kuma-dp'
 2) run this command as a 'root' user to modify the host's iptables and /etc/resolv.conf
    - supply the dedicated username with '--kuma-dp-user'
    - all changes are easly revertible by issuing 'kumactl uninstall transparent-proxy'
    - by default the SSH port tcp/22 will not be redirected to Envoy, but everything else will.
      Use '--exclude-inbound-ports' to provide a comma separated list of ports that should also be excluded

 sudo kumactl install transparent-proxy \
          --kuma-dp-user kuma-dp \
          --exclude-inbound-ports 443

 3) prepare a Dataplane resource yaml like this:

type: Dataplane
mesh: default
name: {{ name }}
networking:
  address: {{ address }}
  inbound:
  - port: {{ port }}
    tags:
      kuma.io/service: demo-client
  transparentProxying:
    redirectPortInbound: 15006
    redirectPortOutbound: 15001

The values in 'transparentProxying' section are the defaults set by this command and if needed be changed by supplying 
'--redirect-inbound-port' and '--redirect-outbound-port' respectively.

 4) the kuma-dp command shall be run with the designated user. 
    - if using systemd to run add 'User=kuma-dp' in the '[Service]' section of the service file
    - leverage 'runuser' similar to (assuming aforementioned yaml):

runuser -u kuma-dp -- \
  /usr/bin/kuma-dp run \
    --cp-address=https://172.19.0.2:5678 \
    --dataplane-token-file=/kuma/token-demo \
    --dataplane-file=/kuma/dpyaml-demo \
    --dataplane-var name=dp-demo \
    --dataplane-var address=172.19.0.4 \
    --dataplane-var port=80  \
    --binary-path /usr/local/bin/envoy

`,
		PreRunE: func(cmd *cobra.Command, _ []string) error {
			cfg.RuntimeStdout = cmd.OutOrStdout()
			cfg.RuntimeStderr = cmd.ErrOrStderr()

			if err := core_config.LoadWithOption(configFile, &cfg, false, true, false, "kuma_transparent_proxy"); err != nil {
				return err
			}

			// Ensure the Set method is called manually if the --kuma-dp-user flag is not specified
			// or if the value was not set in the config file. The Set method contains logic to check
			// for the existence of a user with the default UID "5678". If that does not exist, it
			// checks for the default username "kuma-dp". Since the Cobra library does not call
			// the Set method when --kuma-dp-user is not specified, we need to invoke it manually
			// here to ensure the proper user is set.
			if !cfg.Owner.Changed() {
				if err := cfg.Owner.Set(""); err != nil {
					return err
				}
			}

			return nil
		},
		RunE: func(cmd *cobra.Command, _ []string) error {
			if !cfg.DryRun && runtime.GOOS != "linux" {
				return errors.Errorf("transparent proxy will work only on Linux OSes")
			}

			if cfg.Redirect.DNS.CaptureAll && cfg.Redirect.DNS.Enabled {
				return errors.Errorf("one of --redirect-dns or --redirect-all-dns-traffic should be specified")
			}

			if cfg.Redirect.DNS.CaptureAll {
				cfg.Redirect.DNS.Enabled = true
			}

			if cfg.Ebpf.Enabled {
				if cfg.Ebpf.InstanceIP == "" {
					return errors.Errorf("--ebpf-instance-ip flag has to be specified --ebpf-enabled is provided")
				}

				if cfg.StoreFirewalld {
					fmt.Fprintln(cfg.RuntimeStderr, "# [WARNING] --store-firewalld will be ignored when --ebpf-enabled is being used")
				}

				if cfg.Redirect.DNS.SkipConntrackZoneSplit {
					fmt.Fprintln(cfg.RuntimeStderr, "# [WARNING] --skip-dns-conntrack-zone-split will be ignored when --ebpf-enabled is being used")
				}
			}

			initializedConfig, err := cfg.Initialize(cmd.Context())
			if err != nil {
				return errors.Wrap(err, "failed to initialize config")
			}

			output, err := transparentproxy.Setup(cmd.Context(), initializedConfig)
			if err != nil {
				return errors.Wrap(err, "failed to setup transparent proxy")
			}

			if !cfg.Ebpf.Enabled && cfg.StoreFirewalld {
				if _, err := firewalld.NewIptablesTranslator().
					WithDryRun(cfg.DryRun).
					WithOutput(cfg.RuntimeStdout).
					StoreRules(output); err != nil {
					return err
				}
			}

			if !initializedConfig.DryRun {
				initializedConfig.Logger.Info(
					"transparent proxy setup completed successfully. You can now run kuma-dp with the transparent-proxy feature enabled",
				)
			}

			return nil
		},
	}

	cmd.Flags().BoolVar(&cfg.DryRun, "dry-run", cfg.DryRun, "dry run")
	cmd.Flags().BoolVar(&cfg.Verbose, "verbose", cfg.Verbose, "verbose")
	cmd.Flags().Var(&cfg.IPFamilyMode, "ip-family-mode", "The IP family mode to enable traffic redirection for. Can be 'dualstack' or 'ipv4'")
	cmd.Flags().Var(&cfg.Redirect.Outbound.Port, "redirect-outbound-port", `outbound port redirected to Envoy, as specified in dataplane's "networking.transparentProxying.redirectPortOutbound"`)
	cmd.Flags().BoolVar(&cfg.Redirect.Inbound.Enabled, "redirect-inbound", cfg.Redirect.Inbound.Enabled, "redirect the inbound traffic to the Envoy. Should be disabled for Gateway data plane proxies.")
	cmd.Flags().Var(&cfg.Redirect.Inbound.Port, "redirect-inbound-port", `inbound port redirected to Envoy, as specified in dataplane's "networking.transparentProxying.redirectPortInbound"`)
	cmd.Flags().Var(&cfg.Redirect.Inbound.ExcludePorts, "exclude-inbound-ports", "a comma separated list of inbound ports to exclude from redirect to Envoy")
	cmd.Flags().Var(&cfg.Redirect.Outbound.ExcludePorts, "exclude-outbound-ports", "a comma separated list of outbound ports to exclude from redirect to Envoy")
	cmd.Flags().Var(&cfg.Owner, "kuma-dp-user", fmt.Sprintf("the username or UID of the user that will run kuma-dp. If not provided, the system will search for a user with the default UID ('%s') or the default username ('%s')", consts.OwnerDefaultUID, consts.OwnerDefaultUsername))
	cmd.Flags().Var(&cfg.Owner, "kuma-dp-uid", "the uid of the user that will run kuma-dp")
	cmd.Flags().BoolVar(&cfg.Redirect.DNS.Enabled, "redirect-dns", cfg.Redirect.DNS.Enabled, "redirect only DNS requests targeted to the servers listed in /etc/resolv.conf to a specified port")
	cmd.Flags().BoolVar(&cfg.Redirect.DNS.CaptureAll, "redirect-all-dns-traffic", cfg.Redirect.DNS.CaptureAll, "redirect all DNS traffic to a specified port, unlike --redirect-dns this will not be limited to the dns servers identified in /etc/resolve.conf")
	cmd.Flags().Var(&cfg.Redirect.DNS.Port, "redirect-dns-port", "the port where the DNS agent is listening")
	cmd.Flags().StringVar(&cfg.Redirect.DNS.UpstreamTargetChain, "redirect-dns-upstream-target-chain", cfg.Redirect.DNS.UpstreamTargetChain, "(optional) the iptables chain where the upstream DNS requests should be directed to. It is only applied for IP V4. Use with care.")
	cmd.Flags().BoolVar(&cfg.StoreFirewalld, "store-firewalld", cfg.StoreFirewalld, "store the iptables changes with firewalld")
	cmd.Flags().BoolVar(&cfg.Redirect.DNS.SkipConntrackZoneSplit, "skip-dns-conntrack-zone-split", cfg.Redirect.DNS.SkipConntrackZoneSplit, "skip applying conntrack zone splitting iptables rules")
	cmd.Flags().BoolVar(&cfg.DropInvalidPackets, "drop-invalid-packets", cfg.DropInvalidPackets, "This flag enables dropping of packets in invalid states, improving application stability by preventing them from reaching the backend. This is particularly beneficial during high-throughput requests where out-of-order packets might bypass DNAT. Note: Enabling this flag may introduce slight performance overhead. Weigh the trade-off between connection stability and performance before enabling it.")

	// ebpf
	cmd.Flags().BoolVar(&cfg.Ebpf.Enabled, "ebpf-enabled", cfg.Ebpf.Enabled, "use ebpf instead of iptables to install transparent proxy")
	cmd.Flags().StringVar(&cfg.Ebpf.ProgramsSourcePath, "ebpf-programs-source-path", cfg.Ebpf.ProgramsSourcePath, "path where compiled ebpf programs and other necessary for ebpf mode files can be found")
	cmd.Flags().StringVar(&cfg.Ebpf.InstanceIP, "ebpf-instance-ip", cfg.Ebpf.InstanceIP, "IP address of the instance (pod/vm) where transparent proxy will be installed")
	cmd.Flags().StringVar(&cfg.Ebpf.BPFFSPath, "ebpf-bpffs-path", cfg.Ebpf.BPFFSPath, "the path of the BPF filesystem")
	cmd.Flags().StringVar(&cfg.Ebpf.CgroupPath, "ebpf-cgroup-path", cfg.Ebpf.CgroupPath, "the path of cgroup2")
	cmd.Flags().StringVar(&cfg.Ebpf.TCAttachIface, "ebpf-tc-attach-iface", cfg.Ebpf.TCAttachIface, "name of the interface which TC eBPF programs should be attached to")

	cmd.Flags().StringArrayVar(&cfg.Redirect.Outbound.ExcludePortsForUIDs, "exclude-outbound-ports-for-uids", []string{}, "outbound ports to exclude for specific uids in a format of protocol:ports:uids where protocol and ports can be omitted or have value tcp or udp and ports can be a single value, a list, a range or a combination of all or * and uid can be a value or a range e.g. 53,3000-5000:106-108 would mean exclude ports 53 and from 3000 to 5000 for both TCP and UDP for uids 106, 107, 108")
	cmd.Flags().StringArrayVar(&cfg.Redirect.VNet.Networks, "vnet", cfg.Redirect.VNet.Networks, "virtual networks in a format of interfaceNameRegex:CIDR split by ':' where interface name doesn't have to be exact name e.g. docker0:172.17.0.0/16, br+:172.18.0.0/16, iface:::1/64")
	cmd.Flags().UintVar(&cfg.Wait, "wait", cfg.Wait, "specify the amount of time, in seconds, that the application should wait for the xtables exclusive lock before exiting. If the lock is not available within the specified time, the application will exit with an error")
	cmd.Flags().UintVar(&cfg.WaitInterval, "wait-interval", cfg.WaitInterval, "flag can be used to specify the amount of time, in microseconds, that iptables should wait between each iteration of the lock acquisition loop. This can be useful if the xtables lock is being held by another application for a long time, and you want to reduce the amount of CPU that iptables uses while waiting for the lock")
	cmd.Flags().IntVar(&cfg.Retry.MaxRetries, "max-retries", cfg.Retry.MaxRetries, "flag can be used to specify the maximum number of times to retry an installation before giving up")
	cmd.Flags().DurationVar(&cfg.Retry.SleepBetweenRetries, "sleep-between-retries", cfg.Retry.SleepBetweenRetries, "flag can be used to specify the amount of time to sleep between retries")

	cmd.Flags().BoolVar(&cfg.Log.Enabled, "iptables-logs", cfg.Log.Enabled, "enable logs for iptables rules using the LOG chain. This option activates kernel logging for packets matching the rules, where details about the IP/IPv6 headers are logged. This information can be accessed via dmesg(1) or syslog.")

	cmd.Flags().BoolVar(&cfg.Comments.Disabled, "disable-comments", cfg.Comments.Disabled, "Disable the addition of comments to iptables rules")

	cmd.Flags().StringArrayVar(&cfg.Redirect.Inbound.ExcludePortsForIPs, "exclude-inbound-ips", []string{}, "specify IP addresses (IPv4 or IPv6, with or without CIDR notation) to be excluded from transparent proxy inbound redirection. Examples: '10.0.0.1', '192.168.0.0/24', 'fe80::1', 'fd00::/8'. This flag can be specified multiple times or with multiple addresses separated by commas to exclude multiple IP addresses or ranges.")
	cmd.Flags().StringArrayVar(&cfg.Redirect.Outbound.ExcludePortsForIPs, "exclude-outbound-ips", []string{}, "specify IP addresses (IPv4 or IPv6, with or without CIDR notation) to be excluded from transparent proxy outbound redirection. Examples: '10.0.0.1', '192.168.0.0/24', 'fe80::1', 'fd00::/8'. This flag can be specified multiple times or with multiple addresses separated by commas to exclude multiple IP addresses or ranges.")

	_ = cmd.Flags().MarkDeprecated("redirect-dns-upstream-target-chain", "This flag has no effect anymore. Will be removed in 2.9.x version")
	_ = cmd.Flags().MarkDeprecated("kuma-dp-uid", "please use --kuma-dp-user, which accepts both UIDs and usernames")

	cmd.Flags().StringVar(&configFile, "transparent-proxy-config-file", configFile, "path to the transparent proxy configuration file")

	return cmd
}
