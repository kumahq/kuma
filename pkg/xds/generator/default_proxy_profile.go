package generator

import (
	core_generator "github.com/kumahq/kuma/v3/pkg/core/resources/apis/core/generator"
	policies_generator "github.com/kumahq/kuma/v3/pkg/plugins/policies/core/generator"
	"github.com/kumahq/kuma/v3/pkg/xds/generator/core"
)

// NewDefaultProxyProfile returns the generators that build the Envoy
// configuration of a data plane proxy.
func NewDefaultProxyProfile() core.ResourceGenerator {
	return core.CompositeResourceGenerator{
		AdminProxyGenerator{},
		TransparentProxyGenerator{},
		InboundProxyGenerator{},
		DirectAccessProxyGenerator{},
		DNSGenerator{},
		ZoneProxyListenerGenerator{},
		policies_generator.NewGenerator(),
		core_generator.NewGenerator(),
	}
}
