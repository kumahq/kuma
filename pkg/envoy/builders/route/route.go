package route

import (
	envoy_core "github.com/envoyproxy/go-control-plane/envoy/config/core/v3"
	routev3 "github.com/envoyproxy/go-control-plane/envoy/config/route/v3"
	envoy_wellknown "github.com/envoyproxy/go-control-plane/pkg/wellknown"
	"google.golang.org/protobuf/types/known/structpb"

	. "github.com/kumahq/kuma/pkg/envoy/builders/common"
)

func Metadata(key, value string) Configurer[routev3.Route] {
	return func(r *routev3.Route) error {
		if r.Metadata == nil {
			r.Metadata = &envoy_core.Metadata{}
		}
		if r.Metadata.FilterMetadata == nil {
			r.Metadata.FilterMetadata = map[string]*structpb.Struct{}
		}
		r.Metadata.FilterMetadata[envoy_wellknown.Lua] = &structpb.Struct{
			Fields: map[string]*structpb.Value{
				key: {Kind: &structpb.Value_StringValue{StringValue: value}},
			},
		}
		return nil
	}
}

func HashPolicies(builders []*Builder[routev3.RouteAction_HashPolicy]) Configurer[routev3.Route] {
	return func(r *routev3.Route) error {
		for _, builder := range builders {
			hp, err := builder.Build()
			if err != nil {
				return err
			}
			ra, ok := r.Action.(*routev3.Route_Route)
			if !ok {
				continue
			}
			ra.Route.HashPolicy = append(ra.Route.HashPolicy, hp)
		}
		return nil
	}
}

func AllRoutes(configurer Configurer[routev3.Route]) Configurer[routev3.RouteConfiguration] {
	return func(r *routev3.RouteConfiguration) error {
		for _, vh := range r.VirtualHosts {
			for _, route := range vh.Routes {
				if err := NewModifier(route).Configure(configurer).Modify(); err != nil {
					return err
				}
			}
		}
		return nil
	}
}
