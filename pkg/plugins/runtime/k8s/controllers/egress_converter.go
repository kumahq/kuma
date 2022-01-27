package controllers

import (
	"context"

	"github.com/pkg/errors"
	kube_core "k8s.io/api/core/v1"

	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
	"github.com/kumahq/kuma/pkg/plugins/runtime/k8s/metadata"
)

func (p *PodConverter) EgressFor(
	ctx context.Context, zoneEgress *mesh_proto.ZoneEgress, pod *kube_core.Pod, services []*kube_core.Service,
) error {
	if len(services) != 1 {
		return errors.Errorf("ingress should be matched by exactly one service. Matched %d services", len(services))
	}
	ifaces, err := InboundInterfacesFor(p.Zone, pod, services)
	if err != nil {
		return errors.Wrap(err, "could not generate inbound interfaces")
	}
	if len(ifaces) != 1 {
		return errors.Errorf("generated %d inbound interfaces, expected 1. Interfaces: %v", len(ifaces), ifaces)
	}

	if zoneEgress.Networking == nil {
		zoneEgress.Networking = &mesh_proto.ZoneEgress_Networking{}
	}

	zoneEgress.Networking.Address = pod.Status.PodIP
	zoneEgress.Networking.Port = ifaces[0].Port

	coords, err := p.coordinatesFromAnnotations(pod.Annotations)
	if err != nil {
		return err
	}

	if coords == nil { // if ingress public coordinates were not present in annotations we will try to pick it from service
		if _, err := p.coordinatesFromService(ctx, services[0]); err != nil {
			return err
		}
	}

	adminPort, exist, err := metadata.Annotations(pod.Annotations).GetUint32(metadata.KumaEnvoyAdminPort)
	if err != nil {
		return err
	}
	if exist {
		zoneEgress.Networking.Admin = &mesh_proto.EnvoyAdmin{Port: adminPort}
	}

	return nil
}
