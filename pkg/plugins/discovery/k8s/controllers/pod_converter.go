package controllers

import (
	"strconv"
	"strings"

	"github.com/pkg/errors"

	kube_core "k8s.io/api/core/v1"
	kube_client "sigs.k8s.io/controller-runtime/pkg/client"

	mesh_proto "github.com/Kong/kuma/api/mesh/v1alpha1"
	"github.com/Kong/kuma/pkg/core"
	mesh_k8s "github.com/Kong/kuma/pkg/plugins/resources/k8s/native/api/v1alpha1"
	injector_metadata "github.com/Kong/kuma/pkg/plugins/runtime/k8s/webhooks/injector/metadata"
	util_proto "github.com/Kong/kuma/pkg/util/proto"
)

var (
	converterLog = core.Log.WithName("discovery").WithName("k8s").WithName("pod-to-dataplane-converter")
)

const (
	ingressPort = 10001
)

type PodConverter struct {
	ServiceGetter kube_client.Reader
	ClusterName   string
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

func (p *PodConverter) PodToIngress(dataplane *mesh_k8s.Dataplane, pod *kube_core.Pod) error {
	dataplane.Mesh = MeshFor(pod)
	ingressProto := p.IngressFor(pod)
	spec, err := util_proto.ToMap(ingressProto)
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
		services := pod.GetAnnotations()[injector_metadata.KumaDirectAccess]
		dataplane.Networking.TransparentProxying = &mesh_proto.Dataplane_Networking_TransparentProxying{
			RedirectPort:         injector_metadata.GetTransparentProxyingPort(pod),
			DirectAccessServices: strings.Split(services, ","),
		}
	}

	dataplane.Networking.Address = pod.Status.PodIP
	if injector_metadata.HasGatewayEnabled(pod) {
		gateway, err := GatewayFor(p.ClusterName, pod, services)
		if err != nil {
			return nil, err
		}
		dataplane.Networking.Gateway = gateway
	} else {
		ifaces, err := InboundInterfacesFor(p.ClusterName, pod, services, false)
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

	metrics, err := MetricsFor(pod)
	if err != nil {
		return nil, err
	}
	dataplane.Metrics = metrics

	return dataplane, nil
}

func (p *PodConverter) IngressFor(pod *kube_core.Pod) *mesh_proto.Dataplane {
	var ingressTags map[string]string

	if p.ClusterName != "" {
		ingressTags = map[string]string{
			mesh_proto.ClusterTag: p.ClusterName,
		}
	}

	return &mesh_proto.Dataplane{
		Networking: &mesh_proto.Dataplane_Networking{
			Ingress: &mesh_proto.Dataplane_Networking_Ingress{},
			Address: pod.Status.PodIP,
			Inbound: []*mesh_proto.Dataplane_Networking_Inbound{
				{
					Port: ingressPort,
					Tags: ingressTags,
				},
			},
		},
	}
}

func GatewayFor(clusterName string, pod *kube_core.Pod, services []*kube_core.Service) (*mesh_proto.Dataplane_Networking_Gateway, error) {
	interfaces, err := InboundInterfacesFor(clusterName, pod, services, true)
	if err != nil {
		return nil, err
	}
	return &mesh_proto.Dataplane_Networking_Gateway{
		Tags: interfaces[0].Tags, // InboundInterfacesFor() returns either a non-empty list or an error
	}, nil
}

func MetricsFor(pod *kube_core.Pod) (*mesh_proto.MetricsBackend, error) {
	path := pod.GetAnnotations()[injector_metadata.KumaMetricsPrometheusPath]
	port := pod.GetAnnotations()[injector_metadata.KumaMetricsPrometheusPort]
	if path == "" && port == "" {
		return nil, nil
	}

	cfg := &mesh_proto.PrometheusMetricsBackendConfig{
		Path: path,
	}
	if port != "" {
		portValue, err := strconv.ParseUint(port, 10, 32)
		if err != nil {
			return nil, errors.Wrapf(err, "could not parse port from %s annotation", injector_metadata.KumaMetricsPrometheusPort)
		}
		cfg.Port = uint32(portValue)
	}

	str, err := util_proto.ToStruct(cfg)
	if err != nil {
		return nil, err
	}
	return &mesh_proto.MetricsBackend{
		Type: mesh_proto.MetricsPrometheusType,
		Conf: &str,
	}, nil
}
