package v3

import (
	"context"

	envoy_core "github.com/envoyproxy/go-control-plane/envoy/config/core/v3"
	envoy_types "github.com/envoyproxy/go-control-plane/pkg/cache/types"
	envoy_cache "github.com/envoyproxy/go-control-plane/pkg/cache/v3"
	envoy_resource "github.com/envoyproxy/go-control-plane/pkg/resource/v3"
	"github.com/golang/protobuf/proto"
	"github.com/pkg/errors"

	"github.com/kumahq/kuma/pkg/core"
	model "github.com/kumahq/kuma/pkg/core/xds"
	xds_context "github.com/kumahq/kuma/pkg/xds/context"
	"github.com/kumahq/kuma/pkg/xds/generator"
	xds_hooks "github.com/kumahq/kuma/pkg/xds/hooks"
	xds_sync "github.com/kumahq/kuma/pkg/xds/sync"
	xds_template "github.com/kumahq/kuma/pkg/xds/template"
)

var (
	reconcileLog = core.Log.WithName("xds-server").WithName("reconcile")
)

var _ xds_sync.SnapshotReconciler = &reconciler{}

type reconciler struct {
	generator snapshotGenerator
	cacher    snapshotCacher
}

func (r *reconciler) Clear(proxyId *model.ProxyId) error {
	r.cacher.Clear(&envoy_core.Node{Id: proxyId.String()})
	return nil
}

func (r *reconciler) Reconcile(ctx xds_context.Context, proxy *model.Proxy) error {
	node := &envoy_core.Node{Id: proxy.Id.String()}
	snapshot, err := r.generator.GenerateSnapshot(ctx, proxy)
	if err != nil {
		reconcileLog.Error(err, "failed to generate a snapshot", "node", node, "proxy", proxy)
		return err
	}
	if err := snapshot.Consistent(); err != nil {
		reconcileLog.Error(err, "inconsistent snapshot", "snapshot", snapshot, "proxy", proxy)
	}
	// to avoid assigning a new version every time,
	// compare with the previous snapshot and reuse its version whenever possible,
	// fallback to UUID otherwise
	previous, err := r.cacher.Get(node)
	if err != nil {
		previous = envoy_cache.Snapshot{}
	}
	snapshot = r.autoVersion(previous, snapshot)
	if err := r.cacher.Cache(node, snapshot); err != nil {
		reconcileLog.Error(err, "failed to store snapshot", "snapshot", snapshot, "proxy", proxy)
	}
	return nil
}

func (r *reconciler) autoVersion(old envoy_cache.Snapshot, new envoy_cache.Snapshot) envoy_cache.Snapshot {
	new.Resources[envoy_types.Listener] = reuseVersion(old.Resources[envoy_types.Listener], new.Resources[envoy_types.Listener])
	new.Resources[envoy_types.Route] = reuseVersion(old.Resources[envoy_types.Route], new.Resources[envoy_types.Route])
	new.Resources[envoy_types.Cluster] = reuseVersion(old.Resources[envoy_types.Cluster], new.Resources[envoy_types.Cluster])
	new.Resources[envoy_types.Endpoint] = reuseVersion(old.Resources[envoy_types.Endpoint], new.Resources[envoy_types.Endpoint])
	new.Resources[envoy_types.Secret] = reuseVersion(old.Resources[envoy_types.Secret], new.Resources[envoy_types.Secret])
	return new
}

func reuseVersion(old, new envoy_cache.Resources) envoy_cache.Resources {
	new.Version = old.Version
	if !equalSnapshots(old.Items, new.Items) {
		new.Version = core.NewUUID()
	}
	return new
}

func equalSnapshots(old, new map[string]envoy_types.ResourceWithTTL) bool {
	if len(new) != len(old) {
		return false
	}
	for key, newValue := range new {
		if oldValue, hasOldValue := old[key]; !hasOldValue || !proto.Equal(newValue.Resource, oldValue.Resource) {
			return false
		}
	}
	return true
}

type snapshotGenerator interface {
	GenerateSnapshot(ctx xds_context.Context, proxy *model.Proxy) (envoy_cache.Snapshot, error)
}

type templateSnapshotGenerator struct {
	ProxyTemplateResolver xds_template.ProxyTemplateResolver
	ResourceSetHooks      []xds_hooks.ResourceSetHook
}

func (s *templateSnapshotGenerator) GenerateSnapshot(ctx xds_context.Context, proxy *model.Proxy) (envoy_cache.Snapshot, error) {
	template := s.ProxyTemplateResolver.GetTemplate(proxy)

	gen := generator.ProxyTemplateGenerator{ProxyTemplate: template}

	rs, err := gen.Generate(ctx, proxy)
	if err != nil {
		reconcileLog.Error(err, "failed to generate a snapshot", "proxy", proxy, "template", template)
		return envoy_cache.Snapshot{}, err
	}
	for _, hook := range s.ResourceSetHooks {
		if err := hook.Modify(rs, ctx, proxy); err != nil {
			return envoy_cache.Snapshot{}, errors.Wrapf(err, "could not apply hook %T", hook)
		}
	}

	version := "" // empty value is a sign to other components to generate the version automatically
	out := envoy_cache.Snapshot{
		Resources: [envoy_types.UnknownType]envoy_cache.Resources{
			envoy_types.Endpoint: envoy_cache.NewResources(version, rs.ListOf(envoy_resource.EndpointType).Payloads()),
			envoy_types.Cluster:  envoy_cache.NewResources(version, rs.ListOf(envoy_resource.ClusterType).Payloads()),
			envoy_types.Route:    envoy_cache.NewResources(version, rs.ListOf(envoy_resource.RouteType).Payloads()),
			envoy_types.Listener: envoy_cache.NewResources(version, rs.ListOf(envoy_resource.ListenerType).Payloads()),
			envoy_types.Secret:   envoy_cache.NewResources(version, rs.ListOf(envoy_resource.SecretType).Payloads()),
		},
	}

	return out, nil
}

type snapshotCacher interface {
	Get(*envoy_core.Node) (envoy_cache.Snapshot, error)
	Cache(*envoy_core.Node, envoy_cache.Snapshot) error
	Clear(*envoy_core.Node)
}

type simpleSnapshotCacher struct {
	hasher envoy_cache.NodeHash
	store  envoy_cache.SnapshotCache
}

func (s *simpleSnapshotCacher) Get(node *envoy_core.Node) (envoy_cache.Snapshot, error) {
	return s.store.GetSnapshot(s.hasher.ID(node))
}

func (s *simpleSnapshotCacher) Cache(node *envoy_core.Node, snapshot envoy_cache.Snapshot) error {
	return s.store.SetSnapshot(context.TODO(), s.hasher.ID(node), snapshot)
}

func (s *simpleSnapshotCacher) Clear(node *envoy_core.Node) {
	s.store.ClearSnapshot(s.hasher.ID(node))
}
