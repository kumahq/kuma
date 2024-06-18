package v1alpha1

import mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"

func (m *MeshExternalServiceResource) DestinationName(port uint32) string {
	return m.GetMeta().GetName()
}

func (m *MeshExternalServiceResource) IsReachableFromZone(zone string) bool {
	return m.GetMeta().GetLabels() == nil || m.GetMeta().GetLabels()[mesh_proto.ZoneTag] == "" || m.GetMeta().GetLabels()[mesh_proto.ZoneTag] == zone
}
