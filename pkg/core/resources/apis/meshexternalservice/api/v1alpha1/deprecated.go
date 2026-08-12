package v1alpha1

import (
	"fmt"

	"github.com/kumahq/kuma/v3/pkg/core/kri"
	"github.com/kumahq/kuma/v3/pkg/core/resources/sni"
)

func (t *MeshExternalServiceResource) Deprecations() []string {
	var deprecations []string

	portName := t.Spec.Match.GetName()
	id := kri.WithSectionName(kri.From(t), portName)
	for _, err := range sni.ValidateKRI(id) {
		deprecations = append(deprecations, fmt.Sprintf(
			"Invalid %s SNI (port %q): %s. This is deprecated.",
			MeshExternalServiceResourceTypeDescriptor.Name, portName, err))
	}

	return deprecations
}
