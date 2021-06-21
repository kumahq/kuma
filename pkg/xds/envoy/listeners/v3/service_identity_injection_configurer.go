package v3

import (
	envoy_listener "github.com/envoyproxy/go-control-plane/envoy/config/listener/v3"
	envoy_http_lua "github.com/envoyproxy/go-control-plane/envoy/extensions/filters/http/lua/v3"
	envoy_hcm "github.com/envoyproxy/go-control-plane/envoy/extensions/filters/network/http_connection_manager/v3"

	"github.com/kumahq/kuma/pkg/util/proto"
)

const serviceIdentityInjectionScript = `
function envoy_on_request(request_handle)
  xfcc = request_handle:headers():get("x-forwarded-client-cert")

  if xfcc == nil or xfcc == '' then
    return
  end

  spiffe = nil
  service = nil

  for str in string.gmatch(xfcc, "([^;]+)") do
    uri_match_result = str:match("URI=(%S+)")
    if uri_match_result ~= nil then
      spiffe = uri_match_result
      mesh = spiffe:match("spiffe://(%S+)/")
      service = spiffe:match("spiffe://" .. mesh .. "/(%S+)")
    end
  end

  if spiffe ~= nil then
    request_handle:headers():add("X-Kuma-Forwarded-Client-Cert", spiffe)
  end

  if service ~= nil then
    request_handle:headers():add("X-Kuma-Forwarded-Client-Service", service)
  end

end

function envoy_on_response(handle)
end
`

type ServiceIdentityInjectionConfigurer struct {
}

func (_ *ServiceIdentityInjectionConfigurer) Configure(filterChain *envoy_listener.FilterChain) error {
	config := &envoy_http_lua.Lua{
		InlineCode:  serviceIdentityInjectionScript,
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
