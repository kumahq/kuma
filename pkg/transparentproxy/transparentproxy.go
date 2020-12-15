package transparentproxy

import "github.com/kumahq/kuma/pkg/transparentproxy/istio"

type TransparentProxy interface {
	// returns the stdout and stderr as string and error if such has occurred
	Setup(dryRun bool, excludeInboundPorts string) (string, error)

	// returns the stdout and stderr as string and error if such has occurred
	Cleanup(dryRun bool) (string, error)
}

func GetDefaultTransparentProxy() TransparentProxy {
	return istio.NewIstioTransparentProxy()
}
