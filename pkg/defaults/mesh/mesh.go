package mesh

import (
	"github.com/pkg/errors"

	"github.com/kumahq/kuma/pkg/core"
	"github.com/kumahq/kuma/pkg/core/resources/manager"
)

var log = core.Log.WithName("defaults").WithName("mesh")

func CreateDefaultMeshResources(resManager manager.ResourceManager, meshName string) error {
	log.Info("creating default resources for mesh", "mesh", meshName)
	if err := createDefaultTrafficPermission(resManager, meshName); err != nil {
		return errors.Wrap(err, "could not create default traffic permission")
	}
	if err := createDefaultTrafficRoute(resManager, meshName); err != nil {
		return errors.Wrap(err, "could not create default traffic permission")
	}
	log.Info("default TrafficPermission created", "mesh", meshName, "name", defaultTrafficPermissionName(meshName))
	log.Info("creating Signing Key for mesh", "mesh", meshName)
	if err := createSigningKey(resManager, meshName); err != nil {
		return errors.Wrap(err, "could not create default signing key")
	}
	log.Info("Signing Key created", "mesh", meshName)
	return nil
}
