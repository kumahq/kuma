//go:build !windows
// +build !windows

package install

import (
	"fmt"
	os_user "os/user"
	"runtime"
	"time"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"

	"github.com/kumahq/kuma/pkg/transparentproxy"
	"github.com/kumahq/kuma/pkg/transparentproxy/config"
	"github.com/kumahq/kuma/pkg/transparentproxy/firewalld"
)

type transparentProxyArgs struct {
	DryRun                         bool
	Verbose                        bool
	RedirectPortOutBound           string
	RedirectInbound                bool
	RedirectPortInBound            string
	RedirectPortInBoundV6          string
	ExcludeInboundPorts            string
	ExcludeOutboundPorts           string
	ExcludeOutboundTCPPortsForUIDs []string
	ExcludeOutboundUDPPortsForUIDs []string
	ExcludeOutboundPortsForUIDs    []string
	UID                            string
	User                           string
	RedirectDNS                    bool
	RedirectAllDNSTraffic          bool
	AgentDNSListenerPort           string
	DNSUpstreamTargetChain         string
	StoreFirewalld                 bool
	SkipDNSConntrackZoneSplit      bool
	EbpfEnabled                    bool
	EbpfProgramsSourcePath         string
	EbpfInstanceIP                 string
	EbpfBPFFSPath                  string
	EbpfCgroupPath                 string
	EbpfTCAttachIface              string
	VnetNetworks                   []string
	Wait                           uint
	WaitInterval                   uint
	MaxRetries                     int
	SleepBetweenRetries            time.Duration
}

func newInstallTransparentProxy() *cobra.Command {
	args := transparentProxyArgs{
		DryRun:                         false,
		Verbose:                        false,
		RedirectPortOutBound:           "15001",
		RedirectInbound:                true,
		RedirectPortInBound:            "15006",
		RedirectPortInBoundV6:          "15010",
		ExcludeInboundPorts:            "",
		ExcludeOutboundPorts:           "",
		ExcludeOutboundTCPPortsForUIDs: []string{},
		ExcludeOutboundUDPPortsForUIDs: []string{},
		UID:                            "",
		User:                           "",
		RedirectDNS:                    false,
		RedirectAllDNSTraffic:          false,
		AgentDNSListenerPort:           "15053",
		DNSUpstreamTargetChain:         "RETURN",
		StoreFirewalld:                 false,
		SkipDNSConntrackZoneSplit:      false,
		EbpfEnabled:                    false,
		EbpfProgramsSourcePath:         "/kuma/ebpf",
		EbpfBPFFSPath:                  "/sys/fs/bpf",
		EbpfCgroupPath:                 "/sys/fs/cgroup",
		EbpfTCAttachIface:              "",
		VnetNetworks:                   []string{},
		Wait:                           5,
		WaitInterval:                   0,
		MaxRetries:                     5,
		SleepBetweenRetries:            2 * time.Second,
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

			if args.EbpfEnabled {
				if args.EbpfInstanceIP == "" {
					return errors.Errorf("--ebpf-instance-ip flag has to be specified --ebpf-enabled is provided")
				}

				if args.StoreFirewalld {
					_, _ = cmd.ErrOrStderr().Write([]byte("# [WARNING] --store-firewalld will be ignored when --ebpf-enabled is being used\n"))
				}

				if args.SkipDNSConntrackZoneSplit {
					_, _ = cmd.ErrOrStderr().Write([]byte("# [WARNING] --skip-dns-conntrack-zone-split will be ignored when --ebpf-enabled is being used\n"))
				}
			}
			// Backward compatibility
			if len(args.ExcludeOutboundPorts) > 0 && (len(args.ExcludeOutboundUDPPortsForUIDs) > 0 || len(args.ExcludeOutboundTCPPortsForUIDs) > 0) {
				return errors.Errorf("--exclude-outbound-ports-for-uids set you can't use --exclude-outbound-tcp-ports-for-uids and --exclude-outbound-udp-ports-for-uids anymore")
			}
			if len(args.ExcludeOutboundTCPPortsForUIDs) > 0 {
				_, _ = cmd.ErrOrStderr().Write([]byte("# [WARNING] flag --exclude-outbound-tcp-ports-for-uids is deprecated use --exclude-outbound-ports-for-uids instead\n"))
				for _, v := range args.ExcludeOutboundTCPPortsForUIDs {
					args.ExcludeOutboundPortsForUIDs = append(args.ExcludeOutboundPortsForUIDs, fmt.Sprintf("tcp:%s", v))
				}
			}
			if len(args.ExcludeOutboundUDPPortsForUIDs) > 0 {
				_, _ = cmd.ErrOrStderr().Write([]byte("# [WARNING] flag --exclude-outbound-udp-ports-for-uids is deprecated use --exclude-outbound-ports-for-uids instead\n"))
				for _, v := range args.ExcludeOutboundUDPPortsForUIDs {
					args.ExcludeOutboundPortsForUIDs = append(args.ExcludeOutboundPortsForUIDs, fmt.Sprintf("udp:%s", v))
				}
			}
			if err := configureTransparentProxy(cmd, &args); err != nil {
				return err
			}

			_, _ = cmd.OutOrStdout().Write([]byte("# Transparent proxy set up successfully, you can now run kuma-dp using transparent-proxy.\n"))
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
	cmd.Flags().StringVar(&args.UID, "kuma-dp-uid", args.UID, "the uid of the user that will run kuma-dp")
	cmd.Flags().BoolVar(&args.RedirectDNS, "redirect-dns", args.RedirectDNS, "redirect only DNS requests targeted to the servers listed in /etc/resolv.conf to a specified port")
	cmd.Flags().BoolVar(&args.RedirectAllDNSTraffic, "redirect-all-dns-traffic", args.RedirectAllDNSTraffic, "redirect all DNS traffic to a specified port, unlike --redirect-dns this will not be limited to the dns servers identified in /etc/resolve.conf")
	cmd.Flags().StringVar(&args.AgentDNSListenerPort, "redirect-dns-port", args.AgentDNSListenerPort, "the port where the DNS agent is listening")
	cmd.Flags().StringVar(&args.DNSUpstreamTargetChain, "redirect-dns-upstream-target-chain", args.DNSUpstreamTargetChain, "(optional) the iptables chain where the upstream DNS requests should be directed to. It is only applied for IP V4. Use with care.")
	cmd.Flags().BoolVar(&args.StoreFirewalld, "store-firewalld", args.StoreFirewalld, "store the iptables changes with firewalld")
	cmd.Flags().BoolVar(&args.SkipDNSConntrackZoneSplit, "skip-dns-conntrack-zone-split", args.SkipDNSConntrackZoneSplit, "skip applying conntrack zone splitting iptables rules")

	// ebpf
	cmd.Flags().BoolVar(&args.EbpfEnabled, "ebpf-enabled", args.EbpfEnabled, "use ebpf instead of iptables to install transparent proxy")
	cmd.Flags().StringVar(&args.EbpfProgramsSourcePath, "ebpf-programs-source-path", args.EbpfProgramsSourcePath, "path where compiled ebpf programs and other necessary for ebpf mode files can be found")
	cmd.Flags().StringVar(&args.EbpfInstanceIP, "ebpf-instance-ip", args.EbpfInstanceIP, "IP address of the instance (pod/vm) where transparent proxy will be installed")
	cmd.Flags().StringVar(&args.EbpfBPFFSPath, "ebpf-bpffs-path", args.EbpfBPFFSPath, "the path of the BPF filesystem")
	cmd.Flags().StringVar(&args.EbpfCgroupPath, "ebpf-cgroup-path", args.EbpfCgroupPath, "the path of cgroup2")
	cmd.Flags().StringVar(&args.EbpfTCAttachIface, "ebpf-tc-attach-iface", args.EbpfTCAttachIface, "name of the interface which TC eBPF programs should be attached to")

	cmd.Flags().StringArrayVar(&args.ExcludeOutboundTCPPortsForUIDs, "exclude-outbound-tcp-ports-for-uids", []string{}, "[DEPRECATED (use --exclude-outbound-ports-for-uids)] tcp outbound ports to exclude for specific uids in a format of ports:uids where ports can be a single value, a list, a range or a combination of all and uid can be a value or a range e.g. 53,3000-5000:106-108 would mean exclude ports 53 and from 3000 to 5000 for uids 106, 107, 108")
	cmd.Flags().StringArrayVar(&args.ExcludeOutboundUDPPortsForUIDs, "exclude-outbound-udp-ports-for-uids", []string{}, "[DEPRECATED (use --exclude-outbound-ports-for-uids)] udp outbound ports to exclude for specific uids in a format of ports:uids where ports can be a single value, a list, a range or a combination of all and uid can be a value or a range e.g. 53, 3000-5000:106-108 would mean exclude ports 53 and from 3000 to 5000 for uids 106, 107, 108")
	cmd.Flags().StringArrayVar(&args.ExcludeOutboundPortsForUIDs, "exclude-outbound-ports-for-uids", []string{}, "outbound ports to exclude for specific uids in a format of protocol:ports:uids where protocol and ports can be omitted or have value tcp or udp and ports can be a single value, a list, a range or a combination of all or * and uid can be a value or a range e.g. 53,3000-5000:106-108 would mean exclude ports 53 and from 3000 to 5000 for both TCP and UDP for uids 106, 107, 108")
	cmd.Flags().StringArrayVar(&args.VnetNetworks, "vnet", []string{}, "virtual networks in a format of interfaceNameRegex:CIDR split by ':' where interface name doesn't have to be exact name e.g. docker0:172.17.0.0/16, br+:172.18.0.0/16, iface:::1/64")
	cmd.Flags().UintVar(&args.Wait, "wait", args.Wait, "specify the amount of time, in seconds, that the application should wait for the xtables exclusive lock before exiting. If the lock is not available within the specified time, the application will exit with an error")
	cmd.Flags().UintVar(&args.WaitInterval, "wait-interval", args.WaitInterval, "flag can be used to specify the amount of time, in microseconds, that iptables should wait between each iteration of the lock acquisition loop. This can be useful if the xtables lock is being held by another application for a long time, and you want to reduce the amount of CPU that iptables uses while waiting for the lock")
	cmd.Flags().IntVar(&args.MaxRetries, "max-retries", args.MaxRetries, "flag can be used to specify the maximum number of times to retry an installation before giving up")
	cmd.Flags().DurationVar(&args.SleepBetweenRetries, "sleep-between-retries", args.SleepBetweenRetries, "flag can be used to specify the amount of time to sleep between retries")

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

	uid, gid, err := findUidGid(args.UID, args.User)
	if err != nil {
		return errors.Wrapf(err, "unable to find the kuma-dp user")
	}

	cfg := &config.TransparentProxyConfig{
		DryRun:                    args.DryRun,
		Verbose:                   args.Verbose,
		RedirectPortOutBound:      args.RedirectPortOutBound,
		RedirectInBound:           args.RedirectInbound,
		RedirectPortInBound:       args.RedirectPortInBound,
		RedirectPortInBoundV6:     args.RedirectPortInBoundV6,
		ExcludeInboundPorts:       args.ExcludeInboundPorts,
		ExcludeOutboundPorts:      args.ExcludeOutboundPorts,
		ExcludedOutboundsForUIDs:  args.ExcludeOutboundPortsForUIDs,
		UID:                       uid,
		GID:                       gid,
		RedirectDNS:               args.RedirectDNS,
		RedirectAllDNSTraffic:     args.RedirectAllDNSTraffic,
		AgentDNSListenerPort:      args.AgentDNSListenerPort,
		DNSUpstreamTargetChain:    args.DNSUpstreamTargetChain,
		SkipDNSConntrackZoneSplit: args.SkipDNSConntrackZoneSplit,
		EbpfEnabled:               args.EbpfEnabled,
		EbpfInstanceIP:            args.EbpfInstanceIP,
		EbpfBPFFSPath:             args.EbpfBPFFSPath,
		EbpfCgroupPath:            args.EbpfCgroupPath,
		EbpfTCAttachIface:         args.EbpfTCAttachIface,
		EbpfProgramsSourcePath:    args.EbpfProgramsSourcePath,
		VnetNetworks:              args.VnetNetworks,
		Stdout:                    cmd.OutOrStdout(),
		Stderr:                    cmd.ErrOrStderr(),
		Wait:                      args.Wait,
		WaitInterval:              args.WaitInterval,
		MaxRetries:                args.MaxRetries,
		SleepBetweenRetries:       args.SleepBetweenRetries,
	}
	tp = transparentproxy.V2()

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
