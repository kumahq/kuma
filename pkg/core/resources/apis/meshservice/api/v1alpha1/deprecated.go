package v1alpha1

import (
	"fmt"

	"github.com/kumahq/kuma/v3/pkg/core/kri"
	"github.com/kumahq/kuma/v3/pkg/core/resources/model"
	"github.com/kumahq/kuma/v3/pkg/core/resources/sni"
)

func (t *MeshServiceResource) Deprecations() []string {
	var deprecations []string

	name := model.GetDisplayName(t.GetMeta())

	// `spec.selector.dataplaneTags` was removed in 3.0.0, so a MeshService that
	// only carried it now deserializes with an empty selector and silently
	// matches nothing. Warn instead of letting it fail quietly.
	if t.Spec.Selector.DataplaneRef == nil && t.Spec.Selector.DataplaneLabels == nil {
		deprecations = append(deprecations, fmt.Sprintf(
			"%s resource '%s' has no selector, so it matches no data plane proxies. Set 'spec.selector.dataplaneRef' or 'spec.selector.dataplaneLabels'. The 'spec.selector.dataplaneTags' selector was removed, migrate it to 'dataplaneLabels'.",
			MeshServiceResourceTypeDescriptor.Name, name))
	}

	base := kri.From(t)
	for _, port := range t.Spec.Ports {
		for _, err := range sni.ValidateKRI(kri.WithSectionName(base, port.GetName())) {
			deprecations = append(deprecations, fmt.Sprintf(
				"Invalid %s SNI (port %q): %s. This is deprecated.",
				MeshServiceResourceTypeDescriptor.Name, port.GetName(), err))
		}
	}

	return deprecations
}
