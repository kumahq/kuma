package k8s

import (
	"github.com/pkg/errors"
	kube_ctrl "sigs.k8s.io/controller-runtime"

	"github.com/Kong/kuma/pkg/core"
	core_plugins "github.com/Kong/kuma/pkg/core/plugins"
	"github.com/Kong/kuma/pkg/plugins/discovery/k8s/controllers"
	k8s_runtime "github.com/Kong/kuma/pkg/runtime/k8s"
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
	return addPodReconciler(pc.Config().General.ClusterName, mgr)
}

func addPodReconciler(clusterName string, mgr kube_ctrl.Manager) error {
	reconciler := &controllers.PodReconciler{
		Client:        mgr.GetClient(),
		EventRecorder: mgr.GetEventRecorderFor("k8s.kuma.io/dataplane-generator"),
		Scheme:        mgr.GetScheme(),
		Log:           core.Log.WithName("controllers").WithName("Pod"),
		PodConverter: controllers.PodConverter{
			ServiceGetter: mgr.GetClient(),
			ClusterName:   clusterName,
		},
	}
	return reconciler.SetupWithManager(mgr)
}
