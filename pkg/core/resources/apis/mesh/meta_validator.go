package mesh

import (
	"regexp"

	"github.com/kumahq/kuma/pkg/core/resources/model"
	"github.com/kumahq/kuma/pkg/core/validators"
)

var nameMeshRegexp = regexp.MustCompile("^[0-9a-z-_]*$")

func ValidateMeta(name, mesh string, scope model.ResourceScope) validators.ValidationError {
	var err validators.ValidationError
	if name == "" {
		err.AddViolation("name", "cannot be empty")
	}
	if !nameMeshRegexp.MatchString(name) {
		err.AddViolation("name", "invalid characters. Valid characters are numbers, lowercase latin letters and '-', '_' symbols.")
	}
	err.Add(ValidateMesh(mesh, scope))
	return err
}

func ValidateMesh(mesh string, scope model.ResourceScope) validators.ValidationError {
	var err validators.ValidationError
	if scope == model.ScopeMesh {
		if mesh == "" {
			err.AddViolation("mesh", "cannot be empty")
		}
		if !nameMeshRegexp.MatchString(mesh) {
			err.AddViolation("mesh", "invalid characters. Valid characters are numbers, lowercase latin letters and '-', '_' symbols.")
		}
	}
	return err
}
