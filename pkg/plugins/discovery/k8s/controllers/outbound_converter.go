package controllers

import (
	"context"
	"strings"

	"github.com/pkg/errors"
	kube_core "k8s.io/api/core/v1"
	kube_client "sigs.k8s.io/controller-runtime/pkg/client"

	mesh_proto "github.com/Kong/kuma/api/mesh/v1alpha1"
	mesh_k8s "github.com/Kong/kuma/pkg/plugins/resources/k8s/native/api/v1alpha1"
)

func (p *PodConverter) OutboundInterfacesFor(pod *kube_core.Pod, others []*mesh_k8s.Dataplane) ([]*mesh_proto.Dataplane_Networking_Outbound, error) {
	var outbounds []*mesh_proto.Dataplane_Networking_Outbound
	endpoints := endpointsByService(others)
	for _, serviceTag := range endpoints.Services() {
		service, port, err := p.k8sService(serviceTag)
		if err != nil {
			converterLog.Error(err, "could not get K8S Service for service tag")
			continue // one invalid Dataplane definition should not break the entire mesh
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
					Service: serviceTag,
				})
			}
		} else {
			// generate outbound based on ClusterIP. Transparent Proxy will work only if DNS name that resolves to ClusterIP is used
			outbounds = append(outbounds, &mesh_proto.Dataplane_Networking_Outbound{
				Address: service.Spec.ClusterIP,
				Port:    port,
				Service: serviceTag,
			})
		}
	}
	return outbounds, nil
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
