package context

import mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"

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

func (a AnyToAnyReachableServicesGraph) CanReach(fromTags map[string]string, toSvc string) bool {
	return true
}

func (a AnyToAnyReachableServicesGraph) CanReachFromAny(fromTagSets []mesh_proto.SingleValueTagSet, to string) bool {
	return true
}

var _ ReachableServicesGraph = AnyToAnyReachableServicesGraph{}

func AnyToAnyReachableServicesGraphBuilder(Resources) ReachableServicesGraph {
	return AnyToAnyReachableServicesGraph{}
}
