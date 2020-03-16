package mesh

import (
	"regexp"

	"github.com/Kong/kuma/pkg/core/validators"
)

var nameMeshRegexp = regexp.MustCompile("^[0-9a-z-_]*$")

func ValidateMeta(name, mesh string) validators.ValidationError {
	var err validators.ValidationError
	if !nameMeshRegexp.MatchString(name) {
		err.AddViolation("name", "invalid characters. Valid characters are numbers, lowercase latin letters and '-', '_' symbols.")
	}
	if !nameMeshRegexp.MatchString(mesh) {
		err.AddViolation("mesh", "invalid characters. Valid characters are numbers, lowercase latin letters and '-', '_' symbols.")
	}
	return err
}
