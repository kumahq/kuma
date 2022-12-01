package envoy

import (
	"time"

	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
	policies_api "github.com/kumahq/kuma/pkg/plugins/policies/meshratelimit/api/v1alpha1"
	"github.com/kumahq/kuma/pkg/xds/envoy/tags"
)

type Route struct {
	Match     *HttpMatch
	Modify    *mesh_proto.TrafficRoute_Http_Modify
	RateLimit *RateLimitConfiguration
	Tags      []tags.Tags
	Clusters  []Cluster
}

type MatchType uint32

const (
	Prefix MatchType = 0
	Exact            = 1
	Regex            = 2
)

type StringMatcher struct {
	Value     string
	MatchType MatchType
}
type HttpMatch struct {
	Path    *StringMatcher
	Headers map[string]*StringMatcher
	Method  *StringMatcher
}

type RateLimitConfiguration struct {
	Interval    time.Duration
	Requests    uint32
	OnRateLimit *OnRateLimit
}

type OnRateLimit struct {
	Status  uint32
	Headers []*Headers
}

type Headers struct {
	Key    string
	Value  string
	Append bool
}

func NewRouteFromProto(
	match *mesh_proto.TrafficRoute_Http_Match,
	modify *mesh_proto.TrafficRoute_Http_Modify,
	rl *mesh_proto.RateLimit,
	tags []*mesh_proto.Selector,
	cluster []Cluster,
) Route {
	return Route{
		Match:     matchFromProto(match),
		Modify:    modify,
		RateLimit: rateLimitFromProto(rl),
		Tags:      tagsFromProto(tags),
		Clusters:  cluster,
	}
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
			route.Match = &HttpMatch{}
		}
		if route.Match.Headers == nil {
			route.Match.Headers = make(map[string]*StringMatcher)
		}
		route.Match.Headers[name] = &StringMatcher{
			Value:     regex,
			MatchType: Regex,
		}
	})
}

func WithRateLimit(rl *mesh_proto.RateLimit) NewRouteOpt {
	return newRouteOptFunc(func(route *Route) {
		route.RateLimit = rateLimitFromProto(rl)
	})
}

func WithMeshRateLimit(rl *policies_api.LocalHTTP) NewRouteOpt {
	return newRouteOptFunc(func(route *Route) {
		route.RateLimit = rateLimitFromPolicy(rl)
	})
}

func WithMetadataTags(selectors []*mesh_proto.Selector) NewRouteOpt {
	return newRouteOptFunc(func(route *Route) {
		route.Tags = tagsFromProto(selectors)
	})
}

func tagsFromProto(selectors []*mesh_proto.Selector) []tags.Tags {
	tags := make([]tags.Tags, len(selectors))
	for _, selector := range selectors {
		tags = append(tags, selector.Match)
	}
	return tags
}

func rateLimitFromPolicy(rl *policies_api.LocalHTTP) *RateLimitConfiguration {
	rlc := &RateLimitConfiguration{
		Interval: rl.Interval.Duration,
		Requests: rl.Requests,
	}
	headers := []*Headers{}
	if rl.OnRateLimit != nil {
		for _, header := range rl.OnRateLimit.Headers {
			headers = append(headers, &Headers{
				Key:    header.Key,
				Value:  header.Value,
				Append: *header.Append,
			})
		}
		rlc.OnRateLimit = &OnRateLimit{
			Status:  *rl.OnRateLimit.Status,
			Headers: headers,
		}
	}
	return rlc
}

func rateLimitFromProto(rl *mesh_proto.RateLimit) *RateLimitConfiguration {
	headers := []*Headers{}
	for _, header := range rl.GetConf().GetHttp().GetOnRateLimit().GetHeaders() {
		headers = append(headers, &Headers{
			Key:    header.GetKey(),
			Value:  header.GetValue(),
			Append: header.GetAppend().Value,
		})
	}
	return &RateLimitConfiguration{
		Interval: rl.GetConf().GetHttp().GetInterval().AsDuration(),
		Requests: rl.GetConf().GetHttp().GetRequests(),
		OnRateLimit: &OnRateLimit{
			Status:  rl.GetConf().GetHttp().GetOnRateLimit().GetStatus().GetValue(),
			Headers: headers,
		},
	}
}

func matchFromProto(match *mesh_proto.TrafficRoute_Http_Match) *HttpMatch {
	return &HttpMatch{
		Path:    getStringMatcher(match.GetPath()),
		Headers: getHeadersMatcher(match.GetHeaders()),
	}
}

func getStringMatcher(matcher *mesh_proto.TrafficRoute_Http_Match_StringMatcher) *StringMatcher {
	stringMatcher := &StringMatcher{}
	switch matcher.GetMatcherType().(type) {
	case *mesh_proto.TrafficRoute_Http_Match_StringMatcher_Exact:
		stringMatcher.Value = matcher.GetExact()
		stringMatcher.MatchType = Exact
	case *mesh_proto.TrafficRoute_Http_Match_StringMatcher_Prefix:
		stringMatcher.Value = matcher.GetPrefix()
		stringMatcher.MatchType = Prefix
	case *mesh_proto.TrafficRoute_Http_Match_StringMatcher_Regex:
		stringMatcher.Value = matcher.GetRegex()
		stringMatcher.MatchType = Regex
	}
	return stringMatcher
}

func getHeadersMatcher(headers map[string]*mesh_proto.TrafficRoute_Http_Match_StringMatcher) map[string]*StringMatcher {
	headersMatcher := map[string]*StringMatcher{}
	for headerName, header := range headers {
		headersMatcher[headerName] = getStringMatcher(header)
	}
	return headersMatcher
}
