package datasource

import (
	system_proto "github.com/Kong/kuma/api/system/v1alpha1"
	"github.com/Kong/kuma/pkg/core/validators"
)

func Validate(source *system_proto.DataSource) validators.ValidationError {
	verr := validators.ValidationError{}
	if source == nil || source.Type == nil {
		verr.AddViolation("", "data source has to be chosen. Available sources: secret, file, inline")
	}
	switch source.GetType().(type) {
	case *system_proto.DataSource_Secret:
		if source.GetSecret() == "" {
			verr.AddViolation("secret", "cannot be empty")
		}
	case *system_proto.DataSource_Inline:
		if len(source.GetInline().GetValue()) == 0 {
			verr.AddViolation("inline", "cannot be empty")
		}
	case *system_proto.DataSource_File:
		if source.GetFile() == "" {
			verr.AddViolation("file", "cannot be empty")
		}
	}
	return verr
}
