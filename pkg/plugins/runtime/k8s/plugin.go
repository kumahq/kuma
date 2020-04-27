package k8s

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	kube_core "k8s.io/api/core/v1"
	kube_admission "sigs.k8s.io/controller-runtime/pkg/webhook/admission"

	"github.com/Kong/kuma/pkg/plugins/runtime/k8s/webhooks/injector"

	"github.com/pkg/errors"

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
	k8s_runtime "github.com/Kong/kuma/pkg/runtime/k8s"

	kube_schema "k8s.io/apimachinery/pkg/runtime/schema"
	kube_ctrl "sigs.k8s.io/controller-runtime"
	kube_webhook "sigs.k8s.io/controller-runtime/pkg/webhook"
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
	k8sMeshValidator := k8s_webhooks.NewMeshValidatorWebhook(coreMeshValidator, k8s_resources.DefaultConverter())
	composite.AddValidator(k8sMeshValidator)

	path := "/validate-kuma-io-v1alpha1"
	mgr.GetWebhookServer().Register(path, composite.WebHook())
	log.Info("Registering a validation composite webhook", "path", path)

	mgr.GetWebhookServer().Register("/validate-v1-service", &kube_webhook.Admission{Handler: &k8s_webhooks.ServiceValidator{}})
	log.Info("Registering a validation webhook for v1/Service", "path", "/validate-v1-service")

	secretValidator := &k8s_webhooks.SecretValidator{
		Client: mgr.GetClient(),
	}
	mgr.GetWebhookServer().Register("/validate-v1-secret", &kuba_webhook.Admission{Handler: secretValidator})
	log.Info("Registering a validation webhook for v1/Secret", "path", "/validate-v1-secret")

	return nil
}

func addMutators(mgr kube_ctrl.Manager, rt core_runtime.Runtime) {
	mgr.GetWebhookServer().Register("/inject-sidecar",
		PodMutatingWebhook(injector.New(rt.Config().Runtime.Kubernetes.Injector, mgr.GetClient()).InjectKuma))
}

//func Setup(mgr kube_manager.Manager, cfg *kuma_injector_conf.Config) error {
//	webhookServer := &kube_webhook.Server{
//		Host:    cfg.WebHookServer.Address,
//		Port:    int(cfg.WebHookServer.Port),
//		CertDir: cfg.WebHookServer.CertDir,
//	}
//	if err := mesh_k8s.AddToScheme(mgr.GetScheme()); err != nil {
//		return errors.Wrap(err, "could not add to scheme")
//	}
//
//	if err := k8scnicncfio.AddToScheme(mgr.GetScheme()); err != nil {
//		return errors.Wrap(err, "could not add to scheme")
//	}
//
//	webhookServer.Register("/inject-sidecar", PodMutatingWebhook(injector.New(cfg.Injector, mgr.GetClient()).InjectKuma))
//	webhookServer.WebhookMux.HandleFunc("/healthy", func(resp http.ResponseWriter, _ *http.Request) {
//		resp.WriteHeader(http.StatusOK)
//	})
//	webhookServer.WebhookMux.HandleFunc("/ready", func(resp http.ResponseWriter, _ *http.Request) {
//		resp.WriteHeader(http.StatusOK)
//	})
//	return mgr.Add(webhookServer)
//}

type PodMutator func(*kube_core.Pod) error

func PodMutatingWebhook(mutator PodMutator) *kube_admission.Webhook {
	return &kube_admission.Webhook{
		Handler: &podMutatingHandler{mutator: mutator},
	}
}

type podMutatingHandler struct {
	mutator PodMutator
}

func (h *podMutatingHandler) Handle(ctx context.Context, req kube_webhook.AdmissionRequest) kube_webhook.AdmissionResponse {
	//webhookLog.V(1).Info("received request", "request", req)
	var pod kube_core.Pod
	if err := json.Unmarshal(req.Object.Raw, &pod); err != nil {
		return kube_admission.Errored(http.StatusBadRequest, err)
	}
	if err := h.mutator(&pod); err != nil {
		return kube_admission.Errored(http.StatusInternalServerError, err)
	}
	mutatedRaw, err := json.Marshal(pod)
	if err != nil {
		return kube_admission.Errored(http.StatusInternalServerError, err)
	}
	return kube_admission.PatchResponseFromRaw(req.Object.Raw, mutatedRaw)
}
