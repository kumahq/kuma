package controllers

import (
	"context"
	"fmt"
	"sort"
	"strings"

	"github.com/pkg/errors"

	mesh_proto "github.com/Kong/kuma/api/mesh/v1alpha1"
	"github.com/Kong/kuma/pkg/core"
	mesh_core "github.com/Kong/kuma/pkg/core/resources/apis/mesh"
	util_k8s "github.com/Kong/kuma/pkg/plugins/discovery/k8s/util"
	mesh_k8s "github.com/Kong/kuma/pkg/plugins/resources/k8s/native/api/v1alpha1"
	injector_metadata "github.com/Kong/kuma/pkg/plugins/runtime/k8s/webhooks/injector/metadata"
	util_proto "github.com/Kong/kuma/pkg/util/proto"

	kube_core "k8s.io/api/core/v1"
	kube_client "sigs.k8s.io/controller-runtime/pkg/client"
)

var (
	converterLog = core.Log.WithName("discovery").WithName("k8s").WithName("pod-to-dataplane-converter")
)

type PodConverter struct {
	ServiceGetter kube_client.Reader
}

func (p *PodConverter) PodToDataplane(dataplane *mesh_k8s.Dataplane, pod *kube_core.Pod, services []*kube_core.Service, others []*mesh_k8s.Dataplane) error {
	dataplane.Mesh = MeshFor(pod)
	dataplaneProto, err := p.DataplaneFor(pod, services, others)
	if err != nil {
		return err
	}
	spec, err := util_proto.ToMap(dataplaneProto)
	if err != nil {
		return err
	}
	dataplane.Spec = spec
	return nil
}

func MeshFor(pod *kube_core.Pod) string {
	return injector_metadata.GetMesh(pod)
}

func (p *PodConverter) DataplaneFor(pod *kube_core.Pod, services []*kube_core.Service, others []*mesh_k8s.Dataplane) (*mesh_proto.Dataplane, error) {
	dataplane := &mesh_proto.Dataplane{
		Networking: &mesh_proto.Dataplane_Networking{},
	}
	if injector_metadata.HasTransparentProxyingEnabled(pod) {
		dataplane.Networking.TransparentProxying = &mesh_proto.Dataplane_Networking_TransparentProxying{
			RedirectPort: injector_metadata.GetTransparentProxyingPort(pod),
		}
	}

	dataplane.Networking.Address = pod.Status.PodIP
	if injector_metadata.HasGatewayEnabled(pod) {
		gateway, err := GatewayFor(pod, services)
		if err != nil {
			return nil, err
		}
		dataplane.Networking.Gateway = gateway
	} else {
		ifaces, err := InboundInterfacesFor(pod, services, false)
		if err != nil {
			return nil, err
		}
		dataplane.Networking.Inbound = ifaces
	}

	ofaces, err := p.OutboundInterfacesFor(pod, others)
	if err != nil {
		return nil, err
	}
	dataplane.Networking.Outbound = ofaces

	return dataplane, nil
}

func GatewayFor(pod *kube_core.Pod, services []*kube_core.Service) (*mesh_proto.Dataplane_Networking_Gateway, error) {
	interfaces, err := InboundInterfacesFor(pod, services, true)
	if err != nil {
		return nil, err
	}
	return &mesh_proto.Dataplane_Networking_Gateway{
		Tags: interfaces[0].Tags, // InboundInterfacesFor() returns either a non-empty list or an error
	}, nil
}

func InboundInterfacesFor(pod *kube_core.Pod, services []*kube_core.Service, isGateway bool) ([]*mesh_proto.Dataplane_Networking_Inbound, error) {
	var ifaces []*mesh_proto.Dataplane_Networking_Inbound
	for _, svc := range services {
		for _, svcPort := range svc.Spec.Ports {
			if svcPort.Protocol != "" && svcPort.Protocol != kube_core.ProtocolTCP {
				// ignore non-TCP ports
				continue
			}
			containerPort, err := util_k8s.FindPort(pod, &svcPort)
			if err != nil {
				converterLog.Error(err, "failed to find a container port in a given Pod that would match a given Service port", "namespace", pod.Namespace, "podName", pod.Name, "serviceName", svc.Name, "servicePortName", svcPort.Name)
				// ignore those cases where a Pod doesn't have all the ports a Service has
				continue
			}

			tags := InboundTagsFor(pod, svc, &svcPort, isGateway)

			ifaces = append(ifaces, &mesh_proto.Dataplane_Networking_Inbound{
				Port: uint32(containerPort),
				Tags: tags,
			})
		}
	}
	if len(ifaces) == 0 {
		// Notice that here we return an error immediately
		// instead of leaving Dataplane validation up to a ValidatingAdmissionWebHook.
		// We do it this way in order to provide the most descriptive error message.
		cause := "However, there are no Services that select this Pod."
		if len(services) > 0 {
			cause = "However, this Pod doesn't have any container ports that would satisfy matching Service(s)."
		}
		return nil, errors.Errorf("Kuma requires every Pod in a Mesh to be a part of at least one Service. %s", cause)
	}
	return ifaces, nil
}

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

type Endpoint struct {
	Address string
	Port    uint32
}

type EndpointsByService map[string][]Endpoint

func (e EndpointsByService) Services() []string {
	list := make([]string, 0, len(e))
	for key := range e {
		list = append(list, key)
	}
	sort.Strings(list)
	return list
}

func endpointsByService(dataplanes []*mesh_k8s.Dataplane) EndpointsByService {
	result := EndpointsByService{}
	for _, other := range dataplanes {
		dataplane := &mesh_proto.Dataplane{}
		if err := util_proto.FromMap(other.Spec, dataplane); err != nil {
			converterLog.Error(err, "failed to parse Dataplane", "dataplane", other.Spec)
			continue // one invalid Dataplane definition should not break the entire mesh
		}
		for _, inbound := range dataplane.Networking.GetInbound() {
			svc, ok := inbound.GetTags()[mesh_proto.ServiceTag]
			if !ok {
				continue
			}
			endpoint := Endpoint{
				Port: inbound.Port,
			}
			if inbound.Address != "" {
				endpoint.Address = inbound.Address
			} else {
				endpoint.Address = dataplane.Networking.Address
			}
			result[svc] = append(result[svc], endpoint)
		}
	}
	return result
}

func InboundTagsFor(pod *kube_core.Pod, svc *kube_core.Service, svcPort *kube_core.ServicePort, isGateway bool) map[string]string {
	tags := util_k8s.CopyStringMap(pod.Labels)
	if tags == nil {
		tags = make(map[string]string)
	}
	tags[mesh_proto.ServiceTag] = ServiceTagFor(svc, svcPort)
	// notice that in case of a gateway it might be confusing to see a protocol tag
	// since gateway proxies multiple services each with its own protocol
	if !isGateway {
		tags[mesh_proto.ProtocolTag] = ProtocolTagFor(svc, svcPort)
	}
	return tags
}

func ServiceTagFor(svc *kube_core.Service, svcPort *kube_core.ServicePort) string {
	return fmt.Sprintf("%s.%s.svc:%d", svc.Name, svc.Namespace, svcPort.Port)
}

// ProtocolTagFor infers service protocol from a `<port>.service.kuma.io/protocol` annotation.
func ProtocolTagFor(svc *kube_core.Service, svcPort *kube_core.ServicePort) string {
	protocolAnnotation := fmt.Sprintf("%d.service.kuma.io/protocol", svcPort.Port)
	protocolValue := svc.Annotations[protocolAnnotation]
	if protocolValue == "" {
		// if `<port>.service.kuma.io/protocol` annotation is missing or has an empty value
		// we want Dataplane to have a `protocol: tcp` tag in order to get user's attention
		return mesh_core.ProtocolTCP
	}
	// if `<port>.service.kuma.io/protocol` annotation is present but has an invalid value
	// we still want Dataplane to have a `protocol: <value as is>` tag in order to make it clear
	// to a user that at least `<port>.service.kuma.io/protocol` has an effect
	return protocolValue
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
