package context

import mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"

// ReachableServicesGraph is a graph of services in the mesh.
// We can test whether the DPP of given tags can reach a service.
// This way we can trim the configuration for a DPP, so it won't include unnecessary configuration.
type ReachableServicesGraph interface {
	CanReach(fromTags map[string]string, toSvc string) bool
}

func CanReachFromAny(graph ReachableServicesGraph, fromTagSets []mesh_proto.SingleValueTagSet, to string) bool {
	for _, from := range fromTagSets {
		if graph.CanReach(from, to) {
			return true
		}
	}
	return false
}

type ReachableServicesGraphBuilder func(Resources) ReachableServicesGraph

type AnyToAnyReachableServicesGraph struct {
}

func (a AnyToAnyReachableServicesGraph) CanReach(map[string]string, string) bool {
	return true
}

func AnyToAnyReachableServicesGraphBuilder(Resources) ReachableServicesGraph {
	return AnyToAnyReachableServicesGraph{}
}
