package controllers

import (
	"context"
	"sort"
	"strings"

	"github.com/pkg/errors"
	kube_core "k8s.io/api/core/v1"
	kube_client "sigs.k8s.io/controller-runtime/pkg/client"

	mesh_proto "github.com/Kong/kuma/api/mesh/v1alpha1"
	mesh_k8s "github.com/Kong/kuma/pkg/plugins/resources/k8s/native/api/v1alpha1"
	injector_metadata "github.com/Kong/kuma/pkg/plugins/runtime/k8s/webhooks/injector/metadata"
)

func (p *PodConverter) OutboundInterfacesFor(pod *kube_core.Pod, others []*mesh_k8s.Dataplane) ([]*mesh_proto.Dataplane_Networking_Outbound, error) {
	var outbounds []*mesh_proto.Dataplane_Networking_Outbound
	directAccessServices := directAccessServices(pod)
	endpoints := endpointsByService(others)
	for _, serviceTag := range endpoints.Services() {
		service, port, err := p.k8sService(serviceTag)
		if err != nil {
			converterLog.Error(err, "could not get K8S Service for service tag")
			continue // one invalid Dataplane definition should not break the entire mesh
		}
		if isHeadlessService(service) {
			directAccessServices[serviceTag] = true
		} else {
			// generate outbound based on ClusterIP. Transparent Proxy will work only if DNS name that resolves to ClusterIP is used
			outbounds = append(outbounds, &mesh_proto.Dataplane_Networking_Outbound{
				Address: service.Spec.ClusterIP,
				Port:    port,
				Service: serviceTag,
			})
		}
	}

	directAccessOutbounds := directAccessOutbounds(directAccessServices, endpoints)
	return append(outbounds, directAccessOutbounds...), nil
}

func directAccessServices(pod *kube_core.Pod) map[string]bool {
	result := map[string]bool{}
	servicesRaw := pod.GetAnnotations()[injector_metadata.KumaDirectAccess]
	services := strings.Split(servicesRaw, ",")
	for _, service := range services {
		result[service] = true
	}
	return result
}

// Generate outbound listeners for every endpoint of services.
// This will enable consuming applications via transparent proxy by PodIP instead of ClusterIP of its service
// Generating listener for every endpoint will cause XDS snapshot to be huge therefore it should be used only if really needed
func directAccessOutbounds(services map[string]bool, endpointsByService EndpointsByService) []*mesh_proto.Dataplane_Networking_Outbound {
	var sortedServices []string // service should be sorted so we generate consistent every time
	if services[injector_metadata.KumaDirectAccessAll] {
		sortedServices = endpointsByService.Services()
	} else {
		sortedServices = stringSetToSortedList(services)
	}

	var outbounds []*mesh_proto.Dataplane_Networking_Outbound
	for _, service := range sortedServices {
		// services that are not found will be ignored
		for _, endpoint := range endpointsByService[service] {
			outbounds = append(outbounds, &mesh_proto.Dataplane_Networking_Outbound{
				Address: endpoint.Address,
				Port:    endpoint.Port,
				Service: service,
			})
		}
	}
	return outbounds
}

func isHeadlessService(svc *kube_core.Service) bool {
	return svc.Spec.ClusterIP == "None"
}

func (p *PodConverter) k8sService(serviceTag string) (*kube_core.Service, uint32, error) {
	host, port, err := mesh_proto.ServiceTagValue(serviceTag).HostAndPort()
	if err != nil {
		return nil, 0, errors.Errorf("failed to parse `service` tag %q", serviceTag)
	}
	name, ns, err := ParseServiceFQDN(host)
	if err != nil {
		return nil, 0, errors.Errorf("failed to parse `service` host %q as FQDN", host)
	}

	svc := &kube_core.Service{}
	svcKey := kube_client.ObjectKey{Namespace: ns, Name: name}
	if err := p.ServiceGetter.Get(context.Background(), svcKey, svc); err != nil {
		return nil, 0, errors.Wrapf(err, "failed to get Service %q", svcKey)
	}
	return svc, port, nil
}

func ParseServiceFQDN(host string) (name string, namespace string, err error) {
	// split host into <name>.<namespace>.svc
	segments := strings.Split(host, ".")
	if len(segments) != 3 {
		return "", "", errors.Errorf("service tag in unexpected format")
	}
	name, namespace = segments[0], segments[1]
	return
}

func stringSetToSortedList(set map[string]bool) []string {
	list := make([]string, 0, len(set))
	for key := range set {
		list = append(list, key)
	}
	sort.Strings(list)
	return list
}
