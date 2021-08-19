// +build gateway

package register

import (
	core_mesh "github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
	"github.com/kumahq/kuma/pkg/core/resources/registry"
)

// NOTE: this is non-deterministic in testing. Some tests will import
// the plugin and trigger registration and some won't. This means that
// whether the Gateway types are registered in tests depends on which
// subset of tests are running.
func init() {
	// A Gateway is local to a zone, which means that it propagates in one
	// direction, from a zone CP up to a global CP. The reason for this
	// is that the Kubernetes Gateway API is the native Kubernetes API
	// for Kuma gateways. If we propagated a Universal Gateway resource
	// to a Kubernetes zone, we would need to be able to transform Gateway
	// resources from Universal -> Kubernetes and have to deal with namespace
	// semantics and a lot of other unpleasantness.
	registry.RegisterType(core_mesh.GatewayResourceTypeDescriptor)
}
