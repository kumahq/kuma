package hcm

import (
	envoy_accesslog "github.com/envoyproxy/go-control-plane/envoy/config/accesslog/v3"
	envoy_hcm "github.com/envoyproxy/go-control-plane/envoy/extensions/filters/network/http_connection_manager/v3"
	. "github.com/kumahq/kuma/pkg/envoy/builders/common"
	util_proto "github.com/kumahq/kuma/pkg/util/proto"
	envoy_core "github.com/envoyproxy/go-control-plane/envoy/config/core/v3"
	envoy_wellknown "github.com/envoyproxy/go-control-plane/pkg/wellknown"
	luav3 "github.com/envoyproxy/go-control-plane/envoy/extensions/filters/http/lua/v3"
)

func AccessLog(builder *Builder[envoy_accesslog.AccessLog]) Configurer[envoy_hcm.HttpConnectionManager] {
	return func(hcm *envoy_hcm.HttpConnectionManager) error {
		accessLog, err := builder.Build()
		if err != nil {
			return err
		}
		hcm.AccessLog = append(hcm.AccessLog, accessLog)
		return nil
	}
}

func LuaFilterAddFirst(code string) Configurer[envoy_hcm.HttpConnectionManager] {
	return func(hcm *envoy_hcm.HttpConnectionManager) error {
		marshaled, err := util_proto.MarshalAnyDeterministic(&luav3.Lua{
			DefaultSourceCode: &envoy_core.DataSource{
				Specifier: &envoy_core.DataSource_InlineString{
					InlineString: code,
				},
			},
		})
		if err != nil {
			return err
		}
		hcm.HttpFilters = append([]*envoy_hcm.HttpFilter{{
			Name: envoy_wellknown.Lua,
			ConfigType: &envoy_hcm.HttpFilter_TypedConfig{
				TypedConfig: marshaled,
			},
		}}, hcm.HttpFilters...)
		return nil
	}
}
