package system

import (
	system_proto "github.com/kumahq/kuma/api/system/v1alpha1"
	"github.com/kumahq/kuma/pkg/core/validators"
)

func ValidateDataSource(path validators.PathBuilder, ds *system_proto.DataSource) validators.ValidationError {
	verr := validators.ValidationError{}
	if ds == nil {
		return verr
	}
	if len(ds.GetInline().GetValue()) == 0 && ds.GetInlineString() == "" && ds.GetSecret() == "" && ds.GetFile() == "" {
		verr.AddViolationAt(path, "data source cannot be empty")
	}
	return verr
}
