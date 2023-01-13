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

func (r Routes) Clusters() []Cluster {
	var clusters []Cluster
	for _, route := range r {
		clusters = append(clusters, route.Clusters...)
	}
	return clusters
}

type NewRouteOpt interface {
	apply(route *Route)
}

type newRouteOptFunc func(route *Route)

func (f newRouteOptFunc) apply(route *Route) {
	f(route)
}

func NewRoute(opts ...NewRouteOpt) Route {
	r := Route{}
	for _, opt := range opts {
		opt.apply(&r)
	}
	return r
}

func WithCluster(cluster Cluster) NewRouteOpt {
	return newRouteOptFunc(func(route *Route) {
		route.Clusters = append(route.Clusters, cluster)
	})
}

func WithMatchHeaderRegex(name, regex string) NewRouteOpt {
	return newRouteOptFunc(func(route *Route) {
		if route.Match == nil {
			route.Match = &mesh_proto.TrafficRoute_Http_Match{}
		}
		if route.Match.Headers == nil {
			route.Match.Headers = make(map[string]*mesh_proto.TrafficRoute_Http_Match_StringMatcher)
		}
		route.Match.Headers[name] = &mesh_proto.TrafficRoute_Http_Match_StringMatcher{
			MatcherType: &mesh_proto.TrafficRoute_Http_Match_StringMatcher_Regex{
				Regex: regex,
			},
		}
	})
}

func WithRateLimit(rl *mesh_proto.RateLimit) NewRouteOpt {
	return newRouteOptFunc(func(route *Route) {
		route.RateLimit = rl
	})
}
