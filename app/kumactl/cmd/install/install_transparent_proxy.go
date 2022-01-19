//go:build !windows
// +build !windows

package install

import (
	"bufio"
	"fmt"
	"net"
	"os"
	os_user "os/user"
	"regexp"
	"runtime"
	"strings"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"

	"github.com/kumahq/kuma/pkg/transparentproxy"
	"github.com/kumahq/kuma/pkg/transparentproxy/config"
	"github.com/kumahq/kuma/pkg/transparentproxy/firewalld"
)

type transparentProxyArgs struct {
	DryRun                    bool
	Verbose                   bool
	RedirectPortOutBound      string
	RedirectInbound           bool
	RedirectPortInBound       string
	RedirectPortInBoundV6     string
	ExcludeInboundPorts       string
	ExcludeOutboundPorts      string
	UID                       string
	User                      string
	RedirectDNS               bool
	RedirectAllDNSTraffic     bool
	AgentDNSListenerPort      string
	DNSUpstreamTargetChain    string
	SkipResolvConf            bool
	StoreFirewalld            bool
	KumaCpIP                  net.IP
	SkipDNSConntrackZoneSplit bool
}

var defaultCpIP = net.IPv4(0, 0, 0, 0)

func newInstallTransparentProxy() *cobra.Command {
	args := transparentProxyArgs{
		DryRun:                    false,
		Verbose:                   false,
		RedirectPortOutBound:      "15001",
		RedirectInbound:           true,
		RedirectPortInBound:       "15006",
		RedirectPortInBoundV6:     "15010",
		ExcludeInboundPorts:       "",
		ExcludeOutboundPorts:      "",
		UID:                       "",
		User:                      "",
		RedirectDNS:               false,
		RedirectAllDNSTraffic:     false,
		AgentDNSListenerPort:      "15053",
		DNSUpstreamTargetChain:    "RETURN",
		SkipResolvConf:            false,
		StoreFirewalld:            false,
		KumaCpIP:                  defaultCpIP,
		SkipDNSConntrackZoneSplit: false,
	}
	cmd := &cobra.Command{
		Use:   "transparent-proxy",
		Short: "Install Transparent Proxy pre-requisites on the host",
		Long: `Install Transparent Proxy by modifying the hosts iptables and /etc/resolv.conf.

Follow the following steps to use the Kuma data plane proxy in Transparent Proxy mode:

 1) create a dedicated user for the Kuma data plane proxy, e.g. 'kuma-dp'
 2) run this command as a 'root' user to modify the host's iptables and /etc/resolv.conf
    - supply the dedicated username with '--kuma-dp-'
    - all changes are easly revertible by issuing 'kumactl uninstall transparent-proxy'
    - by default the SSH port tcp/22 will not be redirected to Envoy, but everything else will.
      Use '--exclude-inbound-ports' to provide a comma separated list of ports that should also be excluded
    - this command also creates a backup copy of the modified resolv.conf under /etc/resolv.conf

 sudo kumactl install transparent-proxy \
          --kuma-dp-user kuma-dp \
          --kuma-cp-ip 10.0.0.1 \
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

 5) make sure the kuma-cp is running its DNS service on port 53 by setting the environment variable 'KUMA_DNS_SERVER_PORT=53'

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

			if args.RedirectDNS && !args.SkipResolvConf {
				return errors.Errorf("please set --skip-resolv-conf when using --redirect-dns or --redirect-all-dns-traffic")
			}

			if !args.SkipResolvConf && args.KumaCpIP.String() == defaultCpIP.String() {
				return errors.Errorf("please supply a valid --kuma-cp-ip")
			}

			if err := modifyIpTables(cmd, &args); err != nil {
				return err
			}

			if !args.SkipResolvConf {
				if err := modifyResolvConf(cmd, &args); err != nil {
					return err
				}
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
	cmd.Flags().BoolVar(&args.RedirectAllDNSTraffic, "redirect-all-dns-traffic", args.RedirectAllDNSTraffic, "redirect all DNS requests to a specified port")
	cmd.Flags().StringVar(&args.AgentDNSListenerPort, "redirect-dns-port", args.AgentDNSListenerPort, "the port where the DNS agent is listening")
	cmd.Flags().StringVar(&args.DNSUpstreamTargetChain, "redirect-dns-upstream-target-chain", args.DNSUpstreamTargetChain, "(optional) the iptables chain where the upstream DNS requests should be directed to. It is only applied for IP V4. Use with care.")
	cmd.Flags().BoolVar(&args.SkipResolvConf, "skip-resolv-conf", args.SkipResolvConf, "skip modifying the host `/etc/resolv.conf`")
	cmd.Flags().BoolVar(&args.StoreFirewalld, "store-firewalld", args.StoreFirewalld, "store the iptables changes with firewalld")
	cmd.Flags().IPVar(&args.KumaCpIP, "kuma-cp-ip", args.KumaCpIP, "the IP address of the Kuma CP which exposes the DNS service on port 53.")
	cmd.Flags().BoolVar(&args.SkipDNSConntrackZoneSplit, "skip-dns-conntrack-zone-split", args.SkipDNSConntrackZoneSplit, "skip applying conntrack zone splitting iptables rules")

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

func modifyIpTables(cmd *cobra.Command, args *transparentProxyArgs) error {
	tp := transparentproxy.DefaultTransparentProxy()

	// best effort cleanup before we apply the rules (again?)
	_, err := tp.Cleanup(args.DryRun, args.Verbose)
	if err != nil {
		return errors.Wrapf(err, "unable to invoke cleanup")
	}

	uid, gid, err := findUidGid(args.UID, args.User)
	if err != nil {
		return errors.Wrapf(err, "unable to find the kuma-dp user")
	}

	if !args.DryRun {
		_, _ = cmd.OutOrStdout().Write([]byte("kumactl is about to apply the iptables rules that will enable transparent proxying on the machine. The SSH connection may drop. If that happens, just reconnect again.\n"))
	}
	output, err := tp.Setup(&config.TransparentProxyConfig{
		DryRun:                    args.DryRun,
		Verbose:                   args.Verbose,
		RedirectPortOutBound:      args.RedirectPortOutBound,
		RedirectInBound:           args.RedirectInbound,
		RedirectPortInBound:       args.RedirectPortInBound,
		RedirectPortInBoundV6:     args.RedirectPortInBoundV6,
		ExcludeInboundPorts:       args.ExcludeInboundPorts,
		ExcludeOutboundPorts:      args.ExcludeOutboundPorts,
		UID:                       uid,
		GID:                       gid,
		RedirectDNS:               args.RedirectDNS,
		RedirectAllDNSTraffic:     args.RedirectAllDNSTraffic,
		AgentDNSListenerPort:      args.AgentDNSListenerPort,
		DNSUpstreamTargetChain:    args.DNSUpstreamTargetChain,
		SkipDNSConntrackZoneSplit: args.SkipDNSConntrackZoneSplit,
	})
	if err != nil {
		return errors.Wrap(err, "failed to setup transparent proxy")
	}

	if args.DryRun {
		_, _ = cmd.OutOrStdout().Write([]byte(output))
	} else {
		_, _ = cmd.OutOrStdout().Write([]byte("iptables set to diverge the traffic to Envoy.\n"))
	}

	if args.StoreFirewalld {
		err = storeFirewalld(cmd, args, output)
		if err != nil {
			return err
		}
	}

	return nil
}

func storeFirewalld(cmd *cobra.Command, args *transparentProxyArgs, output string) error {
	translator := firewalld.NewFirewalldIptablesTranslator(args.DryRun)
	parser := regexp.MustCompile(`\* (?P<table>\w*)`)
	rules := map[string][]string{}

	scanner := bufio.NewScanner(strings.NewReader(output))
	table := ""

	scanner.Split(bufio.ScanLines)
	for scanner.Scan() {
		line := scanner.Text()
		if strings.Contains(line, "COMMIT") {
			table = ""
			continue
		}

		matches := parser.FindStringSubmatch(line)
		if len(matches) > 1 {
			table = matches[parser.SubexpIndex("table")]
			continue
		}

		if table != "" {
			rules[table] = append(rules[table], line)
		}
	}

	translated, err := translator.StoreRules(rules)
	if err != nil {
		return err
	}

	if args.DryRun {
		_, _ = cmd.OutOrStdout().Write([]byte("\n\n" + translated + "\n\n"))
	} else {
		_, _ = cmd.OutOrStdout().Write([]byte("iptables saved with firewalld."))
	}

	return nil
}

func modifyResolvConf(cmd *cobra.Command, args *transparentProxyArgs) error {
	kumaCPLine := fmt.Sprintf("nameserver %s", args.KumaCpIP.String())
	content, err := os.ReadFile("/etc/resolv.conf")
	if err != nil {
		return errors.Wrap(err, "unable to open /etc/resolv.conf")
	}
	newcontent := fmt.Sprintf("%s\n%s", kumaCPLine, string(content))

	if args.DryRun {
		_, _ = cmd.OutOrStdout().Write([]byte(newcontent))
		_, _ = cmd.OutOrStdout().Write([]byte("\n"))
		return nil
	}

	if !strings.Contains(string(content), kumaCPLine) {
		err = os.WriteFile("/etc/resolv.conf.kuma-backup", content, 0644)
		if err != nil {
			return errors.Wrap(err, "unable to open /etc/resolv.conf.kuma-backup")
		}
		err = os.WriteFile("/etc/resolv.conf", []byte(newcontent), 0644)
		if err != nil {
			return errors.Wrap(err, "unable to write /etc/resolv.conf")
		}
	}

	_, _ = cmd.OutOrStdout().Write([]byte("/etc/resolv.conf modified. The previous version of the file backed up to /etc/resolv.conf.kuma-backup"))

	return nil
}
