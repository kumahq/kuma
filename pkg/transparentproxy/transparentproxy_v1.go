package transparentproxy

import (
	"github.com/kumahq/kuma/pkg/transparentproxy/istio"
)

func V1() TransparentProxy {
	return &istio.IstioTransparentProxy{}
}
