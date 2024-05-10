package transparentproxy

import (
	"context"
	"fmt"
	"net"
	"os/exec"
	"strconv"
	"strings"

	"github.com/pkg/errors"

	"github.com/kumahq/kuma/pkg/transparentproxy/config"
	"github.com/kumahq/kuma/pkg/transparentproxy/ebpf"
	"github.com/kumahq/kuma/pkg/transparentproxy/iptables"
)

func HasLocalIPv6() (bool, error) {
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

	hasIPv6Address, err := HasLocalIPv6()
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

func ParseUint16(port string) (uint16, error) {
	parsedPort, err := strconv.ParseUint(port, 10, 16)
	if err != nil {
		return 0, fmt.Errorf("value '%s', is not valid uint16", port)
	}

	return uint16(parsedPort), nil
}

func SplitPorts(ports string) ([]uint16, error) {
	ports = strings.TrimSpace(ports)
	if ports == "" {
		return nil, nil
	}

	var result []uint16

	for _, port := range strings.Split(ports, ",") {
		p, err := ParseUint16(port)
		if err != nil {
			return nil, err
		}

		result = append(result, p)
	}

	return result, nil
}

func ParseExcludePortsForUIDs(excludeOutboundPortsForUIDs []string) ([]config.UIDsToPorts, error) {
	var uidsToPorts []config.UIDsToPorts
	for _, excludePort := range excludeOutboundPortsForUIDs {
		parts := strings.Split(excludePort, ":")
		if len(parts) == 0 || len(parts) > 3 {
			return nil, fmt.Errorf("value: '%s' is invalid - format for excluding ports by uids <protocol:>?<ports:>?uids", excludePort)
		}
		var portValuesOrRange string
		var protocolOpts string
		var uidValuesOrRange string
		switch len(parts) {
		case 1:
			protocolOpts = "*"
			portValuesOrRange = "*"
			uidValuesOrRange = parts[0]
		case 2:
			protocolOpts = "*"
			portValuesOrRange = parts[0]
			uidValuesOrRange = parts[1]
		case 3:
			protocolOpts = parts[0]
			portValuesOrRange = parts[1]
			uidValuesOrRange = parts[2]
		}
		if uidValuesOrRange == "*" {
			return nil, errors.New("can't use wildcard '*' for uids")
		}
		if portValuesOrRange == "*" || portValuesOrRange == "" {
			portValuesOrRange = "1-65535"
		}

		if err := validateUintValueOrRange(portValuesOrRange); err != nil {
			return nil, err
		}

		if strings.Contains(uidValuesOrRange, ",") {
			return nil, fmt.Errorf("uid entry invalid:'%s', it should either be a single item or a range", uidValuesOrRange)
		}
		if err := validateUintValueOrRange(uidValuesOrRange); err != nil {
			return nil, err
		}

		var protocols []string
		if protocolOpts == "" || protocolOpts == "*" {
			protocols = []string{"tcp", "udp"}
		} else {
			for _, p := range strings.Split(protocolOpts, ",") {
				pCleaned := strings.ToLower(strings.TrimSpace(p))
				if pCleaned != "tcp" && pCleaned != "udp" {
					return nil, fmt.Errorf("protocol '%s' is invalid or unsupported", pCleaned)
				}
				protocols = append(protocols, pCleaned)
			}
		}
		for _, p := range protocols {
			uidsToPorts = append(uidsToPorts, config.UIDsToPorts{
				Ports:    config.ValueOrRangeList(strings.ReplaceAll(portValuesOrRange, "-", ":")),
				UIDs:     config.ValueOrRangeList(strings.ReplaceAll(uidValuesOrRange, "-", ":")),
				Protocol: p,
			})
		}
	}

	return uidsToPorts, nil
}

func validateUintValueOrRange(valueOrRange string) error {
	elements := strings.Split(valueOrRange, ",")

	for _, element := range elements {
		portRanges := strings.Split(element, "-")

		for _, port := range portRanges {
			_, err := ParseUint16(port)
			if err != nil {
				return errors.Wrapf(err, "values or range '%s' failed validation", valueOrRange)
			}
		}
	}

	return nil
}

func Setup(ctx context.Context, cfg config.InitializedConfig) (string, error) {
	if cfg.Ebpf.Enabled {
		return ebpf.Setup(cfg)
	}

	return iptables.Setup(ctx, cfg)
}

func Cleanup(cfg config.InitializedConfig) (string, error) {
	if cfg.Ebpf.Enabled {
		return ebpf.Cleanup(cfg)
	}

	return iptables.Cleanup(cfg)
}
