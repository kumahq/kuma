package xds

import (
	"time"

	envoy_cluster "github.com/envoyproxy/go-control-plane/envoy/config/cluster/v3"
	envoy_core "github.com/envoyproxy/go-control-plane/envoy/config/core/v3"
	envoy_listener "github.com/envoyproxy/go-control-plane/envoy/config/listener/v3"
	envoy_route "github.com/envoyproxy/go-control-plane/envoy/config/route/v3"
	envoy_hcm "github.com/envoyproxy/go-control-plane/envoy/extensions/filters/network/http_connection_manager/v3"
	envoy_tcp "github.com/envoyproxy/go-control-plane/envoy/extensions/filters/network/tcp_proxy/v3"
	envoy_upstream_http "github.com/envoyproxy/go-control-plane/envoy/extensions/upstreams/http/v3"
	"github.com/pkg/errors"
	"google.golang.org/protobuf/types/known/durationpb"
	kube_meta "k8s.io/apimachinery/pkg/apis/meta/v1"

	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
	core_mesh "github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
	policies_defaults "github.com/kumahq/kuma/pkg/plugins/policies/core/defaults"
	"github.com/kumahq/kuma/pkg/plugins/policies/core/rules"
	api "github.com/kumahq/kuma/pkg/plugins/policies/meshtimeout/api/v1alpha1"
	"github.com/kumahq/kuma/pkg/util/pointer"
	util_proto "github.com/kumahq/kuma/pkg/util/proto"
	clusters_v3 "github.com/kumahq/kuma/pkg/xds/envoy/clusters/v3"
	listeners_v3 "github.com/kumahq/kuma/pkg/xds/envoy/listeners/v3"
)

// DeprecatedListenerConfigurer should be only used for configuring old MeshService outbounds.
// It should be removed after we stop using kuma.io/service tag, and move fully to new MeshService
// Deprecated
type DeprecatedListenerConfigurer struct {
	Rules    rules.Rules
	Protocol core_mesh.Protocol
	Subset   rules.Subset
	Element  rules.Element
}

func (c *DeprecatedListenerConfigurer) ConfigureListener(listener *envoy_listener.Listener) error {
	if listener == nil {
		return nil
	}

	httpTimeouts := func(hcm *envoy_hcm.HttpConnectionManager) error {
		c.configureRequestTimeout(hcm.GetRouteConfig())
		c.configureRequestHeadersTimeout(hcm)
		// old Timeout policy configures idleTimeout on listener while MeshTimeout sets this in cluster
		if hcm.CommonHttpProtocolOptions == nil {
			hcm.CommonHttpProtocolOptions = &envoy_core.HttpProtocolOptions{}
		}

		hcm.CommonHttpProtocolOptions.IdleTimeout = util_proto.Duration(0)
		return nil
	}
	tcpTimeouts := func(proxy *envoy_tcp.TcpProxy) error {
		if conf := c.getConf(c.Element); conf != nil {
			proxy.IdleTimeout = toProtoDurationOrDefault(conf.IdleTimeout, policies_defaults.DefaultIdleTimeout)
		}
		return nil
	}
	for _, filterChain := range listener.FilterChains {
		switch c.Protocol {
		case core_mesh.ProtocolHTTP, core_mesh.ProtocolHTTP2, core_mesh.ProtocolGRPC:
			if err := listeners_v3.UpdateHTTPConnectionManager(filterChain, httpTimeouts); err != nil && !errors.Is(err, &listeners_v3.UnexpectedFilterConfigTypeError{}) {
				return err
			}
		case core_mesh.ProtocolUnknown, core_mesh.ProtocolTCP, core_mesh.ProtocolKafka:
			if err := listeners_v3.UpdateTCPProxy(filterChain, tcpTimeouts); err != nil && !errors.Is(err, &listeners_v3.UnexpectedFilterConfigTypeError{}) {
				return err
			}
		}
	}

	return nil
}

func (c *DeprecatedListenerConfigurer) configureRequestTimeout(routeConfiguration *envoy_route.RouteConfiguration) {
	if routeConfiguration != nil {
		for _, vh := range routeConfiguration.VirtualHosts {
			for _, route := range vh.Routes {
				conf := c.getConf(c.Element.WithKeyValue(rules.RuleMatchesHashTag, route.Name))
				if conf == nil {
					conf = c.getConf(c.Element)
				}
				if conf == nil {
					continue
				}
				ConfigureRouteAction(
					route.GetRoute(),
					pointer.Deref(conf.Http).RequestTimeout,
					pointer.Deref(conf.Http).StreamIdleTimeout,
				)
			}
		}
	}
}

func (c *DeprecatedListenerConfigurer) configureRequestHeadersTimeout(hcm *envoy_hcm.HttpConnectionManager) {
	// For backward, once a user upgrades from an older version we shouldn't set default timeouts.
	// Refer to https://github.com/kumahq/kuma/issues/12033
	deprecatedGetConf := c.legacyGetConf(c.Subset)
	if deprecatedGetConf == nil {
		return
	}

	if conf := c.getConf(c.Element); conf != nil {
		hcm.RequestHeadersTimeout = toProtoDurationOrDefault(
			pointer.Deref(conf.Http).RequestHeadersTimeout,
			policies_defaults.DefaultRequestHeadersTimeout,
		)
	}
}

func (c *DeprecatedListenerConfigurer) getConf(element rules.Element) *api.Conf {
	if c.Rules == nil {
		return &api.Conf{}
	}
	return rules.ComputeConf[api.Conf](c.Rules, element)
}

func (c *DeprecatedListenerConfigurer) legacyGetConf(subset rules.Subset) *api.Conf {
	if c.Rules == nil {
		return &api.Conf{}
	}
	return rules.LegacyComputeConf[api.Conf](c.Rules, subset)
}

type ClusterConfigurer struct {
	ConnectionTimeout         *kube_meta.Duration
	IdleTimeout               *kube_meta.Duration
	HTTPMaxStreamDuration     *kube_meta.Duration
	HTTPMaxConnectionDuration *kube_meta.Duration
	Protocol                  core_mesh.Protocol
}

func ClusterConfigurerFromConf(conf api.Conf, protocol core_mesh.Protocol) ClusterConfigurer {
	return ClusterConfigurer{
		ConnectionTimeout:         conf.ConnectionTimeout,
		IdleTimeout:               conf.IdleTimeout,
		HTTPMaxStreamDuration:     pointer.Deref(conf.Http).MaxStreamDuration,
		HTTPMaxConnectionDuration: pointer.Deref(conf.Http).MaxConnectionDuration,
		Protocol:                  protocol,
	}
}

func (c *ClusterConfigurer) Configure(cluster *envoy_cluster.Cluster) error {
	cluster.ConnectTimeout = toProtoDurationOrDefault(c.ConnectionTimeout, policies_defaults.DefaultConnectTimeout)
	switch c.Protocol {
	case core_mesh.ProtocolHTTP, core_mesh.ProtocolHTTP2:
		err := clusters_v3.UpdateCommonHttpProtocolOptions(cluster, func(options *envoy_upstream_http.HttpProtocolOptions) {
			if options.CommonHttpProtocolOptions == nil {
				options.CommonHttpProtocolOptions = &envoy_core.HttpProtocolOptions{}
			}
			commonHttp := options.CommonHttpProtocolOptions
			commonHttp.IdleTimeout = toProtoDurationOrDefault(c.IdleTimeout, policies_defaults.DefaultIdleTimeout)
			commonHttp.MaxStreamDuration = toProtoDurationOrDefault(c.HTTPMaxStreamDuration, policies_defaults.DefaultMaxStreamDuration)
			commonHttp.MaxConnectionDuration = toProtoDurationOrDefault(c.HTTPMaxConnectionDuration, policies_defaults.DefaultMaxConnectionDuration)
		})
		if err != nil {
			return err
		}
	}
	return nil
}

func ConfigureRouteAction(
	routeAction *envoy_route.RouteAction,
	httpRequestTimeout *kube_meta.Duration,
	httpStreamIdleTimeout *kube_meta.Duration,
) {
	if routeAction == nil {
		return
	}
	routeAction.Timeout = toProtoDurationOrDefault(httpRequestTimeout, policies_defaults.DefaultRequestTimeout)
	if httpStreamIdleTimeout != nil {
		routeAction.IdleTimeout = toProtoDurationOrDefault(httpStreamIdleTimeout, policies_defaults.DefaultStreamIdleTimeout)
	} else if routeAction.IdleTimeout == nil {
		routeAction.IdleTimeout = util_proto.Duration(policies_defaults.DefaultStreamIdleTimeout)
	}
}

func ConfigureGatewayListener(
	conf *api.Conf,
	protocol mesh_proto.MeshGateway_Listener_Protocol,
	listener *envoy_listener.Listener,
) error {
	if listener == nil || conf == nil {
		return nil
	}

	httpTimeouts := func(hcm *envoy_hcm.HttpConnectionManager) error {
		if hcm.CommonHttpProtocolOptions == nil {
			hcm.CommonHttpProtocolOptions = &envoy_core.HttpProtocolOptions{}
		}
		hcm.CommonHttpProtocolOptions.IdleTimeout = toProtoDurationOrDefault(
			pointer.Deref(conf).IdleTimeout,
			policies_defaults.DefaultGatewayIdleTimeout,
		)
		hcm.RequestHeadersTimeout = toProtoDurationOrDefault(
			pointer.Deref(conf.Http).RequestHeadersTimeout,
			policies_defaults.DefaultGatewayRequestHeadersTimeout,
		)
		hcm.StreamIdleTimeout = toProtoDurationOrDefault(
			pointer.Deref(conf.Http).StreamIdleTimeout,
			policies_defaults.DefaultGatewayStreamIdleTimeout,
		)
		if httpConf := pointer.Deref(conf.Http); httpConf.RequestTimeout != nil {
			hcm.RequestTimeout = util_proto.Duration(httpConf.RequestTimeout.Duration)
		}
		return nil
	}
	tcpTimeouts := func(proxy *envoy_tcp.TcpProxy) error {
		if conf != nil {
			proxy.IdleTimeout = toProtoDurationOrDefault(conf.IdleTimeout, policies_defaults.DefaultGatewayIdleTimeout)
		}
		return nil
	}
	for _, filterChain := range listener.FilterChains {
		switch protocol {
		case mesh_proto.MeshGateway_Listener_HTTP, mesh_proto.MeshGateway_Listener_HTTPS:
			if err := listeners_v3.UpdateHTTPConnectionManager(filterChain, httpTimeouts); err != nil && !errors.Is(err, &listeners_v3.UnexpectedFilterConfigTypeError{}) {
				return err
			}
		case mesh_proto.MeshGateway_Listener_TCP, mesh_proto.MeshGateway_Listener_TLS:
			if err := listeners_v3.UpdateTCPProxy(filterChain, tcpTimeouts); err != nil && !errors.Is(err, &listeners_v3.UnexpectedFilterConfigTypeError{}) {
				return err
			}
		}
	}

	return nil
}

func toProtoDurationOrDefault(d *kube_meta.Duration, defaultDuration time.Duration) *durationpb.Duration {
	if d == nil {
		return util_proto.Duration(defaultDuration)
	}
	return util_proto.Duration(d.Duration)
}

var DefaultTimeoutConf = api.Conf{
	ConnectionTimeout: &kube_meta.Duration{Duration: policies_defaults.DefaultConnectTimeout},
	IdleTimeout:       &kube_meta.Duration{Duration: policies_defaults.DefaultIdleTimeout},
	Http: &api.Http{
		RequestTimeout:        &kube_meta.Duration{Duration: policies_defaults.DefaultRequestTimeout},
		StreamIdleTimeout:     &kube_meta.Duration{Duration: policies_defaults.DefaultGatewayStreamIdleTimeout},
		MaxStreamDuration:     &kube_meta.Duration{Duration: policies_defaults.DefaultMaxStreamDuration},
		MaxConnectionDuration: &kube_meta.Duration{Duration: policies_defaults.DefaultConnectTimeout},
		RequestHeadersTimeout: &kube_meta.Duration{Duration: policies_defaults.DefaultRequestHeadersTimeout},
	},
}

type ListenerConfigurer struct {
	Conf     api.Conf
	Protocol core_mesh.Protocol
}

func (rc *ListenerConfigurer) ConfigureListener(listener *envoy_listener.Listener) error {
	if listener == nil {
		return nil
	}

	httpTimeouts := func(hcm *envoy_hcm.HttpConnectionManager) error {
		rc.configureRequestTimeout(hcm.GetRouteConfig())
		rc.configureRequestHeadersTimeout(hcm)
		// old Timeout policy configures idleTimeout on listener while MeshTimeout sets this in cluster
		if hcm.CommonHttpProtocolOptions == nil {
			hcm.CommonHttpProtocolOptions = &envoy_core.HttpProtocolOptions{}
		}

		hcm.CommonHttpProtocolOptions.IdleTimeout = util_proto.Duration(0)
		return nil
	}
	tcpTimeouts := func(proxy *envoy_tcp.TcpProxy) error {
		proxy.IdleTimeout = toProtoDurationOrDefault(rc.Conf.IdleTimeout, policies_defaults.DefaultIdleTimeout)
		return nil
	}
	for _, filterChain := range listener.FilterChains {
		switch rc.Protocol {
		case core_mesh.ProtocolHTTP, core_mesh.ProtocolHTTP2, core_mesh.ProtocolGRPC:
			if err := listeners_v3.UpdateHTTPConnectionManager(filterChain, httpTimeouts); err != nil && !errors.Is(err, &listeners_v3.UnexpectedFilterConfigTypeError{}) {
				return err
			}
		case core_mesh.ProtocolUnknown, core_mesh.ProtocolTCP, core_mesh.ProtocolKafka:
			if err := listeners_v3.UpdateTCPProxy(filterChain, tcpTimeouts); err != nil && !errors.Is(err, &listeners_v3.UnexpectedFilterConfigTypeError{}) {
				return err
			}
		}
	}

	return nil
}

func (rc *ListenerConfigurer) configureRequestHeadersTimeout(hcm *envoy_hcm.HttpConnectionManager) {
	hcm.RequestHeadersTimeout = toProtoDurationOrDefault(
		pointer.Deref(rc.Conf.Http).RequestHeadersTimeout,
		policies_defaults.DefaultRequestHeadersTimeout,
	)
}

func (rc *ListenerConfigurer) configureRequestTimeout(routeConfiguration *envoy_route.RouteConfiguration) {
	if routeConfiguration != nil {
		for _, vh := range routeConfiguration.VirtualHosts {
			for _, route := range vh.Routes {
				ConfigureRouteAction(
					route.GetRoute(),
					pointer.Deref(rc.Conf.Http).RequestTimeout,
					pointer.Deref(rc.Conf.Http).StreamIdleTimeout,
				)
			}
		}
	}
}
