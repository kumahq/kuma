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

func parseUint16(port string) (uint16, error) {
	parsedPort, err := strconv.ParseUint(port, 10, 16)
	if err != nil {
		return 0, fmt.Errorf("value %s, is not valid uint16", port)
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
		p, err := parseUint16(port)
		if err != nil {
			return nil, err
		}

		result = append(result, p)
	}

	return result, nil
}

func (tp *ExperimentalTransparentProxy) Setup(tpConfig *config.TransparentProxyConfig) (string, error) {
	redirectInboundPort, err := parseUint16(tpConfig.RedirectPortInBound)
	if err != nil {
		return "", errors.Wrap(err, "parsing inbound redirect port failed")
	}

	var redirectInboundPortIPv6 uint16

	if tpConfig.RedirectPortInBoundV6 != "" {
		redirectInboundPortIPv6, err = parseUint16(tpConfig.RedirectPortInBoundV6)
		if err != nil {
			return "", errors.Wrap(err, "parsing inbound redirect port IPv6 failed")
		}
	}

	redirectOutboundPort, err := parseUint16(tpConfig.RedirectPortOutBound)
	if err != nil {
		return "", errors.Wrap(err, "parsing outbound redirect port failed")
	}

	agentDNSListenerPort, err := parseUint16(tpConfig.AgentDNSListenerPort)
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

	var excludePortsForUIDs []kumanet_config.UIDsToPorts
	if len(tpConfig.ExcludeOutboundTCPPortsForUIDs) > 0 {
		excludeTCPPortsForUIDs, err := parseExcludePortsForUIDs(tpConfig.ExcludeOutboundTCPPortsForUIDs, "tcp")
		if err != nil {
			return "", errors.Wrap(err, "parsing excluded outbound TCP ports for UIDs failed")
		}
		excludePortsForUIDs = append(excludePortsForUIDs, excludeTCPPortsForUIDs...)
	}

	if len(tpConfig.ExcludeOutboundUDPPortsForUIDs) > 0 {
		excludeUDPPortsForUIDs, err := parseExcludePortsForUIDs(tpConfig.ExcludeOutboundUDPPortsForUIDs, "udp")
		if err != nil {
			return "", errors.Wrap(err, "parsing excluded outbound UDP ports for UIDs failed")
		}
		excludePortsForUIDs = append(excludePortsForUIDs, excludeUDPPortsForUIDs...)
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
				Enabled:             true,
				Port:                redirectOutboundPort,
				ExcludePorts:        excludeOutboundPorts,
				ExcludePortsForUIDs: excludePortsForUIDs,
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

func parseExcludePortsForUIDs(excludeOutboundPortsForUIDs []string, protocol string) ([]kumanet_config.UIDsToPorts, error) {
	var uidsToPorts []kumanet_config.UIDsToPorts
	for _, excludePort := range excludeOutboundPortsForUIDs {
		parts := strings.Split(excludePort, ":")
		if len(parts) != 2 {
			return nil, errors.New("value contains too many \":\" - format for excluding ports by UIDs ports:uids")
		}
		portValuesOrRange := parts[0]
		uidValuesOrRange := parts[1]

		if err := validatePorts(portValuesOrRange); err != nil {
			return nil, err
		}

		if err := validateUids(uidValuesOrRange); err != nil {
			return nil, err
		}

		uidsToPorts = append(uidsToPorts, kumanet_config.UIDsToPorts{
			Ports:    kumanet_config.ValueOrRangeList(portValuesOrRange),
			UIDs:     kumanet_config.ValueOrRangeList(uidValuesOrRange),
			Protocol: protocol,
		})
	}

	return uidsToPorts, nil
}

func validateUids(uidValuesOrRange string) error {
	return validateUintValueOrRange(uidValuesOrRange)
}

func validatePorts(portValuesOrRange string) error {
	return validateUintValueOrRange(portValuesOrRange)
}

func validateUintValueOrRange(valueOrRange string) error {
	elements := strings.Split(valueOrRange, ",")

	for _, element := range elements {
		portRanges := strings.Split(element, "-")

		for _, port := range portRanges {
			_, err := parseUint16(port)
			if err != nil {
				return errors.Wrapf(err, "values or range %s failed validation", valueOrRange)
			}
		}
	}

	return nil
}

func (tp *ExperimentalTransparentProxy) Cleanup(dryRun, verbose bool) (string, error) {
	// TODO implement me
	panic("implement me")
}
