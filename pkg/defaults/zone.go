package defaults

import (
	"context"

	"github.com/go-logr/logr"
	"github.com/pkg/errors"

	kuma_cp "github.com/kumahq/kuma/pkg/config/app/kuma-cp"
	config_core "github.com/kumahq/kuma/pkg/config/core"
	"github.com/kumahq/kuma/pkg/core/resources/apis/system"
	"github.com/kumahq/kuma/pkg/core/resources/manager"
	"github.com/kumahq/kuma/pkg/core/resources/model"
	"github.com/kumahq/kuma/pkg/core/resources/store"
)

func EnsureOnlyOneZoneExists(
	ctx context.Context,
	resManager manager.ResourceManager,
	logger logr.Logger,
	cfg kuma_cp.Config,
) error {
	if cfg.Mode == config_core.Global {
		return nil // Zone creation on Zone CP is local to the specific zone
	}
	zoneName := cfg.Multizone.Zone.Name
	logger.V(1).Info("ensuring Zone resource exists", "name", zoneName)
	zones := &system.ZoneResourceList{}
	if err := resManager.List(ctx, zones); err != nil {
		return errors.Wrap(err, "cannot list zones")
	}
	exists := false
	for _, zone := range zones.Items {
		if zone.GetMeta().GetName() == zoneName {
			exists = true
		} else {
			logger.Info("detected Zone resource with different name than Zone CP name. Deleting. This might happen if you change the name of the Zone CP", "name", zoneName)
			if err := resManager.Delete(ctx, zone, store.DeleteByKey(zone.GetMeta().GetName(), model.NoMesh)); err != nil {
				return errors.Wrap(err, "cannot delete old zone")
			}
		}
	}
	if !exists {
		zone := system.NewZoneResource()
		if err := resManager.Create(ctx, zone, store.CreateByKey(zoneName, model.NoMesh)); err != nil {
			return err
		}
		logger.Info("Zone resource created", "name", zoneName)
	} else {
		logger.V(1).Info("Zone resource already exist", "name", zoneName)
	}
	return nil
}
