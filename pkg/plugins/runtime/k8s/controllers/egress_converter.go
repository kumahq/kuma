package controllers

import (
	"context"

	"github.com/pkg/errors"
	kube_core "k8s.io/api/core/v1"

	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
	"github.com/kumahq/kuma/pkg/plugins/runtime/k8s/metadata"
)

func (p *PodConverter) EgressFor(
	ctx context.Context,
	zoneEgress *mesh_proto.ZoneEgress,
	pod *kube_core.Pod,
	services []*kube_core.Service,
) error {
	if len(services) != 1 {
		return errors.Errorf("egress should be matched by exactly one service. Matched %d services", len(services))
	}
	ifaces, err := p.InboundConverter.InboundInterfacesFor(ctx, p.Zone, pod, services)
	if err != nil {
		return errors.Wrap(err, "could not generate inbound interfaces")
	}
	if len(ifaces) != 1 {
		return errors.Errorf("generated %d inbound interfaces, expected 1. Interfaces: %v", len(ifaces), ifaces)
	}

	if zoneEgress.Networking == nil {
		zoneEgress.Networking = &mesh_proto.ZoneEgress_Networking{}
	}

	zoneEgress.Zone = p.Zone
	zoneEgress.Networking.Address = pod.Status.PodIP
	zoneEgress.Networking.Port = ifaces[0].Port

	adminPort, exist, err := metadata.Annotations(pod.Annotations).GetUint32(metadata.KumaEnvoyAdminPort)
	if err != nil {
		return err
	}
	if exist {
		zoneEgress.Networking.Admin = &mesh_proto.EnvoyAdmin{Port: adminPort}
	}
	zoneEgress.Envoy, err = GetEnvoyConfiguration(p.DeltaXds, metadata.Annotations(pod.Annotations))
	if err != nil {
		return err
	}

	return nil
}
