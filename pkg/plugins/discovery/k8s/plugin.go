package k8s

import (
	"github.com/Kong/kuma/pkg/core"
	"github.com/Kong/kuma/pkg/plugins/discovery/k8s/controllers"
	"github.com/pkg/errors"

	core_plugins "github.com/Kong/kuma/pkg/core/plugins"
	k8s_runtime "github.com/Kong/kuma/pkg/runtime/k8s"

	kube_ctrl "sigs.k8s.io/controller-runtime"
)

var _ core_plugins.DiscoveryPlugin = &plugin{}

type plugin struct{}

func init() {
	core_plugins.Register(core_plugins.Kubernetes, &plugin{})
}

func (p *plugin) StartDiscovering(pc core_plugins.PluginContext, _ core_plugins.PluginConfig) error {
	mgr, ok := k8s_runtime.FromManagerContext(pc.Extensions())
	if !ok {
		return errors.Errorf("k8s controller runtime Manager hasn't been configured")
	}
	// convert Pods into Dataplanes
	return addPodReconciler(mgr)
}

func addPodReconciler(mgr kube_ctrl.Manager) error {
	reconciler := &controllers.PodReconciler{
		Client:        mgr.GetClient(),
		EventRecorder: mgr.GetEventRecorderFor("k8s.kuma.io/dataplane-generator"),
		Scheme:        mgr.GetScheme(),
		Log:           core.Log.WithName("controllers").WithName("Pod"),
	}
	return reconciler.SetupWithManager(mgr)
}
