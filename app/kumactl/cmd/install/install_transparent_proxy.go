package install

import (
	"fmt"
	os_user "os/user"

	"github.com/kumahq/kuma/pkg/transparentproxy/config"

	"runtime"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"

	"github.com/kumahq/kuma/pkg/transparentproxy"
)

type transparenProxyArgs struct {
	DryRun               bool
	RedirectPortOutBound string
	RedirectPortInBound  string
	ExcludeInboundPorts  string
	ExcludeOutboundPorts string
	UID                  string
	User                 string
}

func newInstallTransparentProxy() *cobra.Command {
	args := transparenProxyArgs{
		DryRun:               false,
		RedirectPortOutBound: "15001",
		RedirectPortInBound:  "15006",
		ExcludeInboundPorts:  "",
		ExcludeOutboundPorts: "",
		UID:                  "",
		User:                 "",
	}
	cmd := &cobra.Command{
		Use:   "transparent-proxy",
		Short: "Install Transparent Proxy pre-requisites on the host",
		Long:  `Install Transparent Proxy by modifying the hosts iptables and /etc/resolv.conf`,
		RunE: func(cmd *cobra.Command, _ []string) error {
			if !args.DryRun && runtime.GOOS != "linux" {
				return errors.Errorf("transparent proxy will work only on Linux OSes")
			}

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
			} else {
				fmt.Println("Transparent proxy set up successfully")
			}

			return nil
		},
	}

	cmd.Flags().BoolVar(&args.DryRun, "dry-run", args.DryRun, "dry run")
	cmd.Flags().StringVar(&args.RedirectPortOutBound, "redirect-outbound-port", args.RedirectPortOutBound, "outbound port redirected to Envoy, as specified in dataplane's `networking.transparentProxying.redirectPortOutbound`")
	cmd.Flags().StringVar(&args.RedirectPortInBound, "redirect-inbound-port", args.RedirectPortInBound, "inbound port redirected to Envoy, as specified in dataplane's `networking.transparentProxying.redirectPortInbound`")
	cmd.Flags().StringVar(&args.ExcludeInboundPorts, "exclude-inbound-ports", args.ExcludeInboundPorts, "inbound ports to exclude from redirect to Envoy")
	cmd.Flags().StringVar(&args.ExcludeOutboundPorts, "exclude-outbound-ports", args.ExcludeOutboundPorts, "outbound ports to exclude from redirect to Envoy")
	cmd.Flags().StringVar(&args.User, "kuma-dp-user", args.UID, "the user that will run kuma-dp")
	cmd.Flags().StringVar(&args.UID, "kuma-dp-uid", args.UID, "the UID of the user that will run kuma-dp")

	return cmd
}

func findUidGid(uid, user string) (string, string, error) {
	var u *os_user.User
	var err error

	if u, err = os_user.LookupId(uid); err != nil {
		if u, err = os_user.Lookup(user); err != nil {
			return "", "", errors.Errorf("--kuma-dp-user or --kuma-dp-uid should be set on the command line")
		}
	}

	return u.Uid, u.Gid, nil
}
