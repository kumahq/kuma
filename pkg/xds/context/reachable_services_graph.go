package context

import (
	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
	core_model "github.com/kumahq/kuma/pkg/core/resources/model"
)

// ReachableServicesGraph is a graph of services in the mesh.
// We can test whether the DPP of given tags can reach a service.
// This way we can trim the configuration for a DPP, so it won't include unnecessary configuration.
type ReachableServicesGraph interface {
	CanReach(fromTags map[string]string, toTags map[string]string) bool
	CanReachBackend(fromTags map[string]string, backendIdentifier core_model.TypedResourceIdentifier) bool
}

func CanReachFromAny(graph ReachableServicesGraph, fromTagSets []mesh_proto.SingleValueTagSet, toTags map[string]string) bool {
	for _, fromTags := range fromTagSets {
		if graph.CanReach(fromTags, toTags) {
			return true
		}
	}
	return false
}

func CanReachBackendFromAny(graph ReachableServicesGraph, fromTagSets []mesh_proto.SingleValueTagSet, backendRef core_model.TypedResourceIdentifier) bool {
	for _, fromTags := range fromTagSets {
		if graph.CanReachBackend(fromTags, backendRef) {
			return true
		}
	}
	return false
}

type ReachableServicesGraphBuilder func(meshName string, resources Resources) ReachableServicesGraph

type AnyToAnyReachableServicesGraph struct{}

func (a AnyToAnyReachableServicesGraph) CanReach(map[string]string, map[string]string) bool {
	return true
}

func (a AnyToAnyReachableServicesGraph) CanReachBackend(map[string]string, core_model.TypedResourceIdentifier) bool {
	return true
}

func AnyToAnyReachableServicesGraphBuilder(string, Resources) ReachableServicesGraph {
	return AnyToAnyReachableServicesGraph{}
}
