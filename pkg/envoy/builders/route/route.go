package route

import (
	envoy_core "github.com/envoyproxy/go-control-plane/envoy/config/core/v3"
	routev3 "github.com/envoyproxy/go-control-plane/envoy/config/route/v3"
	envoy_wellknown "github.com/envoyproxy/go-control-plane/pkg/wellknown"
	"google.golang.org/protobuf/types/known/structpb"

	"github.com/kumahq/kuma/pkg/envoy/builders/common"
)

func Metadata(key, value string) common.Configurer[routev3.Route] {
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
