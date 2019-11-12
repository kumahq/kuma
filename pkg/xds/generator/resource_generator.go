package generator

import (
	"github.com/golang/protobuf/proto"
	"github.com/golang/protobuf/ptypes"

	model "github.com/Kong/kuma/pkg/core/xds"
	util_error "github.com/Kong/kuma/pkg/util/error"
	xds_context "github.com/Kong/kuma/pkg/xds/context"
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
		pbany, err := ptypes.MarshalAny(r.Resource)
		util_error.MustNot(err)
		resp.Resources = append(resp.Resources, &envoy.Resource{
			Name:     r.Name,
			Version:  r.Version,
			Resource: pbany,
		})
	}
	return resp
}

type ResourceSet struct {
	// we want to keep resources in the order they were added
	resources []*Resource
	// we want to prevent duplicates
	typeToNamesIndex map[string]map[string]bool
}

func (s *ResourceSet) Contains(name string, resource ResourcePayload) bool {
	names, ok := s.typeToNamesIndex[s.typeName(resource)]
	if !ok {
		return false
	}
	_, ok = names[name]
	return ok
}

func (s *ResourceSet) Add(resources ...*Resource) *ResourceSet {
	for _, resource := range resources {
		if s.Contains(resource.Name, resource.Resource) {
			continue
		}
		s.resources = append(s.resources, resource)
		s.index(resource)
	}
	return s
}

func (s *ResourceSet) typeName(resource ResourcePayload) string {
	return proto.MessageName(resource)
}

func (s *ResourceSet) index(resource *Resource) {
	if s.typeToNamesIndex == nil {
		s.typeToNamesIndex = map[string]map[string]bool{}
	}
	typeName := s.typeName(resource.Resource)
	if s.typeToNamesIndex[typeName] == nil {
		s.typeToNamesIndex[typeName] = map[string]bool{}
	}
	s.typeToNamesIndex[typeName][resource.Name] = true
}

func (s *ResourceSet) List() []*Resource {
	return s.resources
}
