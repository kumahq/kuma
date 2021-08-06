// +build gateway

package register

import (
	"github.com/kumahq/kuma/app/kumactl/pkg/entities"
	core_mesh "github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
	"github.com/kumahq/kuma/pkg/core/resources/registry"
	kds_definitions "github.com/kumahq/kuma/pkg/kds/definitions"
)

// NOTE: this is non-deterministic in testing. Some tests will import
// the plugin and trigger registration and some won't. This means that
// whether the Gateway types are registered in tests depends on which
// subset of tests are running.
func init() {
	registry.RegisterType(core_mesh.NewGatewayResource())
	registry.RegistryListType(&core_mesh.GatewayResourceList{})

	// A Gateway is local to a zone, which means that it propagates in one
	// direction, from a zone CP up to a global CP. The reason for this
	// is that the Kubernetes Gateway API is the native Kubernetes API
	// for Kuma gateways. If we propagated a Universal Gateway resource
	// to a Kubernetes zone, we would need to be able to transform Gateway
	// resources from Universal -> Kubernetes and have to deal with namespace
	// semantics and a lot of other unpleasantness.

	kds_definitions.All = append(kds_definitions.All,
		kds_definitions.KdsDefinition{
			Type:      core_mesh.GatewayType,
			Direction: kds_definitions.FromZoneToGlobal,
		})

	entities.All = append(entities.All, entities.Definition{
		Singular:     "gateway",
		Plural:       "gateways",
		ResourceType: core_mesh.GatewayType,
	})
}
