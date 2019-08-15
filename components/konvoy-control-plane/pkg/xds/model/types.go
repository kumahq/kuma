package model

import (
	"fmt"

	mesh_core "github.com/Kong/konvoy/components/konvoy-control-plane/pkg/core/resources/apis/mesh"
)

type ProxyId struct {
	Name      string
	Namespace string
}

func (id *ProxyId) String() string {
	return fmt.Sprintf("%s.%s", id.Name, id.Namespace)
}

type Proxy struct {
	Id        ProxyId
	Dataplane *mesh_core.DataplaneResource
}
