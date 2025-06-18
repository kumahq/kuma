package route

import (
	"github.com/kumahq/kuma/pkg/envoy/builders/common"
	envoy_route "github.com/envoyproxy/go-control-plane/envoy/config/route/v3"
	"time"
	"google.golang.org/protobuf/types/known/durationpb"
)

func HashPolicy() *common.Builder[envoy_route.RouteAction_HashPolicy] {
	return &common.Builder[envoy_route.RouteAction_HashPolicy]{}
}

func Terminal(terminal bool) common.Configurer[envoy_route.RouteAction_HashPolicy] {
	return func(r *envoy_route.RouteAction_HashPolicy) error {
		r.Terminal = terminal
		return nil
	}
}

func HeaderPolicySpecifier(name string) common.Configurer[envoy_route.RouteAction_HashPolicy] {
	return func(r *envoy_route.RouteAction_HashPolicy) error {
		r.PolicySpecifier = &envoy_route.RouteAction_HashPolicy_Header_{
			Header: &envoy_route.RouteAction_HashPolicy_Header{
				HeaderName: name,
			},
		}
		return nil
	}
}

func CookiePolicySpecifier(name, path string, ttl *time.Duration) common.Configurer[envoy_route.RouteAction_HashPolicy] {
	return func(r *envoy_route.RouteAction_HashPolicy) error {
		var ttlPb *durationpb.Duration
		if ttl != nil {
			ttlPb = durationpb.New(*ttl)
		}
		r.PolicySpecifier = &envoy_route.RouteAction_HashPolicy_Cookie_{
			Cookie: &envoy_route.RouteAction_HashPolicy_Cookie{
				Name: name,
				Ttl:  ttlPb,
				Path: path,
			},
		}
		return nil
	}
}

func ConnectionTypePolicySpecifier(sourceIP bool) common.Configurer[envoy_route.RouteAction_HashPolicy] {
	return func(r *envoy_route.RouteAction_HashPolicy) error {
		r.PolicySpecifier = &envoy_route.RouteAction_HashPolicy_ConnectionProperties_{
			ConnectionProperties: &envoy_route.RouteAction_HashPolicy_ConnectionProperties{
				SourceIp: sourceIP,
			},
		}
		return nil
	}
}

func QueryPolicySpecifier(name string) common.Configurer[envoy_route.RouteAction_HashPolicy] {
	return func(r *envoy_route.RouteAction_HashPolicy) error {
		r.PolicySpecifier = &envoy_route.RouteAction_HashPolicy_QueryParameter_{
			QueryParameter: &envoy_route.RouteAction_HashPolicy_QueryParameter{
				Name: name,
			},
		}
		return nil
	}
}

func FilterStatePolicySpecifier(key string) common.Configurer[envoy_route.RouteAction_HashPolicy] {
	return func(r *envoy_route.RouteAction_HashPolicy) error {
		r.PolicySpecifier = &envoy_route.RouteAction_HashPolicy_FilterState_{
			FilterState: &envoy_route.RouteAction_HashPolicy_FilterState{
				Key: key,
			},
		}
		return nil
	}
}
