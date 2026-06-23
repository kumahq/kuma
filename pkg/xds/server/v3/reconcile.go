package v3

import (
	"context"
	"hash/fnv"
	"strconv"

	envoy_core "github.com/envoyproxy/go-control-plane/envoy/config/core/v3"
	envoy_types "github.com/envoyproxy/go-control-plane/pkg/cache/types"
	envoy_cache "github.com/envoyproxy/go-control-plane/pkg/cache/v3"
	envoy_resource "github.com/envoyproxy/go-control-plane/pkg/resource/v3"
	"github.com/pkg/errors"

	"github.com/kumahq/kuma/v3/pkg/core"
	model "github.com/kumahq/kuma/v3/pkg/core/xds"
	"github.com/kumahq/kuma/v3/pkg/util/maps"
	util_xds "github.com/kumahq/kuma/v3/pkg/util/xds"
	xds_context "github.com/kumahq/kuma/v3/pkg/xds/context"
	"github.com/kumahq/kuma/v3/pkg/xds/generator"
	"github.com/kumahq/kuma/v3/pkg/xds/generator/modifications"
	xds_hooks "github.com/kumahq/kuma/v3/pkg/xds/hooks"
	xds_metrics "github.com/kumahq/kuma/v3/pkg/xds/metrics"
	xds_sync "github.com/kumahq/kuma/v3/pkg/xds/sync"
	xds_template "github.com/kumahq/kuma/v3/pkg/xds/template"
)

var reconcileLog = core.Log.WithName("xds").WithName("reconcile")

var _ xds_sync.SnapshotReconciler = &reconciler{}

type reconciler struct {
	generator      snapshotGenerator
	cacher         snapshotCacher
	statsCallbacks util_xds.StatsCallbacks
	metrics        *xds_metrics.Metrics
}

func (r *reconciler) Clear(proxyId *model.ProxyId) error {
	nodeId := &envoy_core.Node{Id: proxyId.String()}
	r.clearUndeliveredConfigStats(nodeId)
	r.cacher.Clear(nodeId)
	return nil
}

func (r *reconciler) clearUndeliveredConfigStats(nodeId *envoy_core.Node) {
	snap, err := r.cacher.Get(nodeId)
	if err != nil {
		return // already cleared
	}
	for _, res := range snap.Resources {
		if res.Version != "" {
			r.statsCallbacks.DiscardConfig(res.Version)
		}
	}
}

func (r *reconciler) Reconcile(ctx context.Context, xdsCtx xds_context.Context, proxy *model.Proxy) (bool, error) {
	node := &envoy_core.Node{Id: proxy.Id.String()}
	snapshot, err := r.generator.GenerateSnapshot(ctx, xdsCtx, proxy)
	if err != nil {
		return false, errors.Wrapf(err, "failed to generate a snapshot")
	}

	// To avoid assigning a new version every time, compute stable content
	// hashes per resource type and compare against the previous snapshot.
	previous, err := r.cacher.Get(node)
	if err != nil {
		previous = &envoy_cache.Snapshot{}
	}

	snapshot, changed, err := autoVersion(node.Id, previous, snapshot)
	if err != nil {
		return false, errors.Wrap(err, "failed to compute snapshot versions")
	}

	resKey := proxy.Id.ToResourceKey()
	log := reconcileLog.WithValues("proxyName", resKey.Name, "mesh", resKey.Mesh)

	// Validate the resources we reconciled before sending them
	// to Envoy. This ensures that we have as much in-band error
	// information as possible, which is especially useful for tests
	// that don't actually program an Envoy instance.
	if len(changed) == 0 {
		log.V(1).Info("config is the same")
		return false, nil
	}

	for _, resources := range snapshot.Resources {
		for name, resource := range resources.Items {
			if err := validateResource(resource.Resource); err != nil {
				return false, errors.Wrapf(err, "invalid resource %q", name)
			}
		}
	}

	if err := snapshot.Consistent(); err != nil {
		return false, errors.Wrap(err, "inconsistent snapshot")
	}
	log.Info("config has changed", "versions", changed)

	if err := r.cacher.Cache(ctx, node, snapshot); err != nil {
		return false, errors.Wrap(err, "failed to store snapshot")
	}

	if r.metrics != nil {
		resourceTypeNames := [envoy_types.UnknownType]string{
			envoy_types.Endpoint:        "Endpoint",
			envoy_types.Cluster:         "Cluster",
			envoy_types.Route:           "Route",
			envoy_types.ScopedRoute:     "ScopedRoute",
			envoy_types.VirtualHost:     "VirtualHost",
			envoy_types.Listener:        "Listener",
			envoy_types.Secret:          "Secret",
			envoy_types.Runtime:         "Runtime",
			envoy_types.ExtensionConfig: "ExtensionConfig",
			envoy_types.RateLimitConfig: "RateLimitConfig",
		}
		for i, resources := range snapshot.Resources {
			name := resourceTypeNames[i]
			if name == "" || len(resources.Items) == 0 {
				continue
			}
			r.metrics.SnapshotResources.WithLabelValues(name).Observe(float64(len(resources.Items)))
		}
	}

	for _, version := range changed {
		r.statsCallbacks.ConfigReadyForDelivery(version)
	}
	return true, nil
}

func validateResource(r envoy_types.Resource) error {
	switch v := r.(type) {
	// Newer go-control-plane versions have `ValidateAll()` method, that accumulates as many validation errors as possible.
	case interface{ ValidateAll() error }:
		return v.ValidateAll()
	// Older go-control-plane stops validation at the first error.
	case interface{ Validate() error }:
		return v.Validate()
	default:
		return nil
	}
}

// autoVersion computes deterministic FNV-64a content hashes for each resource
// type in n, seeded with nodeId+type-index. Empty slots keep version "".
// The cluster version is folded into the endpoint version when endpoints are
// non-empty to force EDS re-push on cluster changes (prevents warming stalls).
// Returns the versions that changed relative to old.
func autoVersion(nodeId string, old, n *envoy_cache.Snapshot) (*envoy_cache.Snapshot, []string, error) {
	for i := range n.Resources {
		seed := nodeId + strconv.Itoa(i)
		ver, err := resourcesVersion(seed, n.Resources[i].Items)
		if err != nil {
			return nil, nil, errors.Wrapf(err, "failed to hash resources for type %d", i)
		}
		n.Resources[i].Version = ver
	}

	// Fold cluster version into endpoint version so that a cluster change
	// forces an EDS re-push even when endpoint content is identical.
	// Only when endpoints are non-empty; empty slots stay at "".
	if ep := n.Resources[envoy_types.Endpoint].Version; ep != "" {
		n.Resources[envoy_types.Endpoint].Version = mixVersions(ep, n.Resources[envoy_types.Cluster].Version)
	}

	var changed []string
	for i, resource := range n.Resources {
		if old.Resources[i].Version != resource.Version {
			changed = append(changed, resource.Version)
		}
	}

	return n, changed, nil
}

// resourcesVersion returns a hex FNV-64a hash over sorted resource names and
// their deterministic proto serializations, seeded with seed. Returns "" for
// empty slots so that two empty slots compare equal without versioning.
func resourcesVersion(seed string, items map[string]envoy_types.ResourceWithTTL) (string, error) {
	if len(items) == 0 {
		return "", nil
	}
	h := fnv.New64a()
	h.Write([]byte(seed))
	for _, key := range maps.SortedKeys(items) {
		h.Write([]byte(key))
		b, err := envoy_cache.MarshalResource(items[key].Resource)
		if err != nil {
			return "", err
		}
		h.Write(b)
	}
	return strconv.FormatUint(h.Sum64(), 16), nil
}

// mixVersions combines two version strings into a single FNV-64a hash.
func mixVersions(a, b string) string {
	h := fnv.New64a()
	h.Write([]byte(a))
	h.Write([]byte(b))
	return strconv.FormatUint(h.Sum64(), 16)
}

type snapshotGenerator interface {
	GenerateSnapshot(context.Context, xds_context.Context, *model.Proxy) (*envoy_cache.Snapshot, error)
}

type TemplateSnapshotGenerator struct {
	ProxyTemplateResolver xds_template.ProxyTemplateResolver
	ResourceSetHooks      []xds_hooks.ResourceSetHook
}

func (s *TemplateSnapshotGenerator) GenerateSnapshot(ctx context.Context, xdsCtx xds_context.Context, proxy *model.Proxy) (*envoy_cache.Snapshot, error) {
	template := s.ProxyTemplateResolver.GetTemplate(proxy)

	gen := generator.ProxyTemplateGenerator{ProxyTemplate: template}

	rs, err := gen.Generate(ctx, xdsCtx, proxy)
	if err != nil {
		return nil, err
	}
	for _, hook := range s.ResourceSetHooks {
		if err := hook.Modify(rs, xdsCtx, proxy); err != nil {
			return nil, errors.Wrapf(err, "could not apply hook %T", hook)
		}
	}
	if err := modifications.Apply(rs, template.GetConf().GetModifications(), proxy.APIVersion); err != nil {
		return nil, errors.Wrap(err, "could not apply modifications")
	}

	version := "" // empty value is a sign to other components to generate the version automatically
	resources := map[envoy_resource.Type][]envoy_types.Resource{}

	for _, resourceType := range rs.ResourceTypes() {
		resources[resourceType] = append(resources[resourceType], rs.ListOf(resourceType).Payloads()...)
	}

	return envoy_cache.NewSnapshot(version, resources)
}

type snapshotCacher interface {
	Get(*envoy_core.Node) (*envoy_cache.Snapshot, error)
	Cache(context.Context, *envoy_core.Node, *envoy_cache.Snapshot) error
	Clear(*envoy_core.Node)
}

type simpleSnapshotCacher struct {
	hasher envoy_cache.NodeHash
	store  envoy_cache.SnapshotCache
}

func (s *simpleSnapshotCacher) Get(node *envoy_core.Node) (*envoy_cache.Snapshot, error) {
	snap, err := s.store.GetSnapshot(s.hasher.ID(node))
	if snap != nil {
		snapshot, ok := snap.(*envoy_cache.Snapshot)
		if !ok {
			return nil, errors.New("couldn't convert snapshot from cache to envoy Snapshot")
		}
		return snapshot, nil
	}
	return nil, err
}

func (s *simpleSnapshotCacher) Cache(ctx context.Context, node *envoy_core.Node, snapshot *envoy_cache.Snapshot) error {
	return s.store.SetSnapshot(ctx, s.hasher.ID(node), snapshot)
}

func (s *simpleSnapshotCacher) Clear(node *envoy_core.Node) {
	s.store.ClearSnapshot(s.hasher.ID(node))
}
