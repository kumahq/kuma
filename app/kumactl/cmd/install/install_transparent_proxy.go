//go:build !windows
// +build !windows

package install

import (
	"net"
	os_user "os/user"
	"runtime"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"

	"github.com/kumahq/kuma-net/firewalld"

	"github.com/kumahq/kuma/pkg/transparentproxy"
	"github.com/kumahq/kuma/pkg/transparentproxy/config"
)

type transparentProxyArgs struct {
	DryRun                             bool
	Verbose                            bool
	RedirectPortOutBound               string
	RedirectInbound                    bool
	RedirectPortInBound                string
	RedirectPortInBoundV6              string
	ExcludeInboundPorts                string
	ExcludeOutboundPorts               string
	ExcludeOutboundTCPPortsForUIDs     []string
	ExcludeOutboundUDPPortsForUIDs     []string
	UID                                string
	User                               string
	RedirectDNS                        bool
	RedirectAllDNSTraffic              bool
	AgentDNSListenerPort               string
	DNSUpstreamTargetChain             string
	StoreFirewalld                     bool
	SkipDNSConntrackZoneSplit          bool
	ExperimentalTransparentProxyEngine bool
	EbpfEnabled                        bool
	EbpfProgramsSourcePath             string
	EbpfInstanceIP                     string
	EbpfBPFFSPath                      string
}

func newInstallTransparentProxy() *cobra.Command {
	args := transparentProxyArgs{
		DryRun:                             false,
		Verbose:                            false,
		RedirectPortOutBound:               "15001",
		RedirectInbound:                    true,
		RedirectPortInBound:                "15006",
		RedirectPortInBoundV6:              "15010",
		ExcludeInboundPorts:                "",
		ExcludeOutboundPorts:               "",
		ExcludeOutboundTCPPortsForUIDs:     []string{},
		ExcludeOutboundUDPPortsForUIDs:     []string{},
		UID:                                "",
		User:                               "",
		RedirectDNS:                        false,
		RedirectAllDNSTraffic:              false,
		AgentDNSListenerPort:               "15053",
		DNSUpstreamTargetChain:             "RETURN",
		StoreFirewalld:                     false,
		SkipDNSConntrackZoneSplit:          false,
		ExperimentalTransparentProxyEngine: false,
		EbpfEnabled:                        false,
		EbpfProgramsSourcePath:             "/kuma/ebpf",
		EbpfBPFFSPath:                      "/run/kuma/bpf",
	}
	cmd := &cobra.Command{
		Use:   "transparent-proxy",
		Short: "Install Transparent Proxy pre-requisites on the host",
		Long: `Install Transparent Proxy by modifying the hosts iptables.

Follow the following steps to use the Kuma data plane proxy in Transparent Proxy mode:

 1) create a dedicated user for the Kuma data plane proxy, e.g. 'kuma-dp'
 2) run this command as a 'root' user to modify the host's iptables and /etc/resolv.conf
    - supply the dedicated username with '--kuma-dp-uid'
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
		RunE: func(cmd *cobra.Command, _ []string) error {
			if !args.DryRun && runtime.GOOS != "linux" {
				return errors.Errorf("transparent proxy will work only on Linux OSes")
			}

			if args.User == "" && args.UID == "" {
				return errors.Errorf("--kuma-dp-user or --kuma-dp-uid should be supplied")
			}

			if args.RedirectAllDNSTraffic && args.RedirectDNS {
				return errors.Errorf("one of --redirect-dns or --redirect-all-dns-traffic should be specified")
			}

			if args.RedirectAllDNSTraffic {
				args.RedirectDNS = true
			}

			if args.DNSUpstreamTargetChain != "RETURN" {
				_, _ = cmd.ErrOrStderr().Write([]byte("# `--redirect-dns-upstream-target-chain` is deprecated, please avoid using it"))
			}

			if args.EbpfEnabled {
				if args.EbpfInstanceIP == "" {
					return errors.Errorf("--ebpf-instance-ip flag has to be specified --ebpf-enabled is provided")
				}

				if !args.ExperimentalTransparentProxyEngine {
					return errors.Errorf("--experimental-transparent-proxy-engine flag has to be specified when --ebpf-enabled is provided")
				}

				if args.StoreFirewalld {
					_, _ = cmd.ErrOrStderr().Write([]byte("# --store-firewalld will be ignored when --ebpf-enabled is being used"))
				}

				if args.SkipDNSConntrackZoneSplit {
					_, _ = cmd.ErrOrStderr().Write([]byte("# --skip-dns-conntrack-zone-split will be ignored when --ebpf-enabled is being used"))
				}
			}

			if err := configureTransparentProxy(cmd, &args); err != nil {
				return err
			}

			_, _ = cmd.OutOrStdout().Write([]byte("Transparent proxy set up successfully\n"))
			return nil
		},
	}

	cmd.Flags().BoolVar(&args.DryRun, "dry-run", args.DryRun, "dry run")
	cmd.Flags().BoolVar(&args.Verbose, "verbose", args.Verbose, "verbose")
	cmd.Flags().StringVar(&args.RedirectPortOutBound, "redirect-outbound-port", args.RedirectPortOutBound, "outbound port redirected to Envoy, as specified in dataplane's `networking.transparentProxying.redirectPortOutbound`")
	cmd.Flags().BoolVar(&args.RedirectInbound, "redirect-inbound", args.RedirectInbound, "redirect the inbound traffic to the Envoy. Should be disabled for Gateway data plane proxies.")
	cmd.Flags().StringVar(&args.RedirectPortInBound, "redirect-inbound-port", args.RedirectPortInBound, "inbound port redirected to Envoy, as specified in dataplane's `networking.transparentProxying.redirectPortInbound`")
	cmd.Flags().StringVar(&args.RedirectPortInBoundV6, "redirect-inbound-port-v6", args.RedirectPortInBoundV6, "IPv6 inbound port redirected to Envoy, as specified in dataplane's `networking.transparentProxying.redirectPortInboundV6`")
	cmd.Flags().StringVar(&args.ExcludeInboundPorts, "exclude-inbound-ports", args.ExcludeInboundPorts, "a comma separated list of inbound ports to exclude from redirect to Envoy")
	cmd.Flags().StringVar(&args.ExcludeOutboundPorts, "exclude-outbound-ports", args.ExcludeOutboundPorts, "a comma separated list of outbound ports to exclude from redirect to Envoy")
	cmd.Flags().StringVar(&args.User, "kuma-dp-user", args.UID, "the user that will run kuma-dp")
	cmd.Flags().StringVar(&args.UID, "kuma-dp-uid", args.UID, "the UID of the user that will run kuma-dp")
	cmd.Flags().BoolVar(&args.RedirectDNS, "redirect-dns", args.RedirectDNS, "redirect only DNS requests targeted to the servers listed in /etc/resolv.conf to a specified port")
	// Deprecation issue: https://github.com/kumahq/kuma/issues/4759
	cmd.Flags().BoolVar(&args.RedirectAllDNSTraffic, "redirect-all-dns-traffic", args.RedirectAllDNSTraffic, "redirect all DNS traffic to a specified port, unlike --redirect-dns this will not be limited to the dns servers identified in /etc/resolve.conf")
	cmd.Flags().StringVar(&args.AgentDNSListenerPort, "redirect-dns-port", args.AgentDNSListenerPort, "the port where the DNS agent is listening")
	cmd.Flags().StringVar(&args.DNSUpstreamTargetChain, "redirect-dns-upstream-target-chain", args.DNSUpstreamTargetChain, "(optional) the iptables chain where the upstream DNS requests should be directed to. It is only applied for IP V4. Use with care.")
	// Deprecation issue: https://github.com/kumahq/kuma/issues/4759
	_ = cmd.Flags().Bool("skip-resolv-conf", false, "[Deprecated]")
	_ = cmd.Flags().MarkDeprecated("skip-resolv-conf", "we never change resolveConf so this flag has no effect, you can stop using it")
	cmd.Flags().BoolVar(&args.StoreFirewalld, "store-firewalld", args.StoreFirewalld, "store the iptables changes with firewalld")
	_ = cmd.Flags().IP("kuma-cp-ip", net.IPv4(0, 0, 0, 0), "[Deprecated]")
	_ = cmd.Flags().MarkDeprecated("kuma-cp-ip", "Running a DNS inside the CP is not possible anymore")
	cmd.Flags().BoolVar(&args.SkipDNSConntrackZoneSplit, "skip-dns-conntrack-zone-split", args.SkipDNSConntrackZoneSplit, "skip applying conntrack zone splitting iptables rules")
	cmd.Flags().BoolVar(&args.ExperimentalTransparentProxyEngine, "experimental-transparent-proxy-engine", args.ExperimentalTransparentProxyEngine, "use experimental transparent proxy engine")

	// ebpf
	cmd.Flags().BoolVar(&args.EbpfEnabled, "ebpf-enabled", args.EbpfEnabled, "use ebpf instead of iptables to install transparent proxy")
	cmd.Flags().StringVar(&args.EbpfProgramsSourcePath, "ebpf-programs-source-path", args.EbpfProgramsSourcePath, "path where compiled ebpf programs and other necessary for ebpf mode files can be found")
	cmd.Flags().StringVar(&args.EbpfInstanceIP, "ebpf-instance-ip", args.EbpfInstanceIP, "IP address of the instance (pod/vm) where transparent proxy will be installed")
	cmd.Flags().StringVar(&args.EbpfBPFFSPath, "ebpf-bpffs-path", args.EbpfBPFFSPath, "the path of the BPF filesystem")

	cmd.Flags().StringArrayVar(&args.ExcludeOutboundTCPPortsForUIDs, "exclude-outbound-tcp-ports-for-uids", []string{}, "tcp outbound ports to exclude for specific UIDs in a format of ports:uids where both ports and uids can be a single value, a list, a range or a combination of all, e.g. 3000-5000:103,104,106-108 would mean exclude ports from 3000 to 5000 for UIDs 103, 104, 106, 107, 108")
	cmd.Flags().StringArrayVar(&args.ExcludeOutboundUDPPortsForUIDs, "exclude-outbound-udp-ports-for-uids", []string{}, "udp outbound ports to exclude for specific UIDs in a format of ports:uids where both ports and uids can be a single value, a list, a range or a combination of all, e.g. 3000-5000:103,104,106-108 would mean exclude ports from 3000 to 5000 for UIDs 103, 104, 106, 107, 108")

	return cmd
}

func findUidGid(uid, user string) (string, string, error) {
	var u *os_user.User
	var err error

	if u, err = os_user.LookupId(uid); err != nil {
		if user != "" {
			if u, err = os_user.Lookup(user); err != nil {
				return "", "", errors.Errorf("--kuma-dp-user or --kuma-dp-uid should refer to a valid user on the host")
			}
		} else {
			u = &os_user.User{
				Uid: uid,
				Gid: uid,
			}
		}
	}

	return u.Uid, u.Gid, nil
}

func configureTransparentProxy(cmd *cobra.Command, args *transparentProxyArgs) error {
	var tp transparentproxy.TransparentProxy
	if !args.ExperimentalTransparentProxyEngine {
		tp = transparentproxy.DefaultTransparentProxy()

		// best effort cleanup before we apply the rules (again?)
		_, err := tp.Cleanup(args.DryRun, args.Verbose)
		if err != nil {
			return errors.Wrapf(err, "unable to invoke cleanup")
		}
	} else {
		tp = &transparentproxy.ExperimentalTransparentProxy{}
	}

	uid, gid, err := findUidGid(args.UID, args.User)
	if err != nil {
		return errors.Wrapf(err, "unable to find the kuma-dp user")
	}

	cfg := &config.TransparentProxyConfig{
		DryRun:                         args.DryRun,
		Verbose:                        args.Verbose,
		RedirectPortOutBound:           args.RedirectPortOutBound,
		RedirectInBound:                args.RedirectInbound,
		RedirectPortInBound:            args.RedirectPortInBound,
		RedirectPortInBoundV6:          args.RedirectPortInBoundV6,
		ExcludeInboundPorts:            args.ExcludeInboundPorts,
		ExcludeOutboundPorts:           args.ExcludeOutboundPorts,
		ExcludeOutboundTCPPortsForUIDs: args.ExcludeOutboundTCPPortsForUIDs,
		ExcludeOutboundUDPPortsForUIDs: args.ExcludeOutboundUDPPortsForUIDs,
		UID:                            uid,
		GID:                            gid,
		RedirectDNS:                    args.RedirectDNS,
		RedirectAllDNSTraffic:          args.RedirectAllDNSTraffic,
		AgentDNSListenerPort:           args.AgentDNSListenerPort,
		DNSUpstreamTargetChain:         args.DNSUpstreamTargetChain,
		SkipDNSConntrackZoneSplit:      args.SkipDNSConntrackZoneSplit,
		EbpfEnabled:                    args.EbpfEnabled,
		EbpfInstanceIP:                 args.EbpfInstanceIP,
		EbpfBPFFSPath:                  args.EbpfBPFFSPath,
		EbpfProgramsSourcePath:         args.EbpfProgramsSourcePath,
		Stdout:                         cmd.OutOrStdout(),
		Stderr:                         cmd.OutOrStderr(),
	}

	output, err := tp.Setup(cfg)
	if err != nil {
		return errors.Wrap(err, "failed to setup transparent proxy")
	}

	if !args.EbpfEnabled && args.StoreFirewalld {
		if _, err := firewalld.NewIptablesTranslator().
			WithDryRun(args.DryRun).
			WithOutput(cmd.OutOrStdout()).
			StoreRules(output); err != nil {
			return err
		}
	}

	return nil
}
