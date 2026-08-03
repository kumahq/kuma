package envoy

type Route struct {
	Clusters []Cluster
}

func NewRouteFromCluster(cluster Cluster) Route {
	return Route{
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
