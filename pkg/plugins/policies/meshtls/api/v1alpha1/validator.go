package v1alpha1

import (
	"slices"

	common_api "github.com/kumahq/kuma/api/common/v1alpha1"
	common_tls "github.com/kumahq/kuma/api/common/v1alpha1/tls"
	"github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
	"github.com/kumahq/kuma/pkg/core/validators"
	"github.com/kumahq/kuma/pkg/util/pointer"
)

func (r *MeshTLSResource) validate() error {
	var verr validators.ValidationError
	path := validators.RootedAt("spec")
	verr.AddErrorAt(path.Field("targetRef"), validateTop(r.Spec.TargetRef))
	topLevel := pointer.DerefOr(r.Spec.TargetRef, common_api.TargetRef{Kind: common_api.Mesh, UsesSyntacticSugar: true})
	verr.AddErrorAt(path.Field("from"), validateFrom(r.Spec.From, topLevel.Kind))
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
			common_api.Dataplane,
		},
		IsInboundPolicy: true,
	})
	return targetRefErr
}

func validateFrom(from []From, topLevelTargetRef common_api.TargetRefKind) validators.ValidationError {
	var verr validators.ValidationError
	fromPath := validators.Root()

	for idx, fromItem := range from {
		path := fromPath.Index(idx)

		targetRefErr := mesh.ValidateTargetRef(fromItem.TargetRef, &mesh.ValidateTargetRefOpts{
			SupportedKinds: []common_api.TargetRefKind{
				common_api.Mesh,
			},
		})
		verr.AddErrorAt(path.Field("targetRef"), targetRefErr)
		verr.Add(validateDefault(path.Field("default"), fromItem.Default, topLevelTargetRef))
	}
	return verr
}

func validateDefault(path validators.PathBuilder, conf Conf, topLevelTargetRef common_api.TargetRefKind) validators.ValidationError {
	var verr validators.ValidationError

	if conf.Mode != nil {
		if !slices.Contains(allModes, string(*conf.Mode)) {
			verr.AddErrorAt(path.Field("mode"), validators.MakeFieldMustBeOneOfErr("mode", allModes...))
		}
	}

	if len(conf.TlsCiphers) > 0 && topLevelTargetRef != common_api.Mesh {
		verr.AddViolationAt(path.Field("tlsCiphers"), "tlsCiphers can only be defined with top level targetRef kind: Mesh")
	} else if !containsAll(common_tls.AllCiphers, conf.TlsCiphers) {
		verr.AddErrorAt(path.Field("tlsCiphers"), validators.MakeFieldMustBeOneOfErr("tlsCiphers", common_tls.AllCiphers...))
	}

	if conf.TlsVersion != nil {
		if topLevelTargetRef != common_api.Mesh {
			verr.AddViolationAt(path.Field("tlsVersion"), "tlsVersion can only be defined with top level targetRef kind: Mesh")
		} else {
			verr.AddErrorAt(path.Field("tlsVersion"), common_tls.ValidateVersion(conf.TlsVersion))
		}
	}

	return verr
}

func containsAll(main []string, sub common_tls.TlsCiphers) bool {
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
