package mesh

import mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"

func (es *ExternalServiceResource) IsReachableFromZone(zone string) bool {
	return es.Spec.Tags[mesh_proto.ZoneTag] == "" || es.Spec.Tags[mesh_proto.ZoneTag] == zone
}
