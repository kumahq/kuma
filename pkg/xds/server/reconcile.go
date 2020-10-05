package server

import (
	"github.com/kumahq/kuma/pkg/core"
	model "github.com/kumahq/kuma/pkg/core/xds"
	xds_context "github.com/kumahq/kuma/pkg/xds/context"
	"github.com/kumahq/kuma/pkg/xds/generator"

	envoy_core "github.com/envoyproxy/go-control-plane/envoy/api/v2/core"
	envoy_types "github.com/envoyproxy/go-control-plane/pkg/cache/types"
	envoy_cache "github.com/envoyproxy/go-control-plane/pkg/cache/v2"
	envoy_resource "github.com/envoyproxy/go-control-plane/pkg/resource/v2"
)

var (
	reconcileLog = core.Log.WithName("xds-server").WithName("reconcile")
)

type SnapshotReconciler interface {
	Reconcile(ctx xds_context.Context, proxy *model.Proxy) error
	Clear(proxyId *model.ProxyId) error
}

var _ SnapshotReconciler = &reconciler{}

type reconciler struct {
	generator snapshotGenerator
	cacher    snapshotCacher
}

func (r *reconciler) Clear(proxyId *model.ProxyId) error {
	// cache.Clear() operation does not push a new (empty) configuration to Envoy.
	// That is why instead of calling cache.Clear() we set configuration to an empty Snapshot.
	// This fake value will be removed from cache on Envoy disconnect.
	return r.cacher.Cache(&envoy_core.Node{Id: proxyId.String()}, envoy_cache.Snapshot{})
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
	snapshot = r.autoVersion(snapshot)
	if err := r.cacher.Cache(node, snapshot); err != nil {
		reconcileLog.Error(err, "failed to store snapshot", "snapshot", snapshot, "proxy", proxy)
	}
	return nil
}

func (r *reconciler) autoVersion(new envoy_cache.Snapshot) envoy_cache.Snapshot {
	new.Resources[envoy_types.Listener].Version = newUUID()
	new.Resources[envoy_types.Route].Version = newUUID()
	new.Resources[envoy_types.Cluster].Version = newUUID()
	new.Resources[envoy_types.Endpoint].Version = newUUID()
	new.Resources[envoy_types.Secret].Version = newUUID()
	return new
}

type snapshotGenerator interface {
	GenerateSnapshot(ctx xds_context.Context, proxy *model.Proxy) (envoy_cache.Snapshot, error)
}

type templateSnapshotGenerator struct {
	ProxyTemplateResolver proxyTemplateResolver
}

func (s *templateSnapshotGenerator) GenerateSnapshot(ctx xds_context.Context, proxy *model.Proxy) (envoy_cache.Snapshot, error) {
	template := s.ProxyTemplateResolver.GetTemplate(proxy)

	gen := generator.ProxyTemplateGenerator{ProxyTemplate: template}

	rs, err := gen.Generate(ctx, proxy)
	if err != nil {
		reconcileLog.Error(err, "failed to generate a snapshot", "proxy", proxy, "template", template)
		return envoy_cache.Snapshot{}, err
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
	return s.store.SetSnapshot(s.hasher.ID(node), snapshot)
}

func (s *simpleSnapshotCacher) Clear(node *envoy_core.Node) {
	s.store.ClearSnapshot(s.hasher.ID(node))
}
