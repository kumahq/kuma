package mesh

import (
	"context"
	"fmt"
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

func EnsureDefaultMeshResources(ctx context.Context, resManager manager.ResourceManager, meshName string, skippedPolicies []string) error {
	ensureMux.Lock()
	defer ensureMux.Unlock()

	log.Info("ensuring default resources for Mesh exist", "mesh", meshName)

	created, err := ensureDataplaneTokenSigningKey(ctx, resManager, meshName)
	if err != nil {
		return errors.Wrap(err, "could not create default Dataplane Token Signing Key")
	}
	if created {
		resKey := tokens.SigningKeyResourceKey(issuer.DataplaneTokenSigningKeyPrefix(meshName), tokens.DefaultKeyID, meshName)
		log.Info("default Dataplane Token Signing Key created", "mesh", meshName, "name", resKey.Name)
	} else {
		log.Info("Dataplane Token Signing Key already exists", "mesh", meshName)
	}

	if contains(skippedPolicies, "*") {
		log.Info("skipping all default policy creation", "mesh", meshName)
		return nil
	}

	defaultResourceBuilders := map[string]func() model.Resource{
		"allow-all":           defaultTrafficPermissionResource,
		"route-all":           defaultTrafficRouteResource,
		"timeout-all":         DefaultTimeoutResource,
		"circuit-breaker-all": defaultCircuitBreakerResource,
		"retry-all":           defaultRetryResource,
	}

	for prefix, resourceBuilder := range defaultResourceBuilders {
		key := model.ResourceKey{
			Mesh: meshName,
			Name: fmt.Sprintf("%s-%s", prefix, meshName),
		}
		resource := resourceBuilder()

		var msg string
		if !contains(skippedPolicies, string(resource.Descriptor().Name)) {
			err, created := ensureDefaultResource(ctx, resManager, resource, key)
			if err != nil {
				return errors.Wrapf(err, "could not create default %s %q", resource.Descriptor().Name, key.Name)
			}

			msg = fmt.Sprintf("default %s already exists", resource.Descriptor().Name)
			if created {
				msg = fmt.Sprintf("default %s created", resource.Descriptor().Name)
			}
		} else {
			msg = fmt.Sprintf("skipping default %s creation", resource.Descriptor().Name)
		}

		log.Info(msg, "mesh", meshName, "name", key.Name)
	}

	return nil
}

func ensureDefaultResource(ctx context.Context, resManager manager.ResourceManager, res model.Resource, resourceKey model.ResourceKey) (error, bool) {
	err := resManager.Get(ctx, res, store.GetBy(resourceKey))
	if err == nil {
		return nil, false
	}
	if !store.IsResourceNotFound(err) {
		return errors.Wrap(err, "could not retrieve a resource"), false
	}
	if err := resManager.Create(ctx, res, store.CreateBy(resourceKey)); err != nil {
		return errors.Wrap(err, "could not create a resource"), false
	}
	return nil, true
}

func contains(s []string, str string) bool {
	for _, v := range s {
		if v == str {
			return true
		}
	}
	return false
}