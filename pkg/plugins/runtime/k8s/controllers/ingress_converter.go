package controllers

import (
	"context"

	"github.com/pkg/errors"
	kube_core "k8s.io/api/core/v1"

	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
	"github.com/kumahq/kuma/pkg/plugins/runtime/k8s/metadata"
)

// List of priority for picking IP when Service that selects ingress is of type NodePort
// We first try to find ExternalIP and then InternalIP.
// ExternalIP will be available in public clouds like GCP, but not on Kind or Minikube.
// On the other hand, on Kind with multizone, there is a connectivity between clusters using InternalIP.
// Technically there is a risk that we will pick InternalIP and other cluster will try to access it without connectivity between them.
// However, in most cases, LoadBalancer will be used anyways, therefore we accept this risk.
var NodePortAddressPriority = []kube_core.NodeAddressType{
	kube_core.NodeExternalIP,
	kube_core.NodeInternalIP,
}

func (p *PodConverter) IngressFor(
	ctx context.Context, zoneIngress *mesh_proto.ZoneIngress, pod *kube_core.Pod, services []*kube_core.Service,
) error {
	if len(services) != 1 {
		return errors.Errorf("ingress should be matched by exactly one service. Matched %d services", len(services))
	}
	ifaces, err := p.InboundConverter.InboundInterfacesFor(ctx, p.Zone, pod, services)
	if err != nil {
		return errors.Wrap(err, "could not generate inbound interfaces")
	}
	if len(ifaces) != 1 {
		return errors.Errorf("generated %d inbound interfaces, expected 1. Interfaces: %v", len(ifaces), ifaces)
	}

	if zoneIngress.Networking == nil {
		zoneIngress.Networking = &mesh_proto.ZoneIngress_Networking{}
	}

	zoneIngress.Networking.Address = pod.Status.PodIP
	zoneIngress.Networking.Port = ifaces[0].Port

	coords, err := p.coordinatesFromAnnotations(pod.Annotations)
	if err != nil {
		return err
	}

	if coords == nil { // if ingress public coordinates were not present in annotations we will try to pick it from service
		coords, err = p.coordinatesFromService(ctx, services[0])
		if err != nil {
			return err
		}
	}

	if coords != nil {
		zoneIngress.Networking.AdvertisedAddress = coords.address
		zoneIngress.Networking.AdvertisedPort = coords.port
	}

	adminPort, exist, err := metadata.Annotations(pod.Annotations).GetUint32(metadata.KumaEnvoyAdminPort)
	if err != nil {
		return err
	}
	if exist {
		zoneIngress.Networking.Admin = &mesh_proto.EnvoyAdmin{Port: adminPort}
	}
	zoneIngress.Envoy = GetEnvoyConfiguration(p.DeltaXds, metadata.Annotations(pod.Annotations))

	return nil
}

type coordinates struct {
	address string
	port    uint32
}

func (p *PodConverter) coordinatesFromAnnotations(annotations metadata.Annotations) (*coordinates, error) {
	publicAddress, addressExist := annotations.GetString(metadata.KumaIngressPublicAddressAnnotation)
	publicPort, portExist, err := annotations.GetUint32(metadata.KumaIngressPublicPortAnnotation)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to parse annotation %s", metadata.KumaIngressPublicPortAnnotation)
	}
	if addressExist != portExist {
		return nil, errors.Errorf("both %s and %s has to be defined", metadata.KumaIngressPublicAddressAnnotation, metadata.KumaIngressPublicPortAnnotation)
	}
	if addressExist && portExist {
		return &coordinates{
			address: publicAddress,
			port:    publicPort,
		}, nil
	}
	return nil, nil
}

// coordinatesFromService is trying to generate ingress with public address and port using Service that selects the ingress
func (p *PodConverter) coordinatesFromService(ctx context.Context, service *kube_core.Service) (*coordinates, error) {
	switch service.Spec.Type {
	case kube_core.ServiceTypeLoadBalancer:
		return p.coordinatesFromLoadBalancer(service)
	case kube_core.ServiceTypeNodePort:
		return p.coordinatesFromNodePort(ctx, service)
	default:
		converterLog.Info("ingress service type is not public, therefore the public coordinates of the ingress will not be automatically set. Change the ingress service to LoadBalancer or NodePort or override the settings using annotations.")
		return nil, nil
	}
}

func (p *PodConverter) coordinatesFromNodePort(ctx context.Context, service *kube_core.Service) (*coordinates, error) {
	nodes := &kube_core.NodeList{}
	if err := p.NodeGetter.List(ctx, nodes); err != nil {
		return nil, err
	}
	if len(nodes.Items) < 1 { // this should not happen, K8S always has at least one node
		return nil, errors.New("there are no nodes")
	}
	for _, addressType := range NodePortAddressPriority {
		for _, address := range nodes.Items[0].Status.Addresses {
			if address.Type == addressType {
				coords := &coordinates{
					address: address.Address,
					port:    uint32(service.Spec.Ports[0].NodePort),
				}
				return coords, nil
			}
		}
	}
	return nil, errors.New("could not find valid Node address for Ingress publicAddress")
}

func (p *PodConverter) coordinatesFromLoadBalancer(service *kube_core.Service) (*coordinates, error) {
	if len(service.Status.LoadBalancer.Ingress) == 0 {
		converterLog.V(1).Info("load balancer for ingress is not yet ready")
		return nil, nil
	}
	publicAddress := ""
	if service.Status.LoadBalancer.Ingress[0].Hostname != "" {
		publicAddress = service.Status.LoadBalancer.Ingress[0].Hostname
	}
	if service.Status.LoadBalancer.Ingress[0].IP != "" {
		publicAddress = service.Status.LoadBalancer.Ingress[0].IP
	}
	if publicAddress == "" {
		converterLog.V(1).Info("load balancer for ingress is not yet ready. Hostname and IP are empty")
		return nil, nil
	}
	coords := &coordinates{
		address: publicAddress,
		port:    uint32(service.Spec.Ports[0].Port), // service has to have port, otherwise we would not generate inbound
	}
	return coords, nil
}
