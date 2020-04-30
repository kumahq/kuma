package controllers

import (
	"context"
	"fmt"
	"net"
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

func PodToDataplane(dataplane *mesh_k8s.Dataplane, pod *kube_core.Pod, services []*kube_core.Service,
	others []*mesh_k8s.Dataplane, serviceGetter kube_client.Reader) error {
	// pick a Mesh
	dataplane.Mesh = MeshFor(pod)

	// auto-generate Dataplane definition
	dataplaneProto, err := DataplaneFor(pod, services, others, serviceGetter)
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

func DataplaneFor(pod *kube_core.Pod, services []*kube_core.Service, others []*mesh_k8s.Dataplane, serviceGetter kube_client.Reader) (*mesh_proto.Dataplane, error) {
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

	ofaces, err := OutboundInterfacesFor(others, serviceGetter)
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
			if net.ParseIP(svc.Spec.ClusterIP) == nil {
				return nil, errors.Errorf("Kuma requires a Kubernetes Service entity associated with a Pod with a valid IP address in the ClusterIP field. Service %s.%s has a ClusterIP value of %q. At the moment Kuma does not support headless services, to continue please add the missing Service definition or - alternatively - exclude this Pod from the automatic sidecar injection by following the instructions at: https://kuma.io/docs/latest/documentation/dps-and-data-model/#kubernetes", svc.Name, svc.Namespace, svc.Spec.ClusterIP)
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
		// instead of leaving validation up to a ValidatingAdmissionWebHook.
		// We do it this way in order to provide the most descriptive error message.
		cause := "However, there are no Services that select this Pod."
		if len(services) > 0 {
			cause = "However, this Pod doesn't have any container ports that would satisfy matching Service(s)."
		}
		return nil, errors.Errorf("Kuma requires every Pod in a Mesh to be a part of at least one Service. %s", cause)
	}
	return ifaces, nil
}

func OutboundInterfacesFor(others []*mesh_k8s.Dataplane, serviceGetter kube_client.Reader) ([]*mesh_proto.Dataplane_Networking_Outbound, error) {
	var ofaces []*mesh_proto.Dataplane_Networking_Outbound

	allServiceTags := make(map[string]bool)
	for _, other := range others {
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
			allServiceTags[svc] = true
		}
	}
	for _, serviceTag := range stringSetToSortedList(allServiceTags) {
		host, port, err := mesh_proto.ServiceTagValue(serviceTag).HostAndPort()
		if err != nil {
			converterLog.Error(err, "failed to parse `service` tag value", "value", serviceTag)
			continue // one invalid Dataplane definition should not break the entire mesh
		}
		name, ns, err := ParseServiceFQDN(host)
		if err != nil {
			converterLog.Error(err, "failed to parse `service` host as FQDN", "host", host)
			continue // one invalid Dataplane definition should not break the entire mesh
		}

		svc := &kube_core.Service{}
		if err := serviceGetter.Get(context.Background(), kube_client.ObjectKey{Namespace: ns, Name: name}, svc); err != nil {
			converterLog.Error(err, "failed to get Service", "namespace", ns, "name", name)
			continue // one invalid Dataplane definition should not break the entire mesh
		}

		dataplaneIP := svc.Spec.ClusterIP
		dataplanePort := port

		ofaces = append(ofaces, &mesh_proto.Dataplane_Networking_Outbound{
			Address: dataplaneIP,
			Port:    dataplanePort,
			Service: serviceTag,
		})
	}
	return ofaces, nil
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
