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
	k8s "k8s.io/apimachinery/pkg/apis/meta/v1"

	core_mesh "github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
	api "github.com/kumahq/kuma/pkg/plugins/policies/meshtimeout/api/v1alpha1"
	util_proto "github.com/kumahq/kuma/pkg/util/proto"
	clusters "github.com/kumahq/kuma/pkg/xds/envoy/clusters/v3"
	listeners_v3 "github.com/kumahq/kuma/pkg/xds/envoy/listeners/v3"
)

const (
	defaultConnectionTimeout     = "5s"
	defaultIdleTimeout           = "1h"
	defaultRequestTimeout        = "15s"
	defaultStreamIdleTimeout     = "30m"
	defaultMaxStreamDuration     = "0s"
	defaultMaxConnectionDuration = "0s"
)

type Configurer struct {
	Conf     api.Conf
	Protocol core_mesh.Protocol
}

func (c *Configurer) ConfigureListener(filterChain *envoy_listener.FilterChain) error {
	httpTimeouts := func(hcm *envoy_hcm.HttpConnectionManager) error {
		if c.Conf.Http != nil {
			hcm.StreamIdleTimeout = toProtoDurationOrDefault(c.Conf.Http.StreamIdleTimeout, defaultStreamIdleTimeout)
			configureRequestTimeout(hcm.GetRouteConfig(), toProtoDurationOrDefault(c.Conf.Http.RequestTimeout, defaultRequestTimeout))
		} else {
			hcm.StreamIdleTimeout = defaultToProtoDuration(defaultStreamIdleTimeout)
			configureRequestTimeout(hcm.GetRouteConfig(), defaultToProtoDuration(defaultRequestTimeout))
		}
		return nil
	}
	tcpTimeouts := func(proxy *envoy_tcp.TcpProxy) error {
		proxy.IdleTimeout = toProtoDurationOrDefault(c.Conf.IdleTimeout, defaultIdleTimeout)
		return nil
	}
	switch c.Protocol {
	case core_mesh.ProtocolHTTP, core_mesh.ProtocolHTTP2:
		if err := listeners_v3.UpdateHTTPConnectionManager(filterChain, httpTimeouts); err != nil && !errors.Is(err, &listeners_v3.UnexpectedFilterConfigTypeError{}) {
			return err
		}
	case core_mesh.ProtocolUnknown, core_mesh.ProtocolTCP, core_mesh.ProtocolKafka:
		if err := listeners_v3.UpdateTCPProxy(filterChain, tcpTimeouts); err != nil && !errors.Is(err, &listeners_v3.UnexpectedFilterConfigTypeError{}) {
			return err
		}
	}

	return nil
}

func (c *Configurer) ConfigureCluster(cluster *envoy_cluster.Cluster) error {
	cluster.ConnectTimeout = toProtoDurationOrDefault(c.Conf.ConnectionTimeout, defaultConnectionTimeout)
	switch c.Protocol {
	case core_mesh.ProtocolHTTP, core_mesh.ProtocolHTTP2:
		err := clusters.UpdateCommonHttpProtocolOptions(cluster, func(options *envoy_upstream_http.HttpProtocolOptions) {
			if options.CommonHttpProtocolOptions == nil {
				options.CommonHttpProtocolOptions = &envoy_core.HttpProtocolOptions{}
			}
			commonHttp := options.CommonHttpProtocolOptions
			commonHttp.IdleTimeout = toProtoDurationOrDefault(c.Conf.IdleTimeout, defaultIdleTimeout)
			if c.Conf.Http != nil {
				commonHttp.MaxStreamDuration = toProtoDurationOrDefault(c.Conf.Http.MaxStreamDuration, defaultMaxStreamDuration)
				commonHttp.MaxConnectionDuration = toProtoDurationOrDefault(c.Conf.Http.MaxConnectionDuration, defaultMaxConnectionDuration)
			} else {
				commonHttp.MaxStreamDuration = defaultToProtoDuration(defaultMaxStreamDuration)
				commonHttp.MaxConnectionDuration = defaultToProtoDuration(defaultMaxConnectionDuration)
			}
		})
		if err != nil {
			return err
		}
	}
	return nil
}

func configureRequestTimeout(routeConfiguration *envoy_route.RouteConfiguration, timeout *durationpb.Duration) {
	for _, vh := range routeConfiguration.VirtualHosts {
		for _, route := range vh.Routes {
			routeAction := route.GetRoute()
			if routeAction != nil {
				routeAction.Timeout = timeout
			}
		}
	}
}

func toProtoDurationOrDefault(d *k8s.Duration, defaultDuration string) *durationpb.Duration {
	if d == nil {
		return defaultToProtoDuration(defaultDuration)
	}
	return util_proto.Duration(d.Duration)
}

func defaultToProtoDuration(d string) *durationpb.Duration {
	duration, _ := time.ParseDuration(d)
	return util_proto.Duration(duration)
}
