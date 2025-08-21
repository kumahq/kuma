package generator

import (
	"context"
	"fmt"

	"github.com/pkg/errors"

	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
	manager_dataplane "github.com/kumahq/kuma/pkg/core/managers/apis/dataplane"
	core_meta "github.com/kumahq/kuma/pkg/core/metadata"
	core_mesh "github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
	core_xds "github.com/kumahq/kuma/pkg/core/xds"
	"github.com/kumahq/kuma/pkg/plugins/policies/core/xds"
	xds_context "github.com/kumahq/kuma/pkg/xds/context"
	envoy_common "github.com/kumahq/kuma/pkg/xds/envoy"
	envoy_clusters "github.com/kumahq/kuma/pkg/xds/envoy/clusters"
	envoy_listeners "github.com/kumahq/kuma/pkg/xds/envoy/listeners"
	meta "github.com/kumahq/kuma/pkg/xds/generator/metadata"
	"github.com/kumahq/kuma/pkg/xds/generator/model"
)

// Transparent Proxy is based on having 1 IP for cluster (ex. ClusterIP of Service on K8S), so consuming apps by their IP
// is unknown destination from Envoy perspective. Therefore such request will go trough pass_trough cluster and won't be encrypted by mTLS.
// This generates listener for every IP and redirects traffic trough "direct_access" cluster which is configured to encrypt connections.
// Generating listener for every endpoint will cause XDS snapshot to be large therefore it should be used only if really needed.
//
// Second approach to consider was to use FilterChainMatch on catch_all listener with list of all direct access endpoints
// instead of generating outbound listener, but it seemed to not work with Listener#UseOriginalDst
type DirectAccessProxyGenerator struct{}

func DirectAccessEndpointName(endpoint model.Endpoint) string {
	return fmt.Sprintf("%s_%s:%d", meta.DirectAccessClusterName, endpoint.Address, endpoint.Port)
}

func (DirectAccessProxyGenerator) Generate(_ context.Context, _ *core_xds.ResourceSet, xdsCtx xds_context.Context, proxy *core_xds.Proxy) (*core_xds.ResourceSet, error) {
	rs := core_xds.NewResourceSet()

	directAccessServices := proxy.Dataplane.Spec.GetNetworking().GetTransparentProxying().GetDirectAccessServices()
	if !proxy.GetTransparentProxy().Enabled() || len(directAccessServices) == 0 {
		return rs, nil
	}

	svc := proxy.Dataplane.Spec.GetIdentifyingService()
	mesh := xdsCtx.Mesh.Resource.GetMeta().GetName()

	endpoints, err := directAccessEndpoints(proxy.Dataplane, xdsCtx.Mesh.Resources.Dataplanes(), xdsCtx.Mesh.Resource)
	if err != nil {
		return nil, err
	}

	for _, endpoint := range endpoints {
		name := DirectAccessEndpointName(endpoint)

		cluster := xds.NewClusterBuilder().WithService(meta.DirectAccessClusterName).Build()

		loggingBackend := xdsCtx.Mesh.GetLoggingBackend(proxy.Policies.TrafficLogs[core_meta.PassThroughServiceName])

		filterChain := envoy_listeners.NewFilterChainBuilder(proxy.APIVersion, envoy_common.AnonymousResource).
			Configure(envoy_listeners.TcpProxyDeprecated(name, cluster)).
			Configure(envoy_listeners.NetworkAccessLog(mesh, envoy_common.TrafficDirectionOutbound, svc, name, loggingBackend, proxy))

		listener, err := envoy_listeners.NewOutboundListenerBuilder(proxy.APIVersion, endpoint.Address, endpoint.Port, core_xds.SocketAddressProtocolTCP).
			WithOverwriteName(name).
			Configure(envoy_listeners.FilterChain(filterChain)).
			Configure(envoy_listeners.TransparentProxying(proxy)).
			Build()
		if err != nil {
			return nil, err
		}

		rs.Add(&core_xds.Resource{
			Name:     name,
			Origin:   meta.OriginDirectAccess,
			Resource: listener,
		})
	}

	resource, err := envoy_clusters.NewClusterBuilder(proxy.APIVersion, meta.DirectAccessClusterName).
		Configure(envoy_clusters.PassThroughCluster()).
		Configure(envoy_clusters.UnknownDestinationClientSideMTLS(proxy.SecretsTracker, xdsCtx.Mesh.Resource)).
		Configure(envoy_clusters.DefaultTimeout()).
		Build()
	if err != nil {
		return nil, errors.Wrapf(err, "could not generate cluster: %s", meta.DirectAccessClusterName)
	}

	rs.Add(&core_xds.Resource{
		Name:     resource.GetName(),
		Origin:   meta.OriginDirectAccess,
		Resource: resource,
	})

	return rs, nil
}

func directAccessEndpoints(dataplane *core_mesh.DataplaneResource, other *core_mesh.DataplaneResourceList, mesh *core_mesh.MeshResource) (model.Endpoints, error) {
	// collect endpoints that are already created so we don't create 2 listeners with same IP:PORT
	taken := map[model.Endpoint]struct{}{}
	for _, oface := range dataplane.Spec.GetNetworking().GetOutboundInterfaces() {
		endpoint := model.Endpoint{
			Address: oface.DataplaneIP,
			Port:    oface.DataplanePort,
		}
		taken[endpoint] = struct{}{}
	}

	services := map[string]bool{}
	for _, service := range dataplane.Spec.GetNetworking().GetTransparentProxying().GetDirectAccessServices() {
		services[service] = true
	}

	endpoints := map[model.Endpoint]struct{}{}
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
				endpoint := model.Endpoint{
					Address: iface.DataplaneIP,
					Port:    iface.DataplanePort,
				}
				if _, ok := taken[endpoint]; !ok {
					endpoints[endpoint] = struct{}{}
				}
			}
		}
	}

	return model.EndpointsFromMap(endpoints), nil
}
