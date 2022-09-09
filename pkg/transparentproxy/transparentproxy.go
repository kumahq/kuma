package transparentproxy

import (
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
	return &istio.IstioTransparentProxy{}
}
