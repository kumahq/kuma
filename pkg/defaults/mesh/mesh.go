package mesh

import (
	"context"
	"fmt"
	"slices"
	"sync"

	"github.com/pkg/errors"

	config_core "github.com/kumahq/kuma/v3/pkg/config/core"
	"github.com/kumahq/kuma/v3/pkg/core"
	"github.com/kumahq/kuma/v3/pkg/core/resources/apis/system"
	resource_labels "github.com/kumahq/kuma/v3/pkg/core/resources/labels"
	"github.com/kumahq/kuma/v3/pkg/core/resources/manager"
	"github.com/kumahq/kuma/v3/pkg/core/resources/model"
	"github.com/kumahq/kuma/v3/pkg/core/resources/store"
	"github.com/kumahq/kuma/v3/pkg/core/tokens"
	kuma_log "github.com/kumahq/kuma/v3/pkg/log"
)

var log = core.Log.WithName("defaults").WithName("mesh")

// ensureMux protects concurrent EnsureDefaultMeshResources invocation.
// On Kubernetes, EnsureDefaultMeshResources is called both from MeshManager when creating default Mesh and from the MeshController
// When they run concurrently:
// 1 invocation can check that a default resource is absent and then create it.
// 2 invocation can check that the same resource is absent, but it was just created, so it tries to create it which results in error
var ensureMux = sync.Mutex{}

func EnsureDefaultMeshResources(
	ctx context.Context,
	resManager manager.ResourceManager,
	mesh model.Resource,
	skippedPolicies []string,
	extensions context.Context,
	k8sStore bool,
	systemNamespace string,
	cpMode config_core.CpMode,
	cpZone string,
) error {
	ensureMux.Lock()
	defer ensureMux.Unlock()

	meshName := mesh.GetMeta().GetName()
	logger := kuma_log.AddFieldsFromCtx(log, ctx, extensions).WithValues("mesh", meshName)

	logger.Info("ensuring default resources for Mesh exist")

	created, err := ensureDataplaneTokenSigningKey(ctx, resManager, mesh)
	if err != nil {
		return errors.Wrap(err, "could not create default Dataplane Token Signing Key")
	}
	if created {
		resKey := tokens.SigningKeyResourceKey(system.DataplaneTokenSigningKey(meshName), tokens.DefaultKeyID, meshName)
		logger.Info("default Dataplane Token Signing Key created", "name", resKey.Name)
	} else {
		logger.Info("Dataplane Token Signing Key already exists")
	}
	if slices.Contains(skippedPolicies, "*") {
		logger.Info("skipping all default policy creation")
		return nil
	}

	defaultResourceBuilders := map[string]func() model.Resource{
		"mesh-timeout-all":         defaultMeshTimeoutResource,
		"mesh-timeout-to-all":      defaultMeshTimeoutToResource,
		"mesh-circuit-breaker-all": defaultMeshCircuitBreakerResource,
		"mesh-retry-all":           defaultMeshRetryResource,
	}
	for prefix, resourceBuilder := range defaultResourceBuilders {
		resourceName := fmt.Sprintf("%s-%s", prefix, meshName)
		// new policies are created in a kuma system namespace
		if k8sStore {
			resourceName = fmt.Sprintf("%s.%s", resourceName, systemNamespace)
		}
		key := model.ResourceKey{
			Mesh: meshName,
			Name: resourceName,
		}

		resource := resourceBuilder()

		var msg string
		if !slices.Contains(skippedPolicies, string(resource.Descriptor().Name)) {
			err, created := ensureDefaultResource(ctx, resManager, resource, key, cpMode, cpZone, k8sStore, systemNamespace)
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

		logger.Info(msg, "name", key.Name)
	}
	return nil
}

func ensureDefaultResource(
	ctx context.Context,
	resManager manager.ResourceManager,
	res model.Resource,
	resourceKey model.ResourceKey,
	cpMode config_core.CpMode,
	cpZone string,
	k8sStore bool,
	systemNamespace string,
) (error, bool) {
	err := resManager.Get(ctx, res, store.GetBy(resourceKey), store.GetConsistent())
	if err == nil {
		return nil, false
	}
	if !store.IsNotFound(err) {
		return errors.Wrap(err, "could not retrieve a resource"), false
	}
	namespace := resource_labels.UnsetNamespace
	if k8sStore {
		namespace = resource_labels.NewNamespace(systemNamespace, true)
	}
	desired, err := resource_labels.Compute(
		res.Descriptor(),
		res.GetSpec(),
		nil,
		resourceKey.Mesh,
		resourceKey.Name,
		resource_labels.WithMode(cpMode),
		resource_labels.WithZone(cpZone),
		resource_labels.WithK8s(k8sStore),
		resource_labels.WithNamespace(namespace),
	)
	if err != nil {
		return errors.Wrap(err, "could not compute labels for a default resource"), false
	}
	if err := resManager.Create(ctx, res, store.CreateBy(resourceKey), store.CreateWithLabels(desired)); err != nil {
		return errors.Wrap(err, "could not create a resource"), false
	}
	return nil, true
}
