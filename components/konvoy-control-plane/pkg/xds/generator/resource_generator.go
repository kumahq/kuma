package generator

import (
	"fmt"
	"github.com/gogo/protobuf/types"

	util_error "github.com/Kong/konvoy/components/konvoy-control-plane/pkg/util/error"
	envoy "github.com/envoyproxy/go-control-plane/envoy/api/v2"
	"github.com/envoyproxy/go-control-plane/pkg/cache"
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
	Addresses []string
	Ports     []uint32
}

type ResourcePayload = cache.Resource

type Resource struct {
	Name     string
	Version  string
	Resource ResourcePayload
}

type ResourceGenerator interface {
	Generate(*Proxy) ([]*Resource, error)
}

type ResourceList []*Resource

func (rs ResourceList) ToDeltaDiscoveryResponse() *envoy.DeltaDiscoveryResponse {
	resp := &envoy.DeltaDiscoveryResponse{}
	for _, r := range rs {
		pbany, err := types.MarshalAny(r.Resource)
		util_error.MustNot(err)
		resp.Resources = append(resp.Resources, envoy.Resource{
			Name:     r.Name,
			Version:  r.Version,
			Resource: pbany,
		})
	}
	return resp
}
