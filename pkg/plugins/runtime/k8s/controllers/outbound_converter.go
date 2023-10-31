package controllers

import (
	"context"
	"strconv"
	"strings"

	"github.com/pkg/errors"
	kube_core "k8s.io/api/core/v1"
	kube_client "sigs.k8s.io/controller-runtime/pkg/client"

	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
	core_mesh "github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
	mesh_k8s "github.com/kumahq/kuma/pkg/plugins/resources/k8s/native/api/v1alpha1"
)

func (p *PodConverter) OutboundInterfacesFor(
	ctx context.Context,
	pod *kube_core.Pod,
	others []*mesh_k8s.Dataplane,
	reachableServices []string,
) ([]*mesh_proto.Dataplane_Networking_Outbound, error) {
	var outbounds []*mesh_proto.Dataplane_Networking_Outbound

	reachableServicesMap := map[string]bool{}
	for _, service := range reachableServices {
		reachableServicesMap[service] = true
	}

	var dataplanes []*core_mesh.DataplaneResource
	for _, other := range others {
		dp := core_mesh.NewDataplaneResource()
		if err := p.ResourceConverter.ToCoreResource(other, dp); err != nil {
			converterLog.Error(err, "failed to parse Dataplane", "dataplane", other.Spec)
			continue // one invalid Dataplane definition should not break the entire mesh
		}
		dataplanes = append(dataplanes, dp)
	}

	endpoints := endpointsByService(dataplanes)
	for _, serviceTag := range endpoints.Services() {
		service, port, err := k8sService(ctx, serviceTag, p.ServiceGetter)
		if err != nil {
			converterLog.Error(err, "could not get K8S Service for service tag")
			continue // one invalid Dataplane definition should not break the entire mesh
		}
		if len(reachableServices) > 0 && !reachableServicesMap[serviceTag] {
			continue // ignore generating outbound if reachable services are defined and this one is not on the list
		}

		// Do not generate outbounds for service-less
		if isServiceLess(port) {
			continue
		}

		// Do not generate hostnames for ExternalName Service
		if isExternalNameService(service) {
			converterLog.V(1).Info(
				"ignoring outbound generation for unsupported ExternalName Service",
				"name", service.GetName(),
				"namespace", service.GetNamespace(),
			)
			continue
		}

		if isHeadlessService(service) {
			// Generate outbound listeners for every endpoint of services.
			for _, endpoint := range endpoints[serviceTag] {
				if endpoint.Address == pod.Status.PodIP {
					continue // ignore generating outbound for itself, otherwise we've got a conflict with inbound
				}
				outbounds = append(outbounds, &mesh_proto.Dataplane_Networking_Outbound{
					Address: endpoint.Address,
					Port:    endpoint.Port,
					Tags: map[string]string{
						mesh_proto.ServiceTag:  serviceTag,
						mesh_proto.InstanceTag: endpoint.Instance,
					},
				})
			}
		} else {
			// generate outbound based on ClusterIP. Transparent Proxy will work only if DNS name that resolves to ClusterIP is used
			outbounds = append(outbounds, &mesh_proto.Dataplane_Networking_Outbound{
				Address: service.Spec.ClusterIP,
				Port:    port,
				Tags: map[string]string{
					mesh_proto.ServiceTag: serviceTag,
				},
			})
		}
	}
	return outbounds, nil
}

func isHeadlessService(svc *kube_core.Service) bool {
	return svc.Spec.ClusterIP == kube_core.ClusterIPNone
}

// Services of ExternalName type should not have any selectors.
// Kubernetes does not validate this, so in rare cases, a service of
// ExternalName type could point to a workload inside the mesh. If this
// happens, we will add the service to the VIPs config map, but we will
// not be able to obtain its IP address. As a result, the key in the map
// will be incorrect (e.g., "1:"). We do not currently support
// ExternalName services, so we can safely skip them from processing.
func isExternalNameService(svc *kube_core.Service) bool {
	return svc != nil && svc.Spec.Type == kube_core.ServiceTypeExternalName
}

func isServiceLess(port uint32) bool {
	return port == mesh_proto.TCPPortReserved
}

func k8sService(ctx context.Context, serviceTag string, client kube_client.Reader) (*kube_core.Service, uint32, error) {
	name, ns, port, err := parseService(serviceTag)
	if err != nil {
		return nil, 0, errors.Wrapf(err, "failed to parse `service` host %q as FQDN", serviceTag)
	}
	if isServiceLess(port) {
		return nil, port, nil
	}

	svc := &kube_core.Service{}
	svcKey := kube_client.ObjectKey{Namespace: ns, Name: name}
	if err := client.Get(ctx, svcKey, svc); err != nil {
		return nil, 0, errors.Wrapf(err, "failed to get Service %q", svcKey)
	}
	return svc, port, nil
}

func parseService(host string) (string, string, uint32, error) {
	// split host into <name>_<namespace>_svc_<port>
	segments := strings.Split(host, "_")

	var port uint32
	switch len(segments) {
	case 4:
		p, err := strconv.ParseInt(segments[3], 10, 32)
		if err != nil {
			return "", "", 0, err
		}
		port = uint32(p)
	case 3:
		// service less service names have no port, so we just put the reserved
		// one here to note that this service is actually
		port = mesh_proto.TCPPortReserved
	default:
		return "", "", 0, errors.Errorf("service tag in unexpected format")
	}

	name, namespace := segments[0], segments[1]
	return name, namespace, port, nil
}
