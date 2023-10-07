package validators

import (
	"fmt"
	"regexp"

	common_api "github.com/kumahq/kuma/api/common/v1alpha1"
	"github.com/kumahq/kuma/pkg/core/validators"
)

var (
	nameCharacterSet     = regexp.MustCompile("^[0-9a-z-_]*$")
	tagNameCharacterSet  = regexp.MustCompile(`^[a-zA-Z0-9\.\-_:/]*$`)
	tagValueCharacterSet = regexp.MustCompile(`^([a-zA-Z0-9\.\-_:/]*|\*)$`)
)

type ValidateTargetRefOpts struct {
	SupportedKinds []common_api.TargetRefKind
}

func ValidateTargetRef(
	ref common_api.TargetRef,
	opts *ValidateTargetRefOpts,
) validators.ValidationError {
	verr := validators.ValidationError{}
	if ref.Kind == "" {
		verr.AddViolation("kind", "must be set")
		return verr
	}
	if !contains(opts.SupportedKinds, ref.Kind) {
		verr.AddViolation("kind", "value is not supported")
	} else {
		refKind := ref.Kind
		switch refKind {
		case common_api.Mesh:
			if ref.Name != "" {
				verr.AddViolation("name", fmt.Sprintf("using name with kind %v is not yet supported", refKind))
			}
			verr.Add(disallowedField("mesh", ref.Mesh, refKind))
			verr.Add(disallowedField("tags", ref.Tags, refKind))
		case common_api.MeshSubset:
			verr.Add(disallowedField("name", ref.Name, refKind))
			verr.Add(disallowedField("mesh", ref.Mesh, refKind))
			verr.Add(validTags(ref.Tags))
		case common_api.MeshService, common_api.MeshHTTPRoute:
			verr.Add(requiredField("name", ref.Name, refKind))
			verr.Add(validName(ref.Name))
			verr.Add(disallowedField("mesh", ref.Mesh, refKind))
			verr.Add(disallowedField("tags", ref.Tags, refKind))
		case common_api.MeshServiceSubset, common_api.MeshGateway:
			verr.Add(requiredField("name", ref.Name, refKind))
			verr.Add(validName(ref.Name))
			verr.Add(disallowedField("mesh", ref.Mesh, refKind))
			verr.Add(validTags(ref.Tags))
		}
	}
	return verr
}

func validTags(tags map[string]string) validators.ValidationError {
	res := validators.ValidationError{}
	for key, value := range tags {
		if key == "" {
			res.Violations = append(res.Violations, validators.Violation{
				Field: "tags", Message: "tag name must be non-empty",
			})
		}
		if !tagNameCharacterSet.MatchString(key) {
			res.Violations = append(res.Violations, validators.Violation{
				Field:   "tags",
				Message: "tag name must consist of alphanumeric characters, dots, dashes, slashes and underscores",
			})
		}
		if value == "" {
			res.Violations = append(res.Violations, validators.Violation{
				Field: "tags", Message: "tag value must be non-empty",
			})
		}
		if !tagValueCharacterSet.MatchString(value) {
			res.Violations = append(res.Violations, validators.Violation{
				Field:   "tags",
				Message: `tag value must consist of alphanumeric characters, dots, dashes, slashes and underscores or be "*"`,
			})
		}
	}
	return res
}

func validName(value string) validators.ValidationError {
	res := validators.ValidationError{}
	if !nameCharacterSet.MatchString(value) {
		res.Violations = append(res.Violations, validators.Violation{
			Field:   "name",
			Message: "invalid characters. Valid characters are numbers, lowercase latin letters and '-', '_' symbols.",
		})
	}
	return res
}

func contains(array []common_api.TargetRefKind, item common_api.TargetRefKind) bool {
	for _, it := range array {
		if it == item {
			return true
		}
	}
	return false
}

func disallowedField(name string, value interface{}, kind common_api.TargetRefKind) validators.ValidationError {
	res := validators.ValidationError{}
	if isSet(value) {
		res.Violations = append(res.Violations, validators.Violation{
			Field: name, Message: fmt.Sprintf("cannot be set with kind %v", kind),
		})
	}
	return res
}

func requiredField(name string, value interface{}, kind common_api.TargetRefKind) validators.ValidationError {
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
