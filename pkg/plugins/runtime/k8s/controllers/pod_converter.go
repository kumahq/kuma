package controllers

import (
	"context"
	"strings"

	"github.com/pkg/errors"
	kube_core "k8s.io/api/core/v1"
	kube_client "sigs.k8s.io/controller-runtime/pkg/client"

	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
	"github.com/kumahq/kuma/pkg/core"
	k8s_common "github.com/kumahq/kuma/pkg/plugins/common/k8s"
	mesh_k8s "github.com/kumahq/kuma/pkg/plugins/resources/k8s/native/api/v1alpha1"
	"github.com/kumahq/kuma/pkg/plugins/runtime/k8s/metadata"
	util_k8s "github.com/kumahq/kuma/pkg/plugins/runtime/k8s/util"
	util_proto "github.com/kumahq/kuma/pkg/util/proto"
)

var (
	converterLog = core.Log.WithName("discovery").WithName("k8s").WithName("pod-to-dataplane-converter")
)

type PodConverter struct {
	ServiceGetter     kube_client.Reader
	NodeGetter        kube_client.Reader
	ResourceConverter k8s_common.Converter
	Zone              string
}

func (p *PodConverter) PodToDataplane(
	ctx context.Context,
	dataplane *mesh_k8s.Dataplane,
	pod *kube_core.Pod,
	ns *kube_core.Namespace,
	services []*kube_core.Service,
	others []*mesh_k8s.Dataplane,
) error {
	dataplane.Mesh = util_k8s.MeshOf(pod, ns)
	dataplaneProto, err := p.dataplaneFor(ctx, pod, services, others)
	if err != nil {
		return err
	}
	dataplane.SetSpec(dataplaneProto)
	return nil
}

func (p *PodConverter) PodToIngress(ctx context.Context, zoneIngress *mesh_k8s.ZoneIngress, pod *kube_core.Pod, services []*kube_core.Service) error {
	zoneIngressProto := &mesh_proto.ZoneIngress{}
	// Pass the current dataplane so we won't override available services in Ingress section
	if err := p.IngressFor(ctx, zoneIngressProto, pod, services); err != nil {
		return err
	}
	zoneIngress.SetSpec(zoneIngressProto)
	return nil
}

func (p *PodConverter) PodToEgress(zoneEgress *mesh_k8s.ZoneEgress, pod *kube_core.Pod, services []*kube_core.Service) error {
	zoneEgressProto := &mesh_proto.ZoneEgress{}
	// Pass the current dataplane, so we won't override available services in Egress section
	if err := p.EgressFor(zoneEgressProto, pod, services); err != nil {
		return err
	}

	zoneEgress.SetSpec(zoneEgressProto)

	return nil
}

func (p *PodConverter) dataplaneFor(
	ctx context.Context,
	pod *kube_core.Pod,
	services []*kube_core.Service,
	others []*mesh_k8s.Dataplane,
) (*mesh_proto.Dataplane, error) {
	dataplane := &mesh_proto.Dataplane{
		Networking: &mesh_proto.Dataplane_Networking{},
	}
	annotations := metadata.Annotations(pod.Annotations)

	enabled, exist, err := annotations.GetEnabled(metadata.KumaTransparentProxyingAnnotation)
	if err != nil {
		return nil, err
	}
	var reachableServices []string
	if exist && enabled {
		inboundPort, exist, err := annotations.GetUint32(metadata.KumaTransparentProxyingInboundPortAnnotation)
		if err != nil {
			return nil, err
		}
		if !exist {
			return nil, errors.New("transparent proxying inbound port has to be set in transparent mode")
		}
		inboundPortV6, _, err := annotations.GetUint32(metadata.KumaTransparentProxyingInboundPortAnnotationV6)
		if err != nil {
			return nil, err
		}

		outboundPort, exist, err := annotations.GetUint32(metadata.KumaTransparentProxyingOutboundPortAnnotation)
		if err != nil {
			return nil, err
		}
		if !exist {
			return nil, errors.New("transparent proxying outbound port has to be set in transparent mode")
		}
		dataplane.Networking.TransparentProxying = &mesh_proto.Dataplane_Networking_TransparentProxying{
			RedirectPortInbound:   inboundPort,
			RedirectPortOutbound:  outboundPort,
			RedirectPortInboundV6: inboundPortV6,
		}
		if services, _ := annotations.GetString(metadata.KumaDirectAccess); services != "" {
			dataplane.Networking.TransparentProxying.DirectAccessServices = strings.Split(services, ",")
		}
		if reachableServicesRaw, exist := annotations.GetString(metadata.KumaTransparentProxyingReachableServicesAnnotation); exist {
			reachableServices = strings.Split(reachableServicesRaw, ",")
			dataplane.Networking.TransparentProxying.ReachableServices = reachableServices
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

	ofaces, err := p.OutboundInterfacesFor(ctx, pod, others, reachableServices)
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

	adminPort, exist, err := annotations.GetUint32(metadata.KumaEnvoyAdminPort)
	if err != nil {
		return nil, err
	}
	if exist {
		dataplane.Networking.Admin = &mesh_proto.EnvoyAdmin{Port: adminPort}
	}

	return dataplane, nil
}

func GatewayFor(clusterName string, pod *kube_core.Pod, services []*kube_core.Service) (*mesh_proto.Dataplane_Networking_Gateway, error) {
	interfaces, err := InboundInterfacesFor(clusterName, pod, services)
	if err != nil {
		return nil, err
	}
	return &mesh_proto.Dataplane_Networking_Gateway{
		Type: mesh_proto.Dataplane_Networking_Gateway_DELEGATED,
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
		Conf: str,
	}, nil
}
