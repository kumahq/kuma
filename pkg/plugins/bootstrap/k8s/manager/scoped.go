package manager

import (
	"context"
	"fmt"
	"reflect"
	"time"

	runtime_config "github.com/kumahq/kuma/pkg/config/plugins/runtime/k8s"
	"github.com/kumahq/kuma/pkg/core"
	core_plugins "github.com/kumahq/kuma/pkg/core/plugins"
	core_runtime "github.com/kumahq/kuma/pkg/core/runtime"
	"github.com/kumahq/kuma/pkg/plugins/bootstrap/k8s/schema"
	"k8s.io/apimachinery/pkg/api/meta"
	kube_runtime "k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/rest"
	kube_ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/cache"
	"sigs.k8s.io/controller-runtime/pkg/client"
	kube_client "sigs.k8s.io/controller-runtime/pkg/client"
	kube_metricsserver "sigs.k8s.io/controller-runtime/pkg/metrics/server"
	kube_webhook "sigs.k8s.io/controller-runtime/pkg/webhook"
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
	if len(b.Config().Runtime.Kubernetes.WatchNamespaces) > 0 {
		managerOpts.Cache = getCacheConfig(b.Config().Runtime.Kubernetes, systemNamespace)
	}
	mgr, err := kube_ctrl.NewManager(
		restClientConfig,
		managerOpts,
	)
	if err != nil {
		return nil, err
	}

	return &scopedManager{
		Manager: mgr,
		client:  NewScopedClient(mgr.GetClient(), b.Config().Runtime.Kubernetes.WatchNamespaces, systemNamespace),
	}, nil
}

func getCacheConfig(
	cfg *runtime_config.KubernetesRuntimeConfig,
	systemNamespace string,
) cache.Options {
	resyncPeriod := 10 * time.Hour // default resyncPeriod in Kubernetes
	cacheConfig := cache.Options{
		SyncPeriod:        &resyncPeriod,
	}
	cacheConfig.DefaultNamespaces = map[string]cache.Config{}
	// cache only watched namespaces
	for _, namespace := range cfg.WatchNamespaces {
		cacheConfig.DefaultNamespaces[namespace] = cache.Config{}
	}
	// system namespace should be also cached
	cacheConfig.DefaultNamespaces[systemNamespace] = cache.Config{}
	// cni namespace
	if cfg.NodeTaintController.Enabled {
		cacheConfig.DefaultNamespaces[cfg.NodeTaintController.CniNamespace] = cache.Config{}
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

// scopedClient wraps a controller-runtime client and restricts operations to allowed namespaces.
type scopedClient struct {
	client.Client
	watchedNamespaces map[string]struct{}
	systemNamespace   string
}

// NewScopedClient creates a new ScopedClient wrapping the given client
// and restricting operations to the provided namespaces.
func NewScopedClient(cli client.Client, watchedNamespaces []string, systemNamespace string) ScopedClient {
	nsMap := make(map[string]struct{})
	for _, ns := range watchedNamespaces {
		nsMap[ns] = struct{}{}
	}
	// Always include systemNamespace.
	nsMap[systemNamespace] = struct{}{}
	return &scopedClient{
		Client:            cli,
		watchedNamespaces: nsMap,
		systemNamespace:   systemNamespace,
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

// Example: Override List to enforce namespace scoping.
func (s *scopedClient) List(ctx context.Context, list client.ObjectList, opts ...client.ListOption) error {
	lo := &client.ListOptions{}
	for _, o := range opts {
		o.ApplyToList(lo)
	}
	// core.Log.Info("CALLING MY CUSTOM LIST", "watchedNamespaces", s.watchedNamespaces, "lo", lo, "list.GetObjectKind().GroupVersionKind().Kind", list.GetObjectKind().GroupVersionKind().Kind)
	if len(s.watchedNamespaces) == 0 || schema.IsClusterScopeResource(list.GetObjectKind().GroupVersionKind().Kind) {
		return s.Client.List(ctx, list, opts...)
	}
	// core.Log.Info("CALLING MY CUSTOM LIST if watchedNamespaces is more than 0")

	if lo.Namespace != "" {
		if _, allowed := s.watchedNamespaces[lo.Namespace]; !allowed {
			return fmt.Errorf("namespace %q is not a part of the mesh", lo.Namespace)
		}
		// core.Log.Info("CALLING MY CUSTOM LIST namespace is in the request", "lo", lo.Namespace)
		return s.Client.List(ctx, list, opts...)
	}

	var allItems []kube_runtime.Object
	for namespace := range s.watchedNamespaces {
		// Create a new list instance of the same type as 'list'
		tempList, err := newListOfType(list)
		if err != nil {
			return err
		}

		// Append an option to query this specific namespace.
		optsForNS := append(opts, client.InNamespace(namespace))
		if err := s.Client.List(ctx, tempList, optsForNS...); err != nil {
			return err
		}
		// core.Log.Info("CALLING MY CUSTOM LIST tempList", "tempList", tempList)

		// Extract the items from the temporary list.
		items, err := meta.ExtractList(tempList)
		if err != nil {
			return err
		}
		// core.Log.Info("CALLING MY CUSTOM LIST items", "items", items)
		allItems = append(allItems, items...)
	}

	// Set all aggregated items into the original list.
	if err := meta.SetList(list, allItems); err != nil {
		return err
	}
	// core.Log.Info("CALLING MY CUSTOM LIST list", "list", list)
	return nil
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
