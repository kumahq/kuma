package k8s

import (
	"github.com/pkg/errors"

	"github.com/Kong/kuma/pkg/core"
	core_plugins "github.com/Kong/kuma/pkg/core/plugins"
	core_runtime "github.com/Kong/kuma/pkg/core/runtime"
	k8s_controllers "github.com/Kong/kuma/pkg/plugins/runtime/k8s/controllers"
	k8s_runtime "github.com/Kong/kuma/pkg/runtime/k8s"

	kube_ctrl "sigs.k8s.io/controller-runtime"
)

var (
	log = core.Log.WithName("plugin").WithName("runtime").WithName("k8s")
)

var _ core_plugins.RuntimePlugin = &plugin{}

type plugin struct{}

func init() {
	core_plugins.Register(core_plugins.Kubernetes, &plugin{})
}

func (p *plugin) Customize(rt core_runtime.Runtime) error {
	mgr, ok := k8s_runtime.FromManagerContext(rt.Extensions())
	if !ok {
		return errors.Errorf("k8s controller runtime Manager hasn't been configured")
	}

	return addControllers(mgr, rt)
}

func addControllers(mgr kube_ctrl.Manager, rt core_runtime.Runtime) error {
	return addNamespaceReconciler(mgr, rt)
}

func addNamespaceReconciler(mgr kube_ctrl.Manager, rt core_runtime.Runtime) error {
	reconciler := &k8s_controllers.NamespaceReconciler{
		Client:              mgr.GetClient(),
		Log:                 core.Log.WithName("controllers").WithName("Namespace"),
		SystemNamespace:     rt.Config().Store.Kubernetes.SystemNamespace,
		ResourceManager:     rt.ResourceManager(),
		DefaultMeshTemplate: rt.Config().Defaults.MeshProto(),
	}
	return reconciler.SetupWithManager(mgr)
}
