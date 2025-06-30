package zone

import (
	"context"

	"github.com/pkg/errors"

	"github.com/kumahq/kuma/pkg/core/resources/apis/system"
	"github.com/kumahq/kuma/pkg/core/resources/model"
	"github.com/kumahq/kuma/pkg/core/resources/store"
	"github.com/kumahq/kuma/pkg/core/validators"
)

type Validator struct {
	Store store.ResourceStore
}

func (v *Validator) ValidateDelete(ctx context.Context, name string) error {
	zi := system.NewZoneInsightResource()
	validationErr := &validators.ValidationError{}
	if err := v.Store.Get(ctx, zi, store.GetByKey(name, model.NoMesh)); err != nil {
		if store.IsNotFound(err) {
			// if ZoneInsight isn't found we allow to delete Zone since
			// there is no information about Online/Offline status
			return nil
		}
		return errors.Wrap(err, "unable to get ZoneInsight")
	}
	if zi.Spec.IsOnline() {
		validationErr.AddViolation("zone", "unable to delete Zone, Zone CP is still connected, please shut it down first")
		return validationErr
	}
	return nil
}
