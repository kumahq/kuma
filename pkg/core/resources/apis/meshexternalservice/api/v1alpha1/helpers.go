package v1alpha1

import (
	"fmt"
	"strings"
)

func (m *MeshExternalServiceResource) DestinationName(port uint32) string {
    return fmt.Sprintf("%s_svc_%d", strings.ReplaceAll(m.GetMeta().GetName(), ".", "_"), port)
}
