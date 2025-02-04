package mesh

import (
	"regexp"

	core_model "github.com/kumahq/kuma/pkg/core/resources/model"
	"github.com/kumahq/kuma/pkg/core/validators"
)

var (
	backwardCompatRegexp = regexp.MustCompile(`^[0-9a-z-_.]*$`)
	backwardCompatErrMsg = "invalid characters. Valid characters are numbers, lowercase latin letters and '-', '_', '.' symbols."
)

var (
	identifierRegexp = regexp.MustCompile(`^[a-z0-9]([-a-z0-9]*[a-z0-9])?(\.[a-z0-9]([-a-z0-9]*[a-z0-9])?)*$`)
	identifierErrMsg = "invalid characters. A lowercase RFC 1123 subdomain must consist of lower case alphanumeric " +
		"characters, '-' or '.', and must start and end with an alphanumeric character"
)

func ValidateMeta(m core_model.ResourceMeta, scope core_model.ResourceScope) validators.ValidationError {
	var err validators.ValidationError
	err.AddError("name", validateIdentifier(m.GetName(), identifierRegexp, identifierErrMsg))
	err.Add(ValidateMesh(m.GetMesh(), scope))
	return err
}

// ValidateMesh checks that resource's mesh matches the old regex (with '_'). Even if user creates entirely new resource,
// we can't check resource's mesh against the new regex, because Mesh resource itself can be old and contain '_' in its name.
// All new Mesh resources will have their name validated against new regex.
func ValidateMesh(mesh string, scope core_model.ResourceScope) validators.ValidationError {
	var err validators.ValidationError
	if scope == core_model.ScopeMesh {
		err.AddError("mesh", validateIdentifier(mesh, backwardCompatRegexp, backwardCompatErrMsg))
	}
	return err
}

func validateIdentifier(identifier string, r *regexp.Regexp, errMsg string) validators.ValidationError {
	var err validators.ValidationError
	switch {
	case identifier == "":
		err.AddViolation("", "cannot be empty")
	case len(identifier) > 253:
		err.AddViolation("", "value length must less or equal 253")
	case !r.MatchString(identifier):
		err.AddViolation("", errMsg)
	}
	return err
}
