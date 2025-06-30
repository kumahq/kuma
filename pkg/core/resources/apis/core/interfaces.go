package core

import core_mesh "github.com/kumahq/kuma/pkg/core/resources/apis/mesh"

// Port is a common abstraction for a destination port. It provides us basic information about port
type Port interface {
	// GetName returns port name or stringified port value. This can be used when building KRI of a destination. Name cannot be empty
	GetName() string
	// GetValue returns port value
	GetValue() int32
	// GetProtocol return standardized protocol name of a port
	GetProtocol() core_mesh.Protocol
}

// Destination interface creates abstraction for Kuma destinations like MeshService, MeshMultiZoneService or MeshExternalService
type Destination interface {
	// GetPorts returns all ports from a destination
	GetPorts() []Port
	// FindPortByName return single port and information if port was found. This method accepts either port name or
	// stringified version of port value. Can be used to find destination port struct based on destination KRI
	FindPortByName(name string) (Port, bool)
}
