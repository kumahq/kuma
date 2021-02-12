package mesh

import (
	"sync"

	"github.com/pkg/errors"

	"github.com/kumahq/kuma/pkg/core"
	"github.com/kumahq/kuma/pkg/core/resources/manager"
	"github.com/kumahq/kuma/pkg/tokens/builtin/issuer"
)

var log = core.Log.WithName("defaults").WithName("mesh")

// ensureMux protects concurrent EnsureDefaultMeshResources invocation.
// On Kubernetes, EnsureDefaultMeshResources is called both from MeshManager when creating default Mesh and from the MeshController
// When they run concurrently:
// 1 invocation can check that TrafficPermission is absent and then create it.
// 2 invocation can check that TrafficPermission is absent, but it was just created, so it tries to created it which results in error
var ensureMux = sync.Mutex{}

func EnsureDefaultMeshResources(resManager manager.ResourceManager, meshName string) error {
	ensureMux.Lock()
	defer ensureMux.Unlock()
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

	err, created = ensureDefaultTimeout(resManager, meshName)
	if err != nil {
		log.Error(err, "")
		return errors.Wrap(err, "could not create default Timeout")
	}
	if created {
		log.Info("default Timeout created", "mesh", meshName, "name", defaultTimeoutKey(meshName).Name)
	} else {
		log.Info("default Timeout already exist", "mesh", meshName, "name", defaultTimeoutKey(meshName).Name)
	}

	created, err = ensureDataplaneTokenSigningKey(resManager, meshName)
	if err != nil {
		return errors.Wrap(err, "could not create default Dataplane Token Signing Key")
	}
	if created {
		log.Info("default Signing Key created", "mesh", meshName, "name", issuer.SigningKeyResourceKey(issuer.DataplaneTokenPrefix, meshName).Name)
	} else {
		log.Info("default Signing Key already exist", "mesh", meshName, "name", issuer.SigningKeyResourceKey(issuer.DataplaneTokenPrefix, meshName).Name)
	}

	created, err = ensureEnvoyAdminClientSigningKey(resManager, meshName)
	if err != nil {
		return errors.Wrap(err, "could not create default Envoy Admin Client Token Signing Key")
	}
	if created {
		log.Info("default Signing Key created", "mesh", meshName, "name", issuer.SigningKeyResourceKey(issuer.EnvoyAdminClientTokenPrefix, meshName).Name)
	} else {
		log.Info("default Signing Key already exist", "mesh", meshName, "name", issuer.SigningKeyResourceKey(issuer.EnvoyAdminClientTokenPrefix, meshName).Name)
	}
	return nil
}
