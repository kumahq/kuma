package controllers

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/pkg/errors"
	kube_core "k8s.io/api/core/v1"
	kube_labels "k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/types"
	kube_client "sigs.k8s.io/controller-runtime/pkg/client"

	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
	core_mesh "github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
	"github.com/kumahq/kuma/pkg/plugins/runtime/k8s/metadata"
	util_k8s "github.com/kumahq/kuma/pkg/plugins/runtime/k8s/util"
)

type InboundConverter struct {
	NameExtractor    NameExtractor
	NodeGetter       kube_client.Reader
	NodeLabelsToCopy []string
}

func inboundForService(zone string, pod *kube_core.Pod, service *kube_core.Service, nodeLabels map[string]string) []*mesh_proto.Dataplane_Networking_Inbound {
	var ifaces []*mesh_proto.Dataplane_Networking_Inbound
	for i := range service.Spec.Ports {
		svcPort := service.Spec.Ports[i]
		if svcPort.Protocol != "" && svcPort.Protocol != kube_core.ProtocolTCP {
			// ignore non-TCP ports
			continue
		}
		containerPort, container, err := util_k8s.FindPort(pod, &svcPort)
		if err != nil {
			converterLog.Error(err, "failed to find a container port in a given Pod that would match a given Service port", "namespace", pod.Namespace, "podName", pod.Name, "serviceName", service.Name, "servicePortName", svcPort.Name)
			// ignore those cases where a Pod doesn't have all the ports a Service has
			continue
		}

		tags := InboundTagsForService(zone, pod, service, &svcPort, nodeLabels)
		state := mesh_proto.Dataplane_Networking_Inbound_Ready
		health := mesh_proto.Dataplane_Networking_Inbound_Health{
			Ready: true,
		}

		// if container is not equal nil then port is explicitly defined as containerPort so we're able
		// to figure out which container implements which service. Since we know container we can check its status
		// and map it to the Dataplane health
		if container != nil {
			if cs := util_k8s.FindContainerStatus(container.Name, pod.Status.ContainerStatuses); cs != nil && !cs.Ready {
				state = mesh_proto.Dataplane_Networking_Inbound_NotReady
				health.Ready = false
			}
		}

		// also we're checking whether kuma-sidecar container is ready
		if cs := util_k8s.FindContainerOrInitContainerStatus(
			util_k8s.KumaSidecarContainerName,
			pod.Status.ContainerStatuses,
			pod.Status.InitContainerStatuses,
		); cs != nil && !cs.Ready {
			state = mesh_proto.Dataplane_Networking_Inbound_NotReady
			health.Ready = false
		}

		if pod.DeletionTimestamp != nil { // pod is in Termination state
			state = mesh_proto.Dataplane_Networking_Inbound_NotReady
			health.Ready = false
		}

		if !kube_labels.SelectorFromSet(service.Spec.Selector).Matches(kube_labels.Set(pod.Labels)) {
			state = mesh_proto.Dataplane_Networking_Inbound_Ignored
			health.Ready = false
		}

		ifaces = append(ifaces, &mesh_proto.Dataplane_Networking_Inbound{
			Port:   uint32(containerPort),
			Name:   svcPort.Name,
			Tags:   tags,
			State:  state,
			Health: &health, // write health for backwards compatibility with Kuma 2.5 and older
		})
	}

	return ifaces
}

func inboundForServiceless(zone string, pod *kube_core.Pod, name string, nodeLabels map[string]string) *mesh_proto.Dataplane_Networking_Inbound {
	// The Pod does not have any services associated with it, just get the data from the Pod itself

	// We still need that extra listener with a service because it is required in many places of the code (e.g. mTLS)
	// TCPPortReserved, is a special port that will never be allocated from the TCP/IP stack. We use it as special
	// designator that this is actually a service-less inbound.

	// NOTE: It is cleaner to implement an equivalent of Gateway which is inbound-less dataplane. However such approch
	// will create lots of code changes to account for this other type of dataplne (we already have GW and Ingress),
	// including GUI and CLI changes

	tags := InboundTagsForPod(zone, pod, name, nodeLabels)
	state := mesh_proto.Dataplane_Networking_Inbound_Ready
	health := mesh_proto.Dataplane_Networking_Inbound_Health{
		Ready: true,
	}

	for _, container := range pod.Spec.Containers {
		if container.Name != util_k8s.KumaSidecarContainerName {
			if cs := util_k8s.FindContainerStatus(container.Name, pod.Status.ContainerStatuses); cs != nil && !cs.Ready {
				state = mesh_proto.Dataplane_Networking_Inbound_NotReady
				health.Ready = false
			}
		}
	}

	// also we're checking whether kuma-sidecar container is ready
	if cs := util_k8s.FindContainerOrInitContainerStatus(
		util_k8s.KumaSidecarContainerName,
		pod.Status.ContainerStatuses,
		pod.Status.InitContainerStatuses,
	); cs != nil && !cs.Ready {
		state = mesh_proto.Dataplane_Networking_Inbound_NotReady
		health.Ready = false
	}

	return &mesh_proto.Dataplane_Networking_Inbound{
		Port:   mesh_proto.TCPPortReserved,
		Tags:   tags,
		State:  state,
		Health: &health, // write health for backwards compatibility with Kuma 2.5 and older
	}
}

func (i *InboundConverter) InboundInterfacesFor(ctx context.Context, zone string, pod *kube_core.Pod, services []*kube_core.Service) ([]*mesh_proto.Dataplane_Networking_Inbound, error) {
	nodeLabels, err := i.getNodeLabelsToCopy(ctx, pod.Spec.NodeName)
	if err != nil {
		return nil, err
	}

	var ifaces []*mesh_proto.Dataplane_Networking_Inbound
	for _, svc := range services {
		// Services of ExternalName type should not have any selectors.
		// Kubernetes does not validate this, so in rare cases, a service of
		// ExternalName type could point to a workload inside the mesh. If this
		// happens, we would incorrectly generate inbounds including
		// ExternalName service. We do not currently support ExternalName
		// services, so we can safely skip them from processing.
		if svc.Spec.Type != kube_core.ServiceTypeExternalName {
			ifaces = append(ifaces, inboundForService(zone, pod, svc, nodeLabels)...)
		}
	}

	if len(ifaces) == 0 {
		name, _, err := i.NameExtractor.Name(ctx, pod)
		if err != nil {
			return nil, err
		}

		ifaces = append(ifaces, inboundForServiceless(zone, pod, name, nodeLabels))
	}
	return ifaces, nil
}

func (i *InboundConverter) getNodeLabelsToCopy(ctx context.Context, nodeName string) (map[string]string, error) {
	if len(i.NodeLabelsToCopy) == 0 || nodeName == "" {
		return map[string]string{}, nil
	}
	node := &kube_core.Node{}
	if err := i.NodeGetter.Get(ctx, types.NamespacedName{Name: nodeName}, node); err != nil {
		return nil, errors.Wrapf(err, "unable to get Node %s", nodeName)
	}
	nodeLabels := make(map[string]string, len(i.NodeLabelsToCopy))
	for _, label := range i.NodeLabelsToCopy {
		if value, exists := node.Labels[label]; exists {
			nodeLabels[label] = value
		}
	}
	return nodeLabels, nil
}

func InboundTagsForService(zone string, pod *kube_core.Pod, svc *kube_core.Service, svcPort *kube_core.ServicePort, nodeLabels map[string]string) map[string]string {
	logger := converterLog.WithValues("pod", pod.Name, "namespace", pod.Namespace)
	tags := map[string]string{}
	var ignoredLabels []string
	for key, value := range pod.Labels {
		if value == "" {
			continue
		}
		if strings.Contains(key, "kuma.io/") {
			ignoredLabels = append(ignoredLabels, key)
			continue
		}
		tags[key] = value
	}
	if len(ignoredLabels) > 0 {
		logger.V(1).Info("ignoring internal labels when converting labels to tags", "label", strings.Join(ignoredLabels, ","))
	}

	tags[mesh_proto.KubeNamespaceTag] = pod.Namespace
	tags[mesh_proto.KubeServiceTag] = svc.Name
	tags[mesh_proto.KubePortTag] = strconv.Itoa(int(svcPort.Port))
	tags[mesh_proto.ServiceTag] = util_k8s.ServiceTag(kube_client.ObjectKeyFromObject(svc), &svcPort.Port)
	if zone != "" {
		tags[mesh_proto.ZoneTag] = zone
	}
	for key, value := range nodeLabels {
		tags[key] = value
	}
	// For provided gateway we should ignore the protocol tag
	protocol := ProtocolTagFor(svc, svcPort)
	if enabled, _, _ := metadata.Annotations(pod.Annotations).GetEnabled(metadata.KumaGatewayAnnotation); enabled && protocol != core_mesh.ProtocolTCP {
		logger.Info("ignoring non TCP appProtocol or annotation as provided gateway only supports 'tcp'", "appProtocol", protocol)
	} else {
		tags[mesh_proto.ProtocolTag] = protocol
	}
	if isHeadlessService(svc) {
		tags[mesh_proto.InstanceTag] = pod.Name
	}
	return tags
}

// ProtocolTagFor infers service protocol from a `<port>.service.kuma.io/protocol` annotation or `appProtocol`.
func ProtocolTagFor(svc *kube_core.Service, svcPort *kube_core.ServicePort) string {
	var protocolValue string
	protocolAnnotation := fmt.Sprintf("%d.service.kuma.io/protocol", svcPort.Port)

	if svcPort.AppProtocol != nil {
		protocolValue = *svcPort.AppProtocol
		// `appProtocol` can be any protocol and if we don't explicitly support
		// it, let the default below take effect
		if core_mesh.ParseProtocol(protocolValue) == core_mesh.ProtocolUnknown {
			protocolValue = ""
		}
	}

	if explicitKumaProtocol, ok := svc.Annotations[protocolAnnotation]; ok && protocolValue == "" {
		protocolValue = explicitKumaProtocol
	}

	if protocolValue == "" {
		// if `appProtocol` or `<port>.service.kuma.io/protocol` is missing or has an empty value
		// we want Dataplane to have a `protocol: tcp` tag in order to get user's attention
		protocolValue = core_mesh.ProtocolTCP
	}

	// if `<port>.service.kuma.io/protocol` field is present but has an invalid value
	// we still want Dataplane to have a `protocol: <lowercase value>` tag in order to make it clear
	// to a user that at least `<port>.service.kuma.io/protocol` has an effect
	return strings.ToLower(protocolValue)
}

func InboundTagsForPod(zone string, pod *kube_core.Pod, name string, nodeLabels map[string]string) map[string]string {
	tags := util_k8s.CopyStringMap(pod.Labels)
	for key, value := range tags {
		if value == "" {
			delete(tags, key)
		}
	}
	if tags == nil {
		tags = make(map[string]string)
	}
	tags[mesh_proto.KubeNamespaceTag] = pod.Namespace
	tags[mesh_proto.ServiceTag] = fmt.Sprintf("%s_%s_svc", name, pod.Namespace)
	if zone != "" {
		tags[mesh_proto.ZoneTag] = zone
	}
	tags[mesh_proto.ProtocolTag] = core_mesh.ProtocolTCP
	tags[mesh_proto.InstanceTag] = pod.Name
	for key, value := range nodeLabels {
		tags[key] = value
	}

	return tags
}
