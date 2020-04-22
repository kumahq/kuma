package k8s

import (
	core_plugins "github.com/Kong/kuma/pkg/core/plugins"
	core_runtime "github.com/Kong/kuma/pkg/core/runtime"
	"github.com/Kong/kuma/pkg/core/runtime/component"

	kube_runtime "k8s.io/apimachinery/pkg/runtime"
	kube_rest "k8s.io/client-go/rest"
	kube_ctrl "sigs.k8s.io/controller-runtime"
	kube_cache "sigs.k8s.io/controller-runtime/pkg/cache"
	kube_client "sigs.k8s.io/controller-runtime/pkg/client"

	k8s_runtime "github.com/Kong/kuma/pkg/runtime/k8s"
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
			Scheme: scheme,
			NewClient: func(_ kube_cache.Cache, config *kube_rest.Config, options kube_client.Options) (kube_client.Client, error) {
				// Use client without cache for two reasons
				// 1) K8S Cached client does not support chunking (pagination)
				// 2) We maintain our cache in Kuma so we don't want to have duplicated entries in the memory
				return kube_client.New(config, options)
			},
			// Admission WebHook Server
			Host:    b.Config().Runtime.Kubernetes.AdmissionServer.Address,
			Port:    int(b.Config().Runtime.Kubernetes.AdmissionServer.Port),
			CertDir: b.Config().Runtime.Kubernetes.AdmissionServer.CertDir,
		},
	)
	if err != nil {
		return err
	}
	b.WithComponentManager(&kubeComponentManager{mgr})
	b.WithExtensions(k8s_runtime.NewManagerContext(b.Extensions(), mgr))
	return nil
}

type kubeComponentManager struct {
	kube_ctrl.Manager
}

var _ component.Manager = &kubeComponentManager{}

func (k *kubeComponentManager) Add(components ...component.Component) error {
	for _, c := range components {
		if err := k.Manager.Add(c); err != nil {
			return err
		}
	}
	return nil
}
