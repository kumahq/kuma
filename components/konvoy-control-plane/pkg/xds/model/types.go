package model

import (
	"fmt"

	mesh_proto "github.com/Kong/konvoy/components/konvoy-control-plane/api/mesh/v1alpha1"
)

type ProxyId struct {
	Name      string
	Namespace string
}

func (id *ProxyId) String() string {
	return fmt.Sprintf("%s.%s", id.Name, id.Namespace)
}

type Proxy struct {
	Id       ProxyId
	Workload Workload
}

type Workload struct {
	Version   string
	Endpoints []mesh_proto.InboundInterface
}
