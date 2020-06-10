package k8s

import (
	"fmt"
	"strings"

	"github.com/Kong/kuma/pkg/core/secrets/manager"

	"github.com/pkg/errors"
	kube_schema "k8s.io/apimachinery/pkg/runtime/schema"
	kube_ctrl "sigs.k8s.io/controller-runtime"
	kube_webhook "sigs.k8s.io/controller-runtime/pkg/webhook"

	"github.com/Kong/kuma/pkg/core"
	managers_mesh "github.com/Kong/kuma/pkg/core/managers/apis/mesh"
	core_plugins "github.com/Kong/kuma/pkg/core/plugins"
	mesh_core "github.com/Kong/kuma/pkg/core/resources/apis/mesh"
	core_model "github.com/Kong/kuma/pkg/core/resources/model"
	core_registry "github.com/Kong/kuma/pkg/core/resources/registry"
	core_runtime "github.com/Kong/kuma/pkg/core/runtime"
	k8s_resources "github.com/Kong/kuma/pkg/plugins/resources/k8s"
	mesh_k8s "github.com/Kong/kuma/pkg/plugins/resources/k8s/native/api/v1alpha1"
	k8s_registry "github.com/Kong/kuma/pkg/plugins/resources/k8s/native/pkg/registry"
	k8s_controllers "github.com/Kong/kuma/pkg/plugins/runtime/k8s/controllers"
	k8s_webhooks "github.com/Kong/kuma/pkg/plugins/runtime/k8s/webhooks"
	"github.com/Kong/kuma/pkg/plugins/runtime/k8s/webhooks/injector"
	k8s_runtime "github.com/Kong/kuma/pkg/runtime/k8s"
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

	if err := addValidators(mgr, rt); err != nil {
		return err
	}

	addMutators(mgr, rt)

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
		CNIEnabled:          rt.Config().Runtime.Kubernetes.Injector.CNIEnabled,
		ResourceManager:     rt.ResourceManager(),
		DefaultMeshTemplate: rt.Config().Defaults.MeshProto(),
	}
	return reconciler.SetupWithManager(mgr)
}

func addMeshReconciler(mgr kube_ctrl.Manager, rt core_runtime.Runtime) error {
	reconciler := &k8s_controllers.MeshReconciler{
		Client:          mgr.GetClient(),
		Reader:          mgr.GetAPIReader(),
		Log:             core.Log.WithName("controllers").WithName("Mesh"),
		Scheme:          mgr.GetScheme(),
		Converter:       k8s_resources.DefaultConverter(),
		CaManagers:      rt.CaManagers(),
		SystemNamespace: rt.Config().Store.Kubernetes.SystemNamespace,
		ResourceManager: rt.ResourceManager(),
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

func addValidators(mgr kube_ctrl.Manager, rt core_runtime.Runtime) error {
	composite := k8s_webhooks.CompositeValidator{}

	handler := k8s_webhooks.NewValidatingWebhook(k8s_resources.DefaultConverter(), core_registry.Global(), k8s_registry.Global())
	composite.AddValidator(handler)

	coreMeshValidator := managers_mesh.MeshValidator{CaManagers: rt.CaManagers()}
	k8sMeshValidator := k8s_webhooks.NewMeshValidatorWebhook(coreMeshValidator, k8s_resources.DefaultConverter(), rt.ResourceManager())
	composite.AddValidator(k8sMeshValidator)

	path := "/validate-kuma-io-v1alpha1"
	mgr.GetWebhookServer().Register(path, composite.WebHook())
	log.Info("Registering a validation composite webhook", "path", path)

	mgr.GetWebhookServer().Register("/validate-v1-service", &kube_webhook.Admission{Handler: &k8s_webhooks.ServiceValidator{}})
	log.Info("Registering a validation webhook for v1/Service", "path", "/validate-v1-service")

	secretValidator := &k8s_webhooks.SecretValidator{
		Client:    mgr.GetClient(),
		Validator: manager.NewSecretValidator(rt.CaManagers(), rt.ResourceStore()),
	}
	mgr.GetWebhookServer().Register("/validate-v1-secret", &kube_webhook.Admission{Handler: secretValidator})
	log.Info("Registering a validation webhook for v1/Secret", "path", "/validate-v1-secret")

	return nil
}

func addMutators(mgr kube_ctrl.Manager, rt core_runtime.Runtime) {
	kumaInjector := injector.New(
		rt.Config().Runtime.Kubernetes.Injector,
		rt.Config().ApiServer.Catalog.ApiServer.Url,
		mgr.GetClient(),
	)
	mgr.GetWebhookServer().Register("/inject-sidecar", k8s_webhooks.PodMutatingWebhook(kumaInjector.InjectKuma))

	ownerRefMutator := &k8s_webhooks.OwnerReferenceMutator{
		Client:       mgr.GetClient(),
		CoreRegistry: core_registry.Global(),
		K8sRegistry:  k8s_registry.Global(),
		Converter:    k8s_resources.DefaultConverter(),
		Scheme:       mgr.GetScheme(),
	}
	mgr.GetWebhookServer().Register("/owner-reference-kuma-io-v1alpha1", &kube_webhook.Admission{Handler: ownerRefMutator})
}
