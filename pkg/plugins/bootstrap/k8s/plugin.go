package k8s

import (
	core_plugins "github.com/Kong/kuma/pkg/core/plugins"
	core_runtime "github.com/Kong/kuma/pkg/core/runtime"
	k8s_runtime "github.com/Kong/kuma/pkg/runtime/k8s"

	kube_runtime "k8s.io/apimachinery/pkg/runtime"
	kube_ctrl "sigs.k8s.io/controller-runtime"
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
			// Admission WebHook Server
			Host:    b.Config().Runtime.Kubernetes.AdmissionServer.Address,
			Port:    int(b.Config().Runtime.Kubernetes.AdmissionServer.Port),
			CertDir: b.Config().Runtime.Kubernetes.AdmissionServer.CertDir,
		},
	)
	if err != nil {
		return err
	}
	b.WithComponentManager(mgr)
	b.WithExtensions(k8s_runtime.NewManagerContext(b.Extensions(), mgr))
	return nil
}
