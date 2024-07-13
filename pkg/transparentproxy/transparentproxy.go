package transparentproxy

import (
	"context"
	"fmt"
	"net"
	"os/exec"
	"strconv"
	"strings"

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

func Setup(ctx context.Context, cfg config.InitializedConfig) (string, error) {
	if cfg.IPv4.Ebpf.Enabled {
		return ebpf.Setup(cfg.IPv4)
	}

	return iptables.Setup(ctx, cfg)
}

func Cleanup(ctx context.Context, cfg config.InitializedConfig) (string, error) {
	if cfg.IPv4.Ebpf.Enabled {
		return ebpf.Cleanup(cfg.IPv4)
	}

	return "", iptables.Cleanup(ctx, cfg)
}
