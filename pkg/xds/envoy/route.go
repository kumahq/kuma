package envoy

import (
	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
)

type Route struct {
	Match    *mesh_proto.TrafficRoute_Http_Match
	Clusters []Cluster
}

func NewRouteFromCluster(cluster Cluster) Route {
	return Route{
		Match:    nil,
		Clusters: []Cluster{cluster},
	}
}

type Routes []Route

func (r Routes) Clusters() (clusters []Cluster) {
	for _, route := range r {
		clusters = append(clusters, route.Clusters...)
	}
	return
}
