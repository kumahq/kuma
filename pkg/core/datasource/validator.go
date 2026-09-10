package datasource

import (
	system_proto "github.com/kumahq/kuma/v3/api/system/v1alpha1"
	"github.com/kumahq/kuma/v3/pkg/core/validators"
)

func Validate(source *system_proto.DataSource) validators.ValidationError {
	verr := validators.ValidationError{}
	if !source.IsSet() {
		verr.AddViolation("", "data source has to be chosen. Available sources: secret, file, inline")
	}
	switch {
	case source.HasSecret():
		if source.GetSecret() == "" {
			verr.AddViolation("secret", "cannot be empty")
		}
	case source.HasInline():
		if len(source.GetInline().GetValue()) == 0 {
			verr.AddViolation("inline", "cannot be empty")
		}
	case source.HasFile():
		if source.GetFile() == "" {
			verr.AddViolation("file", "cannot be empty")
		}
	}
	return verr
}
