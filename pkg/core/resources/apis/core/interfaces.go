package core

import core_mesh "github.com/kumahq/kuma/pkg/core/resources/apis/mesh"

type Port interface {
	GetName() string
	GetValue() uint32
	GetNameOrStringifyPort() string
	GetProtocol() core_mesh.Protocol
}

type Destination interface {
	GetPorts() []Port
	FindPortByName(name string) (Port, bool)
}
