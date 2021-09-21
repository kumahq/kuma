package k8s

import (
	"context"
	"os"
	"strconv"
	"time"

	"github.com/pkg/errors"
	kube_core "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/meta"
	kube_runtime "k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/rest"
	kube_ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/cache"
	kube_client "sigs.k8s.io/controller-runtime/pkg/client"
	kube_manager "sigs.k8s.io/controller-runtime/pkg/manager"

	"github.com/kumahq/kuma/pkg/core"
	core_plugins "github.com/kumahq/kuma/pkg/core/plugins"
	core_runtime "github.com/kumahq/kuma/pkg/core/runtime"
	"github.com/kumahq/kuma/pkg/core/runtime/component"
	kuma_kube_cache "github.com/kumahq/kuma/pkg/plugins/bootstrap/k8s/cache"
	"github.com/kumahq/kuma/pkg/plugins/bootstrap/k8s/xds/hooks"
	k8s_common "github.com/kumahq/kuma/pkg/plugins/common/k8s"
	k8s_extensions "github.com/kumahq/kuma/pkg/plugins/extensions/k8s"
	"github.com/kumahq/kuma/pkg/plugins/resources/k8s"
)

var _ core_plugins.BootstrapPlugin = &plugin{}

var log = core.Log.WithName("plugins").WithName("bootstrap").WithName("k8s")

type plugin struct{}

func init() {
	core_plugins.Register(core_plugins.Kubernetes, &plugin{})
}

func (p *plugin) BeforeBootstrap(b *core_runtime.Builder, _ core_plugins.PluginConfig) error {
	scheme, err := NewScheme()
	if err != nil {
		return err
	}
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

	secretClient, err := secretClient(b.AppCtx(), b.Config().Store.Kubernetes.SystemNamespace, config, scheme, mgr.GetRESTMapper())
	if err != nil {
		return err
	}

	b.WithComponentManager(&kubeComponentManager{mgr})
	b.WithExtensions(k8s_extensions.NewManagerContext(b.Extensions(), mgr))
	b.WithExtensions(k8s_extensions.NewSecretClientContext(b.Extensions(), secretClient))
	if expTime := b.Config().Runtime.Kubernetes.MarshalingCacheExpirationTime; expTime > 0 {
		b.WithExtensions(k8s_extensions.NewResourceConverterContext(b.Extensions(), k8s.NewCachingConverter(expTime)))
	} else {
		b.WithExtensions(k8s_extensions.NewResourceConverterContext(b.Extensions(), k8s.NewSimpleConverter()))
	}
	b.WithExtensions(k8s_extensions.NewCompositeValidatorContext(b.Extensions(), &k8s_common.CompositeValidator{}))
	return nil
}

// We need separate client for Secrets, because we don't have (get/list/watch) RBAC for all namespaces / cluster scope.
// Kubernetes cache lists resources under the hood from all Namespace unless we specify the "Namespace" in Options.
// If we try to use regular cached client for Secrets then we will see following error: E1126 10:42:52.097662       1 reflector.go:178] pkg/mod/k8s.io/client-go@v0.18.9/tools/cache/reflector.go:125: Failed to list *v1.Secret: secrets is forbidden: User "system:serviceaccount:kuma-system:kuma-control-plane" cannot list resource "secrets" in API group "" at the cluster scope
// We cannot specify this Namespace parameter for the main cache in ControllerManager because it affect all the resources, therefore we need separate client with cache for Secrets.
// The alternative was to use non-cached client, but it had performance problems.
func secretClient(appCtx context.Context, systemNamespace string, config *rest.Config, scheme *kube_runtime.Scheme, restMapper meta.RESTMapper) (kube_client.Client, error) {
	resyncPeriod := 10 * time.Hour // default resyncPeriod in Kubernetes
	kubeCache, err := kuma_kube_cache.New(config, cache.Options{
		Scheme:    scheme,
		Mapper:    restMapper,
		Resync:    &resyncPeriod,
		Namespace: systemNamespace,
	})
	if err != nil {
		return nil, err
	}

	// We are listing secrets by our custom "type", therefore we need to add index by this field into cache
	err = kubeCache.IndexField(context.Background(), &kube_core.Secret{}, "type", func(object kube_runtime.Object) []string {
		secret := object.(*kube_core.Secret)
		return []string{string(secret.Type)}
	})
	if err != nil {
		return nil, errors.Wrap(err, "could not add index of Secret cache by field 'type'")
	}

	// According to ControllerManager code, cache needs to start before all the Runnables (our Components)
	// So we need separate go routine to start a cache and then wait for cache
	go func() {
		if err := kubeCache.Start(appCtx.Done()); err != nil {
			// According to implementations, there is no case when error is returned. It just for the Runnable contract.
			log.Error(err, "could not start the secret k8s cache")
		}
	}()

	if ok := kubeCache.WaitForCacheSync(appCtx.Done()); !ok {
		// ControllerManager ignores case when WaitForCacheSync returns false.
		// It might be a better idea to return an error and stop the Control Plane altogether, but sticking to return error for now.
		core.Log.Error(errors.New("could not sync secret cache"), "failed to wait for cache")
	}

	return kube_manager.DefaultNewClient(kubeCache, config, kube_client.Options{Scheme: scheme, Mapper: restMapper})
}

func (p *plugin) AfterBootstrap(b *core_runtime.Builder, _ core_plugins.PluginConfig) error {
	apiServerAddress := os.Getenv("KUBERNETES_SERVICE_HOST")
	port := os.Getenv("KUBERNETES_SERVICE_PORT")
	apiServerPort, err := strconv.ParseUint(port, 10, 32)
	if err != nil {
		return errors.Wrapf(err, "could not parse KUBERNETES_SERVICE_PORT environment variable")
	}

	b.XDSHooks().AddResourceSetHook(hooks.NewApiServerBypass(apiServerAddress, uint32(apiServerPort)))

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
