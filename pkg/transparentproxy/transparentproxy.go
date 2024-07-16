package transparentproxy

import (
	"context"
	"fmt"
	"net"
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

func Cleanup(ctx context.Context, cfg config.InitializedConfig) error {
	if cfg.IPv4.Ebpf.Enabled {
		_, err := ebpf.Cleanup(cfg.IPv4)
		return err
	}

	return iptables.Cleanup(ctx, cfg)
}
