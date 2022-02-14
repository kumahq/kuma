package generator

import (
	"fmt"
	"sort"

	"github.com/pkg/errors"

	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
	manager_dataplane "github.com/kumahq/kuma/pkg/core/managers/apis/dataplane"
	core_mesh "github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
	core_xds "github.com/kumahq/kuma/pkg/core/xds"
	xds_context "github.com/kumahq/kuma/pkg/xds/context"
	envoy_common "github.com/kumahq/kuma/pkg/xds/envoy"
	envoy_clusters "github.com/kumahq/kuma/pkg/xds/envoy/clusters"
	envoy_listeners "github.com/kumahq/kuma/pkg/xds/envoy/listeners"
)

// OriginDirectAccess is a marker to indicate by which ProxyGenerator resources were generated.
const OriginDirectAccess = "direct-access"

// Transparent Proxy is based on having 1 IP for cluster (ex. ClusterIP of Service on K8S), so consuming apps by their IP
// is unknown destination from Envoy perspective. Therefore such request will go trough pass_trough cluster and won't be encrypted by mTLS.
// This generates listener for every IP and redirect traffic trough "direct_access" cluster which is configured to encrypt connections.
// Generating listener for every endpoint will cause XDS snapshot to be large therefore it should be used only if really needed.
//
// Second approach to consider was to use FilterChainMatch on catch_all listener with list of all direct access endpoints
// instead of generating outbound listener, but it seemed to not work with Listener#UseOriginalDst
type DirectAccessProxyGenerator struct {
}

func (_ DirectAccessProxyGenerator) Generate(ctx xds_context.Context, proxy *core_xds.Proxy) (*core_xds.ResourceSet, error) {
	tproxy := proxy.Dataplane.Spec.Networking.GetTransparentProxying()
	resources := core_xds.NewResourceSet()
	if tproxy.GetRedirectPortOutbound() == 0 || tproxy.GetRedirectPortInbound() == 0 || len(tproxy.GetDirectAccessServices()) == 0 {
		return resources, nil
	}

	sourceService := proxy.Dataplane.Spec.GetIdentifyingService()
	meshName := ctx.Mesh.Resource.GetMeta().GetName()

	endpoints, err := directAccessEndpoints(proxy.Dataplane, ctx.Mesh.Resources.Dataplanes(), ctx.Mesh.Resource)
	if err != nil {
		return nil, err
	}

	for _, endpoint := range endpoints {
		name := fmt.Sprintf("direct_access_%s:%d", endpoint.Address, endpoint.Port)
		listener, err := envoy_listeners.NewListenerBuilder(proxy.APIVersion).
			Configure(envoy_listeners.OutboundListener(name, endpoint.Address, endpoint.Port, core_xds.SocketAddressProtocolTCP)).
			Configure(envoy_listeners.FilterChain(envoy_listeners.NewFilterChainBuilder(proxy.APIVersion).
				Configure(envoy_listeners.TcpProxy(name, envoy_common.NewCluster(envoy_common.WithService("direct_access")))).
				Configure(envoy_listeners.NetworkAccessLog(
					meshName,
					envoy_common.TrafficDirectionOutbound,
					sourceService,
					name,
					ctx.Mesh.GetLoggingBackend(proxy.Policies.TrafficLogs[core_mesh.PassThroughService]),
					proxy,
				)))).
			Configure(envoy_listeners.TransparentProxying(proxy.Dataplane.Spec.Networking.GetTransparentProxying())).
			Build()
		if err != nil {
			return nil, err
		}
		resources.Add(&core_xds.Resource{
			Name:     name,
			Origin:   OriginDirectAccess,
			Resource: listener,
		})
	}

	directAccessCluster, err := envoy_clusters.NewClusterBuilder(proxy.APIVersion).
		Configure(envoy_clusters.PassThroughCluster("direct_access")).
		Configure(envoy_clusters.UnknownDestinationClientSideMTLS(ctx.Mesh.Resource)).
		Build()
	if err != nil {
		return nil, errors.Wrapf(err, "could not generate cluster: direct_access")
	}
	resources.Add(&core_xds.Resource{
		Name:     directAccessCluster.GetName(),
		Origin:   OriginDirectAccess,
		Resource: directAccessCluster,
	})
	return resources, nil
}

func directAccessEndpoints(dataplane *core_mesh.DataplaneResource, other *core_mesh.DataplaneResourceList, mesh *core_mesh.MeshResource) (Endpoints, error) {
	// collect endpoints that are already created so we don't create 2 listeners with same IP:PORT
	takenEndpoints, err := takenEndpoints(dataplane)
	if err != nil {
		return nil, err
	}

	services := map[string]bool{}
	for _, service := range dataplane.Spec.Networking.TransparentProxying.DirectAccessServices {
		services[service] = true
	}

	endpoints := map[Endpoint]bool{}
	for _, dp := range other.Items {
		if dp.Meta.GetName() == dataplane.Meta.GetName() { // skip itself
			continue
		}
		inbounds, err := manager_dataplane.AdditionalInbounds(dp, mesh)
		if err != nil {
			return nil, err
		}
		for _, inbound := range append(inbounds, dp.Spec.GetNetworking().GetInbound()...) {
			service := inbound.Tags[mesh_proto.ServiceTag]
			if services["*"] || services[service] {
				iface := dp.Spec.GetNetworking().ToInboundInterface(inbound)
				endpoint := Endpoint{
					Address: iface.DataplaneIP,
					Port:    iface.DataplanePort,
				}
				if !takenEndpoints[endpoint] {
					endpoints[endpoint] = true
				}
			}
		}
	}
	return fromMap(endpoints), nil
}

func takenEndpoints(dataplane *core_mesh.DataplaneResource) (map[Endpoint]bool, error) {
	takenEndpoints := map[Endpoint]bool{}
	for _, oface := range dataplane.Spec.GetNetworking().GetOutboundInterfaces() {
		endpoint := Endpoint{
			Address: oface.DataplaneIP,
			Port:    oface.DataplanePort,
		}
		takenEndpoints[endpoint] = true
	}
	return takenEndpoints, nil
}

type Endpoint struct {
	Address string
	Port    uint32
}

type Endpoints []Endpoint

func fromMap(endpointsMap map[Endpoint]bool) Endpoints {
	endpoints := Endpoints{}
	for endpoint := range endpointsMap {
		endpoints = append(endpoints, endpoint)
	}
	sort.Stable(endpoints) // sort for consistent envoy config
	return endpoints
}

func (a Endpoints) Len() int      { return len(a) }
func (a Endpoints) Swap(i, j int) { a[i], a[j] = a[j], a[i] }
func (a Endpoints) Less(i, j int) bool {
	if a[i].Address == a[j].Address {
		return a[i].Port < a[j].Port
	}
	return a[i].Address < a[j].Address
}
