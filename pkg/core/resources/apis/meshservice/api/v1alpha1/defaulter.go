package v1alpha1

import (
	"strconv"

	"k8s.io/apimachinery/pkg/util/intstr"
)

func (m *MeshServiceResource) Default() error {
	for i := range m.Spec.Ports {
		if m.Spec.Ports[i].Name == "" {
			m.Spec.Ports[i].Name = strconv.Itoa(int(m.Spec.Ports[i].Port))
		}
		if m.Spec.Ports[i].TargetPort.Type == intstr.Int && m.Spec.Ports[i].TargetPort.IntVal == 0 {
			m.Spec.Ports[i].TargetPort = intstr.FromInt32(int32(m.Spec.Ports[i].Port))
		}
	}
	return nil
}
