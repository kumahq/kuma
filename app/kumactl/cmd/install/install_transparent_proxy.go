//go:build !windows
// +build !windows

package install

import (
	"fmt"
	"os"
	"runtime"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"

	core_config "github.com/kumahq/kuma/pkg/config"
	"github.com/kumahq/kuma/pkg/transparentproxy"
	"github.com/kumahq/kuma/pkg/transparentproxy/config"
	"github.com/kumahq/kuma/pkg/transparentproxy/consts"
	"github.com/kumahq/kuma/pkg/transparentproxy/firewalld"
)

const (
	flagHelp                       = "help"
	flagDryRun                     = "dry-run"
	flagTransparentProxyConfig     = "config"
	flagTransparentProxyConfigFile = "config-file"
	flagIptablesExecutables        = "iptables-executables"
)

const (
	WarningDryRunRunningAsNonRoot = "# [WARNING] [dry-run]: running this command as a non-root user may lead to unpredictable results"
)

func newInstallTransparentProxy() *cobra.Command {
	var configValue string
	var configFile string

	cfg := config.DefaultConfig()
	cfgLoader := core_config.NewLoader(&cfg).WithEnvVarsLoading("KUMA_TRANSPARENT_PROXY")

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
		// Disable automatic flag parsing to ensure that our custom order of precedence
		// for configuring the transparent proxy is preserved (we'll manually parse flags)
		DisableFlagParsing: true,
		// We want to display usage information only when the user provides unknown flags. Therefore,
		// we manually handle these errors during flag parsing. Setting `SilenceUsage` to true
		// prevents automatic usage display for other types of errors
		SilenceUsage: true,
		PreRun: func(cmd *cobra.Command, args []string) {
			// This command does not use any of the global flags defined in the root `kumactl`
			// command (--api-timeout, --log-level, and --no-config). To avoid confusing users
			// or creating false expectations that these flags would have an effect, we hide them
			// from the flag list for this command
			cmd.Flags().VisitAll(func(flag *pflag.Flag) {
				if cmd.LocalFlags().Lookup(flag.Name) == nil {
					flag.Hidden = true
				}
			})
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg.RuntimeStdout = cmd.OutOrStdout()
			cfg.RuntimeStderr = cmd.ErrOrStderr()

			parseConfigFlags := func(flag *pflag.Flag, value string) error {
				switch flag.Name {
				case flagHelp, flagDryRun, flagTransparentProxyConfig, flagTransparentProxyConfigFile:
					return flag.Value.Set(value)
				default:
					return nil
				}
			}

			// To ensure the correct order of precedence, we first need to parse the `--config` and
			// `--config-file` flags if they are set. Additionally, since `DisableFlagParsing`
			// is enabled, we must handle the `--help` flag manually. During this parsing, if any
			// unknown flag errors are encountered, we wrap the error with a usage message because
			// the automatic usage display is disabled to prevent it from appearing for other
			// types of errors
			if err := cmd.Flags().ParseAll(args, parseConfigFlags); err != nil {
				return errors.Errorf("%s\n\n%s", err, cmd.UsageString())
			}

			// With `DisableFlagParsing` enabled, we are responsible for manually parsing all flags,
			// including the `--help` flag
			if ok, _ := cmd.Flags().GetBool(flagHelp); ok {
				return cmd.Help()
			}

			switch {
			case runtime.GOOS != "linux" && !cfg.DryRun:
				return errors.New("transparent proxy is supported only on Linux systems")
			case runtime.GOOS == "linux" && os.Geteuid() != 0 && !cfg.DryRun:
				return errors.New("you need to have root privileges to run this command")
			case runtime.GOOS == "linux" && os.Geteuid() != 0 && cfg.DryRun:
				fmt.Fprintln(cfg.RuntimeStderr, WarningDryRunRunningAsNonRoot)
			}

			if configValue == "-" && configFile == "" {
				return errors.Errorf(
					"provided value '-' via flag '--%s' is invalid; to provide config via stdin, use '--%s -'",
					flagTransparentProxyConfig,
					flagTransparentProxyConfigFile,
				)
			}

			// After parsing the config flags, we load the configuration, which involves parsing
			// the provided YAML or JSON, and including environment variables if present
			if err := cfgLoader.Load(cmd.InOrStdin(), []byte(configValue), configFile); err != nil {
				return errors.Wrap(err, "failed to load configuration from provided input")
			}

			// Finally, we parse the remaining CLI flags, as they have the highest priority in our
			// order of precedence. Since unknown flag errors were handled earlier, we'll catch here
			// other parsing errors for all other flags
			if err := cmd.Flags().Parse(args); err != nil {
				return err
			}

			// Ensure the Set method is called manually if the --kuma-dp-user flag is not specified
			// or if the value was not set in the config file. The Set method contains logic to check
			// for the existence of a user with the default UID "5678". If that does not exist, it
			// checks for the default username "kuma-dp". Since the Cobra library does not call
			// the Set method when --kuma-dp-user is not specified, we need to invoke it manually
			// here to ensure the proper user is set.
			if !cfg.KumaDPUser.Changed() {
				if err := cfg.KumaDPUser.Set(""); err != nil {
					return errors.Wrap(err, "failed to set default owner for transparent proxy")
				}
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
	cmd.Flags().Var(&cfg.KumaDPUser, "kuma-dp-user", fmt.Sprintf("the username or UID of the user that will run kuma-dp. If not provided, the system will search for a user with the default UID ('%s') or the default username ('%s')", consts.OwnerDefaultUID, consts.OwnerDefaultUsername))
	cmd.Flags().Var(&cfg.KumaDPUser, "kuma-dp-uid", "the uid of the user that will run kuma-dp")
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
	cmd.Flags().Var(&cfg.Retry.SleepBetweenRetries, "sleep-between-retries", "flag can be used to specify the amount of time to sleep between retries")

	cmd.Flags().BoolVar(&cfg.Log.Enabled, "iptables-logs", cfg.Log.Enabled, "enable logs for iptables rules using the LOG chain. This option activates kernel logging for packets matching the rules, where details about the IP/IPv6 headers are logged. This information can be accessed via dmesg(1) or syslog.")

	cmd.Flags().BoolVar(&cfg.Comments.Disabled, "disable-comments", cfg.Comments.Disabled, "Disable the addition of comments to iptables rules")

	cmd.Flags().StringArrayVar(&cfg.Redirect.Inbound.ExcludePortsForIPs, "exclude-inbound-ips", []string{}, "specify IP addresses (IPv4 or IPv6, with or without CIDR notation) to be excluded from transparent proxy inbound redirection. Examples: '10.0.0.1', '192.168.0.0/24', 'fe80::1', 'fd00::/8'. This flag can be specified multiple times or with multiple addresses separated by commas to exclude multiple IP addresses or ranges.")
	cmd.Flags().StringArrayVar(&cfg.Redirect.Outbound.ExcludePortsForIPs, "exclude-outbound-ips", []string{}, "specify IP addresses (IPv4 or IPv6, with or without CIDR notation) to be excluded from transparent proxy outbound redirection. Examples: '10.0.0.1', '192.168.0.0/24', 'fe80::1', 'fd00::/8'. This flag can be specified multiple times or with multiple addresses separated by commas to exclude multiple IP addresses or ranges.")

	cmd.Flags().BoolVar(
		&cfg.Redirect.Inbound.InsertRedirectInsteadOfAppend,
		"redirect-inbound-insert-instead-of-append",
		cfg.Redirect.Inbound.InsertRedirectInsteadOfAppend,
		fmt.Sprintf(
			"for inbound traffic, by default, the last applied iptables rule in the 'PREROUTING' chain of the 'nat' table redirects traffic to our custom chain ('%s') for handling transparent proxying. If there is an existing rule in this chain that redirects traffic to another chain, our default behavior of appending the rule would cause it to be added after the existing one, making our rule ineffective. Specifying this flag changes the behavior to insert the rule at the beginning of the chain, ensuring our rule takes precedence. Note that if the '--vnet' flag is also specified, the default behavior is already to insert the rule, so using this flag will not change that behavior",
			fmt.Sprintf("%s_%s", cfg.Redirect.NamePrefix, cfg.Redirect.Inbound.ChainName),
		),
	)
	cmd.Flags().BoolVar(
		&cfg.Redirect.Outbound.InsertRedirectInsteadOfAppend,
		"redirect-outbound-insert-instead-of-append",
		cfg.Redirect.Outbound.InsertRedirectInsteadOfAppend,
		fmt.Sprintf(
			"for outbound traffic, by default, the last applied iptables rule in the 'OUTPUT' chain of the 'nat' table redirects traffic to our custom chain ('%s'), where it is processed for transparent proxying. However, if there is an existing rule in this chain that already redirects traffic to another chain, our default behavior of appending the rule will cause our rule to be added after the existing one, effectively ignoring it. When this flag is specified, it changes the behavior from appending to inserting the rule at the beginning of the chain, ensuring that our iptables rule takes precedence",
			fmt.Sprintf("%s_%s", cfg.Redirect.NamePrefix, cfg.Redirect.Outbound.ChainName),
		),
	)

	cmd.Flags().Var(
		&cfg.Executables,
		flagIptablesExecutables,
		fmt.Sprintf(
			"specify custom paths for iptables executables in the format name:path[,name:path...]. Valid names are 'iptables', 'iptables-save', 'iptables-restore', 'ip6tables', 'ip6tables-save', and 'ip6tables-restore'. You must provide all three executables for each IP version you want to customize (IPv4 or IPv6), meaning if you configure one for IPv6 (e.g., 'ip6tables'), you must also specify 'ip6tables-save' and 'ip6tables-restore'. Partial configurations for either IPv4 or IPv6 are not allowed. Configuration values can be set through a combination of sources: config file (via --%s or --%s), environment variables, and the '--%s' flag. For example, you can specify 'ip6tables' in the config file, 'ip6tables-save' as an environment variable, and 'ip6tables-restore' via the '--%[3]s' flag. [WARNING] Provided paths are not extensively validated, so ensure you specify correct paths and that the executables are actual iptables binaries to avoid misconfigurations and unexpected behavior",
			flagTransparentProxyConfig,
			flagTransparentProxyConfigFile,
			flagIptablesExecutables,
		),
	)

	cmd.Flags().StringVar(&configValue, flagTransparentProxyConfig, configValue, "transparent proxy configuration provided in YAML or JSON format")
	cmd.Flags().StringVar(&configFile, flagTransparentProxyConfigFile, configFile, "path to the file containing the transparent proxy configuration in YAML or JSON format")

	_ = cmd.Flags().MarkDeprecated("redirect-dns-upstream-target-chain", "This flag has no effect anymore. Will be removed in 2.9.x version")
	_ = cmd.Flags().MarkDeprecated("kuma-dp-uid", "please use --kuma-dp-user, which accepts both UIDs and usernames")

	// Manually define the `--help` flag since we are handling flag parsing ourselves due to
	// `DisableFlagParsing` being set
	_ = cmd.Flags().BoolP(flagHelp, "h", false, "show this help message")

	return cmd
}
