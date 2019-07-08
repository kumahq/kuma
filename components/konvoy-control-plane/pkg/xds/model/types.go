package model

import (
	"fmt"
	core_discovery "github.com/Kong/konvoy/components/konvoy-control-plane/pkg/core/discovery"
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

type WorkloadMeta struct {
	Name      string
	Namespace string
	Labels    map[string]string
}

type Workload struct {
	Meta      WorkloadMeta
	Version   string
	Endpoints []core_discovery.WorkloadEndpoint
}
