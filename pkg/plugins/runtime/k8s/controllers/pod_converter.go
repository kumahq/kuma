package controllers

import (
	"context"
	"fmt"
	"regexp"

	"github.com/pkg/errors"
	kube_core "k8s.io/api/core/v1"
	kube_client "sigs.k8s.io/controller-runtime/pkg/client"

	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
	"github.com/kumahq/kuma/pkg/core"
	core_mesh "github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
	k8s_common "github.com/kumahq/kuma/pkg/plugins/common/k8s"
	mesh_k8s "github.com/kumahq/kuma/pkg/plugins/resources/k8s/native/api/v1alpha1"
	"github.com/kumahq/kuma/pkg/plugins/runtime/k8s/metadata"
	util_k8s "github.com/kumahq/kuma/pkg/plugins/runtime/k8s/util"
	util_proto "github.com/kumahq/kuma/pkg/util/proto"
)

var (
	converterLog          = core.Log.WithName("discovery").WithName("k8s").WithName("pod-to-dataplane-converter")
	metricsAggregateRegex = regexp.MustCompile(metadata.KumaMetricsPrometheusAggregatePattern)
)

type PodConverter struct {
	ServiceGetter       kube_client.Reader
	NodeGetter          kube_client.Reader
	ResourceConverter   k8s_common.Converter
	InboundConverter    InboundConverter
	Zone                string
	KubeOutboundsAsVIPs bool
}

func (p *PodConverter) PodToDataplane(
	ctx context.Context,
	dataplane *mesh_k8s.Dataplane,
	pod *kube_core.Pod,
	ns *kube_core.Namespace,
	services []*kube_core.Service,
	others []*mesh_k8s.Dataplane,
) error {
	dataplane.Mesh = util_k8s.MeshOfByAnnotation(pod, ns)
	dataplaneProto, err := p.dataplaneFor(ctx, pod, services, others)
	if err != nil {
		return err
	}
	dataplane.SetSpec(dataplaneProto)
	return nil
}

func (p *PodConverter) PodToIngress(ctx context.Context, zoneIngress *mesh_k8s.ZoneIngress, pod *kube_core.Pod, services []*kube_core.Service) error {
	logger := converterLog.WithValues("ZoneIngress.name", zoneIngress.Name, "Pod.name", pod.Name)
	// Start with the existing ZoneIngress spec so we won't override available services in Ingress section
	zoneIngressRes := core_mesh.NewZoneIngressResource()
	if err := p.ResourceConverter.ToCoreResource(zoneIngress, zoneIngressRes); err != nil {
		logger.Error(err, "unable to convert ZoneIngress k8s object into core resource")
		return err
	}

	if err := p.IngressFor(ctx, zoneIngressRes.Spec, pod, services); err != nil {
		return err
	}

	zoneIngress.SetSpec(zoneIngressRes.Spec)
	return nil
}

func (p *PodConverter) PodToEgress(ctx context.Context, zoneEgress *mesh_k8s.ZoneEgress, pod *kube_core.Pod, services []*kube_core.Service) error {
	logger := converterLog.WithValues("ZoneEgress.name", zoneEgress.Name, "Pod.name", pod.Name)
	// Start with the existing ZoneEgress spec
	zoneEgressRes := core_mesh.NewZoneEgressResource()
	if err := p.ResourceConverter.ToCoreResource(zoneEgress, zoneEgressRes); err != nil {
		logger.Error(err, "unable to convert ZoneEgress k8s object into core resource")
		return err
	}

	if err := p.EgressFor(ctx, zoneEgressRes.Spec, pod, services); err != nil {
		return err
	}

	zoneEgress.SetSpec(zoneEgressRes.Spec)
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

		tpEnabledIPMode, ipModeExists := annotations.GetStringWithDefault(metadata.IpFamilyModeDualStack,
			metadata.KumaTransparentProxyingIPFamilyMode)
		inboundPortV6, v6PortExists, _ := annotations.GetUint32(metadata.KumaTransparentProxyingInboundPortAnnotationV6)
		ipMode := mesh_proto.Dataplane_Networking_TransparentProxying_DualStack
		if ipModeExists {
			switch tpEnabledIPMode {
			case metadata.IpFamilyModeDualStack:
				ipMode = mesh_proto.Dataplane_Networking_TransparentProxying_DualStack
			case metadata.IpFamilyModeIPv4:
				ipMode = mesh_proto.Dataplane_Networking_TransparentProxying_IPv4
			case metadata.IpFamilyModeIPv6:
				fallthrough
			default:
				return nil, errors.Errorf("invalid ip family mode '%s'", ipMode)
			}
		} else if v6PortExists && inboundPortV6 == 0 {
			// an existing pod that was created before the introduction of the `KumaTransparentProxyingIPFamilyMode` annotation
			// can disable ipv6 using such annotation
			ipMode = mesh_proto.Dataplane_Networking_TransparentProxying_IPv4
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
			IpFamilyMode:          ipMode,
		}

		if directAccessServices, exist := annotations.GetList(metadata.KumaDirectAccess); exist {
			dataplane.Networking.TransparentProxying.DirectAccessServices = directAccessServices
		}
		if reachableServicesValue, exist := annotations.GetList(metadata.KumaTransparentProxyingReachableServicesAnnotation); exist {
			dataplane.Networking.TransparentProxying.ReachableServices = reachableServicesValue
			reachableServices = reachableServicesValue
		}
	}

	dataplane.Networking.Address = pod.Status.PodIP

	gwType, exist := annotations.GetString(metadata.KumaGatewayAnnotation)
	if exist {
		switch gwType {
		case "enabled":
			gateway, err := p.GatewayByServiceFor(ctx, p.Zone, pod, services)
			if err != nil {
				return nil, err
			}
			dataplane.Networking.Gateway = gateway
		case "provided":
			gateway, err := p.GatewayByDeploymentFor(ctx, p.Zone, pod, services)
			if err != nil {
				return nil, err
			}
			dataplane.Networking.Gateway = gateway
		default:
			return nil, errors.Errorf("invalid delegated gateway type '%s'", gwType)
		}
	} else {
		ifaces, err := p.InboundConverter.InboundInterfacesFor(ctx, p.Zone, pod, services)
		if err != nil {
			return nil, err
		}
		dataplane.Networking.Inbound = ifaces
	}

	if !p.KubeOutboundsAsVIPs {
		ofaces, err := p.OutboundInterfacesFor(ctx, pod, others, reachableServices)
		if err != nil {
			return nil, err
		}
		dataplane.Networking.Outbound = ofaces
	}

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

func (p *PodConverter) GatewayByServiceFor(ctx context.Context, clusterName string, pod *kube_core.Pod, services []*kube_core.Service) (*mesh_proto.Dataplane_Networking_Gateway, error) {
	interfaces, err := p.InboundConverter.InboundInterfacesFor(ctx, clusterName, pod, services)
	if err != nil {
		return nil, err
	}
	return &mesh_proto.Dataplane_Networking_Gateway{
		Type: mesh_proto.Dataplane_Networking_Gateway_DELEGATED,
		Tags: interfaces[0].Tags, // InboundInterfacesFor() returns either a non-empty list or an error
	}, nil
}

func (p *PodConverter) GatewayByDeploymentFor(ctx context.Context, clusterName string, pod *kube_core.Pod, services []*kube_core.Service) (*mesh_proto.Dataplane_Networking_Gateway, error) {
	namespace := pod.GetObjectMeta().GetNamespace()
	deployment, kind, err := p.InboundConverter.NameExtractor.Name(ctx, pod)
	if err != nil {
		return nil, err
	}
	if kind != "Deployment" {
		return p.GatewayByServiceFor(ctx, clusterName, pod, services)
	}
	return &mesh_proto.Dataplane_Networking_Gateway{
		Type: mesh_proto.Dataplane_Networking_Gateway_DELEGATED,
		Tags: map[string]string{"kuma.io/service-name": fmt.Sprintf("%s_%s_svc", deployment, namespace)},
	}, nil
}

func MetricsFor(pod *kube_core.Pod) (*mesh_proto.MetricsBackend, error) {
	path, _ := metadata.Annotations(pod.Annotations).GetString(metadata.KumaMetricsPrometheusPath)
	port, exist, err := metadata.Annotations(pod.Annotations).GetUint32(metadata.KumaMetricsPrometheusPort)
	if err != nil {
		return nil, err
	}

	aggregate, err := MetricsAggregateFor(pod)
	if err != nil {
		return nil, err
	}
	if path == "" && !exist && aggregate == nil {
		return nil, nil
	}
	cfg := &mesh_proto.PrometheusMetricsBackendConfig{
		Path:      path,
		Port:      port,
		Aggregate: aggregate,
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

func MetricsAggregateFor(pod *kube_core.Pod) ([]*mesh_proto.PrometheusAggregateMetricsConfig, error) {
	aggregateConfigNames := make(map[string]bool)
	for key := range pod.Annotations {
		matchedGroups := metricsAggregateRegex.FindStringSubmatch(key)
		if len(matchedGroups) == 3 {
			// first group is service name and second one of (port|path|enabled)
			aggregateConfigNames[matchedGroups[1]] = true
		}
	}
	if len(aggregateConfigNames) == 0 {
		return nil, nil
	}

	var aggregateConfig []*mesh_proto.PrometheusAggregateMetricsConfig
	for app := range aggregateConfigNames {
		enabled, exist, err := metadata.Annotations(pod.Annotations).GetEnabled(fmt.Sprintf(metadata.KumaMetricsPrometheusAggregateEnabled, app))
		if err != nil {
			return nil, err
		}
		if exist && !enabled {
			enabled = false
		} else {
			enabled = true
		}
		path, _ := metadata.Annotations(pod.Annotations).GetStringWithDefault("/metrics", fmt.Sprintf(metadata.KumaMetricsPrometheusAggregatePath, app))

		address, addressExists := metadata.Annotations(pod.Annotations).GetString(fmt.Sprintf(metadata.KumaMetricsPrometheusAggregateAddress, app))

		port, portExist, err := metadata.Annotations(pod.Annotations).GetUint32(fmt.Sprintf(metadata.KumaMetricsPrometheusAggregatePort, app))
		if err != nil {
			return nil, err
		}
		if !portExist && enabled {
			return nil, errors.New("port needs to be specified for metrics scraping")
		}

		config := &mesh_proto.PrometheusAggregateMetricsConfig{
			Name:    app,
			Path:    path,
			Port:    port,
			Enabled: util_proto.Bool(enabled),
		}
		if addressExists {
			config.Address = address
		}
		aggregateConfig = append(aggregateConfig, config)
	}
	return aggregateConfig, nil
}
