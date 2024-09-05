package v1alpha1

import (
	"slices"

	common_api "github.com/kumahq/kuma/api/common/v1alpha1"
	common_tls "github.com/kumahq/kuma/api/common/v1alpha1/tls"
	"github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
	"github.com/kumahq/kuma/pkg/core/validators"
)

func (r *MeshTLSResource) validate() error {
	var verr validators.ValidationError
	path := validators.RootedAt("spec")
	verr.AddErrorAt(path.Field("targetRef"), validateTop(r.Spec.TargetRef))
	verr.AddErrorAt(path.Field("from"), validateFrom(r.Spec.From))
	return verr.OrNil()
}

func validateTop(targetRef *common_api.TargetRef) validators.ValidationError {
	if targetRef == nil {
		return validators.ValidationError{}
	}
	targetRefErr := mesh.ValidateTargetRef(*targetRef, &mesh.ValidateTargetRefOpts{
		SupportedKinds: []common_api.TargetRefKind{
			common_api.Mesh,
			common_api.MeshSubset,
			common_api.MeshService,
			common_api.MeshServiceSubset,
		},
	})
	return targetRefErr
}

func validateFrom(from []From) validators.ValidationError {
	var verr validators.ValidationError
	fromPath := validators.Root()

	for idx, fromItem := range from {
		path := fromPath.Index(idx)

		defaultField := path.Field("default")
		verr.Add(validateDefault(defaultField, fromItem.Default))
	}
	return verr
}

func validateDefault(path validators.PathBuilder, conf Conf) validators.ValidationError {
	var verr validators.ValidationError

	if conf.Mode != nil {
		if !slices.Contains(allModes, string(*conf.Mode)) {
			verr.AddErrorAt(path.Field("mode"), validators.MakeFieldMustBeOneOfErr("mode", allModes...))
		}
	}

	if !containsAll(allCiphers, conf.TlsCiphers) {
		verr.AddErrorAt(path.Field("tlsCiphers"), validators.MakeFieldMustBeOneOfErr("tlsCiphers", allCiphers...))
	}

	if conf.TlsVersion != nil {
		verr.AddErrorAt(path.Field("version"), common_tls.ValidateVersion(conf.TlsVersion))
	}

	return verr
}

func containsAll(main []string, sub TlsCiphers) bool {
	elementMap := make(map[string]bool)

	for _, element := range main {
		elementMap[element] = true
	}

	for _, element := range sub {
		if !elementMap[string(element)] {
			return false
		}
	}

	return true
}
