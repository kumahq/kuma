package server

import (
	"fmt"

	model_controllers "github.com/Kong/konvoy/components/konvoy-control-plane/model/controllers"
	util_k8s "github.com/Kong/konvoy/components/konvoy-control-plane/pkg/util/k8s"
	"github.com/Kong/konvoy/components/konvoy-control-plane/pkg/xds/generator"
	"github.com/Kong/konvoy/components/konvoy-control-plane/pkg/xds/model"
	envoy "github.com/envoyproxy/go-control-plane/envoy/api/v2"
	envoy_auth "github.com/envoyproxy/go-control-plane/envoy/api/v2/auth"
	envoy_core "github.com/envoyproxy/go-control-plane/envoy/api/v2/core"
	"github.com/envoyproxy/go-control-plane/pkg/cache"
	k8s_core "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
)

var (
	reconcileLog = ctrl.Log.WithName("xds-server").WithName("reconcile")
)

type reconciler struct {
	generator snapshotGenerator
	cacher    snapshotCacher
}

// Make sure that reconciler implements all relevant interfaces
var (
	_ model_controllers.PodObserver = &reconciler{}
)

func (r *reconciler) OnUpdate(pod *k8s_core.Pod) error {
	proxyId := model.ProxyId{Name: pod.Name, Namespace: pod.Namespace}
	return r.reconcile(
		&envoy_core.Node{Id: proxyId.String()},
		&model.Proxy{
			Id: proxyId,
			Workload: model.Workload{
				Version:   fmt.Sprintf("v%d", pod.Generation),
				Addresses: []string{pod.Status.PodIP},
				Ports:     util_k8s.GetTcpPorts(pod),
			},
		})
}

func (r *reconciler) OnDelete(name types.NamespacedName) error {
	proxyId := model.ProxyId{Name: name.Name, Namespace: name.Namespace}
	r.cacher.Clear(&envoy_core.Node{Id: proxyId.String()})
	return nil
}

func (r *reconciler) reconcile(node *envoy_core.Node, proxy *model.Proxy) error {
	snapshot, err := r.generator.GenerateSnapshot(proxy)
	if err != nil {
		return err
	}
	if err := snapshot.Consistent(); err != nil {
		reconcileLog.Error(err, "inconsistent snapshot", "snapshot", snapshot)
	}
	if err := r.cacher.Cache(node, snapshot); err != nil {
		reconcileLog.Error(err, "failed to store snapshot", "snapshot", snapshot)
	}
	return nil
}

type snapshotGenerator interface {
	GenerateSnapshot(proxy *model.Proxy) (cache.Snapshot, error)
}

type templateSnapshotGenerator struct {
	ProxyTemplateResolver proxyTemplateResolver
}

func (s *templateSnapshotGenerator) GenerateSnapshot(proxy *model.Proxy) (cache.Snapshot, error) {
	gen := generator.TemplateProxyGenerator{ProxyTemplate: s.ProxyTemplateResolver.GetTemplate(proxy)}

	rs, err := gen.Generate(proxy)
	if err != nil {
		return cache.Snapshot{}, err
	}

	listeners := []cache.Resource{}
	routes := []cache.Resource{}
	clusters := []cache.Resource{}
	endpoints := []cache.Resource{}
	secrets := []cache.Resource{}

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
	out := cache.Snapshot{
		Endpoints: cache.NewResources(version, endpoints),
		Clusters:  cache.NewResources(version, clusters),
		Routes:    cache.NewResources(version, routes),
		Listeners: cache.NewResources(version, listeners),
		Secrets:   cache.NewResources(version, secrets),
	}

	return out, nil
}

type snapshotCacher interface {
	Cache(*envoy_core.Node, cache.Snapshot) error
	Clear(*envoy_core.Node)
}

type simpleSnapshotCacher struct {
	hasher cache.NodeHash
	store  cache.SnapshotCache
}

func (s *simpleSnapshotCacher) Cache(node *envoy_core.Node, snapshot cache.Snapshot) error {
	return s.store.SetSnapshot(s.hasher.ID(node), snapshot)
}

func (s *simpleSnapshotCacher) Clear(node *envoy_core.Node) {
	s.store.ClearSnapshot(s.hasher.ID(node))
}
