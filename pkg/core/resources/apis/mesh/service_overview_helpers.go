package mesh

import mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"

func (t *ServiceOverviewResource) Status() Status {
	switch t.Spec.Status {
	case mesh_proto.ServiceInsight_Service_partially_degraded:
		return PartiallyDegraded
	case mesh_proto.ServiceInsight_Service_online:
		return Online
	default:
		return Offline
	}
}
