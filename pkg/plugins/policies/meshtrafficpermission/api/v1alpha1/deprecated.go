package v1alpha1

import "github.com/kumahq/kuma/api/common/v1alpha1"

func (t *MeshTrafficPermissionResource) Deprecations() []string {
	for _, f := range t.Spec.From {
		if f.GetTargetRef().Kind == v1alpha1.MeshService {
			return []string{"MeshService value for 'from[].targetRef.kind' is deprecated, use MeshSubset with 'kuma.io/service' instead"}
		}
	}
	return nil
}
