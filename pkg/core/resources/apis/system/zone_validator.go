package system

import (
	core_model "github.com/kumahq/kuma/v3/pkg/core/resources/model"
	"github.com/kumahq/kuma/v3/pkg/core/validators"
)

func (t *ZoneResource) Validate() error {
	var verr validators.ValidationError
	if meta := t.GetMeta(); meta != nil {
		verr.Add(validators.ValidateRFC1035Name(validators.RootedAt("name"), core_model.GetDisplayName(meta)))
	}
	return verr.OrNil()
}
