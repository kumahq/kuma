package mesh

import (
	"github.com/pkg/errors"

	"github.com/kumahq/kuma/pkg/core"
	"github.com/kumahq/kuma/pkg/core/resources/manager"
	"github.com/kumahq/kuma/pkg/tokens/builtin/issuer"
)

var log = core.Log.WithName("defaults").WithName("mesh")

func EnsureDefaultMeshResources(resManager manager.ResourceManager, meshName string) error {
	log.Info("ensuring default resources for Mesh exist", "mesh", meshName)

	err, created := ensureDefaultTrafficPermission(resManager, meshName)
	if err != nil {
		return errors.Wrap(err, "could not create default TrafficPermission")
	}
	if created {
		log.Info("default TrafficPermission created", "mesh", meshName, "name", defaultTrafficPermissionKey(meshName).Name)
	} else {
		log.Info("default TrafficPermission already exist", "mesh", meshName, "name", defaultTrafficPermissionKey(meshName).Name)
	}

	err, created = ensureDefaultTrafficRoute(resManager, meshName)
	if err != nil {
		return errors.Wrap(err, "could not create default TrafficRoute")
	}
	if created {
		log.Info("default TrafficRoute created", "mesh", meshName, "name", defaultTrafficRouteKey(meshName).Name)
	} else {
		log.Info("default TrafficRoute already exist", "mesh", meshName, "name", defaultTrafficRouteKey(meshName).Name)
	}

	err, created = ensureSigningKey(resManager, meshName)
	if err != nil {
		return errors.Wrap(err, "could not create default Signing Key")
	}
	if created {
		log.Info("default Signing Key created", "mesh", meshName, "name", issuer.SigningKeyResourceKey(meshName).Name)
	} else {
		log.Info("default Signing Key already exist", "mesh", meshName, "name", issuer.SigningKeyResourceKey(meshName).Name)
	}
	return nil
}
