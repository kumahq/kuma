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
		Long:  `Install Transparent Proxy by modifying the hosts iptables and /etc/resolv.conf`,
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
					fmt.Println(output)
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
					fmt.Println(newcontent)
				} else {
					content, err := ioutil.ReadFile("/etc/resolv.conf")
					if err != nil {
						return errors.Wrap(err, "uanble to open /etc/resolv.conf")
					}
					fmt.Println(string(content))
				}
			}

			fmt.Println("Transparent proxy set up successfully")
			return nil
		},
	}

	cmd.Flags().BoolVar(&args.DryRun, "dry-run", args.DryRun, "dry run")
	cmd.Flags().BoolVar(&args.ModifyIptables, "modify-iptables", args.ModifyIptables, "modify the host iptables to redirect the traffic to Envoy")
	cmd.Flags().StringVar(&args.RedirectPortOutBound, "redirect-outbound-port", args.RedirectPortOutBound, "outbound port redirected to Envoy, as specified in dataplane's `networking.transparentProxying.redirectPortOutbound`")
	cmd.Flags().StringVar(&args.RedirectPortInBound, "redirect-inbound-port", args.RedirectPortInBound, "inbound port redirected to Envoy, as specified in dataplane's `networking.transparentProxying.redirectPortInbound`")
	cmd.Flags().StringVar(&args.ExcludeInboundPorts, "exclude-inbound-ports", args.ExcludeInboundPorts, "inbound ports to exclude from redirect to Envoy")
	cmd.Flags().StringVar(&args.ExcludeOutboundPorts, "exclude-outbound-ports", args.ExcludeOutboundPorts, "outbound ports to exclude from redirect to Envoy")
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
