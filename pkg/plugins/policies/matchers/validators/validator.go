package validators

import (
	"fmt"

	common_proto "github.com/kumahq/kuma/api/common/v1alpha1"
	"github.com/kumahq/kuma/pkg/core/validators"
)

type ValidateTargetRefOpts struct {
	SupportedKinds []common_proto.TargetRef_Kind
}

func ValidateTargetRef(
	ref *common_proto.TargetRef,
	opts *ValidateTargetRefOpts,
) validators.ValidationError {
	verr := validators.ValidationError{}
	if ref == nil {
		verr.AddViolation("", "you need to set a targetRef")
		return verr
	}
	if ref.Kind == "" {
		verr.AddViolation("kind", "must be set")
		return verr
	}
	if !contains(opts.SupportedKinds, ref.GetKindEnum()) {
		verr.AddViolation("kind", "value is not supported")
	} else {
		refKind := ref.GetKindEnum()
		switch refKind {
		case common_proto.TargetRef_Mesh:
			if ref.Name != "" {
				verr.AddViolation("name", fmt.Sprintf("using name with kind %v is not yet supported", refKind))
			}
			verr.Add(disallowedField("mesh", ref.Mesh, refKind))
			verr.Add(disallowedField("tags", ref.Tags, refKind))
		case common_proto.TargetRef_MeshSubset:
			verr.Add(disallowedField("name", ref.Name, refKind))
			verr.Add(disallowedField("mesh", ref.Mesh, refKind))
		case common_proto.TargetRef_MeshService:
			verr.Add(requiredField("name", ref.Name, refKind))
			verr.Add(disallowedField("mesh", ref.Mesh, refKind))
			verr.Add(disallowedField("tags", ref.Tags, refKind))
		case common_proto.TargetRef_MeshServiceSubset:
			verr.Add(requiredField("name", ref.Name, refKind))
			verr.Add(disallowedField("mesh", ref.Mesh, refKind))
		case common_proto.TargetRef_MeshGatewayRoute:
			verr.Add(requiredField("name", ref.Name, refKind))
			verr.Add(disallowedField("mesh", ref.Mesh, refKind))
		case common_proto.TargetRef_MeshHTTPRoute:
			verr.AddViolation("kind", fmt.Sprintf("%v is not yet supported", refKind))
		}
	}
	return verr
}

func contains(array []common_proto.TargetRef_Kind, item common_proto.TargetRef_Kind) bool {
	for _, it := range array {
		if it == item {
			return true
		}
	}
	return false
}

func disallowedField(name string, value interface{}, kind common_proto.TargetRef_Kind) validators.ValidationError {
	res := validators.ValidationError{}
	if isSet(value) {
		res.Violations = append(res.Violations, validators.Violation{
			Field: name, Message: fmt.Sprintf("cannot be set with kind %v", kind),
		})
	}
	return res
}

func requiredField(name string, value interface{}, kind common_proto.TargetRef_Kind) validators.ValidationError {
	res := validators.ValidationError{}
	if !isSet(value) {
		res.Violations = append(res.Violations, validators.Violation{
			Field: name, Message: fmt.Sprintf("must be set with kind %v", kind),
		})
	}
	return res
}

func isSet(value interface{}) bool {
	if v, ok := value.(string); ok {
		return v != ""
	}
	if v, ok := value.(map[string]string); ok {
		return len(v) > 0
	}
	return false
}
