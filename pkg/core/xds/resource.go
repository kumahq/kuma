package xds

import (
	"sort"

	envoy "github.com/envoyproxy/go-control-plane/envoy/api/v2"
	envoy_types "github.com/envoyproxy/go-control-plane/pkg/cache/types"
	envoy_resource "github.com/envoyproxy/go-control-plane/pkg/resource/v2"
	"github.com/golang/protobuf/proto"
	"github.com/golang/protobuf/ptypes"
)

// ResourcePayload is a convenience type alias.
type ResourcePayload = envoy_types.Resource

// Resource represents a generic xDS resource with name and version.
type Resource struct {
	Name     string
	Origin   string
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

func (rs ResourceList) Payloads() []ResourcePayload {
	var payloads []ResourcePayload
	for _, res := range rs {
		payloads = append(payloads, res.Resource)
	}
	return payloads
}

func (rs ResourceList) Len() int      { return len(rs) }
func (rs ResourceList) Swap(i, j int) { rs[i], rs[j] = rs[j], rs[i] }
func (rs ResourceList) Less(i, j int) bool {
	return rs[i].Name < rs[j].Name
}

// ResourceSet represents a set of generic xDS resources.
type ResourceSet struct {
	// we want to prevent duplicates
	typeToNamesIndex map[string]map[string]*Resource
}

func NewResourceSet() *ResourceSet {
	set := &ResourceSet{}
	set.typeToNamesIndex = map[string]map[string]*Resource{}
	return set
}

func (s *ResourceSet) ListOf(typ string) ResourceList {
	list := ResourceList{}
	for _, resource := range s.typeToNamesIndex[typ] {
		list = append(list, resource)
	}
	sort.Stable(list)
	return list
}

func (s *ResourceSet) Contains(name string, resource ResourcePayload) bool {
	names, ok := s.typeToNamesIndex[s.typeName(resource)]
	if !ok {
		return false
	}
	_, ok = names[name]
	return ok
}

func (s *ResourceSet) Empty() bool {
	for _, resourceMap := range s.typeToNamesIndex {
		if len(resourceMap) != 0 {
			return false
		}
	}
	return true
}

func (s *ResourceSet) Add(resources ...*Resource) *ResourceSet {
	for _, resource := range resources {
		if s.typeToNamesIndex[s.typeName(resource.Resource)] == nil {
			s.typeToNamesIndex[s.typeName(resource.Resource)] = map[string]*Resource{}
		}
		s.typeToNamesIndex[s.typeName(resource.Resource)][resource.Name] = resource
	}
	return s
}

func (s *ResourceSet) Remove(typ string, name string) {
	if s.typeToNamesIndex[typ] != nil {
		delete(s.typeToNamesIndex[typ], name)
	}
}

func (s *ResourceSet) Resources(typ string) map[string]*Resource {
	return s.typeToNamesIndex[typ]
}

func (s *ResourceSet) AddSet(set *ResourceSet) *ResourceSet {
	if set == nil {
		return s
	}
	for typ, resources := range set.typeToNamesIndex {
		if s.typeToNamesIndex[typ] == nil {
			s.typeToNamesIndex[typ] = map[string]*Resource{}
		}
		for name, resource := range resources {
			s.typeToNamesIndex[typ][name] = resource
		}
	}
	return s
}

func (s *ResourceSet) typeName(resource ResourcePayload) string {
	return "type.googleapis.com/" + proto.MessageName(resource)
}

func (s *ResourceSet) List() ResourceList {
	if s == nil {
		return nil
	}
	list := ResourceList{}
	list = append(list, s.ListOf(envoy_resource.EndpointType)...)
	list = append(list, s.ListOf(envoy_resource.ClusterType)...)
	list = append(list, s.ListOf(envoy_resource.RouteType)...)
	list = append(list, s.ListOf(envoy_resource.ListenerType)...)
	list = append(list, s.ListOf(envoy_resource.SecretType)...)
	return list
}
