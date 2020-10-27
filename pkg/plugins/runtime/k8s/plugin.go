package k8s

import (
	"fmt"
	"strings"

	"github.com/kumahq/kuma/pkg/plugins/resources/k8s"
	"github.com/kumahq/kuma/pkg/plugins/runtime/k8s/controllers"

	config_core "github.com/kumahq/kuma/pkg/config/core"

	"github.com/kumahq/kuma/pkg/core/secrets/manager"

	"github.com/pkg/errors"
	kube_schema "k8s.io/apimachinery/pkg/runtime/schema"
	kube_ctrl "sigs.k8s.io/controller-runtime"
	kube_webhook "sigs.k8s.io/controller-runtime/pkg/webhook"

	"github.com/kumahq/kuma/pkg/core"
	managers_mesh "github.com/kumahq/kuma/pkg/core/managers/apis/mesh"
	core_plugins "github.com/kumahq/kuma/pkg/core/plugins"
	mesh_core "github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
	core_model "github.com/kumahq/kuma/pkg/core/resources/model"
	core_registry "github.com/kumahq/kuma/pkg/core/resources/registry"
	core_runtime "github.com/kumahq/kuma/pkg/core/runtime"
	k8s_extensions "github.com/kumahq/kuma/pkg/plugins/extensions/k8s"
	mesh_k8s "github.com/kumahq/kuma/pkg/plugins/resources/k8s/native/api/v1alpha1"
	k8s_registry "github.com/kumahq/kuma/pkg/plugins/resources/k8s/native/pkg/registry"
	k8s_controllers "github.com/kumahq/kuma/pkg/plugins/runtime/k8s/controllers"
	k8s_webhooks "github.com/kumahq/kuma/pkg/plugins/runtime/k8s/webhooks"
	"github.com/kumahq/kuma/pkg/plugins/runtime/k8s/webhooks/injector"
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
	mgr, ok := k8s_extensions.FromManagerContext(rt.Extensions())
	if !ok {
		return errors.Errorf("k8s controller runtime Manager hasn't been configured")
	}

	converter, ok := k8s_extensions.FromResourceConverterContext(rt.Extensions())
	if !ok {
		return errors.Errorf("k8s resource converter hasn't been configured")
	}

	if err := addControllers(mgr, rt, converter); err != nil {
		return err
	}

	// Mutators and Validators convert resources from Request (not from the Store)
	// these resources doesn't have ResourceVersion, we can't cache them
	simpleConverter := k8s.NewSimpleConverter()
	if err := addValidators(mgr, rt, simpleConverter); err != nil {
		return err
	}

	if err := addMutators(mgr, rt, simpleConverter); err != nil {
		return err
	}

	if err := addDefaulters(mgr, simpleConverter); err != nil {
		return err
	}

	return nil
}

func addControllers(mgr kube_ctrl.Manager, rt core_runtime.Runtime, converter k8s_extensions.Converter) error {
	if err := addNamespaceReconciler(mgr, rt); err != nil {
		return err
	}
	if err := addMeshReconciler(mgr, rt, converter); err != nil {
		return err
	}
	if err := addPodReconciler(mgr, rt, converter); err != nil {
		return err
	}
	return nil
}

func addNamespaceReconciler(mgr kube_ctrl.Manager, rt core_runtime.Runtime) error {
	reconciler := &k8s_controllers.NamespaceReconciler{
		Client:     mgr.GetClient(),
		Log:        core.Log.WithName("controllers").WithName("Namespace"),
		CNIEnabled: rt.Config().Runtime.Kubernetes.Injector.CNIEnabled,
	}
	return reconciler.SetupWithManager(mgr)
}

func addMeshReconciler(mgr kube_ctrl.Manager, rt core_runtime.Runtime, converter k8s_extensions.Converter) error {
	if rt.Config().Mode == config_core.Remote {
		return nil
	}
	reconciler := &k8s_controllers.MeshReconciler{
		Client:          mgr.GetClient(),
		Reader:          mgr.GetAPIReader(),
		Log:             core.Log.WithName("controllers").WithName("Mesh"),
		Scheme:          mgr.GetScheme(),
		Converter:       converter,
		CaManagers:      rt.CaManagers(),
		SystemNamespace: rt.Config().Store.Kubernetes.SystemNamespace,
		ResourceManager: rt.ResourceManager(),
	}
	return reconciler.SetupWithManager(mgr)
}

func addPodReconciler(mgr kube_ctrl.Manager, rt core_runtime.Runtime, converter k8s_extensions.Converter) error {
	reconciler := &controllers.PodReconciler{
		Client:        mgr.GetClient(),
		EventRecorder: mgr.GetEventRecorderFor("k8s.kuma.io/dataplane-generator"),
		Scheme:        mgr.GetScheme(),
		Log:           core.Log.WithName("controllers").WithName("Pod"),
		PodConverter: controllers.PodConverter{
			ServiceGetter:     mgr.GetClient(),
			Zone:              rt.Config().Multicluster.Remote.Zone,
			ResourceConverter: converter,
		},
		SystemNamespace: rt.Config().Store.Kubernetes.SystemNamespace,
	}
	return reconciler.SetupWithManager(mgr)
}

func addDefaulters(mgr kube_ctrl.Manager, converter k8s_extensions.Converter) error {
	if err := mesh_k8s.AddToScheme(mgr.GetScheme()); err != nil {
		return errors.Wrapf(err, "could not add %q to scheme", mesh_k8s.GroupVersion)
	}

	addDefaulter(mgr, mesh_k8s.GroupVersion.WithKind("Mesh"),
		func() core_model.Resource {
			return &mesh_core.MeshResource{}
		}, converter)

	return nil
}

func addDefaulter(
	mgr kube_ctrl.Manager,
	gvk kube_schema.GroupVersionKind,
	factory func() core_model.Resource,
	converter k8s_extensions.Converter,
) {
	wh := k8s_webhooks.DefaultingWebhookFor(factory, converter)
	path := generateDefaulterPath(gvk)
	log.Info("Registering a defaulting webhook", "GVK", gvk, "path", path)
	mgr.GetWebhookServer().Register(path, wh)
}

func generateDefaulterPath(gvk kube_schema.GroupVersionKind) string {
	return fmt.Sprintf("/default-%s-%s-%s", strings.ReplaceAll(gvk.Group, ".", "-"), gvk.Version, strings.ToLower(gvk.Kind))
}

func addValidators(mgr kube_ctrl.Manager, rt core_runtime.Runtime, converter k8s_extensions.Converter) error {
	composite := k8s_webhooks.CompositeValidator{}

	handler := k8s_webhooks.NewValidatingWebhook(converter, core_registry.Global(), k8s_registry.Global(), rt.Config().Mode)
	composite.AddValidator(handler)

	coreMeshValidator := managers_mesh.MeshValidator{CaManagers: rt.CaManagers()}
	k8sMeshValidator := k8s_webhooks.NewMeshValidatorWebhook(coreMeshValidator, converter, rt.ResourceManager())
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

func addMutators(mgr kube_ctrl.Manager, rt core_runtime.Runtime, converter k8s_extensions.Converter) error {
	if rt.Config().Mode != config_core.Global {
		kumaInjector, err := injector.New(
			rt.Config().Runtime.Kubernetes.Injector,
			rt.Config().ApiServer.Catalog.Bootstrap.Url,
			mgr.GetClient(),
			converter,
		)
		if err != nil {
			return err
		}
		mgr.GetWebhookServer().Register("/inject-sidecar", k8s_webhooks.PodMutatingWebhook(kumaInjector.InjectKuma))
	}

	ownerRefMutator := &k8s_webhooks.OwnerReferenceMutator{
		Client:       mgr.GetClient(),
		CoreRegistry: core_registry.Global(),
		K8sRegistry:  k8s_registry.Global(),
		Scheme:       mgr.GetScheme(),
	}
	mgr.GetWebhookServer().Register("/owner-reference-kuma-io-v1alpha1", &kube_webhook.Admission{Handler: ownerRefMutator})
	return nil
}
