package mesh

import (
	"context"
	"sync"

	"github.com/pkg/errors"

	"github.com/kumahq/kuma/pkg/core"
	"github.com/kumahq/kuma/pkg/core/resources/manager"
	"github.com/kumahq/kuma/pkg/core/resources/model"
	"github.com/kumahq/kuma/pkg/core/resources/store"
	"github.com/kumahq/kuma/pkg/core/tokens"
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

	err, created := ensureDefaultResource(resManager, defaultTrafficPermissionResource, defaultTrafficPermissionKey(meshName))
	if err != nil {
		return errors.Wrap(err, "could not create default TrafficPermission")
	}
	if created {
		log.Info("default TrafficPermission created", "mesh", meshName, "name", defaultTrafficPermissionKey(meshName).Name)
	} else {
		log.Info("default TrafficPermission already exist", "mesh", meshName, "name", defaultTrafficPermissionKey(meshName).Name)
	}

	err, created = ensureDefaultResource(resManager, defaultTrafficRouteResource, defaultTrafficRouteKey(meshName))
	if err != nil {
		return errors.Wrap(err, "could not create default TrafficRoute")
	}
	if created {
		log.Info("default TrafficRoute created", "mesh", meshName, "name", defaultTrafficRouteKey(meshName).Name)
	} else {
		log.Info("default TrafficRoute already exist", "mesh", meshName, "name", defaultTrafficRouteKey(meshName).Name)
	}

	err, created = ensureDefaultResource(resManager, defaultTimeoutResource, defaultTimeoutKey(meshName))
	if err != nil {
		return errors.Wrap(err, "could not create default Timeout")
	}
	if created {
		log.Info("default Timeout created", "mesh", meshName, "name", defaultTimeoutKey(meshName).Name)
	} else {
		log.Info("default Timeout already exist", "mesh", meshName, "name", defaultTimeoutKey(meshName).Name)
	}

	err, created = ensureDefaultResource(resManager, defaultCircuitBreakerResource, defaultCircuitBreakerKey(meshName))
	if err != nil {
		return errors.Wrap(err, "could not create default CircuitBreaker")
	}
	if created {
		log.Info("default CircuitBreaker created", "mesh", meshName, "name", defaultCircuitBreakerKey(meshName).Name)
	} else {
		log.Info("default CircuitBreaker already exist", "mesh", meshName, "name", defaultCircuitBreakerKey(meshName).Name)
	}

	err, created = ensureDefaultResource(resManager, defaultRetryResource, defaultRetryKey(meshName))
	if err != nil {
		return errors.Wrap(err, "could not create default Retry")
	}
	if created {
		log.Info("default Retry created", "mesh", meshName, "name", defaultRetryKey(meshName).Name)
	} else {
		log.Info("default Retry already exist", "mesh", meshName, "name", defaultRetryKey(meshName).Name)
	}

	created, err = ensureDataplaneTokenSigningKey(resManager, meshName)
	if err != nil {
		return errors.Wrap(err, "could not create default Dataplane Token Signing Key")
	}
	if created {
		resKey := tokens.SigningKeyResourceKey(issuer.DataplaneTokenSigningKeyPrefix(meshName), tokens.DefaultSerialNumber, meshName)
		log.Info("default Dataplane Token Signing Key created", "mesh", meshName, "name", resKey.Name)
	} else {
		log.Info("Dataplane Token Signing Key already exists", "mesh", meshName)
	}
	return nil
}

func ensureDefaultResource(resManager manager.ResourceManager, res model.Resource, resourceKey model.ResourceKey) (err error, created bool) {
	err = resManager.Get(context.Background(), res, store.GetBy(resourceKey))
	if err == nil {
		return nil, false
	}
	if !store.IsResourceNotFound(err) {
		return errors.Wrap(err, "could not retrieve a resource"), false
	}
	if err := resManager.Create(context.Background(), res, store.CreateBy(resourceKey)); err != nil {
		return errors.Wrap(err, "could not create a resource"), false
	}
	return nil, true
}
