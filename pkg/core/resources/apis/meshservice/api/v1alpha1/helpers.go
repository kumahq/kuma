package v1alpha1

import (
	"fmt"
	"strconv"
	"strings"

	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
)

func (m *MeshServiceResource) DestinationName(port uint32) string {
	return fmt.Sprintf("%s_svc_%d", strings.ReplaceAll(m.GetMeta().GetName(), ".", "_"), port)
}

func (m *MeshServiceResource) FindPort(port uint32) (Port, bool) {
	for _, p := range m.Spec.Ports {
		if p.Port == port {
			return p, true
		}
	}
	return Port{}, false
}

func (m *MeshServiceResource) IsLocalMeshService(localZone string) bool {
	if len(m.GetMeta().GetLabels()) == 0 {
		return true // no labels mean that it's a local resource
	}
	resZone, ok := m.GetMeta().GetLabels()[mesh_proto.ZoneTag]
	if !ok {
		return true // no zone label mean that it's a local resource
	}
	return resZone == localZone
}

func PortFromDestinationName(destination string) uint32 {
	np := strings.Split(destination, "_svc_")
	if len(np) != 2 {
		return 0
	}
	port, err := strconv.Atoi(np[1])
	if err != nil {
		return 0
	}
	return uint32(port)
}