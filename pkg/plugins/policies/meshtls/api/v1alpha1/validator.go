package v1alpha1

import (
	"slices"

	common_api "github.com/kumahq/kuma/v3/api/common/v1alpha1"
	common_tls "github.com/kumahq/kuma/v3/api/common/v1alpha1/tls"
	"github.com/kumahq/kuma/v3/pkg/core/resources/apis/mesh"
	"github.com/kumahq/kuma/v3/pkg/core/validators"
	"github.com/kumahq/kuma/v3/pkg/plugins/policies/core/rules/inbound"
	"github.com/kumahq/kuma/v3/pkg/util/pointer"
)

func (r *MeshTLSResource) validate() error {
	var verr validators.ValidationError
	path := validators.RootedAt("spec")
	verr.AddErrorAt(path.Field("targetRef"), r.validateTop(r.Spec.TargetRef, inbound.AffectsInbounds(r.Spec)))
	topLevel := r.Spec.TargetRef.ToTargetRef()
	verr.AddErrorAt(path.Field("rules"), validateRules(pointer.Deref(r.Spec.Rules), topLevel.Kind))
	return verr.OrNil()
}

func (r *MeshTLSResource) validateTop(targetRef *common_api.TopLevelTargetRef, isInboundPolicy bool) validators.ValidationError {
	if targetRef == nil {
		return validators.ValidationError{}
	}
	targetRefErr := mesh.ValidateTargetRef(targetRef.ToTargetRef(), &mesh.ValidateTargetRefOpts{
		SupportedKinds: []common_api.TargetRefKind{
			common_api.Mesh,
			common_api.Dataplane,
		},
		IsInboundPolicy: isInboundPolicy,
	})
	return targetRefErr
}

func validateRules(rules []Rule, topLevelTargetRef common_api.TargetRefKind) validators.ValidationError {
	var verr validators.ValidationError
	for idx, rulesItem := range rules {
		path := validators.Root().Index(idx)
		verr.Add(validateDefault(path.Field("default"), rulesItem.Default, topLevelTargetRef))
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

	if len(pointer.Deref(conf.TlsCiphers)) > 0 && topLevelTargetRef != common_api.Mesh {
		verr.AddViolationAt(path.Field("tlsCiphers"), "tlsCiphers can only be defined with top level targetRef kind: Mesh")
	} else if !containsAll(common_tls.AllCiphers, pointer.Deref(conf.TlsCiphers)) {
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

func containsAll(main []string, sub []common_tls.TlsCipher) bool {
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
