package k8s

import (
	"fmt"
	"strings"

	"github.com/pkg/errors"

	"github.com/Kong/kuma/pkg/core"
	core_plugins "github.com/Kong/kuma/pkg/core/plugins"
	mesh_core "github.com/Kong/kuma/pkg/core/resources/apis/mesh"
	core_model "github.com/Kong/kuma/pkg/core/resources/model"
	core_runtime "github.com/Kong/kuma/pkg/core/runtime"
	k8s_resources "github.com/Kong/kuma/pkg/plugins/resources/k8s"
	mesh_k8s "github.com/Kong/kuma/pkg/plugins/resources/k8s/native/api/v1alpha1"
	k8s_controllers "github.com/Kong/kuma/pkg/plugins/runtime/k8s/controllers"
	k8s_webhooks "github.com/Kong/kuma/pkg/plugins/runtime/k8s/webhooks"
	k8s_runtime "github.com/Kong/kuma/pkg/runtime/k8s"

	kube_schema "k8s.io/apimachinery/pkg/runtime/schema"
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

	if err := addControllers(mgr, rt); err != nil {
		return err
	}

	return addDefaulters(mgr)
}

func addControllers(mgr kube_ctrl.Manager, rt core_runtime.Runtime) error {
	if err := addNamespaceReconciler(mgr, rt); err != nil {
		return err
	}
	return addMeshReconciler(mgr, rt)
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

func addMeshReconciler(mgr kube_ctrl.Manager, rt core_runtime.Runtime) error {
	reconciler := &k8s_controllers.MeshReconciler{
		Client:           mgr.GetClient(),
		Reader:           mgr.GetAPIReader(),
		Log:              core.Log.WithName("controllers").WithName("Mesh"),
		Scheme:           mgr.GetScheme(),
		Converter:        k8s_resources.DefaultConverter(),
		BuiltinCaManager: rt.BuiltinCaManager(),
		SystemNamespace:  rt.Config().Store.Kubernetes.SystemNamespace,
		ResourceManager:  rt.ResourceManager(),
	}
	return reconciler.SetupWithManager(mgr)
}

func addDefaulters(mgr kube_ctrl.Manager) error {
	if err := mesh_k8s.AddToScheme(mgr.GetScheme()); err != nil {
		return errors.Wrapf(err, "could not add %q to scheme", mesh_k8s.GroupVersion)
	}

	addDefaulter(mgr, mesh_k8s.GroupVersion.WithKind("Mesh"),
		func() core_model.Resource {
			return &mesh_core.MeshResource{}
		})

	return nil
}

func addDefaulter(mgr kube_ctrl.Manager, gvk kube_schema.GroupVersionKind, factory func() core_model.Resource) {
	wh := k8s_webhooks.DefaultingWebhookFor(factory, k8s_resources.DefaultConverter())
	path := generateDefaulterPath(gvk)
	log.Info("Registering a defaulting webhook", "GVK", gvk, "path", path)
	mgr.GetWebhookServer().Register(path, wh)
}

func generateDefaulterPath(gvk kube_schema.GroupVersionKind) string {
	return fmt.Sprintf("/default-%s-%s-%s", strings.Replace(gvk.Group, ".", "-", -1), gvk.Version, strings.ToLower(gvk.Kind))
}
