package generator

import (
	"github.com/gogo/protobuf/types"

	model "github.com/Kong/konvoy/components/konvoy-control-plane/pkg/core/xds"
	util_error "github.com/Kong/konvoy/components/konvoy-control-plane/pkg/util/error"
	xds_context "github.com/Kong/konvoy/components/konvoy-control-plane/pkg/xds/context"
	envoy "github.com/envoyproxy/go-control-plane/envoy/api/v2"
	"github.com/envoyproxy/go-control-plane/pkg/cache"
)

type ResourcePayload = cache.Resource

type Resource struct {
	Name     string
	Version  string
	Resource ResourcePayload
}

type ResourceGenerator interface {
	Generate(xds_context.Context, *model.Proxy) ([]*Resource, error)
}

type CompositeResourceGenerator []ResourceGenerator

func (c CompositeResourceGenerator) Generate(ctx xds_context.Context, proxy *model.Proxy) ([]*Resource, error) {
	resources := make([]*Resource, 0)
	for _, gen := range c {
		rs, err := gen.Generate(ctx, proxy)
		if err != nil {
			return nil, err
		}
		resources = append(resources, rs...)
	}
	return resources, nil
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
