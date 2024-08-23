package xds

import (
	"sort"

	envoy_sd "github.com/envoyproxy/go-control-plane/envoy/service/discovery/v3"
	envoy_types "github.com/envoyproxy/go-control-plane/pkg/cache/types"
	protov1 "github.com/golang/protobuf/proto"
	"google.golang.org/protobuf/types/known/anypb"

	core_mesh "github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
	core_model "github.com/kumahq/kuma/pkg/core/resources/model"
)

// ResourcePayload is a convenience type alias.
type ResourcePayload = envoy_types.Resource

// Resource represents a generic xDS resource with name and version.
type Resource struct {
	Name           string
	Origin         string
	Resource       ResourcePayload
	ResourceOrigin *core_model.TypedResourceIdentifier
	Protocol       core_mesh.Protocol
}

// ResourceList represents a list of generic xDS resources.
type ResourceList []*Resource

func (rs ResourceList) ToDeltaDiscoveryResponse() (*envoy_sd.DeltaDiscoveryResponse, error) {
	resp := &envoy_sd.DeltaDiscoveryResponse{}
	for _, r := range rs {
		pbany, err := anypb.New(protov1.MessageV2(r.Resource))
		if err != nil {
			return nil, err
		}
		resp.Resources = append(resp.Resources, &envoy_sd.Resource{
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

// ResourceTypes returns names of all the distinct resource types in the set.
func (s *ResourceSet) ResourceTypes() []string {
	var typeNames []string

	for typeName := range s.typeToNamesIndex {
		typeNames = append(typeNames, typeName)
	}

	return typeNames
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
	return "type.googleapis.com/" + string(protov1.MessageV2(resource).ProtoReflect().Descriptor().FullName())
}

func (s *ResourceSet) List() ResourceList {
	if s == nil {
		return nil
	}

	types := s.ResourceTypes()
	list := ResourceList{}

	sort.Strings(types) // Deterministic for test output.

	for _, name := range types {
		list = append(list, s.ListOf(name)...)
	}

	return list
}

func (s *ResourceSet) IndexByOrigin() map[core_model.TypedResourceIdentifier]map[string][]*Resource {
	byOwner := map[core_model.TypedResourceIdentifier]map[string][]*Resource{}
	for typ, nameToRes := range s.typeToNamesIndex {
		for _, resource := range nameToRes {
			if resource.ResourceOrigin == nil {
				continue
			}
			resOwner := *resource.ResourceOrigin
			if byOwner[resOwner] == nil {
				byOwner[resOwner] = map[string][]*Resource{}
			}
			byOwner[resOwner][typ] = append(byOwner[resOwner][typ], resource)
		}
	}
	return byOwner
}
