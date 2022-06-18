package transparentproxy

import (
	"fmt"
	"net"
	"os/exec"
	"strconv"
	"strings"

	"github.com/kumahq/kuma-net/iptables/builder"
	kumanet_config "github.com/kumahq/kuma-net/iptables/config"

	"github.com/kumahq/kuma/pkg/transparentproxy/config"
	"github.com/kumahq/kuma/pkg/transparentproxy/istio/tools/istio-iptables/pkg/constants"
)

var _ TransparentProxy = &ExperimentalTransparentProxy{}

type ExperimentalTransparentProxy struct{}

func splitPorts(ports string) ([]uint16, error) {
	var result []uint16

	for _, port := range strings.Split(ports, ",") {
		p, err := strconv.ParseUint(port, 10, 16)
		if err != nil {
			return nil, fmt.Errorf("port (%s), is not valid uint16: %w", port, err)
		}

		result = append(result, uint16(p))
	}

	return result, nil
}

func hasLocalIPv6() (bool, error) {
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return false, err
	}

	for _, a := range addrs {
		if ipnet, ok := a.(*net.IPNet); ok &&
			!ipnet.IP.IsLoopback() &&
			ipnet.IP.To4() == nil {
			return true, nil
		}
	}

	return false, nil
}

func shouldEnableIPv6() (bool, error) {
	hasIPv6Address, err := hasLocalIPv6()
	if !hasIPv6Address || err != nil {
		return false, err
	}

	// We are executing this command to work around the problem with COS_CONTAINERD
	// image which is being used on GKE nodes. This image is missing "ip6tables_nat"
	// kernel module which is adding `nat` table, so we are checking if this table
	// exists and if so, we are assuming we can safely proceed with ip6tables
	// ref. https://github.com/kumahq/kuma/issues/2046
	err = exec.Command(constants.IP6TABLES, "-t", constants.NAT, "-L").Run()

	return err == nil, nil
}

func (tp *ExperimentalTransparentProxy) Setup(tpConfig *config.TransparentProxyConfig) (string, error) {
	redirectInboundPort, err := strconv.ParseUint(tpConfig.RedirectPortInBound, 10, 16)
	if err != nil {
		return "", fmt.Errorf("inbound redirect port (%s), is not valid uint16: %w",
			tpConfig.RedirectPortInBound, err)
	}

	var redirectInboundPortIPv6 uint64

	if tpConfig.RedirectPortInBoundV6 != "" {
		redirectInboundPortIPv6, err = strconv.ParseUint(tpConfig.RedirectPortInBoundV6, 10, 16)
		if err != nil {
			return "", fmt.Errorf("inbound redirect port IPv6 (%s), is not valid uint16: %w",
				tpConfig.RedirectPortInBound, err)
		}
	}

	redirectOutboundPort, err := strconv.ParseUint(tpConfig.RedirectPortOutBound, 10, 16)
	if err != nil {
		return "", fmt.Errorf("outbound redirect port (%s), is not valid uint16: %w",
			tpConfig.RedirectPortOutBound, err)
	}

	agentDNSListenerPort, err := strconv.ParseUint(tpConfig.AgentDNSListenerPort, 10, 16)
	if err != nil {
		return "", fmt.Errorf("outbound redirect port (%s), is not valid uint16: %w",
			tpConfig.RedirectPortOutBound, err)
	}

	var excludeInboundPorts []uint16
	if tpConfig.ExcludeInboundPorts != "" {
		excludeInboundPorts, err = splitPorts(tpConfig.ExcludeInboundPorts)
		if err != nil {
			return "", fmt.Errorf("cannot parse inbound ports to exclude: %w", err)
		}
	}

	var excludeOutboundPorts []uint16
	if tpConfig.ExcludeOutboundPorts != "" {
		excludeOutboundPorts, err = splitPorts(tpConfig.ExcludeOutboundPorts)
		if err != nil {
			return "", fmt.Errorf("cannot parse outbound ports to exclude: %w", err)
		}
	}

	ipv6, err := shouldEnableIPv6()
	if err != nil {
		return "", fmt.Errorf("cannot verify if IPv6 should be enabled: %w", err)
	}

	cfg := kumanet_config.Config{
		Owner: kumanet_config.Owner{
			UID: tpConfig.UID,
		},
		Redirect: kumanet_config.Redirect{
			NamePrefix: "KUMA_",
			Inbound: kumanet_config.TrafficFlow{
				Port:         uint16(redirectInboundPort),
				PortIPv6:     uint16(redirectInboundPortIPv6),
				ExcludePorts: excludeInboundPorts,
			},
			Outbound: kumanet_config.TrafficFlow{
				Port:         uint16(redirectOutboundPort),
				ExcludePorts: excludeOutboundPorts,
			},
			DNS: kumanet_config.DNS{
				Enabled:            tpConfig.RedirectAllDNSTraffic,
				Port:               uint16(agentDNSListenerPort),
				ConntrackZoneSplit: !tpConfig.SkipDNSConntrackZoneSplit,
			},
		},
		IPv6:    ipv6,
		Verbose: tpConfig.Verbose,
	}

	if tpConfig.DryRun {
		return builder.BuildIPTables(cfg, ipv6)
	}

	return builder.RestoreIPTables(cfg)
}

func (tp *ExperimentalTransparentProxy) Cleanup(dryRun, verbose bool) (string, error) {
	// TODO implement me
	panic("implement me")
}
