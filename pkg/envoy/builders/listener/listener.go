package listener

import (
	envoy_accesslog "github.com/envoyproxy/go-control-plane/envoy/config/accesslog/v3"
	envoy_listener "github.com/envoyproxy/go-control-plane/envoy/config/listener/v3"
	routev3 "github.com/envoyproxy/go-control-plane/envoy/config/route/v3"
	envoy_hcm "github.com/envoyproxy/go-control-plane/envoy/extensions/filters/network/http_connection_manager/v3"
	envoy_tcp "github.com/envoyproxy/go-control-plane/envoy/extensions/filters/network/tcp_proxy/v3"

	"github.com/kumahq/kuma/pkg/core/kri"
	. "github.com/kumahq/kuma/pkg/envoy/builders/common"
	listeners_v3 "github.com/kumahq/kuma/pkg/xds/envoy/listeners/v3"
)

func AccessLog(builder *Builder[envoy_accesslog.AccessLog]) Configurer[envoy_listener.Listener] {
	return func(l *envoy_listener.Listener) error {
		accessLog, err := builder.Build()
		if err != nil {
			return err
		}
		for _, chain := range l.FilterChains {
			if err := listeners_v3.UpdateHTTPConnectionManager(chain, func(hcm *envoy_hcm.HttpConnectionManager) error {
				hcm.AccessLog = append(hcm.AccessLog, accessLog)
				return nil
			}); err != nil {
				return err
			}
			if err := listeners_v3.UpdateTCPProxy(chain, func(tcpProxy *envoy_tcp.TcpProxy) error {
				tcpProxy.AccessLog = append(tcpProxy.AccessLog, accessLog)
				return nil
			}); err != nil {
				return err
			}
		}
		return nil
	}
}

func ForEachRoute(fn func(hcm *envoy_hcm.HttpConnectionManager, route *routev3.Route, id kri.Identifier) error) Configurer[envoy_listener.Listener] {
	return func(l *envoy_listener.Listener) error {
		for _, fc := range l.FilterChains {
			err := listeners_v3.UpdateHTTPConnectionManager(fc, func(hcm *envoy_hcm.HttpConnectionManager) error {
				for _, vh := range hcm.GetRouteConfig().VirtualHosts {
					for _, route := range vh.Routes {
						if !kri.IsValid(route.Name) {
							continue
						}

						id, err := kri.FromString(route.Name)
						if err != nil {
							return err
						}
						if err := fn(hcm, route, id); err != nil {
							return err
						}
					}
				}
				return nil
			})
			if err != nil {
				return err
			}
		}
		return nil
	}
}
