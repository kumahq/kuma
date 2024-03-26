package v1alpha1

import (
	"fmt"
	"strings"
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
