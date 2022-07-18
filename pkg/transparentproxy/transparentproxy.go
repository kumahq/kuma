package transparentproxy

import (
	"strconv"
	"strings"

	"github.com/pkg/errors"

	"github.com/kumahq/kuma/pkg/transparentproxy/config"
	"github.com/kumahq/kuma/pkg/transparentproxy/istio"
)

type IptablesTranslator interface {
	// store iptables rules
	// accepts a map of slices, the map key is the iptables table
	// and the slices are the list of the iptables rules in that table
	// returns the generated translated rules as a single string
	StoreRules(rules map[string][]string) (string, error)
}

type TransparentProxy interface {
	// returns the stdout and stderr as string and an error if such has occurred
	Setup(cfg *config.TransparentProxyConfig) (string, error)

	// returns the stdout and stderr as string and an error if such has occurred
	Cleanup(dryRun, verbose bool) (string, error)
}

func DefaultTransparentProxy() TransparentProxy {
	return istio.NewIstioTransparentProxy()
}

func ParsePort(port string) (uint16, error) {
	parsedPort, err := strconv.ParseUint(port, 10, 16)
	if err != nil {
		return 0, errors.Wrapf(err, "port (%s), is not valid uint16", port)
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
		p, err := strconv.ParseUint(port, 10, 16)
		if err != nil {
			return nil, errors.Wrapf(err, "port (%s), is not valid uint16", port)
		}

		result = append(result, uint16(p))
	}

	return result, nil
}
