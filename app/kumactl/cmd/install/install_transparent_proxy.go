package install

import (
	"fmt"
	"io/ioutil"
	"net"
	os_user "os/user"
	"runtime"
	"strings"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"

	"github.com/kumahq/kuma/pkg/transparentproxy"
	"github.com/kumahq/kuma/pkg/transparentproxy/config"
)

type transparenProxyArgs struct {
	DryRun               bool
	ModifyIptables       bool
	RedirectPortOutBound string
	RedirectPortInBound  string
	ExcludeInboundPorts  string
	ExcludeOutboundPorts string
	UID                  string
	User                 string
	ModifyResolvConf     bool
	KumaCpIP             net.IP
}

func newInstallTransparentProxy() *cobra.Command {
	args := transparenProxyArgs{
		DryRun:               false,
		ModifyIptables:       true,
		RedirectPortOutBound: "15001",
		RedirectPortInBound:  "15006",
		ExcludeInboundPorts:  "",
		ExcludeOutboundPorts: "",
		UID:                  "",
		User:                 "",
		ModifyResolvConf:     true,
		KumaCpIP:             net.IPv4(0, 0, 0, 0),
	}
	cmd := &cobra.Command{
		Use:   "transparent-proxy",
		Short: "Install Transparent Proxy pre-requisites on the host",
		Long: `Install Transparent Proxy by modifying the hosts iptables and /etc/resolv.conf.

Follow the following steps to use the Kuma data plane proxy in Transparent Proxy mode:

 1) create a dedicated user for the Kuma data plane proxy, e.g. 'kuma-dp'
 2) run this command as a 'root' user to modify the host's iptables and /etc/resolv.conf
    - supply the dedicated username with '--kuma-dp-'
    - all changes are easly revertable by issuing 'kumactl uninstall transparent-proxy'
    - by default the SSH port tcp/22 will not be redirected to Envoy, but everything else will.
      Use '--exclude-inbound-ports' to provide a comma separated list of ports that should also be excluded
    - this command also creates a backup copy of the modified resolv.conf under /etc/resolv.conf
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

			if args.ModifyIptables && args.User == "" && args.UID == "" {
				return errors.Errorf("--kuma-dp-user or --kuma-dp-uid should be supplied")
			}

			if args.ModifyResolvConf && args.KumaCpIP.String() == net.IPv4(0, 0, 0, 0).String() {
				return errors.Errorf("please supply a valid `--kuma-cp-ip`")
			}

			if args.ModifyIptables {
				tp := transparentproxy.GetDefaultTransparentProxy()

				uid, gid, err := findUidGid(args.UID, args.User)
				if err != nil {
					return errors.Wrapf(err, "unable to find the kuma-dp user")
				}

				output, err := tp.Setup(&config.TransparentProxyConfig{
					DryRun:               args.DryRun,
					RedirectPortOutBound: args.RedirectPortOutBound,
					RedirectPortInBound:  args.RedirectPortInBound,
					ExcludeInboundPorts:  args.ExcludeInboundPorts,
					ExcludeOutboundPorts: args.ExcludeOutboundPorts,
					UID:                  uid,
					GID:                  gid,
				})
				if err != nil {
					return errors.Wrap(err, "failed to setup transparent proxy")
				}

				if args.DryRun {
					_, _ = cmd.OutOrStdout().Write([]byte(output))
				}
			}

			if args.ModifyResolvConf {
				kumaCPLine := fmt.Sprintf("nameserver %s", args.KumaCpIP.String())
				content, err := ioutil.ReadFile("/etc/resolv.conf")
				if err != nil {
					return errors.Wrap(err, "unable to open /etc/resolv.conf")
				}
				newcontent := fmt.Sprintf("%s\n%s", kumaCPLine, string(content))

				if !args.DryRun && !strings.Contains(string(content), kumaCPLine) {
					err = ioutil.WriteFile("/etc/resolv.conf.kuma", content, 0644)
					if err != nil {
						return errors.Wrap(err, "unable to open /etc/resolv.conf.kuma")
					}
					err = ioutil.WriteFile("/etc/resolv.conf", []byte(newcontent), 0644)
					if err != nil {
						return errors.Wrap(err, "unable to write /etc/resolv.conf")
					}
				}

				if args.DryRun {
					_, _ = cmd.OutOrStdout().Write([]byte(newcontent))
					_, _ = cmd.OutOrStdout().Write([]byte("\n"))
				} else {
					content, err := ioutil.ReadFile("/etc/resolv.conf")
					if err != nil {
						return errors.Wrap(err, "uanble to open /etc/resolv.conf")
					}
					_, _ = cmd.OutOrStdout().Write(content)
					_, _ = cmd.OutOrStdout().Write([]byte("\n"))
				}
			}

			_, _ = cmd.OutOrStdout().Write([]byte("Transparent proxy set up successfully\n"))
			return nil
		},
	}

	cmd.Flags().BoolVar(&args.DryRun, "dry-run", args.DryRun, "dry run")
	cmd.Flags().BoolVar(&args.ModifyIptables, "modify-iptables", args.ModifyIptables, "modify the host iptables to redirect the traffic to Envoy")
	cmd.Flags().StringVar(&args.RedirectPortOutBound, "redirect-outbound-port", args.RedirectPortOutBound, "outbound port redirected to Envoy, as specified in dataplane's `networking.transparentProxying.redirectPortOutbound`")
	cmd.Flags().StringVar(&args.RedirectPortInBound, "redirect-inbound-port", args.RedirectPortInBound, "inbound port redirected to Envoy, as specified in dataplane's `networking.transparentProxying.redirectPortInbound`")
	cmd.Flags().StringVar(&args.ExcludeInboundPorts, "exclude-inbound-ports", args.ExcludeInboundPorts, "a comma separated list of inbound ports to exclude from redirect to Envoy")
	cmd.Flags().StringVar(&args.ExcludeOutboundPorts, "exclude-outbound-ports", args.ExcludeOutboundPorts, "a comma separated list of outbound ports to exclude from redirect to Envoy")
	cmd.Flags().StringVar(&args.User, "kuma-dp-user", args.UID, "the user that will run kuma-dp")
	cmd.Flags().StringVar(&args.UID, "kuma-dp-uid", args.UID, "the UID of the user that will run kuma-dp")
	cmd.Flags().BoolVar(&args.ModifyResolvConf, "modify-resolv-conf", args.ModifyResolvConf, "modify the host `/etc/resolv.conf` to allow `.mesh` resolution through kuma-cp")
	cmd.Flags().IPVar(&args.KumaCpIP, "kuma-cp-ip", args.KumaCpIP, "the ")

	return cmd
}

func findUidGid(uid, user string) (string, string, error) {
	var u *os_user.User
	var err error

	if u, err = os_user.LookupId(uid); err != nil {
		if u, err = os_user.Lookup(user); err != nil {
			return "", "", errors.Errorf("--kuma-dp-user or --kuma-dp-uid should refer to a valid user on the host")
		}
	}

	return u.Uid, u.Gid, nil
}
