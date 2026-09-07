package v1alpha1

import (
	"fmt"

	"github.com/kumahq/kuma/v3/pkg/core/kri"
	"github.com/kumahq/kuma/v3/pkg/core/resources/sni"
)

func (t *MeshMultiZoneServiceResource) Deprecations() []string {
	var deprecations []string

	base := kri.From(t)
	for _, port := range t.Spec.Ports {
		for _, err := range sni.ValidateKRI(kri.WithSectionName(base, port.GetName())) {
			deprecations = append(deprecations, fmt.Sprintf(
				"Invalid %s SNI (port %q): %s. This is deprecated.",
				MeshMultiZoneServiceResourceTypeDescriptor.Name, port.GetName(), err))
		}
	}

	return deprecations
}
