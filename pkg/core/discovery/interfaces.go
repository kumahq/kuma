package discovery

import (
	mesh_core "github.com/Kong/kuma/pkg/core/resources/apis/mesh"
	"github.com/Kong/kuma/pkg/core/resources/model"
)

// DiscoverySource is a source of discovery information, i.e. Services and Workloads.
type DiscoverySource interface {
	AddConsumer(DiscoveryConsumer)
}

// DiscoveryConsumer is a consumer of discovery information, i.e. Services and Workloads.
type DiscoveryConsumer interface {
	DataplaneDiscoveryConsumer
}

type DataplaneDiscoveryConsumer interface {
	OnDataplaneUpdate(*mesh_core.DataplaneResource) error
	OnDataplaneDelete(model.ResourceKey) error
}
