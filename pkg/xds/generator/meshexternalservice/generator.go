package meshexternalservice

import (
	"context"
	"strings"

	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"

	"github.com/kumahq/kuma/pkg/core/resources/apis/meshexternalservice/api/v1alpha1"
	core_xds "github.com/kumahq/kuma/pkg/core/xds"
	xds_context "github.com/kumahq/kuma/pkg/xds/context"
	envoy_common "github.com/kumahq/kuma/pkg/xds/envoy"
	envoy_clusters "github.com/kumahq/kuma/pkg/xds/envoy/clusters"
	v3 "github.com/kumahq/kuma/pkg/xds/envoy/clusters/v3"
	"github.com/kumahq/kuma/pkg/xds/envoy/listeners"
	"github.com/kumahq/kuma/pkg/xds/envoy/names"
	generator_core "github.com/kumahq/kuma/pkg/xds/generator/core"
)

// OriginMeshExternalService is a marker to indicate by which ProxyGenerator resources were generated.
const OriginMeshExternalService = "meshexternalservice"

var extensions map[string]Extension

type Generator struct{}

type Extension interface {
	Generate(config *apiextensionsv1.JSON) (*core_xds.ResourceSet, error)
}

var _ generator_core.ResourceGenerator = Generator{}

func RegisterExtension(name string, extension Extension) {
	if extensions[name] != nil {
		panic("extension " + name + " already registered")
	}
	extensions[name] = extension
}

// Generate generates MeshExternalService related resources
func (g Generator) Generate(
	ctx context.Context,
	_ *core_xds.ResourceSet,
	xdsCtx xds_context.Context,
	proxy *core_xds.Proxy,
) (*core_xds.ResourceSet, error) {
	resources := core_xds.NewResourceSet()
	meshExternalServices := xdsCtx.Mesh.Resources.MeshExternalServices().Items

	for _, mes := range meshExternalServices {
		res, err := g.generateResources(mes, proxy, xdsCtx)
		if err != nil {
			return nil, err
		}

		resources.AddSet(res)
	}

	return resources, nil
}

func (g Generator) generateResources(mes *v1alpha1.MeshExternalServiceResource, proxy *core_xds.Proxy, xdsCtx xds_context.Context) (*core_xds.ResourceSet, error) {
	resources := core_xds.NewResourceSet()

	if mes.Spec.Extension != nil {
		res, err := extensions[mes.Spec.Extension.Type].Generate(mes.Spec.Extension.Config) // pass tls / endpoints as well
		if err != nil {
			return nil, err
		}
		resources.AddSet(res)
	} else {
		cluster, err := createCluster(mes, proxy, xdsCtx)
		if err != nil {
			return nil, err
		}
		resources.Add(&core_xds.Resource{
			Name:     cluster.GetName(),
			Origin:   OriginMeshExternalService,
			Resource: cluster,
		})

		listener, err := createListener(mes, proxy.APIVersion)
		if err != nil {
			return nil, err
		}
		resources.Add(&core_xds.Resource{
			Name:     listener.GetName(),
			Origin:   OriginMeshExternalService,
			Resource: listener,
		})
	}

	return resources, nil
}

func createCluster(mes *v1alpha1.MeshExternalServiceResource, proxy *core_xds.Proxy, xdsCtx xds_context.Context) (envoy_common.NamedResource, error) {
	name := mes.Meta.GetName()
	clusterName := names.GetMeshExternalServiceClusterName(name)
	builder := envoy_clusters.NewClusterBuilder(proxy.APIVersion, clusterName)

	var endpoints []core_xds.Endpoint
	for _, e := range mes.Spec.Endpoints {
		if strings.HasPrefix(e.Address, "unix://") {
			endpoints = append(endpoints, core_xds.Endpoint{
				UnixDomainPath: e.Address,
			})
		} else {
			endpoints = append(endpoints, core_xds.Endpoint{
				Target: e.Address,
				Port:   uint32(*e.Port),
			})
		}
	}
	builder.Configure(
		envoy_clusters.ClusterBuilderOptFunc(func(builder *envoy_clusters.ClusterBuilder) {
			builder.AddConfigurer(&v3.ProvidedEndpointClusterConfigurer{
				Name:                           clusterName,
				Endpoints:                      endpoints,
				HasIPv6:                        proxy.Dataplane.IsIPv6(),
				AllowMixingIpAndNonIpEndpoints: true,
			})
			builder.AddConfigurer(&v3.AltStatNameConfigurer{})
		}))

	if mes.Spec.Tls != nil {
		builder.Configure(envoy_clusters.MeshExternalServiceTLS(mes.Spec.Tls, xdsCtx.Mesh.DataSourceLoader, xdsCtx.Mesh.Resource.GetMeta().GetName()))
	}

	return builder.Build()
}

func createListener(mes *v1alpha1.MeshExternalServiceResource, version core_xds.APIVersion) (envoy_common.NamedResource, error) {
	name := mes.Meta.GetName()
	address := mes.Status.Vip.Value // is this always available right away?
	port := mes.Spec.Match.Port
	protocol := mes.Spec.Match.Protocol
	listenerName := names.GetMeshExternalServiceListenerName(name)
	clusterName := names.GetMeshExternalServiceClusterName(name)

	builder := listeners.NewInboundListenerBuilder(version, address, uint32(port), core_xds.SocketAddressProtocolTCP)
	builder.WithOverwriteName(listenerName)

	filterChainBuilder := listeners.NewFilterChainBuilder(version, names.GetMeshExternalServiceFilterChainName(name))

	switch protocol {
	case v1alpha1.HttpProtocol:
		filterChainBuilder.Configure(listeners.HttpConnectionManager(clusterName, false)) //
	case v1alpha1.Http2Protocol:
		filterChainBuilder.Configure(listeners.HttpConnectionManager(clusterName, false)) //
	case v1alpha1.GrpcProtocol:
		filterChainBuilder.Configure(listeners.HttpConnectionManager(clusterName, false)) //
		filterChainBuilder.Configure(listeners.GrpcStats())
	case v1alpha1.TcpProtocol:
		fallthrough
	default:
		filterChainBuilder.Configure(listeners.TCPProxy(listenerName))
	}

	builder.Configure(listeners.FilterChain(filterChainBuilder))

	return builder.Build()
}
