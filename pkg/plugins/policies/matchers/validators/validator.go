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
		// Implicitly treat nil targetRef as kind Mesh
		return verr
	}
	if !contains(opts.SupportedKinds, ref.GetKindEnum()) {
		verr.AddViolation("kind", "value is not supported")
	} else {
		switch ref.GetKindEnum() {
		case common_proto.TargetRef_Mesh:
			if len(ref.Tags) != 0 {
				verr.AddViolation("tags", fmt.Sprintf("could not be set with kind %v", common_proto.TargetRef_Mesh))
			}
			if len(ref.Mesh) != 0 {
				verr.AddViolation("mesh", fmt.Sprintf("could not be set with kind %v", common_proto.TargetRef_Mesh))
			}
		case common_proto.TargetRef_MeshSubset:
			if len(ref.Mesh) != 0 {
				verr.AddViolation("mesh", fmt.Sprintf("could not be set with kind %v", common_proto.TargetRef_MeshSubset))
			}
			if ref.Tags != nil && len(ref.Tags) == 0 {
				verr.AddViolation("tags", "cannot be empty")
			}
		case common_proto.TargetRef_MeshService:
			if len(ref.Tags) != 0 {
				verr.AddViolation("tags", fmt.Sprintf("could not be set with kind %v", common_proto.TargetRef_MeshService))
			}
			if len(ref.Name) == 0 {
				verr.AddViolation("name", "cannot be empty")
			}
		case common_proto.TargetRef_MeshServiceSubset:
			if len(ref.Name) == 0 {
				verr.AddViolation("name", "cannot be empty")
			}
			if ref.Tags != nil && len(ref.Tags) == 0 {
				verr.AddViolation("tags", "cannot be empty")
			}
		case common_proto.TargetRef_MeshGatewayRoute:
			if len(ref.Name) == 0 {
				verr.AddViolation("name", "cannot be empty")
			}
			if len(ref.Mesh) != 0 {
				verr.AddViolation("mesh", fmt.Sprintf("could not be set with kind %v", common_proto.TargetRef_MeshGatewayRoute))
			}
		case common_proto.TargetRef_MeshHTTPRoute:
			if len(ref.Name) == 0 {
				verr.AddViolation("name", "cannot be empty")
			}
			if len(ref.Mesh) != 0 {
				verr.AddViolation("mesh", fmt.Sprintf("could not be set with kind %v", common_proto.TargetRef_MeshHTTPRoute))
			}
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
