package k8s

import (
	"context"
	"time"

	"github.com/pkg/errors"
	kube_core "k8s.io/api/core/v1"
	kube_runtime "k8s.io/apimachinery/pkg/runtime"
	kube_ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/cache"
	kube_client "sigs.k8s.io/controller-runtime/pkg/client"
	kube_manager "sigs.k8s.io/controller-runtime/pkg/manager"

	"github.com/kumahq/kuma/pkg/core"
	kuma_kube_cache "github.com/kumahq/kuma/pkg/plugins/bootstrap/k8s/cache"

	"github.com/kumahq/kuma/pkg/plugins/resources/k8s"

	core_plugins "github.com/kumahq/kuma/pkg/core/plugins"
	core_runtime "github.com/kumahq/kuma/pkg/core/runtime"
	"github.com/kumahq/kuma/pkg/core/runtime/component"
	k8s_extensions "github.com/kumahq/kuma/pkg/plugins/extensions/k8s"
)

var _ core_plugins.BootstrapPlugin = &plugin{}

type plugin struct{}

func init() {
	core_plugins.Register(core_plugins.Kubernetes, &plugin{})
}

func (p *plugin) BeforeBootstrap(b *core_runtime.Builder, _ core_plugins.PluginConfig) error {
	scheme := kube_runtime.NewScheme()
	config := kube_ctrl.GetConfigOrDie()
	mgr, err := kube_ctrl.NewManager(
		config,
		kube_ctrl.Options{
			Scheme:   scheme,
			NewCache: kuma_kube_cache.New,
			// Admission WebHook Server
			Host:                    b.Config().Runtime.Kubernetes.AdmissionServer.Address,
			Port:                    int(b.Config().Runtime.Kubernetes.AdmissionServer.Port),
			CertDir:                 b.Config().Runtime.Kubernetes.AdmissionServer.CertDir,
			LeaderElection:          true,
			LeaderElectionID:        "kuma-cp-leader",
			LeaderElectionNamespace: b.Config().Store.Kubernetes.SystemNamespace,
		},
	)
	if err != nil {
		return err
	}

	resyncPeriod := 10 * time.Hour
	cache, err := kuma_kube_cache.New(config, cache.Options{Scheme: scheme, Mapper: mgr.GetRESTMapper(), Resync: &resyncPeriod, Namespace: b.Config().Store.Kubernetes.SystemNamespace})
	if err != nil {
		return err
	}
	if err := kube_core.AddToScheme(mgr.GetScheme()); err != nil {
		return errors.Wrapf(err, "could not add %q to scheme", kube_core.SchemeGroupVersion)
	}
	err = cache.IndexField(context.Background(), &kube_core.Secret{}, "type", func(object kube_runtime.Object) []string {
		secret := object.(*kube_core.Secret)
		return []string{string(secret.Type)}
	})
	if err != nil {
		return err
	}

	go func() {
		stopCh := new(chan struct{})
		if err := cache.Start(*stopCh); err != nil {
			panic(err)
		}
	}()

	stopCh := new(chan struct{})
	if ok := cache.WaitForCacheSync(*stopCh); !ok {
		core.Log.Info("could not sync cache")
	}

	writeObj, err := kube_manager.DefaultNewClient(cache, config, kube_client.Options{Scheme: scheme, Mapper: mgr.GetRESTMapper()})
	if err != nil {
		return err
	}

	b.WithComponentManager(&kubeComponentManager{mgr})
	b.WithExtensions(k8s_extensions.NewManagerContext(b.Extensions(), mgr))
	b.WithExtensions(k8s_extensions.NewNonCachedClientContext(b.Extensions(), writeObj))
	if expTime := b.Config().Runtime.Kubernetes.MarshalingCacheExpirationTime; expTime > 0 {
		b.WithExtensions(k8s_extensions.NewResourceConverterContext(b.Extensions(), k8s.NewCachingConverter(expTime)))
	} else {
		b.WithExtensions(k8s_extensions.NewResourceConverterContext(b.Extensions(), k8s.NewSimpleConverter()))
	}
	return nil
}

func (p *plugin) AfterBootstrap(b *core_runtime.Builder, _ core_plugins.PluginConfig) error {
	return nil
}

type kubeComponentManager struct {
	kube_ctrl.Manager
}

var _ component.Manager = &kubeComponentManager{}

// Extra check that component.Component implements LeaderElectionRunnable so the leader election works so we won't break leader election on K8S when refactoring component.Component
var _ kube_manager.LeaderElectionRunnable = component.ComponentFunc(func(i <-chan struct{}) error {
	return nil
})

func (k *kubeComponentManager) Add(components ...component.Component) error {
	for _, c := range components {
		if err := k.Manager.Add(c); err != nil {
			return err
		}
	}
	return nil
}
