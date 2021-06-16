package envoy

import (
	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
)

type Route struct {
	Match     *mesh_proto.TrafficRoute_Http_Match
	Modify    *mesh_proto.TrafficRoute_Http_Modify
	RateLimit *mesh_proto.RateLimit
	Clusters  []Cluster
}

func NewRouteFromCluster(cluster Cluster) Route {
	return Route{
		Match:    nil,
		Modify:   nil,
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
