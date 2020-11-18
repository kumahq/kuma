package k8s

import (
	kube_runtime "k8s.io/apimachinery/pkg/runtime"
	kube_ctrl "sigs.k8s.io/controller-runtime"
	kube_manager "sigs.k8s.io/controller-runtime/pkg/manager"

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

func (p *plugin) Bootstrap(b *core_runtime.Builder, _ core_plugins.PluginConfig) error {
	scheme := kube_runtime.NewScheme()
	mgr, err := kube_ctrl.NewManager(
		kube_ctrl.GetConfigOrDie(),
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
	b.WithComponentManager(&kubeComponentManager{mgr})
	b.WithExtensions(k8s_extensions.NewManagerContext(b.Extensions(), mgr))
	if expTime := b.Config().Runtime.Kubernetes.MarshalingCacheExpirationTime; expTime > 0 {
		b.WithExtensions(k8s_extensions.NewResourceConverterContext(b.Extensions(), k8s.NewCachingConverter(expTime)))
	} else {
		b.WithExtensions(k8s_extensions.NewResourceConverterContext(b.Extensions(), k8s.NewSimpleConverter()))
	}
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
