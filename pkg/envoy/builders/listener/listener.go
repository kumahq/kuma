package listener

import (
	envoy_accesslog "github.com/envoyproxy/go-control-plane/envoy/config/accesslog/v3"
	envoy_listener "github.com/envoyproxy/go-control-plane/envoy/config/listener/v3"
	routev3 "github.com/envoyproxy/go-control-plane/envoy/config/route/v3"
	envoy_hcm "github.com/envoyproxy/go-control-plane/envoy/extensions/filters/network/http_connection_manager/v3"
	envoy_tcp "github.com/envoyproxy/go-control-plane/envoy/extensions/filters/network/tcp_proxy/v3"
	"github.com/envoyproxy/go-control-plane/pkg/wellknown"

	"github.com/kumahq/kuma/pkg/core/kri"
	. "github.com/kumahq/kuma/pkg/envoy/builders/common"
	bldrs_route "github.com/kumahq/kuma/pkg/envoy/builders/route"
	listeners_v3 "github.com/kumahq/kuma/pkg/xds/envoy/listeners/v3"
)

func AccessLogs(builders []*Builder[envoy_accesslog.AccessLog]) Configurer[envoy_listener.Listener] {
	return func(l *envoy_listener.Listener) error {
		accessLogs := []*envoy_accesslog.AccessLog{}
		for _, b := range builders {
			al, err := b.Build()
			if err != nil {
				return err
			}
			accessLogs = append(accessLogs, al)
		}
		for _, chain := range l.FilterChains {
			if err := listeners_v3.UpdateHTTPConnectionManager(chain, func(hcm *envoy_hcm.HttpConnectionManager) error {
				hcm.AccessLog = append(hcm.AccessLog, accessLogs...)
				return nil
			}); err != nil {
				return err
			}
			if err := listeners_v3.UpdateTCPProxy(chain, func(tcpProxy *envoy_tcp.TcpProxy) error {
				tcpProxy.AccessLog = append(tcpProxy.AccessLog, accessLogs...)
				return nil
			}); err != nil {
				return err
			}
		}
		return nil
	}
}

func HCM(hcmConfigurer Configurer[envoy_hcm.HttpConnectionManager]) Configurer[envoy_listener.Listener] {
	return func(l *envoy_listener.Listener) error {
		for _, chain := range l.FilterChains {
			if err := listeners_v3.UpdateHTTPConnectionManager(chain, func(hcm *envoy_hcm.HttpConnectionManager) error {
				return NewModifier(hcm).Configure(hcmConfigurer).Modify()
			}); err != nil {
				return err
			}
		}
		return nil
	}
}

func FilterChains(configurer Configurer[envoy_listener.FilterChain]) Configurer[envoy_listener.Listener] {
	return func(l *envoy_listener.Listener) error {
		for _, fc := range l.FilterChains {
			if err := NewModifier(fc).Configure(configurer).Modify(); err != nil {
				return err
			}
		}
		return nil
	}
}

func RoutesOnFilterChain(configurer Configurer[routev3.Route]) Configurer[envoy_listener.FilterChain] {
	return func(fc *envoy_listener.FilterChain) error {
		return listeners_v3.UpdateHTTPConnectionManager(fc, func(hcm *envoy_hcm.HttpConnectionManager) error {
			return NewModifier(hcm.GetRouteConfig()).
				Configure(bldrs_route.AllRoutes(configurer)).
				Modify()
		})
	}
}

func Routes(configurers map[kri.Identifier]Configurer[routev3.Route]) Configurer[envoy_listener.Listener] {
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

						cfgurer, ok := configurers[id]
						if !ok {
							continue
						}

						if err := NewModifier(route).Configure(cfgurer).Modify(); err != nil {
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

// TraverseRoutes calls visitFn on every HTTPConnectionManager route passing a copy of the Route object.
// The function can be used exclusively for read-only purposed.
func TraverseRoutes(l *envoy_listener.Listener, visitFn func(route *routev3.Route)) error {
	for _, fc := range l.FilterChains {
		for _, filter := range fc.Filters {
			if filter.Name != wellknown.HTTPConnectionManager {
				continue
			}
			var hcm *envoy_hcm.HttpConnectionManager
			if msg, err := filter.GetTypedConfig().UnmarshalNew(); err != nil {
				return err
			} else {
				hcm = msg.(*envoy_hcm.HttpConnectionManager)
			}
			for _, vh := range hcm.GetRouteConfig().VirtualHosts {
				for _, route := range vh.Routes {
					visitFn(route)
				}
			}
		}
	}
	return nil
}
