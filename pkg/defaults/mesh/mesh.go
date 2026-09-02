package mesh

import (
	"context"
	"fmt"
	"maps"
	"slices"
	"sync"

	"github.com/go-logr/logr"
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
	"github.com/kumahq/kuma/v3/pkg/plugins/policies/meshtimeout/api/v1alpha1"
	"github.com/kumahq/kuma/v3/pkg/util/pointer"
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
	reconcileExistingOnly bool,
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
	// A migration that cannot complete is logged and skipped rather than
	// failing CP startup: DefaultComponent retries EnsureDefaultMeshResources
	// on a budget, and a persistently-failing migration would otherwise take
	// the whole control plane down.
	if err := migrateCombinedMeshTimeoutDefaults(ctx, resManager, meshName, k8sStore, systemNamespace, cpMode, cpZone, logger); err != nil {
		logger.Error(err, "could not migrate legacy combined default MeshTimeout resources")
	}
	if err := migrateGatewayMeshTimeoutDefaults(ctx, resManager, meshName, k8sStore, systemNamespace, logger); err != nil {
		logger.Error(err, "could not migrate legacy gateway-specific default MeshTimeout resources")
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
			err, created := ensureDefaultResource(ctx, resManager, resource, key, cpMode, cpZone, k8sStore, systemNamespace, reconcileExistingOnly)
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

// migrateCombinedMeshTimeoutDefaults strips the outbound 'to' config from the
// mesh-wide default MeshTimeout resources persisted by CP versions that
// predate the rules/to split (rules and to are mutually exclusive, see
// validator.go). Without this, an existing mesh would keep an invalid
// combined resource forever: ensureDefaultResource only heals labels on
// existing resources, and Update() re-validates the resource on every call,
// so a legacy combined resource would fail label-healing outright, and would
// keep getting rejected by KDS peers on versions that still enforce the
// same constraint.
func migrateCombinedMeshTimeoutDefaults(
	ctx context.Context,
	resManager manager.ResourceManager,
	meshName string,
	k8sStore bool,
	systemNamespace string,
	cpMode config_core.CpMode,
	cpZone string,
	logger logr.Logger,
) error {
	for _, prefix := range []string{"mesh-timeout-all"} {
		resourceName := fmt.Sprintf("%s-%s", prefix, meshName)
		if k8sStore {
			resourceName = fmt.Sprintf("%s.%s", resourceName, systemNamespace)
		}
		key := model.ResourceKey{Mesh: meshName, Name: resourceName}

		existing := v1alpha1.NewMeshTimeoutResource()
		if err := resManager.Get(ctx, existing, store.GetBy(key), store.GetConsistent()); err != nil {
			if store.IsNotFound(err) {
				continue
			}
			return errors.Wrapf(err, "could not retrieve default MeshTimeout %q", key.Name)
		}
		if len(pointer.Deref(existing.Spec.To)) == 0 {
			continue // already migrated, or an operator-modified rules-only resource
		}

		// The outbound half of the legacy combined resource is about to be
		// stripped: create its replacement first so a crash between the two
		// writes never leaves the mesh without mesh-wide outbound defaults.
		toResourceName := fmt.Sprintf("mesh-timeout-to-all-%s", meshName)
		if k8sStore {
			toResourceName = fmt.Sprintf("%s.%s", toResourceName, systemNamespace)
		}
		toKey := model.ResourceKey{Mesh: meshName, Name: toResourceName}
		if err, _ := ensureDefaultResource(ctx, resManager, defaultMeshTimeoutToResource(), toKey, cpMode, cpZone, k8sStore, systemNamespace, false); err != nil {
			return errors.Wrapf(err, "could not create default MeshTimeout %q", toKey.Name)
		}

		existing.Spec.To = nil
		// 2.14 and earlier kept the inbound half of the combined default in
		// 'from', a field the current struct no longer has, so it's silently
		// dropped on unmarshal. Restore it from the current defaults, whose
		// values are unchanged, so the migrated resource isn't left with
		// neither 'to' nor 'rules' (invalid, see validator.go).
		if len(pointer.Deref(existing.Spec.Rules)) == 0 {
			rules := defaultMeshTimeoutRules()
			existing.Spec.Rules = &rules
		}
		if err := resManager.Update(ctx, existing); err != nil {
			return errors.Wrapf(err, "could not migrate default MeshTimeout %q", key.Name)
		}
		logger.Info("migrated legacy combined default MeshTimeout, outbound defaults now live in a separate resource", "name", key.Name)
	}
	return nil
}

// migrateGatewayMeshTimeoutDefaults removes the gateway-specific default
// MeshTimeout resources persisted by CP versions that split sidecar/gateway
// defaults into separate resources. A single mesh-wide default now applies to
// every proxy type, so the gateway-specific resources are deleted rather than
// migrated. Any legacy proxyTypes restriction on the surviving
// mesh-timeout-all/mesh-timeout-to-all resources needs no migration: the
// field no longer exists on TargetRef, so it's silently dropped on the next
// unmarshal, and the resource already applies mesh-wide.
func migrateGatewayMeshTimeoutDefaults(
	ctx context.Context,
	resManager manager.ResourceManager,
	meshName string,
	k8sStore bool,
	systemNamespace string,
	logger logr.Logger,
) error {
	for _, prefix := range []string{"mesh-gateways-timeout-all", "mesh-gateways-timeout-to-all"} {
		resourceName := fmt.Sprintf("%s-%s", prefix, meshName)
		if k8sStore {
			resourceName = fmt.Sprintf("%s.%s", resourceName, systemNamespace)
		}
		key := model.ResourceKey{Mesh: meshName, Name: resourceName}
		if err := resManager.Delete(ctx, v1alpha1.NewMeshTimeoutResource(), store.DeleteBy(key)); err != nil {
			if store.IsNotFound(err) {
				continue
			}
			return errors.Wrapf(err, "could not delete legacy default MeshTimeout %q", key.Name)
		}
		logger.Info("deleted legacy gateway-specific default MeshTimeout, a single default now applies to every proxy type", "name", key.Name)
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
	reconcileExistingOnly bool,
) (error, bool) {
	computeLabels := func(spec model.ResourceSpec, existing map[string]string) (map[string]string, error) {
		namespace := resource_labels.UnsetNamespace
		if k8sStore {
			namespace = resource_labels.NewNamespace(systemNamespace, true)
		}
		return resource_labels.Compute(
			res.Descriptor(),
			spec,
			existing,
			resourceKey.Mesh,
			resourceKey.Name,
			resource_labels.WithMode(cpMode),
			resource_labels.WithZone(cpZone),
			resource_labels.WithK8s(k8sStore),
			resource_labels.WithNamespace(namespace),
		)
	}

	// Get into a fresh instance rather than the desired 'res': 'res' already
	// carries the builder's default spec, and unmarshalling merges onto
	// whatever fields are already set instead of replacing them, so reusing
	// it here could blend the desired spec with a differently-shaped stored
	// one (e.g. a legacy resource a migration failed to normalize).
	existing := res.Descriptor().NewObject()
	err := resManager.Get(ctx, existing, store.GetBy(resourceKey), store.GetConsistent())
	if err == nil {
		desired, err := computeLabels(existing.GetSpec(), existing.GetMeta().GetLabels())
		if err != nil {
			return errors.Wrap(err, "could not compute labels for a default resource"), false
		}
		if maps.Equal(existing.GetMeta().GetLabels(), desired) {
			return nil, false
		}
		// Older CP versions persisted these without computed labels. Rewrite them in place.
		if err := resManager.Update(ctx, existing, store.UpdateWithLabels(desired)); err != nil {
			return errors.Wrap(err, "could not reconcile labels of a default resource"), false
		}
		return nil, false
	}
	if !store.IsNotFound(err) {
		return errors.Wrap(err, "could not retrieve a resource"), false
	}
	if reconcileExistingOnly {
		// Boot-time reconciliation only heals labels of existing default
		// resources; it must not recreate ones an operator deleted.
		return nil, false
	}
	desired, err := computeLabels(res.GetSpec(), nil)
	if err != nil {
		return errors.Wrap(err, "could not compute labels for a default resource"), false
	}
	if err := resManager.Create(ctx, res, store.CreateBy(resourceKey), store.CreateWithLabels(desired)); err != nil {
		return errors.Wrap(err, "could not create a resource"), false
	}
	return nil, true
}
