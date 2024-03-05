package framework

import (
	"net"
	"strconv"
	"strings"

	"github.com/gruntwork-io/terratest/modules/shell"
	"github.com/gruntwork-io/terratest/modules/testing"
	"github.com/pkg/errors"
)

func GetPublishedDockerPorts(
	t testing.TestingT,
	container string,
	ports []uint32,
) (map[uint32]uint32, error) {
	result := map[uint32]uint32{}
	for _, port := range ports {
		cmd := shell.Command{
			Command: "docker",
			Args:    []string{"port", container, strconv.Itoa(int(port))},
		}
		out, err := shell.RunCommandAndGetStdOutE(t, cmd)
		if err != nil {
			return nil, err
		}
		addresses := strings.Split(out, "\n")
		if len(addresses) < 1 {
			return nil, errors.Errorf("there are no addresses for port %d", port)
		}
		addr := addresses[0]
		// on CircleCI, we get the ipv6 address in the format of ":::port",
		// which is not parsable by the "net.SplitHostPort"
		if strings.HasPrefix(addr, ":::") {
			addr = "[::]:" + addr[3:]
		}
		_, pubPortStr, err := net.SplitHostPort(addr)
		if err != nil {
			return nil, err
		}
		pubPort, _ := strconv.ParseInt(pubPortStr, 10, 32)
		result[port] = uint32(pubPort)
	}
	return result, nil
}
