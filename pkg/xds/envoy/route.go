package envoy

import (
	"github.com/golang/protobuf/ptypes/any"

	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
)

type Route struct {
	Match                *mesh_proto.TrafficRoute_Http_Match
	RateLimit            *mesh_proto.RateLimit
	Clusters             []Cluster
	TypedPerFilterConfig map[string]*any.Any
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
