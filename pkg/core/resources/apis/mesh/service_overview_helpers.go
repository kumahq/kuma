package mesh

import mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"

func (t *ServiceOverviewResource) GetStatus() Status {
	switch t.Spec.Status {
	case mesh_proto.ServiceInsight_Service_partially_degraded:
		return PartiallyDegraded
	case mesh_proto.ServiceInsight_Service_online:
		return Online
	case mesh_proto.ServiceInsight_Service_offline:
		fallthrough
	default:
		return Offline
	}
}
