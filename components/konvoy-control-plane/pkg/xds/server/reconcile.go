package server

import (
	"github.com/Kong/konvoy/components/konvoy-control-plane/pkg/core"
	core_discovery "github.com/Kong/konvoy/components/konvoy-control-plane/pkg/core/discovery"
	"github.com/Kong/konvoy/components/konvoy-control-plane/pkg/xds/generator"
	"github.com/Kong/konvoy/components/konvoy-control-plane/pkg/xds/model"
	envoy "github.com/envoyproxy/go-control-plane/envoy/api/v2"
	envoy_auth "github.com/envoyproxy/go-control-plane/envoy/api/v2/auth"
	envoy_core "github.com/envoyproxy/go-control-plane/envoy/api/v2/core"
	envoy_cache "github.com/envoyproxy/go-control-plane/pkg/cache"
)

var (
	reconcileLog = core.Log.WithName("xds-server").WithName("reconcile")
)

type reconciler struct {
	generator snapshotGenerator
	cacher    snapshotCacher
}

// Make sure that reconciler implements all relevant interfaces
var (
	_ core_discovery.DiscoveryConsumer = &reconciler{}
)

func (r *reconciler) OnWorkloadUpdate(info *core_discovery.WorkloadInfo) error {
	proxyId := model.ProxyId{Name: info.Workload.Id.Name, Namespace: info.Workload.Id.Namespace}
	return r.reconcile(
		&envoy_core.Node{Id: proxyId.String()},
		&model.Proxy{
			Id: proxyId,
			Workload: model.Workload{
				Meta: model.WorkloadMeta{
					Namespace: info.Workload.Id.Namespace,
					Name:      info.Workload.Id.Name,
					Labels:    info.Workload.Meta.Labels,
				},
				Version:   info.Desc.Version,
				Endpoints: info.Desc.Endpoints,
			},
		})
}
func (r *reconciler) OnWorkloadDelete(name core.NamespacedName) error {
	proxyId := model.ProxyId{Name: name.Name, Namespace: name.Namespace}
	r.cacher.Clear(&envoy_core.Node{Id: proxyId.String()})
	return nil
}
func (r *reconciler) OnServiceUpdate(_ *core_discovery.ServiceInfo) error {
	return nil
}
func (r *reconciler) OnServiceDelete(_ core.NamespacedName) error {
	return nil
}

func (r *reconciler) reconcile(node *envoy_core.Node, proxy *model.Proxy) error {
	snapshot, err := r.generator.GenerateSnapshot(proxy)
	if err != nil {
		reconcileLog.Error(err, "failed to generate a snapshot", "node", node, "proxy", proxy)
		return err
	}
	if err := snapshot.Consistent(); err != nil {
		reconcileLog.Error(err, "inconsistent snapshot", "snapshot", snapshot, "proxy", proxy)
	}
	if err := r.cacher.Cache(node, snapshot); err != nil {
		reconcileLog.Error(err, "failed to store snapshot", "snapshot", snapshot, "proxy", proxy)
	}
	return nil
}

type snapshotGenerator interface {
	GenerateSnapshot(proxy *model.Proxy) (envoy_cache.Snapshot, error)
}

type templateSnapshotGenerator struct {
	ProxyTemplateResolver proxyTemplateResolver
}

func (s *templateSnapshotGenerator) GenerateSnapshot(proxy *model.Proxy) (envoy_cache.Snapshot, error) {
	template := s.ProxyTemplateResolver.GetTemplate(proxy)

	gen := generator.TemplateProxyGenerator{ProxyTemplate: template}

	rs, err := gen.Generate(proxy)
	if err != nil {
		reconcileLog.Error(err, "failed to generate a snapshot", "proxy", proxy, "template", template)
		return envoy_cache.Snapshot{}, err
	}

	listeners := []envoy_cache.Resource{}
	routes := []envoy_cache.Resource{}
	clusters := []envoy_cache.Resource{}
	endpoints := []envoy_cache.Resource{}
	secrets := []envoy_cache.Resource{}

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

	version := proxy.Workload.Version
	out := envoy_cache.Snapshot{
		Endpoints: envoy_cache.NewResources(version, endpoints),
		Clusters:  envoy_cache.NewResources(version, clusters),
		Routes:    envoy_cache.NewResources(version, routes),
		Listeners: envoy_cache.NewResources(version, listeners),
		Secrets:   envoy_cache.NewResources(version, secrets),
	}

	return out, nil
}

type snapshotCacher interface {
	Cache(*envoy_core.Node, envoy_cache.Snapshot) error
	Clear(*envoy_core.Node)
}

type simpleSnapshotCacher struct {
	hasher envoy_cache.NodeHash
	store  envoy_cache.SnapshotCache
}

func (s *simpleSnapshotCacher) Cache(node *envoy_core.Node, snapshot envoy_cache.Snapshot) error {
	return s.store.SetSnapshot(s.hasher.ID(node), snapshot)
}

func (s *simpleSnapshotCacher) Clear(node *envoy_core.Node) {
	s.store.ClearSnapshot(s.hasher.ID(node))
}
