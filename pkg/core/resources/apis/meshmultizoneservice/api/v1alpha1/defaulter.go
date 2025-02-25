package v1alpha1

import (
	"strconv"
)

func (m *MeshMultiZoneServiceResource) Default() error {
	for i := range m.Spec.Ports {
		if m.Spec.Ports[i].Name == "" {
			m.Spec.Ports[i].Name = strconv.Itoa(int(m.Spec.Ports[i].Port))
		}
	}
	return nil
}
