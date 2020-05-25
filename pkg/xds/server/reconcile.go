package server

import (
	"github.com/golang/protobuf/proto"

	"github.com/Kong/kuma/pkg/core"
	model "github.com/Kong/kuma/pkg/core/xds"
	xds_context "github.com/Kong/kuma/pkg/xds/context"
	"github.com/Kong/kuma/pkg/xds/generator"

	envoy "github.com/envoyproxy/go-control-plane/envoy/api/v2"
	envoy_auth "github.com/envoyproxy/go-control-plane/envoy/api/v2/auth"
	envoy_core "github.com/envoyproxy/go-control-plane/envoy/api/v2/core"
	envoy_types "github.com/envoyproxy/go-control-plane/pkg/cache/types"
	envoy_cache "github.com/envoyproxy/go-control-plane/pkg/cache/v2"
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
		new.Version = newUUID()
	}
	return new
}

func equalSnapshots(old, new map[string]envoy_types.Resource) bool {
	if len(new) != len(old) {
		return false
	}
	for key, newValue := range new {
		if oldValue, hasOldValue := old[key]; !hasOldValue || !proto.Equal(newValue, oldValue) {
			return false
		}
	}
	return true
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

	listeners := []envoy_types.Resource{}
	routes := []envoy_types.Resource{}
	clusters := []envoy_types.Resource{}
	endpoints := []envoy_types.Resource{}
	secrets := []envoy_types.Resource{}

	for _, r := range rs {
		switch r.Resource.(type) {
		case *envoy.Listener:
			listeners = append(listeners, r.Resource)
		case *envoy.RouteConfiguration:
			routes = append(routes, r.Resource)
		case *envoy.Cluster:
			clusters = append(clusters, r.Resource)
		case *envoy.ClusterLoadAssignment:
			endpoints = append(endpoints, r.Resource)
		case *envoy_auth.Secret:
			secrets = append(secrets, r.Resource)
		default:
		}
	}

	version := "" // empty value is a sign to other components to generate the version automatically
	out := envoy_cache.Snapshot{
		Resources: [envoy_types.UnknownType]envoy_cache.Resources{
			envoy_types.Endpoint: envoy_cache.NewResources(version, endpoints),
			envoy_types.Cluster:  envoy_cache.NewResources(version, clusters),
			envoy_types.Route:    envoy_cache.NewResources(version, routes),
			envoy_types.Listener: envoy_cache.NewResources(version, listeners),
			envoy_types.Secret:   envoy_cache.NewResources(version, secrets),
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
