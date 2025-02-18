package manager

import (
	"context"
	"fmt"
	"reflect"
	"time"

	"k8s.io/apimachinery/pkg/api/meta"
	kube_runtime "k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/rest"
	kube_ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/cache"
	"sigs.k8s.io/controller-runtime/pkg/client"
	kube_client "sigs.k8s.io/controller-runtime/pkg/client"
	kube_metricsserver "sigs.k8s.io/controller-runtime/pkg/metrics/server"
	kube_webhook "sigs.k8s.io/controller-runtime/pkg/webhook"

	"github.com/kumahq/kuma/pkg/core"
	core_plugins "github.com/kumahq/kuma/pkg/core/plugins"
	core_registry "github.com/kumahq/kuma/pkg/core/resources/registry"
	core_runtime "github.com/kumahq/kuma/pkg/core/runtime"
	kuma_v1alpha1 "github.com/kumahq/kuma/pkg/plugins/resources/k8s/native/api/v1alpha1"
	"github.com/kumahq/kuma/pkg/plugins/resources/k8s/native/pkg/model"
	k8s_registry "github.com/kumahq/kuma/pkg/plugins/resources/k8s/native/pkg/registry"
)

// scopedManager wraps a manager.Manager and overrides GetClient.
type scopedManager struct {
	kube_ctrl.Manager
	client ScopedClient
}

// NewScopedManager creates a new manager wrapper that returns a scoped client.
func NewScopedManager(b *core_runtime.Builder, cfg core_plugins.PluginConfig, restClientConfig *rest.Config, scheme *kube_runtime.Scheme) (kube_ctrl.Manager, error) {
	systemNamespace := b.Config().Store.Kubernetes.SystemNamespace
	managerOpts := kube_ctrl.Options{
		Scheme: scheme,
		// Admission WebHook Server
		WebhookServer: kube_webhook.NewServer(kube_webhook.Options{
			Host:    b.Config().Runtime.Kubernetes.AdmissionServer.Address,
			Port:    int(b.Config().Runtime.Kubernetes.AdmissionServer.Port),
			CertDir: b.Config().Runtime.Kubernetes.AdmissionServer.CertDir,
		}),
		LeaderElection:          true,
		LeaderElectionID:        "cp-leader-lease",
		LeaderElectionNamespace: systemNamespace,
		Logger:                  core.Log.WithName("kube-manager"),
		LeaseDuration:           &b.Config().Runtime.Kubernetes.LeaderElection.LeaseDuration.Duration,
		RenewDeadline:           &b.Config().Runtime.Kubernetes.LeaderElection.RenewDeadline.Duration,
		// Disable metrics bind address as we use kube metrics registry directly.
		Metrics: kube_metricsserver.Options{
			BindAddress: "0",
		},
	}
	watchNamespaces := map[string]struct{}{}
	if len(b.Config().Runtime.Kubernetes.WatchNamespaces) > 0 {
		for _, ns := range b.Config().Runtime.Kubernetes.WatchNamespaces {
			watchNamespaces[ns] = struct{}{}
		}
		// always include systemNamespace.
		watchNamespaces[systemNamespace] = struct{}{}
		// add cni namespace if taint controller enabled
		if b.Config().Runtime.Kubernetes.NodeTaintController.Enabled {
			watchNamespaces[b.Config().Runtime.Kubernetes.NodeTaintController.CniNamespace] = struct{}{}
		}
	}

	if len(watchNamespaces) > 0 {
		managerOpts.Cache = getCacheConfig(watchNamespaces)
	}
	mgr, err := kube_ctrl.NewManager(
		restClientConfig,
		managerOpts,
	)
	if err != nil {
		return nil, err
	}

	resourceTypeToScope, err := getResourceTypesScope(mgr)
	if err != nil {
		return nil, err
	}

	return &scopedManager{
		Manager: mgr,
		client:  NewScopedClient(mgr.GetClient(), watchNamespaces, resourceTypeToScope),
	}, nil
}

func getResourceTypesScope(mgr kube_ctrl.Manager) (ResourceTypeToScope, error) {
	resourceTypeToScope := ResourceTypeToScope{}
	types := core_registry.Global().ObjectTypes()
	knownTypes := mgr.GetScheme().KnownTypes(kuma_v1alpha1.GroupVersion)
	for _, t := range types {
		if _, ok := knownTypes[string(t)]; !ok {
			continue
		}
		coreObj, err := core_registry.Global().NewObject(t)
		if err != nil {
			return nil, err
		}
		obj, err := k8s_registry.Global().NewObject(coreObj.GetSpec())
		if err != nil {
			return nil, err
		}

		// <ResourceType>List is not in the registry but needs to be listed
		switch obj.Scope() {
		case model.ScopeCluster:
			resourceTypeToScope[ResourceType(t)] = IsClusterScope(true)
			resourceTypeToScope[ResourceType(fmt.Sprintf("%s%s", t, "List"))] = IsClusterScope(true)
		case model.ScopeNamespace:
			resourceTypeToScope[ResourceType(t)] = IsClusterScope(false)
			resourceTypeToScope[ResourceType(fmt.Sprintf("%s%s", t, "List"))] = IsClusterScope(false)
		}
	}
	return resourceTypeToScope, nil
}

func getCacheConfig(
	watchNamespaces map[string]struct{},
) cache.Options {
	resyncPeriod := 10 * time.Hour // default resyncPeriod in Kubernetes
	cacheConfig := cache.Options{
		SyncPeriod: &resyncPeriod,
	}
	cacheConfig.DefaultNamespaces = map[string]cache.Config{}
	// cache only watched namespaces
	for ns := range watchNamespaces {
		cacheConfig.DefaultNamespaces[ns] = cache.Config{}
	}
	return cacheConfig
}

// Override GetClient to return the scoped client.
func (s *scopedManager) GetClient() kube_client.Client {
	return s.client
}

// ScopedClient defines the methods we need from the Kubernetes client.
type ScopedClient interface {
	client.Client
}

type (
	ResourceType        string
	IsClusterScope      bool
	ResourceTypeToScope map[ResourceType]IsClusterScope
)

// scopedClient wraps a controller-runtime client and restricts operations to allowed namespaces.
type scopedClient struct {
	client.Client
	watchedNamespaces   map[string]struct{}
	resourceTypeToScope ResourceTypeToScope
}

// NewScopedClient creates a new ScopedClient wrapping the given client
// and restricting operations to the provided namespaces.
func NewScopedClient(cli client.Client, watchNamespaces map[string]struct{}, resourceTypeToScope ResourceTypeToScope) ScopedClient {
	if len(watchNamespaces) == 0 {
		return &scopedClient{
			Client: cli,
		}
	}
	return &scopedClient{
		Client:              cli,
		watchedNamespaces:   watchNamespaces,
		resourceTypeToScope: resourceTypeToScope,
	}
}

func (s *scopedClient) Get(ctx context.Context, key client.ObjectKey, obj client.Object, opts ...client.GetOption) error {
	if len(s.watchedNamespaces) == 0 {
		return s.Client.Get(ctx, key, obj, opts...)
	}
	ns := obj.GetNamespace()
	if ns != "" {
		if _, allowed := s.watchedNamespaces[ns]; !allowed {
			return fmt.Errorf("namespace %q is not a part of the mesh", ns)
		}
	}
	return s.Client.Get(ctx, key, obj, opts...)
}

func (s *scopedClient) Create(ctx context.Context, obj client.Object, opts ...client.CreateOption) error {
	if len(s.watchedNamespaces) == 0 {
		return s.Client.Create(ctx, obj, opts...)
	}
	ns := obj.GetNamespace()
	if ns != "" {
		if _, allowed := s.watchedNamespaces[ns]; !allowed {
			return fmt.Errorf("namespace %q is not a part of the mesh", ns)
		}
	}
	return s.Client.Create(ctx, obj, opts...)
}

func (s *scopedClient) Delete(ctx context.Context, obj client.Object, opts ...client.DeleteOption) error {
	if len(s.watchedNamespaces) == 0 {
		return s.Client.Delete(ctx, obj, opts...)
	}
	ns := obj.GetNamespace()
	if ns != "" {
		if _, allowed := s.watchedNamespaces[ns]; !allowed {
			return fmt.Errorf("namespace %q is not a part of the mesh", ns)
		}
	}
	return s.Client.Delete(ctx, obj, opts...)
}

func (s *scopedClient) Update(ctx context.Context, obj client.Object, opts ...client.UpdateOption) error {
	if len(s.watchedNamespaces) == 0 {
		return s.Client.Update(ctx, obj, opts...)
	}
	ns := obj.GetNamespace()
	if ns != "" {
		if _, allowed := s.watchedNamespaces[ns]; !allowed {
			return fmt.Errorf("namespace %q is not a part of the mesh", ns)
		}
	}
	return s.Client.Update(ctx, obj, opts...)
}

func (s *scopedClient) Patch(ctx context.Context, obj client.Object, patch client.Patch, opts ...client.PatchOption) error {
	if len(s.watchedNamespaces) == 0 {
		return s.Client.Patch(ctx, obj, patch, opts...)
	}
	ns := obj.GetNamespace()
	if ns != "" {
		if _, allowed := s.watchedNamespaces[ns]; !allowed {
			return fmt.Errorf("namespace %q is not a part of the mesh", ns)
		}
	}
	return s.Client.Patch(ctx, obj, patch, opts...)
}

func (s *scopedClient) List(ctx context.Context, list client.ObjectList, opts ...client.ListOption) error {
	if len(s.watchedNamespaces) == 0 || s.isClusterScope(list.GetObjectKind().GroupVersionKind().Kind) {
		return s.Client.List(ctx, list, opts...)
	}
	lo := &client.ListOptions{}
	for _, o := range opts {
		o.ApplyToList(lo)
	}
	if lo.Namespace != "" {
		if _, allowed := s.watchedNamespaces[lo.Namespace]; !allowed {
			return fmt.Errorf("namespace %q is not a part of the mesh", lo.Namespace)
		}
		return s.Client.List(ctx, list, opts...)
	}

	var allItems []kube_runtime.Object
	for namespace := range s.watchedNamespaces {
		// Create a new list instance of the same type as 'list'
		tempList, err := newListOfType(list)
		if err != nil {
			return err
		}
		optsForNS := append(opts, client.InNamespace(namespace))
		if err := s.Client.List(ctx, tempList, optsForNS...); err != nil {
			return err
		}
		items, err := meta.ExtractList(tempList)
		if err != nil {
			return err
		}
		allItems = append(allItems, items...)
	}

	// Set all aggregated items into the original list.
	if err := meta.SetList(list, allItems); err != nil {
		return err
	}
	return nil
}

func (s *scopedClient) isClusterScope(resourceType string) IsClusterScope {
	isClusterScope, exist := s.resourceTypeToScope[ResourceType(resourceType)]
	if !exist {
		return false
	}
	return isClusterScope
}

// newListOfType creates a new empty list of the same concrete type as 'list'.
// It assumes that 'list' is a pointer to a struct.
func newListOfType(list client.ObjectList) (client.ObjectList, error) {
	typ := reflect.TypeOf(list)
	if typ.Kind() != reflect.Ptr {
		return nil, fmt.Errorf("expected list to be a pointer, got %T", list)
	}
	// Create a new instance of the same type.
	newList := reflect.New(typ.Elem()).Interface().(client.ObjectList)
	return newList, nil
}
