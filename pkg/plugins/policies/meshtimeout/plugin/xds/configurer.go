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

	core_mesh "github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
	api "github.com/kumahq/kuma/pkg/plugins/policies/meshtimeout/api/v1alpha1"
	"github.com/kumahq/kuma/pkg/util/pointer"
	util_proto "github.com/kumahq/kuma/pkg/util/proto"
	clusters_v3 "github.com/kumahq/kuma/pkg/xds/envoy/clusters/v3"
	listeners_v3 "github.com/kumahq/kuma/pkg/xds/envoy/listeners/v3"
)

const (
	defaultConnectionTimeout     = time.Second * 5
	defaultIdleTimeout           = time.Hour
	defaultRequestTimeout        = time.Second * 15
	defaultStreamIdleTimeout     = time.Minute * 30
	defaultMaxStreamDuration     = 0
	defaultMaxConnectionDuration = 0
)

type ListenerConfigurer struct {
	Conf     api.Conf
	Protocol core_mesh.Protocol
}

func (c *ListenerConfigurer) ConfigureListener(listener *envoy_listener.Listener) error {
	if listener == nil {
		return nil
	}

	httpTimeouts := func(hcm *envoy_hcm.HttpConnectionManager) error {
		c.configureRequestTimeout(hcm.GetRouteConfig())
		// old Timeout policy configures idleTimeout on listener while MeshTimeout sets this in cluster
		if hcm.CommonHttpProtocolOptions == nil {
			hcm.CommonHttpProtocolOptions = &envoy_core.HttpProtocolOptions{}
		}
		hcm.CommonHttpProtocolOptions.IdleTimeout = util_proto.Duration(0)
		return nil
	}
	tcpTimeouts := func(proxy *envoy_tcp.TcpProxy) error {
		proxy.IdleTimeout = toProtoDurationOrDefault(c.Conf.IdleTimeout, defaultIdleTimeout)
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
	cluster.ConnectTimeout = toProtoDurationOrDefault(c.ConnectionTimeout, defaultConnectionTimeout)
	switch c.Protocol {
	case core_mesh.ProtocolHTTP, core_mesh.ProtocolHTTP2:
		err := clusters_v3.UpdateCommonHttpProtocolOptions(cluster, func(options *envoy_upstream_http.HttpProtocolOptions) {
			if options.CommonHttpProtocolOptions == nil {
				options.CommonHttpProtocolOptions = &envoy_core.HttpProtocolOptions{}
			}
			commonHttp := options.CommonHttpProtocolOptions
			commonHttp.IdleTimeout = toProtoDurationOrDefault(c.IdleTimeout, defaultIdleTimeout)
			commonHttp.MaxStreamDuration = toProtoDurationOrDefault(c.HTTPMaxStreamDuration, defaultMaxStreamDuration)
			commonHttp.MaxConnectionDuration = toProtoDurationOrDefault(c.HTTPMaxConnectionDuration, defaultMaxConnectionDuration)
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
	routeAction.Timeout = toProtoDurationOrDefault(httpRequestTimeout, defaultRequestTimeout)
	if httpStreamIdleTimeout != nil {
		routeAction.IdleTimeout = toProtoDurationOrDefault(httpStreamIdleTimeout, defaultStreamIdleTimeout)
	} else if routeAction.IdleTimeout == nil {
		routeAction.IdleTimeout = util_proto.Duration(defaultStreamIdleTimeout)
	}
}

func (c *ListenerConfigurer) configureRequestTimeout(routeConfiguration *envoy_route.RouteConfiguration) {
	if routeConfiguration != nil {
		for _, vh := range routeConfiguration.VirtualHosts {
			for _, route := range vh.Routes {
				ConfigureRouteAction(
					route.GetRoute(),
					pointer.Deref(c.Conf.Http).RequestTimeout,
					pointer.Deref(c.Conf.Http).StreamIdleTimeout,
				)
			}
		}
	}
}

func toProtoDurationOrDefault(d *kube_meta.Duration, defaultDuration time.Duration) *durationpb.Duration {
	if d == nil {
		return util_proto.Duration(defaultDuration)
	}
	return util_proto.Duration(d.Duration)
}
