package controllers

import (
	"strings"

	"github.com/kumahq/kuma/pkg/core/resources/model"
	"github.com/kumahq/kuma/pkg/plugins/runtime/k8s/metadata"

	"github.com/pkg/errors"

	kube_core "k8s.io/api/core/v1"
	kube_client "sigs.k8s.io/controller-runtime/pkg/client"

	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
	"github.com/kumahq/kuma/pkg/core"
	mesh_k8s "github.com/kumahq/kuma/pkg/plugins/resources/k8s/native/api/v1alpha1"
	util_proto "github.com/kumahq/kuma/pkg/util/proto"
)

var (
	converterLog = core.Log.WithName("discovery").WithName("k8s").WithName("pod-to-dataplane-converter")
)

type PodConverter struct {
	ServiceGetter kube_client.Reader
	Zone          string
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

func (p *PodConverter) PodToIngress(dataplane *mesh_k8s.Dataplane, pod *kube_core.Pod, services []*kube_core.Service) error {
	dataplane.Mesh = MeshFor(pod)
	ingressProto, err := p.IngressFor(pod, services)
	if err != nil {
		return err
	}
	spec, err := util_proto.ToMap(ingressProto)
	if err != nil {
		return err
	}
	dataplane.Spec = spec
	return nil
}

func MeshFor(pod *kube_core.Pod) string {
	mesh, exist := metadata.Annotations(pod.Annotations).GetString(metadata.KumaMeshAnnotation)
	if !exist || mesh == "" {
		return model.DefaultMesh
	}
	return mesh
}

func (p *PodConverter) DataplaneFor(pod *kube_core.Pod, services []*kube_core.Service, others []*mesh_k8s.Dataplane) (*mesh_proto.Dataplane, error) {
	dataplane := &mesh_proto.Dataplane{
		Networking: &mesh_proto.Dataplane_Networking{},
	}
	annotations := metadata.Annotations(pod.Annotations)

	enabled, exist, err := annotations.GetEnabled(metadata.KumaTransparentProxyingAnnotation)
	if err != nil {
		return nil, err
	}
	if exist && enabled {
		inboundPort, exist, err := annotations.GetUint32(metadata.KumaTransparentProxyingInboundPortAnnotation)
		if err != nil {
			return nil, err
		}
		if !exist {
			return nil, errors.New("transparent proxying inbound port has to be set in transparent mode")
		}
		outboundPort, exist, err := annotations.GetUint32(metadata.KumaTransparentProxyingOutboundPortAnnotation)
		if err != nil {
			return nil, err
		}
		if !exist {
			return nil, errors.New("transparent proxying outbound port has to be set in transparent mode")
		}
		dataplane.Networking.TransparentProxying = &mesh_proto.Dataplane_Networking_TransparentProxying{
			RedirectPortInbound:  inboundPort,
			RedirectPortOutbound: outboundPort,
		}
		if services, _ := annotations.GetString(metadata.KumaDirectAccess); services != "" {
			dataplane.Networking.TransparentProxying.DirectAccessServices = strings.Split(services, ",")
		}
	}

	dataplane.Networking.Address = pod.Status.PodIP

	enabled, exist, err = annotations.GetEnabled(metadata.KumaGatewayAnnotation)
	if err != nil {
		return nil, err
	}
	if exist && enabled {
		gateway, err := GatewayFor(p.Zone, pod, services)
		if err != nil {
			return nil, err
		}
		dataplane.Networking.Gateway = gateway
	} else {
		ifaces, err := InboundInterfacesFor(p.Zone, pod, services)
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

	probes, err := ProbesFor(pod)
	if err != nil {
		return nil, err
	}
	dataplane.Probes = probes

	return dataplane, nil
}

func (p *PodConverter) IngressFor(pod *kube_core.Pod, services []*kube_core.Service) (*mesh_proto.Dataplane, error) {
	ifaces, err := InboundInterfacesFor(p.Zone, pod, services)
	if err != nil {
		return nil, errors.Wrap(err, "could not generate inbound interfaces")
	}
	if len(ifaces) != 1 {
		return nil, errors.Errorf("generated %d inbound interfaces, expected 1. Interfaces: %v", len(ifaces), ifaces)
	}

	return &mesh_proto.Dataplane{
		Networking: &mesh_proto.Dataplane_Networking{
			Ingress: &mesh_proto.Dataplane_Networking_Ingress{},
			Address: pod.Status.PodIP,
			Inbound: ifaces,
		},
	}, nil
}

func GatewayFor(clusterName string, pod *kube_core.Pod, services []*kube_core.Service) (*mesh_proto.Dataplane_Networking_Gateway, error) {
	interfaces, err := InboundInterfacesFor(clusterName, pod, services)
	if err != nil {
		return nil, err
	}
	return &mesh_proto.Dataplane_Networking_Gateway{
		Tags: interfaces[0].Tags, // InboundInterfacesFor() returns either a non-empty list or an error
	}, nil
}

func MetricsFor(pod *kube_core.Pod) (*mesh_proto.MetricsBackend, error) {
	path, _ := metadata.Annotations(pod.Annotations).GetString(metadata.KumaMetricsPrometheusPath)
	port, exist, err := metadata.Annotations(pod.Annotations).GetUint32(metadata.KumaMetricsPrometheusPort)
	if err != nil {
		return nil, err
	}
	if path == "" && !exist {
		return nil, nil
	}
	cfg := &mesh_proto.PrometheusMetricsBackendConfig{
		Path: path,
		Port: port,
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
