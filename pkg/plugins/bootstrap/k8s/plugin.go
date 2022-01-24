package k8s

import (
	"context"
	"encoding/json"
	"os"
	"strconv"
	"time"

	"github.com/operator-framework/operator-lib/leader"
	"github.com/pkg/errors"
	"golang.org/x/sync/errgroup"
	kube_core "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	kube_runtime "k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/rest"
	kube_ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/cache"
	kube_client "sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/cluster"
	kube_ctrlr "sigs.k8s.io/controller-runtime/pkg/controller"
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
			Host:           b.Config().Runtime.Kubernetes.AdmissionServer.Address,
			Port:           int(b.Config().Runtime.Kubernetes.AdmissionServer.Port),
			CertDir:        b.Config().Runtime.Kubernetes.AdmissionServer.CertDir,
			LeaderElection: false,

			// Disable metrics bind address as we serve metrics some other way.
			MetricsBindAddress: "0",
		},
	)
	if err != nil {
		return err
	}

	systemNamespace := b.Config().Store.Kubernetes.SystemNamespace

	secretClient, err := createSecretClient(b.AppCtx(), scheme, systemNamespace, config, mgr.GetRESTMapper())
	if err != nil {
		return err
	}

	// We can't pass kubeComponentManager into NewManagerContext because of
	// conflicting method signatures.
	kubeManagerWrapper := &kubeManagerWrapper{
		Manager:     mgr,
		controllers: nil,
	}
	b.WithExtensions(k8s_extensions.NewManagerContext(b.Extensions(), kubeManagerWrapper))

	kcm := &kubeComponentManager{
		kubeManagerWrapper:         kubeManagerWrapper,
		oldLeaderElectionNamespace: systemNamespace,
		leaderComponents:           nil,
	}
	b.WithComponentManager(kcm)

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
func createSecretClient(appCtx context.Context, scheme *kube_runtime.Scheme, systemNamespace string, config *rest.Config, restMapper meta.RESTMapper) (kube_client.Client, error) {
	resyncPeriod := 10 * time.Hour // default resyncPeriod in Kubernetes
	kubeCache, err := cache.New(config, cache.Options{
		Scheme:    scheme,
		Mapper:    restMapper,
		Resync:    &resyncPeriod,
		Namespace: systemNamespace,
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

	return cluster.DefaultNewClient(kubeCache, config, kube_client.Options{
		Scheme: scheme,
		Mapper: restMapper,
	})
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

func (p *plugin) Name() core_plugins.PluginName {
	return core_plugins.Kubernetes
}

func (p *plugin) Order() int {
	return core_plugins.EnvironmentPreparingOrder
}

// kubeManagerWrapper exists in order to intercept Manager.Add calls on
// controllers so that we can start them independently of Manager.Start.
type kubeManagerWrapper struct {
	kube_ctrl.Manager
	controllers []kube_ctrlr.Controller
}

// Add calls Manager.Add and passes the runnable.
// If the runnable is a Controller, it will not be immediately Added.
// Controllers must be explicitly added by calling AddControllers.
func (kmw *kubeManagerWrapper) Add(runnable kube_manager.Runnable) error {
	if ctrler, ok := runnable.(kube_ctrlr.Controller); ok {
		kmw.controllers = append(kmw.controllers, ctrler)
		return nil
	}
	return kmw.Manager.Add(runnable)
}

// AddControllers calls Manager.Add on any accumulated controllers.
func (kmw *kubeManagerWrapper) AddControllersToManager() error {
	for _, c := range kmw.controllers {
		if err := kmw.Manager.Add(c); err != nil {
			return errors.Wrap(err, "add controller error")
		}
	}
	return nil
}

type kubeComponentManager struct {
	*kubeManagerWrapper
	oldLeaderElectionNamespace string
	leaderComponents           []component.Component
}

var _ component.Manager = &kubeComponentManager{}

type leaderAnnotation struct {
	HolderIdentity       string `json:"holderIdentity"`
	LeaseDurationSeconds int    `json:"leaseDurationSeconds"`
	AcquireTime          string `json:"acquireTime"`
	RenewTime            string `json:"renewTime"`
	LeaderTransitions    int    `json:"leaderTransistions"`
}

var blockerHolderId = "cp-leader-lock-transition"
var oldLeaderConfigMapName = "kuma-cp-leader"

func makeOldLockAnnotation() string {
	nowStr := time.Now().Format(time.RFC3339)
	annot := &leaderAnnotation{
		HolderIdentity:       blockerHolderId,
		LeaseDurationSeconds: 99999999999999999,
		AcquireTime:          nowStr,
		RenewTime:            nowStr,
		LeaderTransitions:    0,
	}

	annotJson, _ := json.Marshal(annot)
	return string(annotJson)
}

// startLeaderComponents adds all components that need leader election to the
// Manager. The Manager should be running already, any components added after
// that are started immediately.
func (cm *kubeComponentManager) startLeaderComponents() error {
	for _, c := range cm.leaderComponents {
		if err := cm.Manager.Add(&componentRunnableAdaptor{Component: c}); err != nil {
			return errors.Wrap(err, "add component error")
		}
	}
	return cm.kubeManagerWrapper.AddControllersToManager()
}

// Previous versions of kuma-cp used a timeout lock for leader election. We now
// keep the election for the lifetime of the pod. This function forces any previous
// style leader to see itself as having lost its election, and locks out any
// further old leaders.
//
// Only call this after acquiring new-style leader election, so as to only contend
// with old leaders over old locks.
func (cm *kubeComponentManager) forceTakeOldLock(ctx context.Context) error {
	log.Info("checking for deprecated leader locks")
	client := cm.Manager.GetClient()
	ns := cm.oldLeaderElectionNamespace

	pod := &kube_core.Pod{}
	if err := client.Get(ctx, kube_client.ObjectKey{
		Namespace: ns,
		Name:      os.Getenv("POD_NAME"),
	}, pod); err != nil {
		log.Error(err, "unable to retrieve this pod")
		return err
	}

	owner := &metav1.OwnerReference{
		APIVersion: "v1",
		Kind:       "Pod",
		Name:       pod.ObjectMeta.Name,
		UID:        pod.ObjectMeta.UID,
	}

	var mustWait = false

	for {
		newLock := &kube_core.ConfigMap{
			ObjectMeta: metav1.ObjectMeta{
				Name:            oldLeaderConfigMapName,
				Namespace:       ns,
				OwnerReferences: []metav1.OwnerReference{*owner},
				Annotations: map[string]string{
					"control-plane.alpha.kubernetes.io/leader": makeOldLockAnnotation(),
				},
			},
		}

		// Numerous potential races between new and old CP leader in this loop. Just keep
		// trying to grab the lock until it succeeds. Since the old leader will
		// politely die when we acquire lock, and we are relentless, we will eventually
		// prevail.

		err := client.Create(ctx, newLock)
		switch {
		case err == nil:
			// Acquired old lock.
			if mustWait {
				log.Info("waiting 30 seconds for old leader to terminate")
				time.Sleep(30 * time.Second)
			}
			return nil
		case apierrors.IsAlreadyExists(err):
			log.Info("existing deprecated lock found; stealing")
			mustWait = true

			existing := &kube_core.ConfigMap{}
			key := kube_client.ObjectKey{Namespace: ns, Name: oldLeaderConfigMapName}
			err = client.Get(ctx, key, existing)
			if err != nil {
				log.Error(err, "error reading old lock; trying again")
				break
			}

			err := client.Delete(ctx, existing)
			if err != nil {
				log.Error(err, "error deleting old lock; trying again")
			}
		default:
			log.Error(err, "error creating ConfigMap; trying again")
		}
		time.Sleep(1 * time.Second)
	}
}

func (cm *kubeComponentManager) Start(done <-chan struct{}) error {
	baseCtx, cancel := context.WithCancel(context.Background())
	go func() {
		defer cancel()
		<-done
	}()

	eg, ctx := errgroup.WithContext(baseCtx)

	eg.Go(func() error {
		return cm.Manager.Start(ctx)
	})

	eg.Go(func() error {
		// The manager is always elected but this lets us wait until we're sure
		// the other components have started, so we can wait until the Manager
		// _would_ do leader election
		select {
		case <-ctx.Done():
			return nil
		case <-cm.Manager.Elected():
		}

		if err := leader.Become(ctx, "cp-leader"); err != nil {
			return errors.Wrap(err, "leader lock failure")
		}

		// This CP will now be leader. But first, destroy deprecated leader lock,
		// forcing any old leaders to restart as non-leaders.
		if err := cm.forceTakeOldLock(ctx); err != nil {
			return errors.Wrap(err, "error attempting to clean up deprecated lock")
		}

		return cm.startLeaderComponents()
	})

	return errors.Wrap(eg.Wait(), "error running Kubernetes Manager")
}

// Extra check that component.Component implements LeaderElectionRunnable so the leader election works so we won't break leader election on K8S when refactoring component.Component
var _ kube_manager.LeaderElectionRunnable = component.ComponentFunc(func(i <-chan struct{}) error {
	return nil
})

func (k *kubeComponentManager) Add(components ...component.Component) error {
	for _, c := range components {
		if c.NeedLeaderElection() {
			k.leaderComponents = append(k.leaderComponents, c)
		} else if err := k.Manager.Add(&componentRunnableAdaptor{Component: c}); err != nil {
			return err
		}
	}
	return nil
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

var _ kube_manager.LeaderElectionRunnable = &componentRunnableAdaptor{}
var _ kube_manager.Runnable = &componentRunnableAdaptor{}
