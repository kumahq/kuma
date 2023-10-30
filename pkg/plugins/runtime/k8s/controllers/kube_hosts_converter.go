package controllers

import (
	"context"

	kube_core "k8s.io/api/core/v1"
	kube_client "sigs.k8s.io/controller-runtime/pkg/client"

	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
	core_mesh "github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
	resources_manager "github.com/kumahq/kuma/pkg/core/resources/manager"
	core_store "github.com/kumahq/kuma/pkg/core/resources/store"
	"github.com/kumahq/kuma/pkg/dns/vips"
)

func KubeHosts(
	ctx context.Context,
	client kube_client.Client,
	manager resources_manager.ReadOnlyResourceManager,
	mesh string,
) (*vips.VirtualOutboundMeshView, error) {
	dpps := &core_mesh.DataplaneResourceList{}
	if err := manager.List(ctx, dpps, core_store.ListByMesh(mesh)); err != nil {
		return nil, err
	}

	view := vips.NewEmptyVirtualOutboundView()

	endpoints := endpointsByService(dpps.Items)
	for _, serviceTag := range endpoints.Services() {
		service, port, err := k8sService(ctx, serviceTag, client)
		if err != nil {
			converterLog.Error(err, "could not get K8S Service for service tag")
			continue // one invalid Dataplane definition should not break the entire mesh
		}

		// Services of ExternalName type should not have any selectors.
		// Kubernetes does not validate this, so in rare cases, a service of
		// ExternalName type could point to a workload inside the mesh. If this
		// happens, we will add the service to the VIPs config map, but we will
		// not be able to obtain its IP address. As a result, the key in the map
		// will be incorrect (e.g., "1:"). We do not currently support
		// ExternalName services, so we can safely skip them from processing.
		if service.Spec.Type == kube_core.ServiceTypeExternalName {
			converterLog.V(1).Info(
				"ignoring hostnames for unsupported ExternalName Service",
				"name", service.GetName(),
				"namespace", service.GetNamespace(),
			)
			continue
		}

		// Do not generate outbounds for service-less
		if isServiceLess(port) {
			continue
		}

		if isHeadlessService(service) {
			// Generate outbound listeners for every endpoint of services.
			for _, endpoint := range endpoints[serviceTag] {
				hostnameEntry := vips.NewHostEntry(endpoint.Address)
				err := view.Add(hostnameEntry, vips.OutboundEntry{
					Port: port,
					TagSet: map[string]string{
						mesh_proto.ServiceTag:  serviceTag,
						mesh_proto.InstanceTag: endpoint.Instance,
					},
					Origin: vips.OriginKube,
				})
				if err != nil {
					return nil, err
				}
			}
		} else {
			// generate outbound based on ClusterIP. Transparent Proxy will work only if DNS name that resolves to ClusterIP is used
			hostnameEntry := vips.NewHostEntry(service.Spec.ClusterIP)
			err := view.Add(hostnameEntry, vips.OutboundEntry{
				Port: port,
				TagSet: map[string]string{
					mesh_proto.ServiceTag: serviceTag,
				},
				Origin: vips.OriginKube,
			})
			if err != nil {
				return nil, err
			}
		}
	}
	return view, nil
}
