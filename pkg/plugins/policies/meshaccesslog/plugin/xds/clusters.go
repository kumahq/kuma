package xds

import (
	core_xds "github.com/kumahq/kuma/v2/pkg/core/xds"
	"github.com/kumahq/kuma/v2/pkg/plugins/policies/meshaccesslog/plugin/metadata"
	"github.com/kumahq/kuma/v2/pkg/xds/envoy/clusters"
)

type LoggingEndpoint struct {
	Address  string
	Port     uint32
	UseHTTP2 bool
	// SocketPath is non-empty when routing through a kuma-dp Unix socket.
	SocketPath string
	// BackendName is the resolved OTel backend name, used for cluster naming when SocketPath is set.
	BackendName string
}

func xdsEndpoint(endpoint LoggingEndpoint) core_xds.Endpoint {
	if endpoint.SocketPath != "" {
		return core_xds.Endpoint{UnixDomainPath: endpoint.SocketPath}
	}
	return core_xds.Endpoint{
		Target: endpoint.Address,
		Port:   endpoint.Port,
	}
}

func AddLogBackendConf(backendEndpoints EndpointAccumulator, rs *core_xds.ResourceSet, proxy *core_xds.Proxy) error {
	for backendEndpoint := range backendEndpoints.endpoints {
		endpoint := xdsEndpoint(backendEndpoint)

		clusterName := backendEndpoints.ClusterForEndpoint(backendEndpoint)
		builder := clusters.NewClusterBuilder(proxy.APIVersion, string(clusterName)).
			Configure(clusters.ProvidedEndpointCluster(proxy.Dataplane.IsIPv6(), endpoint)).
			Configure(clusters.ClientSideTLS([]core_xds.Endpoint{endpoint})).
			Configure(clusters.DefaultTimeout())
		if backendEndpoint.UseHTTP2 {
			builder.Configure(clusters.Http2())
		}
		res, err := builder.Build()
		if err != nil {
			return err
		}

		rs.Add(&core_xds.Resource{Name: string(clusterName), Origin: metadata.OriginMeshAccessLog, Resource: res})
	}

	return nil
}
