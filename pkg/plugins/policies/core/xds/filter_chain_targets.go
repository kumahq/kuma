package xds

import (
	envoy_cluster "github.com/envoyproxy/go-control-plane/envoy/config/cluster/v3"
	envoy_listener "github.com/envoyproxy/go-control-plane/envoy/config/listener/v3"
	envoy_hcm "github.com/envoyproxy/go-control-plane/envoy/extensions/filters/network/http_connection_manager/v3"
	envoy_tcp "github.com/envoyproxy/go-control-plane/envoy/extensions/filters/network/tcp_proxy/v3"

	core_meta "github.com/kumahq/kuma/v3/pkg/core/metadata"
)

const (
	httpConnectionManagerFilterName = "envoy.filters.network.http_connection_manager"
	tcpProxyFilterName              = "envoy.filters.network.tcp_proxy"
)

// GatherFilterChainTargets returns the resolved target cluster for each filter chain
// that forwards traffic directly to a named cluster.
func GatherFilterChainTargets(listeners Listeners, clusters Clusters) map[*envoy_listener.FilterChain]*envoy_cluster.Cluster {
	targets := map[*envoy_listener.FilterChain]*envoy_cluster.Cluster{}
	clusterByName := gatherAllClusters(clusters)

	for _, listener := range gatherAllListeners(listeners) {
		for _, filterChain := range listener.GetFilterChains() {
			clusterName, _ := filterChainTargetInfo(filterChain)
			if clusterName == "" {
				continue
			}
			cluster, ok := clusterByName[clusterName]
			if !ok {
				continue
			}
			targets[filterChain] = cluster
		}
	}

	return targets
}

// FilterChainProtocol returns the protocol configured on the filter chain's terminal proxy filter.
func FilterChainProtocol(filterChain *envoy_listener.FilterChain) core_meta.Protocol {
	_, protocol := filterChainTargetInfo(filterChain)
	return protocol
}

func gatherAllListeners(listeners Listeners) []*envoy_listener.Listener {
	seen := map[*envoy_listener.Listener]struct{}{}
	all := []*envoy_listener.Listener{}

	add := func(listener *envoy_listener.Listener) {
		if listener == nil {
			return
		}
		if _, ok := seen[listener]; ok {
			return
		}
		seen[listener] = struct{}{}
		all = append(all, listener)
	}

	for _, listener := range listeners.Inbound {
		add(listener)
	}
	for _, listener := range listeners.Outbound {
		add(listener)
	}
	add(listeners.Egress)
	for _, listener := range listeners.ZoneIngress {
		add(listener)
	}
	for _, listener := range listeners.ZoneEgress {
		add(listener)
	}
	for _, listener := range listeners.Gateway {
		add(listener)
	}
	add(listeners.Ipv4Passthrough)
	add(listeners.Ipv6Passthrough)
	for _, listener := range listeners.DirectAccess {
		add(listener)
	}
	add(listeners.Prometheus)

	return all
}

func gatherAllClusters(clusters Clusters) map[string]*envoy_cluster.Cluster {
	all := map[string]*envoy_cluster.Cluster{}

	add := func(cluster *envoy_cluster.Cluster) {
		if cluster == nil {
			return
		}
		all[cluster.GetName()] = cluster
	}

	for _, cluster := range clusters.Inbound {
		add(cluster)
	}
	for _, cluster := range clusters.Outbound {
		add(cluster)
	}
	for _, splitClusters := range clusters.OutboundSplit {
		for _, cluster := range splitClusters {
			add(cluster)
		}
	}
	for _, cluster := range clusters.Gateway {
		add(cluster)
	}
	for _, cluster := range clusters.Egress {
		add(cluster)
	}
	add(clusters.Prometheus)

	return all
}

func filterChainTargetInfo(filterChain *envoy_listener.FilterChain) (string, core_meta.Protocol) {
	for _, filter := range filterChain.GetFilters() {
		typedConfig := filter.GetTypedConfig()
		if typedConfig == nil {
			continue
		}

		msg, err := typedConfig.UnmarshalNew()
		if err != nil {
			continue
		}

		switch filter.GetName() {
		case httpConnectionManagerFilterName:
			hcm, ok := msg.(*envoy_hcm.HttpConnectionManager)
			if !ok {
				continue
			}
			return httpConnectionManagerClusterName(hcm), core_meta.ProtocolHTTP
		case tcpProxyFilterName:
			tcpProxy, ok := msg.(*envoy_tcp.TcpProxy)
			if !ok {
				continue
			}
			return tcpProxy.GetCluster(), core_meta.ProtocolTCP
		}
	}

	return "", core_meta.ProtocolUnknown
}

func httpConnectionManagerClusterName(hcm *envoy_hcm.HttpConnectionManager) string {
	routeConfig := hcm.GetRouteConfig()
	if routeConfig == nil {
		return ""
	}

	for _, virtualHost := range routeConfig.GetVirtualHosts() {
		for _, route := range virtualHost.GetRoutes() {
			if clusterName := route.GetRoute().GetCluster(); clusterName != "" {
				return clusterName
			}
		}
	}

	return ""
}
