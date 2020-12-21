package transparentproxy

import (
	"github.com/kumahq/kuma/pkg/transparentproxy/config"
	"github.com/kumahq/kuma/pkg/transparentproxy/istio"
)

type TransparentProxy interface {
	// returns the stdout and stderr as string and error if such has occurred
	Setup(cfg *config.TransparentProxyConfig) (string, error)

	// returns the stdout and stderr as string and error if such has occurred
	Cleanup(dryRun bool) (string, error)
}

func DefaultTransparentProxy() TransparentProxy {
	return istio.NewIstioTransparentProxy()
}
