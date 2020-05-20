package xds

import (
	"github.com/golang/protobuf/proto"
	"github.com/golang/protobuf/ptypes"

	envoy "github.com/envoyproxy/go-control-plane/envoy/api/v2"
	envoy_types "github.com/envoyproxy/go-control-plane/pkg/cache/types"
)

// ResourcePayload is a convenience type alias.
type ResourcePayload = envoy_types.Resource

type NamedResourcePayload interface {
	ResourcePayload
	GetName() string
}

// Resource represents a generic xDS resource with name and version.
type Resource struct {
	Name     string
	Version  string
	Resource ResourcePayload
}

// ResourceList represents a list of generic xDS resources.
type ResourceList []*Resource

func (rs ResourceList) ToDeltaDiscoveryResponse() (*envoy.DeltaDiscoveryResponse, error) {
	resp := &envoy.DeltaDiscoveryResponse{}
	for _, r := range rs {
		pbany, err := ptypes.MarshalAny(r.Resource)
		if err != nil {
			return nil, err
		}
		resp.Resources = append(resp.Resources, &envoy.Resource{
			Name:     r.Name,
			Version:  r.Version,
			Resource: pbany,
		})
	}
	return resp, nil
}

func (rs ResourceList) ToIndex() map[string]ResourcePayload {
	if len(rs) == 0 {
		return nil
	}
	index := make(map[string]ResourcePayload)
	for _, resource := range rs {
		index[resource.Name] = resource.Resource
	}
	return index
}

// ResourceSet represents a set of generic xDS resources.
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

func (s *ResourceSet) AddNamed(namedPayloads ...NamedResourcePayload) *ResourceSet {
	for _, namedPayload := range namedPayloads {
		s.Add(&Resource{
			Name:     namedPayload.GetName(),
			Resource: namedPayload,
		})
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
