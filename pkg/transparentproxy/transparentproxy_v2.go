package transparentproxy

import (
	"fmt"
	"net"
	"os/exec"
	"strconv"
	"strings"

	"github.com/pkg/errors"

	"github.com/kumahq/kuma/pkg/transparentproxy/config"
)

var _ TransparentProxy = &TransparentProxyV2{}

type TransparentProxyV2 struct{}

func V2() TransparentProxy {
	return &TransparentProxyV2{}
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

// ShouldEnableIPv6 checks if system supports IPv6. The port has a value of
// RedirectPortInBoundV6 and when equals 0 means that IPv6 was disabled by the user.
func ShouldEnableIPv6(port uint16) (bool, error) {
	if port == 0 {
		return false, nil
	}

	hasIPv6Address, err := hasLocalIPv6()
	if !hasIPv6Address || err != nil {
		return false, err
	}

	// We are executing this command to work around the problem with COS_CONTAINERD
	// image which is being used on GKE nodes. This image is missing "ip6tables_nat"
	// kernel module which is adding `nat` table, so we are checking if this table
	// exists and if so, we are assuming we can safely proceed with ip6tables
	// ref. https://github.com/kumahq/kuma/issues/2046
	err = exec.Command("ip6tables", "-t", "nat", "-L").Run()

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

func (tp *TransparentProxyV2) Setup(tpConfig *config.TransparentProxyConfig) (string, error) {
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

	var excludePortsForUIDs []config.UIDsToPorts
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

	ipv6, err := ShouldEnableIPv6(redirectInboundPortIPv6)
	if err != nil {
		return "", errors.Wrap(err, "cannot verify if IPv6 should be enabled")
	}

	cfg := config.Config{
		Owner: config.Owner{
			UID: tpConfig.UID,
		},
		Redirect: config.Redirect{
			NamePrefix: "KUMA_",
			Inbound: config.TrafficFlow{
				Enabled:      tpConfig.RedirectInBound,
				Port:         redirectInboundPort,
				PortIPv6:     redirectInboundPortIPv6,
				ExcludePorts: excludeInboundPorts,
			},
			Outbound: config.TrafficFlow{
				Enabled:             true,
				Port:                redirectOutboundPort,
				ExcludePorts:        excludeOutboundPorts,
				ExcludePortsForUIDs: excludePortsForUIDs,
			},
			DNS: config.DNS{
				Enabled:             tpConfig.RedirectDNS,
				CaptureAll:          tpConfig.RedirectAllDNSTraffic,
				Port:                agentDNSListenerPort,
				UpstreamTargetChain: tpConfig.DNSUpstreamTargetChain,
				ConntrackZoneSplit:  !tpConfig.SkipDNSConntrackZoneSplit,
			},
			VNet: config.VNet{
				Networks: tpConfig.VnetNetworks,
			},
		},
		Ebpf: config.Ebpf{
			Enabled:            tpConfig.EbpfEnabled,
			InstanceIP:         tpConfig.EbpfInstanceIP,
			BPFFSPath:          tpConfig.EbpfBPFFSPath,
			CgroupPath:         tpConfig.EbpfCgroupPath,
			TCAttachIface:      tpConfig.EbpfTCAttachIface,
			ProgramsSourcePath: tpConfig.EbpfProgramsSourcePath,
		},
		RuntimeStdout: tpConfig.Stdout,
		RuntimeStderr: tpConfig.Stderr,
		IPv6:          ipv6,
		Verbose:       tpConfig.Verbose,
		DryRun:        tpConfig.DryRun,
	}

	return Setup(cfg)
}

func parseExcludePortsForUIDs(excludeOutboundPortsForUIDs []string, protocol string) ([]config.UIDsToPorts, error) {
	var uidsToPorts []config.UIDsToPorts
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

		uidsToPorts = append(uidsToPorts, config.UIDsToPorts{
			Ports:    config.ValueOrRangeList(portValuesOrRange),
			UIDs:     config.ValueOrRangeList(uidValuesOrRange),
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

func (tp *TransparentProxyV2) Cleanup(tpConfig *config.TransparentProxyConfig) (string, error) {
	return Cleanup(config.Config{
		Ebpf: config.Ebpf{
			Enabled:   tpConfig.EbpfEnabled,
			BPFFSPath: tpConfig.EbpfBPFFSPath,
		},
		RuntimeStdout: tpConfig.Stdout,
		RuntimeStderr: tpConfig.Stderr,
		Verbose:       tpConfig.Verbose,
		DryRun:        tpConfig.DryRun,
	})
}
