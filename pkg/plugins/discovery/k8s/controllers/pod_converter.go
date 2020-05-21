package controllers

import (
	mesh_proto "github.com/Kong/kuma/api/mesh/v1alpha1"
	"github.com/Kong/kuma/pkg/core"
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
