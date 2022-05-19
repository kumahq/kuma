package dataplane

import (
	"github.com/pkg/errors"

	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
	core_mesh "github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
)

func AdditionalInbounds(dataplane *core_mesh.DataplaneResource, mesh *core_mesh.MeshResource) ([]*mesh_proto.Dataplane_Networking_Inbound, error) {
	var inbounds []*mesh_proto.Dataplane_Networking_Inbound
	inbound, err := PrometheusInbound(dataplane, mesh)
	if err != nil {
		return nil, err
	}
	if inbound != nil {
		inbounds = append(inbounds, inbound)
	}
	return inbounds, nil
}

func PrometheusInbound(dataplane *core_mesh.DataplaneResource, mesh *core_mesh.MeshResource) (*mesh_proto.Dataplane_Networking_Inbound, error) {
	cfg, err := dataplane.GetPrometheusConfig(mesh)
	if err != nil {
		return nil, errors.Wrap(err, "could not parse prometheus config")
	} else if cfg != nil { // metrics are on
		inbound := &mesh_proto.Dataplane_Networking_Inbound{
			Address:     dataplane.Spec.GetNetworking().GetAddress(),
			Port:        cfg.Port,
			ServicePort: 0, // this should be Admin API port. For now it does not matter what is value here. If needed we can extract this from the DataplaneMetadata
			Tags:        cfg.Tags,
		}
		return inbound, nil
	}
	return nil, nil
}

func AdditionalServices(dataplane *core_mesh.DataplaneResource, mesh *core_mesh.MeshResource) ([]string, error) {
	var services []string
	inbounds, err := AdditionalInbounds(dataplane, mesh)
	if err != nil {
		return nil, err
	}
	for _, inbound := range inbounds {
		services = append(services, inbound.Tags[mesh_proto.ServiceTag])
	}
	return services, nil
}
