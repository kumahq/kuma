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
	kube_metrics "sigs.k8s.io/controller-runtime/pkg/metrics"
	kube_metricsserver "sigs.k8s.io/controller-runtime/pkg/metrics/server"
	kube_webhook "sigs.k8s.io/controller-runtime/pkg/webhook"

	config_core "github.com/kumahq/kuma/pkg/config/core"
	"github.com/kumahq/kuma/pkg/core"
	core_plugins "github.com/kumahq/kuma/pkg/core/plugins"
	core_runtime "github.com/kumahq/kuma/pkg/core/runtime"
	"github.com/kumahq/kuma/pkg/core/runtime/component"
	core_metrics "github.com/kumahq/kuma/pkg/metrics"
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

func (p *plugin) BeforeBootstrap(b *core_runtime.Builder, cfg core_plugins.PluginConfig) error {
	if b.Config().Environment != config_core.KubernetesEnvironment {
		return nil
	}
	scheme, err := NewScheme()
	if err != nil {
		return err
	}
	restClientConfig := kube_ctrl.GetConfigOrDie()
	restClientConfig.QPS = float32(b.Config().Runtime.Kubernetes.ClientConfig.Qps)
	restClientConfig.Burst = b.Config().Runtime.Kubernetes.ClientConfig.BurstQps

	systemNamespace := b.Config().Store.Kubernetes.SystemNamespace
	mgr, err := kube_ctrl.NewManager(
		restClientConfig,
		kube_ctrl.Options{
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
		},
	)
	if err != nil {
		return err
	}

	secretClient, err := createSecretClient(b.AppCtx(), scheme, systemNamespace, restClientConfig, mgr.GetRESTMapper())
	if err != nil {
		return err
	}

	b.WithExtensions(k8s_extensions.NewManagerContext(b.Extensions(), mgr))
	b.WithComponentManager(&kubeComponentManager{Manager: mgr})

	b.WithExtensions(k8s_extensions.NewSecretClientContext(b.Extensions(), secretClient))
	if expTime := b.Config().Runtime.Kubernetes.MarshalingCacheExpirationTime.Duration; expTime > 0 {
		b.WithExtensions(k8s_extensions.NewResourceConverterContext(b.Extensions(), k8s.NewCachingConverter(expTime)))
	} else {
		b.WithExtensions(k8s_extensions.NewResourceConverterContext(b.Extensions(), k8s.NewSimpleConverter()))
	}
	b.WithExtensions(k8s_extensions.NewCompositeValidatorContext(b.Extensions(), &k8s_common.CompositeValidator{}))
	zoneName := core_metrics.ZoneNameOrMode(b.Config().Mode, b.Config().Multizone.Zone.Name)
	metrics, err := core_metrics.NewMetricsOfRegistererGatherer(zoneName, kube_metrics.Registry)
	if err != nil {
		return err
	}
	b.WithMetrics(metrics)
	return nil
}

// We need separate client for Secrets, because we don't have (get/list/watch) RBAC for all namespaces / cluster scope.
// Kubernetes cache lists resources under the hood from all Namespace unless we specify the "Namespace" in Options.
// If we try to use regular cached client for Secrets then we will see following error: E1126 10:42:52.097662       1 reflector.go:178] pkg/mod/k8s.io/client-go@v0.18.9/tools/cache/reflector.go:125: Failed to list *v1.Secret: secrets is forbidden: User "system:serviceaccount:kuma-system:kuma-control-plane" cannot list resource "secrets" in API group "" at the cluster scope
// We cannot specify this Namespace parameter for the main cache in ControllerManager because it affect all the resources, therefore we need separate client with cache for Secrets.
// The alternative was to use non-cached client, but it had performance problems.
func createSecretClient(appCtx context.Context, scheme *kube_runtime.Scheme, systemNamespace string, config *rest.Config, restMapper meta.RESTMapper) (kube_client.Client, error) {
	resyncPeriod := 10 * time.Hour // default resyncPeriod in Kubernetes
	kubeCache, err := cache.New(config, cache.Options{
		Scheme:            scheme,
		Mapper:            restMapper,
		SyncPeriod:        &resyncPeriod,
		DefaultNamespaces: map[string]cache.Config{systemNamespace: {}},
	})
	if err != nil {
		return nil, err
	}

	// We are listing secrets by our custom "type", therefore we need to add index by this field into cache
	err = kubeCache.IndexField(appCtx, &kube_core.Secret{}, "type", func(object kube_client.Object) []string {
		secret := object.(*kube_core.Secret)
		return []string{string(secret.Type)}
	})
	if err != nil {
		return nil, errors.Wrap(err, "could not add index of Secret cache by field 'type'")
	}

	// According to ControllerManager code, cache needs to start before all the Runnables (our Components)
	// So we need separate go routine to start a cache and then wait for cache
	go func() {
		if err := kubeCache.Start(appCtx); err != nil {
			// According to implementations, there is no case when error is returned. It just for the Runnable contract.
			log.Error(err, "could not start the secret k8s cache")
		}
	}()

	if ok := kubeCache.WaitForCacheSync(appCtx); !ok {
		// ControllerManager ignores case when WaitForCacheSync returns false.
		// It might be a better idea to return an error and stop the Control Plane altogether, but sticking to return error for now.
		core.Log.Error(errors.New("could not sync secret cache"), "failed to wait for cache")
	}

	return kube_client.New(config, kube_client.Options{
		Scheme: scheme,
		Mapper: restMapper,
		Cache: &kube_client.CacheOptions{
			Reader: kubeCache,
		},
	})
}

func (p *plugin) AfterBootstrap(b *core_runtime.Builder, _ core_plugins.PluginConfig) error {
	if b.Config().Environment != config_core.KubernetesEnvironment {
		return nil
	}
	apiServerAddress := os.Getenv("KUBERNETES_SERVICE_HOST")
	port := os.Getenv("KUBERNETES_SERVICE_PORT")
	apiServerPort, err := strconv.ParseUint(port, 10, 32)
	if err != nil {
		return errors.Wrapf(err, "could not parse KUBERNETES_SERVICE_PORT environment variable")
	}

	b.XDS().Hooks.AddResourceSetHook(hooks.NewApiServerBypass(apiServerAddress, uint32(apiServerPort)))

	return nil
}

func (p *plugin) Name() core_plugins.PluginName {
	return core_plugins.Kubernetes
}

func (p *plugin) Order() int {
	return core_plugins.EnvironmentPreparingOrder
}

type kubeComponentManager struct {
	kube_ctrl.Manager
	components []component.Component
}

var _ component.Manager = &kubeComponentManager{}

const (
	// See https://github.com/kubernetes-sigs/controller-runtime/blob/785762383bc52f4b309dbc9d8f8e9239ff391198/pkg/manager/internal.go#L604
	leaderElectionLost = "leader election lost"
)

func (cm *kubeComponentManager) Start(done <-chan struct{}) error {
	ctx, cancel := context.WithCancel(context.Background())
	go func() {
		defer cancel()
		<-done
	}()

	defer cm.waitForDone()

	if err := cm.Manager.Start(ctx); err != nil {
		if err.Error() == leaderElectionLost {
			cm.GetLogger().Info("leader election lost, stopping")
			return nil
		}
		return errors.Wrap(err, "error running Kubernetes Manager")
	}
	return nil
}

// Extra check that component.Component implements LeaderElectionRunnable so the leader election works so we won't break leader election on K8S when refactoring component.Component
var _ kube_manager.LeaderElectionRunnable = component.ComponentFunc(func(i <-chan struct{}) error {
	return nil
})

func (k *kubeComponentManager) Add(components ...component.Component) error {
	for _, c := range components {
		k.components = append(k.components, c)
		if err := k.Manager.Add(&componentRunnableAdaptor{Component: c}); err != nil {
			return err
		}
	}
	return nil
}

func (k *kubeComponentManager) Ready() bool {
	for _, c := range k.components {
		if rc, ok := c.(component.ReadyComponent); ok && !rc.Ready() {
			return false
		}
	}
	return true
}

func (k *kubeComponentManager) waitForDone() {
	for _, c := range k.components {
		if gc, ok := c.(component.GracefulComponent); ok {
			gc.WaitForDone()
		}
	}
}

// This adaptor is required unless component.Component takes a context as input
type componentRunnableAdaptor struct {
	component.Component
}

func (c componentRunnableAdaptor) Start(ctx context.Context) error {
	return c.Component.Start(ctx.Done())
}

func (c componentRunnableAdaptor) NeedLeaderElection() bool {
	return c.Component.NeedLeaderElection()
}

func (c componentRunnableAdaptor) Ready() bool {
	if ready, ok := c.Component.(component.ReadyComponent); ok {
		return ready.Ready()
	}
	return true
}

var (
	_ kube_manager.LeaderElectionRunnable = &componentRunnableAdaptor{}
	_ kube_manager.Runnable               = &componentRunnableAdaptor{}
)
