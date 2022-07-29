package transparentproxy

import (
	"fmt"
	"net"
	"os/exec"
	"strconv"
	"strings"

	"github.com/pkg/errors"

	kumanet_tproxy "github.com/kumahq/kuma-net/transparent-proxy"
	kumanet_config "github.com/kumahq/kuma-net/transparent-proxy/config"

	"github.com/kumahq/kuma/pkg/transparentproxy/config"
	"github.com/kumahq/kuma/pkg/transparentproxy/istio/tools/istio-iptables/pkg/constants"
)

var _ TransparentProxy = &ExperimentalTransparentProxy{}

type ExperimentalTransparentProxy struct{}

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

func ShouldEnableIPv6() (bool, error) {
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

func parsePort(port string) (uint16, error) {
	parsedPort, err := strconv.ParseUint(port, 10, 16)
	if err != nil {
		return 0, fmt.Errorf("port %s, is not valid uint16", port)
	}

	return uint16(parsedPort), nil
}

func splitPorts(ports string) ([]uint16, error) {
	ports = strings.TrimSpace(ports)
	if ports == "" {
		return nil, nil
	}

	var result []uint16

	for _, port := range strings.Split(ports, ",") {
		p, err := parsePort(port)
		if err != nil {
			return nil, err
		}

		result = append(result, p)
	}

	return result, nil
}

func (tp *ExperimentalTransparentProxy) Setup(tpConfig *config.TransparentProxyConfig) (string, error) {
	redirectInboundPort, err := parsePort(tpConfig.RedirectPortInBound)
	if err != nil {
		return "", errors.Wrap(err, "parsing inbound redirect port failed")
	}

	var redirectInboundPortIPv6 uint16

	if tpConfig.RedirectPortInBoundV6 != "" {
		redirectInboundPortIPv6, err = parsePort(tpConfig.RedirectPortInBoundV6)
		if err != nil {
			return "", errors.Wrap(err, "parsing inbound redirect port IPv6 failed")
		}
	}

	redirectOutboundPort, err := parsePort(tpConfig.RedirectPortOutBound)
	if err != nil {
		return "", errors.Wrap(err, "parsing outbound redirect port failed")
	}

	agentDNSListenerPort, err := parsePort(tpConfig.AgentDNSListenerPort)
	if err != nil {
		return "", errors.Wrap(err, "parsing agent DNS listener port failed")
	}

	var excludeInboundPorts []uint16
	if tpConfig.ExcludeInboundPorts != "" {
		excludeInboundPorts, err = splitPorts(tpConfig.ExcludeInboundPorts)
		if err != nil {
			return "", errors.Wrap(err, "cannot parse inbound ports to exclude")
		}
	}

	var excludeOutboundPorts []uint16
	if tpConfig.ExcludeOutboundPorts != "" {
		excludeOutboundPorts, err = splitPorts(tpConfig.ExcludeOutboundPorts)
		if err != nil {
			return "", errors.Wrap(err, "cannot parse outbound ports to exclude")
		}
	}

	ipv6, err := ShouldEnableIPv6()
	if err != nil {
		return "", errors.Wrap(err, "cannot verify if IPv6 should be enabled")
	}

	cfg := kumanet_config.Config{
		Owner: kumanet_config.Owner{
			UID: tpConfig.UID,
		},
		Redirect: kumanet_config.Redirect{
			NamePrefix: "KUMA_",
			Inbound: kumanet_config.TrafficFlow{
				Enabled:      tpConfig.RedirectInBound,
				Port:         redirectInboundPort,
				PortIPv6:     redirectInboundPortIPv6,
				ExcludePorts: excludeInboundPorts,
			},
			Outbound: kumanet_config.TrafficFlow{
				Enabled:      true,
				Port:         redirectOutboundPort,
				ExcludePorts: excludeOutboundPorts,
			},
			DNS: kumanet_config.DNS{
				Enabled:            tpConfig.RedirectAllDNSTraffic,
				Port:               agentDNSListenerPort,
				ConntrackZoneSplit: !tpConfig.SkipDNSConntrackZoneSplit,
			},
		},
		Ebpf: kumanet_config.Ebpf{
			Enabled:            tpConfig.EbpfEnabled,
			InstanceIP:         tpConfig.EbpfInstanceIP,
			BPFFSPath:          tpConfig.EbpfBPFFSPath,
			ProgramsSourcePath: tpConfig.EbpfProgramsSourcePath,
		},
		RuntimeStdout: tpConfig.Stdout,
		RuntimeStderr: tpConfig.Stderr,
		IPv6:          ipv6,
		Verbose:       tpConfig.Verbose,
		DryRun:        tpConfig.DryRun,
	}

	return kumanet_tproxy.Setup(cfg)
}

func (tp *ExperimentalTransparentProxy) Cleanup(dryRun, verbose bool) (string, error) {
	// TODO implement me
	panic("implement me")
}
