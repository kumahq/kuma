package v3

import (
	envoy_listener "github.com/envoyproxy/go-control-plane/envoy/config/listener/v3"
	envoy_http_lua "github.com/envoyproxy/go-control-plane/envoy/extensions/filters/http/lua/v3"
	envoy_hcm "github.com/envoyproxy/go-control-plane/envoy/extensions/filters/network/http_connection_manager/v3"

	"github.com/kumahq/kuma/pkg/util/proto"
	"github.com/kumahq/kuma/pkg/xds/envoy/filters/lua"
)

type ServiceIdentityInjectionConfigurer struct {
}

func (_ *ServiceIdentityInjectionConfigurer) Configure(filterChain *envoy_listener.FilterChain) error {
	script, err := lua.ServiceIdentityInjectionScript()
	if err != nil {
		return err
	}

	config := &envoy_http_lua.Lua{
		InlineCode:  script,
		SourceCodes: nil,
	}

	pbst, err := proto.MarshalAnyDeterministic(config)
	if err != nil {
		return err
	}

	return UpdateHTTPConnectionManager(filterChain, func(manager *envoy_hcm.HttpConnectionManager) error {
		manager.ForwardClientCertDetails = envoy_hcm.HttpConnectionManager_SANITIZE_SET
		manager.SetCurrentClientCertDetails = &envoy_hcm.HttpConnectionManager_SetCurrentClientCertDetails{
			Uri: true,
		}
		manager.HttpFilters = append([]*envoy_hcm.HttpFilter{
			{
				Name: "envoy.filters.http.lua",
				ConfigType: &envoy_hcm.HttpFilter_TypedConfig{
					TypedConfig: pbst,
				},
			},
		}, manager.HttpFilters...)
		return nil
	})
}
