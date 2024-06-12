package v1alpha1

import (
	"fmt"
	"strconv"
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
