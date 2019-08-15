package controllers

import (
	"fmt"

	core_model "github.com/Kong/konvoy/components/konvoy-control-plane/pkg/core/resources/model"

	mesh_proto "github.com/Kong/konvoy/components/konvoy-control-plane/api/mesh/v1alpha1"
	injector_metadata "github.com/Kong/konvoy/components/konvoy-control-plane/app/konvoy-injector/pkg/injector/metadata"
	util_k8s "github.com/Kong/konvoy/components/konvoy-control-plane/pkg/plugins/discovery/k8s/util"
	mesh_k8s "github.com/Kong/konvoy/components/konvoy-control-plane/pkg/plugins/resources/k8s/native/api/v1alpha1"
	util_proto "github.com/Kong/konvoy/components/konvoy-control-plane/pkg/util/proto"

	kube_core "k8s.io/api/core/v1"
)

func PodToDataplane(pod *kube_core.Pod, services []*kube_core.Service, dataplane *mesh_k8s.Dataplane) error {
	// pick a Mesh
	dataplane.Mesh = MeshFor(pod)

	// auto-generate Dataplane definition
	dataplaneProto, err := DataplaneFor(pod, services)
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
	if mesh := pod.Labels[mesh_k8s.MeshLabel]; mesh != "" {
		return mesh
	}
	return core_model.DefaultMesh
}

func DataplaneFor(pod *kube_core.Pod, services []*kube_core.Service) (*mesh_proto.Dataplane, error) {
	dataplane := &mesh_proto.Dataplane{
		Networking: &mesh_proto.Dataplane_Networking{},
	}
	if injector_metadata.HasTransparentProxyingEnabled(pod) {
		dataplane.Networking.TransparentProxying = &mesh_proto.Dataplane_Networking_TransparentProxying{
			RedirectPort: injector_metadata.GetTransparentProxyingPort(pod),
		}
	}
	for _, svc := range services {
		for _, svcPort := range svc.Spec.Ports {
			if svcPort.Protocol != "" && svcPort.Protocol != kube_core.ProtocolTCP {
				// ignore non-TCP ports
				continue
			}
			containerPort, err := util_k8s.FindPort(pod, &svcPort)
			if err != nil {
				// ignore those cases where a Pod doesn't have all the ports a Service has
				continue
			}

			iface := mesh_proto.InboundInterface{
				WorkloadAddress: pod.Status.PodIP,
				WorkloadPort:    uint32(containerPort),
				ServicePort:     uint32(svcPort.Port),
			}
			tags := InboundTagsFor(pod, svc)

			dataplane.Networking.Inbound = append(dataplane.Networking.Inbound, &mesh_proto.Dataplane_Networking_Inbound{
				Interface: iface.String(),
				Tags:      tags,
			})
		}
	}
	return dataplane, nil
}

func InboundTagsFor(pod *kube_core.Pod, svc *kube_core.Service) map[string]string {
	tags := util_k8s.CopyStringMap(pod.Labels)
	if tags == nil {
		tags = make(map[string]string)
	}
	tags[mesh_proto.ServiceTag] = ServiceTagFor(svc)
	return tags
}

func ServiceTagFor(svc *kube_core.Service) string {
	return fmt.Sprintf("%s.%s.svc", svc.Name, svc.Namespace)
}
