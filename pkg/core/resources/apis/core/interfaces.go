package core

import (
	core_meta "github.com/kumahq/kuma/pkg/core/metadata"
	core_model "github.com/kumahq/kuma/pkg/core/resources/model"
)

// Port is a common abstraction for a destination port. It provides us basic information about port
type Port interface {
	// GetName returns port name or stringified port value. This can be used when building KRI of a destination. Name cannot be empty
	GetName() string
	// GetValue returns port value
	GetValue() int32
	// GetProtocol return standardized protocol name of a port
	GetProtocol() core_meta.Protocol
}

// Destination interface creates abstraction for Kuma destinations like MeshService, MeshMultiZoneService or MeshExternalService
type Destination interface {
	core_model.Resource

	// GetPorts returns all ports from a destination
	GetPorts() []Port
	// FindPortByName return single port and information if port was found. This method accepts either port name or
	// stringified version of port value. It Can be used to find destination port struct based on destination KRI.
	// This method always checks both port value and port name and treats them equally. This is needed to hande backendRef
	// and reachableBackendRef where you can only specify port value. More on this in issue: https://github.com/kumahq/kuma/issues/11738
	FindPortByName(name string) (Port, bool)
}

// DestinationList is a list wrapper for Destination resources. Implementations should embed
// `core_model.ResourceList` and provide items via `GetDestinations`. All items must be the same
// destination kind, for example all MeshService or all MeshExternalService. The order should
// be stable for a single read to make indexing and hashing predictable
type DestinationList interface {
	core_model.ResourceList
	// GetDestinations returns all Destination resources in the list.
	GetDestinations() []Destination
}
