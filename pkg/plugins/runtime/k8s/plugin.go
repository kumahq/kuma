package k8s

import (
	"fmt"

	"github.com/Masterminds/semver/v3"
	"github.com/pkg/errors"
	"k8s.io/client-go/discovery"
	kube_ctrl "sigs.k8s.io/controller-runtime"
	kube_webhook "sigs.k8s.io/controller-runtime/pkg/webhook"
	kube_admission "sigs.k8s.io/controller-runtime/pkg/webhook/admission"

	config_core "github.com/kumahq/kuma/pkg/config/core"
	"github.com/kumahq/kuma/pkg/config/core/resources/store"
	"github.com/kumahq/kuma/pkg/core"
	externalservice "github.com/kumahq/kuma/pkg/core/managers/apis/external_service"
	"github.com/kumahq/kuma/pkg/core/managers/apis/ratelimit"
	"github.com/kumahq/kuma/pkg/core/managers/apis/zone"
	core_plugins "github.com/kumahq/kuma/pkg/core/plugins"
	core_registry "github.com/kumahq/kuma/pkg/core/resources/registry"
	core_runtime "github.com/kumahq/kuma/pkg/core/runtime"
	"github.com/kumahq/kuma/pkg/core/secrets/manager"
	"github.com/kumahq/kuma/pkg/dns"
	k8s_common "github.com/kumahq/kuma/pkg/plugins/common/k8s"
	k8s_extensions "github.com/kumahq/kuma/pkg/plugins/extensions/k8s"
	"github.com/kumahq/kuma/pkg/plugins/resources/k8s"
	k8s_registry "github.com/kumahq/kuma/pkg/plugins/resources/k8s/native/pkg/registry"
	k8s_controllers "github.com/kumahq/kuma/pkg/plugins/runtime/k8s/controllers"
	k8s_webhooks "github.com/kumahq/kuma/pkg/plugins/runtime/k8s/webhooks"
	"github.com/kumahq/kuma/pkg/plugins/runtime/k8s/webhooks/injector"
)

var (
	log                                                = core.Log.WithName("plugin").WithName("runtime").WithName("k8s")
	sidecarContainerVersion                            = semver.New(1, 29, 0, "", "")
	_                       core_plugins.RuntimePlugin = &plugin{}
)

type plugin struct{}

func init() {
	core_plugins.Register(core_plugins.Kubernetes, &plugin{})
}

func (p *plugin) Customize(rt core_runtime.Runtime) error {
	if rt.Config().Environment != config_core.KubernetesEnvironment {
		return nil
	}
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

	return nil
}

func addControllers(mgr kube_ctrl.Manager, rt core_runtime.Runtime, converter k8s_common.Converter) error {
	if err := addNamespaceReconciler(mgr, rt); err != nil {
		return err
	}
	if err := addServiceReconciler(mgr); err != nil {
		return err
	}
	if err := addMeshServiceReconciler(mgr, converter); err != nil {
		return err
	}
	if err := addMeshReconciler(mgr, rt); err != nil {
		return err
	}
	if err := addGatewayReconcilers(mgr, rt, converter); err != nil {
		return err
	}
	if err := addPodReconciler(mgr, rt, converter); err != nil {
		return err
	}
	if err := addPodStatusReconciler(mgr, rt, converter); err != nil {
		return err
	}
	if err := addDNS(mgr, rt, converter); err != nil {
		return err
	}

	nodeTaintController := rt.Config().Runtime.Kubernetes.NodeTaintController
	if nodeTaintController.Enabled {
		if err := addCniNodeTaintReconciler(mgr, nodeTaintController.CniApp, nodeTaintController.CniNamespace); err != nil {
			return err
		}
	}

	return nil
}

func addCniNodeTaintReconciler(mgr kube_ctrl.Manager, cniApp string, cniNamespace string) error {
	reconciler := &k8s_controllers.CniNodeTaintReconciler{
		Client:       mgr.GetClient(),
		Log:          core.Log.WithName("controllers").WithName("NodeTaint"),
		CniApp:       cniApp,
		CniNamespace: cniNamespace,
	}

	return reconciler.SetupWithManager(mgr)
}

func addNamespaceReconciler(mgr kube_ctrl.Manager, rt core_runtime.Runtime) error {
	reconciler := &k8s_controllers.NamespaceReconciler{
		Client:     mgr.GetClient(),
		Log:        core.Log.WithName("controllers").WithName("Namespace"),
		CNIEnabled: rt.Config().Runtime.Kubernetes.Injector.CNIEnabled,
	}
	return reconciler.SetupWithManager(mgr)
}

func addServiceReconciler(mgr kube_ctrl.Manager) error {
	reconciler := &k8s_controllers.ServiceReconciler{
		Client: mgr.GetClient(),
		Log:    core.Log.WithName("controllers").WithName("Service"),
	}
	return reconciler.SetupWithManager(mgr)
}

func addMeshServiceReconciler(mgr kube_ctrl.Manager, converter k8s_common.Converter) error {
	reconciler := &k8s_controllers.MeshServiceReconciler{
		Client:            mgr.GetClient(),
		Log:               core.Log.WithName("controllers").WithName("MeshService"),
		Scheme:            mgr.GetScheme(),
		EventRecorder:     mgr.GetEventRecorderFor("k8s.kuma.io/mesh-service-generator"),
		ResourceConverter: converter,
	}
	return reconciler.SetupWithManager(mgr)
}

func addMeshReconciler(mgr kube_ctrl.Manager, rt core_runtime.Runtime) error {
	if rt.Config().IsFederatedZoneCP() {
		return nil
	}
	defaultsReconciller := &k8s_controllers.MeshReconciler{
		ResourceManager:            rt.ResourceManager(),
		Log:                        core.Log.WithName("controllers").WithName("mesh-defaults"),
		Extensions:                 rt.Extensions(),
		CreateMeshRoutingResources: rt.Config().Defaults.CreateMeshRoutingResources,
		K8sStore:                   rt.Config().Store.Type == store.KubernetesStore,
		SystemNamespace:            rt.Config().Store.Kubernetes.SystemNamespace,
		CaManagers:                 rt.CaManagers(),
	}
	if err := defaultsReconciller.SetupWithManager(mgr); err != nil {
		return errors.Wrap(err, "could not setup mesh defaults reconciller")
	}
	return nil
}

func addPodReconciler(mgr kube_ctrl.Manager, rt core_runtime.Runtime, converter k8s_common.Converter) error {
	reconciler := &k8s_controllers.PodReconciler{
		Client:        mgr.GetClient(),
		EventRecorder: mgr.GetEventRecorderFor("k8s.kuma.io/dataplane-generator"),
		Scheme:        mgr.GetScheme(),
		Log:           core.Log.WithName("controllers").WithName("Pod"),
		PodConverter: k8s_controllers.PodConverter{
			ServiceGetter: mgr.GetClient(),
			NodeGetter:    mgr.GetClient(),
			InboundConverter: k8s_controllers.InboundConverter{
				NameExtractor: k8s_controllers.NameExtractor{
					ReplicaSetGetter: mgr.GetClient(),
					JobGetter:        mgr.GetClient(),
				},
				NodeGetter:       mgr.GetClient(),
				NodeLabelsToCopy: rt.Config().Runtime.Kubernetes.Injector.NodeLabelsToCopy,
			},
			Zone:                rt.Config().Multizone.Zone.Name,
			ResourceConverter:   converter,
			KubeOutboundsAsVIPs: rt.Config().Experimental.KubeOutboundsAsVIPs,
		},
		ResourceConverter:            converter,
		SystemNamespace:              rt.Config().Store.Kubernetes.SystemNamespace,
		IgnoredServiceSelectorLabels: rt.Config().Runtime.Kubernetes.Injector.IgnoredServiceSelectorLabels,
	}
	return reconciler.SetupWithManager(mgr, rt.Config().Runtime.Kubernetes.ControllersConcurrency.PodController)
}

func addPodStatusReconciler(mgr kube_ctrl.Manager, rt core_runtime.Runtime, converter k8s_common.Converter) error {
	reconciler := &k8s_controllers.PodStatusReconciler{
		Client:            mgr.GetClient(),
		EventRecorder:     mgr.GetEventRecorderFor("k8s.kuma.io/dataplane-jobs-syncer"),
		Scheme:            mgr.GetScheme(),
		Log:               core.Log.WithName("controllers").WithName("Pod"),
		ResourceConverter: converter,
		EnvoyAdminClient:  rt.EnvoyAdminClient(),
	}
	return reconciler.SetupWithManager(mgr)
}

func addDNS(mgr kube_ctrl.Manager, rt core_runtime.Runtime, converter k8s_common.Converter) error {
	if rt.Config().Mode == config_core.Global {
		return nil
	}
	zone := ""
	if rt.Config().Multizone != nil && rt.Config().Multizone.Zone != nil {
		zone = rt.Config().Multizone.Zone.Name
	}
	vipsAllocator, err := dns.NewVIPsAllocator(
		rt.ResourceManager(),
		rt.ConfigManager(),
		*rt.Config().DNSServer,
		rt.Config().Experimental,
		zone,
		rt.Metrics(),
	)
	if err != nil {
		return err
	}
	reconciler := &k8s_controllers.ConfigMapReconciler{
		Client:              mgr.GetClient(),
		EventRecorder:       mgr.GetEventRecorderFor("k8s.kuma.io/vips-generator"),
		Scheme:              mgr.GetScheme(),
		Log:                 core.Log.WithName("controllers").WithName("ConfigMap"),
		ResourceManager:     rt.ResourceManager(),
		VIPsAllocator:       vipsAllocator,
		SystemNamespace:     rt.Config().Store.Kubernetes.SystemNamespace,
		ResourceConverter:   converter,
		KubeOutboundsAsVIPs: rt.Config().Experimental.KubeOutboundsAsVIPs,
	}
	if err := reconciler.SetupWithManager(mgr); err != nil {
		return err
	}
	return nil
}

func addValidators(mgr kube_ctrl.Manager, rt core_runtime.Runtime, converter k8s_common.Converter) error {
	composite, ok := k8s_extensions.FromCompositeValidatorContext(rt.Extensions())
	if !ok {
		return errors.Errorf("could not find composite validator in the extensions context")
	}

	resourceAdmissionChecker := k8s_webhooks.ResourceAdmissionChecker{
		AllowedUsers:                 append(rt.Config().Runtime.Kubernetes.AllowedUsers, rt.Config().Runtime.Kubernetes.ServiceAccountName, "system:serviceaccount:kube-system:generic-garbage-collector"),
		Mode:                         rt.Config().Mode,
		FederatedZone:                rt.Config().IsFederatedZoneCP(),
		DisableOriginLabelValidation: rt.Config().Multizone.Zone.DisableOriginLabelValidation,
		SystemNamespace:              rt.Config().Store.Kubernetes.SystemNamespace,
		ZoneName:                     rt.Config().Multizone.Zone.Name,
	}
	handler := k8s_webhooks.NewValidatingWebhook(converter, core_registry.Global(), k8s_registry.Global(), resourceAdmissionChecker)
	composite.AddValidator(handler)

	k8sMeshValidator := k8s_webhooks.NewMeshValidatorWebhook(rt.ResourceValidators().Mesh, converter, rt.Config().Store.UnsafeDelete)
	composite.AddValidator(k8sMeshValidator)

	k8sDataplaneValidator := k8s_webhooks.NewDataplaneValidatorWebhook(rt.ResourceValidators().Dataplane, converter, rt.ResourceManager())
	composite.AddValidator(k8sDataplaneValidator)

	rateLimitValidator := ratelimit.RateLimitValidator{
		Store: rt.ResourceStore(),
	}
	k8sRateLimitValidator := k8s_webhooks.NewRateLimitValidatorWebhook(rateLimitValidator, converter)
	composite.AddValidator(k8sRateLimitValidator)

	externalServiceValidator := externalservice.ExternalServiceValidator{
		Store: rt.ResourceStore(),
	}
	k8sExternalServiceValidator := k8s_webhooks.NewExternalServiceValidatorWebhook(externalServiceValidator, converter)
	composite.AddValidator(k8sExternalServiceValidator)

	coreZoneValidator := zone.Validator{Store: rt.ResourceStore()}
	k8sZoneValidator := k8s_webhooks.NewZoneValidatorWebhook(coreZoneValidator, rt.Config().Store.UnsafeDelete)
	composite.AddValidator(k8sZoneValidator)

	for _, validator := range gatewayValidators(rt, converter) {
		composite.AddValidator(validator)
	}

	composite.AddValidator(&k8s_webhooks.ContainerPatchValidator{
		SystemNamespace: rt.Config().Store.Kubernetes.SystemNamespace,
	})

	mgr.GetWebhookServer().Register("/validate-kuma-io-v1alpha1", composite.IntoWebhook(mgr.GetScheme()))
	mgr.GetWebhookServer().Register("/validate-v1-service", &kube_webhook.Admission{Handler: &k8s_webhooks.ServiceValidator{Decoder: kube_admission.NewDecoder(mgr.GetScheme())}})

	client, ok := k8s_extensions.FromSecretClientContext(rt.Extensions())
	if !ok {
		return errors.Errorf("secret client hasn't been configured")
	}
	secretValidator := &k8s_webhooks.SecretValidator{
		Decoder:      kube_admission.NewDecoder(mgr.GetScheme()),
		Client:       client,
		Validator:    manager.NewSecretValidator(rt.CaManagers(), rt.ResourceStore()),
		UnsafeDelete: rt.Config().Store.UnsafeDelete,
		CpMode:       rt.Config().Mode,
	}
	mgr.GetWebhookServer().Register("/validate-v1-secret", &kube_webhook.Admission{Handler: secretValidator})

	return nil
}

func addMutators(mgr kube_ctrl.Manager, rt core_runtime.Runtime, converter k8s_common.Converter) error {
	if rt.Config().Mode != config_core.Global {
		address := fmt.Sprintf("https://%s.%s:%d", rt.Config().Runtime.Kubernetes.ControlPlaneServiceName, rt.Config().Store.Kubernetes.SystemNamespace, rt.Config().DpServer.Port)
		kubeConfig := mgr.GetConfig()
		discClient, err := discovery.NewDiscoveryClientForConfig(kubeConfig)
		if err != nil {
			return err
		}
		k8sVersion, err := discClient.ServerVersion()
		if err != nil {
			return err
		}
		var sidecarContainersEnabled bool
		if v, err := semver.NewVersion(
			fmt.Sprintf("%s.%s.0", k8sVersion.Major, k8sVersion.Minor),
		); err == nil && !v.LessThan(sidecarContainerVersion) {
			sidecarContainersEnabled = rt.Config().Experimental.SidecarContainers
		} else if rt.Config().Experimental.SidecarContainers {
			log.Info("WARNING: sidecarContainers feature is enabled but Kubernetes server does not support it")
		}
		kumaInjector, err := injector.New(
			rt.Config().Runtime.Kubernetes.Injector,
			address,
			mgr.GetClient(),
			sidecarContainersEnabled,
			converter,
			rt.Config().GetEnvoyAdminPort(),
			rt.Config().Store.Kubernetes.SystemNamespace,
		)
		if err != nil {
			return err
		}
		mgr.GetWebhookServer().Register("/inject-sidecar", k8s_webhooks.PodMutatingWebhook(kumaInjector.InjectKuma))
	}

	ownerRefMutator := &k8s_webhooks.OwnerReferenceMutator{
		Client:                 mgr.GetClient(),
		CoreRegistry:           core_registry.Global(),
		K8sRegistry:            k8s_registry.Global(),
		Scheme:                 mgr.GetScheme(),
		Decoder:                kube_admission.NewDecoder(mgr.GetScheme()),
		SkipMeshOwnerReference: rt.Config().Runtime.Kubernetes.SkipMeshOwnerReference,
		CpMode:                 rt.Config().Mode,
	}
	mgr.GetWebhookServer().Register("/owner-reference-kuma-io-v1alpha1", &kube_webhook.Admission{Handler: ownerRefMutator})

	resourceAdmissionChecker := k8s_webhooks.ResourceAdmissionChecker{
		AllowedUsers:                 append(rt.Config().Runtime.Kubernetes.AllowedUsers, rt.Config().Runtime.Kubernetes.ServiceAccountName, "system:serviceaccount:kube-system:generic-garbage-collector"),
		Mode:                         rt.Config().Mode,
		FederatedZone:                rt.Config().IsFederatedZoneCP(),
		DisableOriginLabelValidation: rt.Config().Multizone.Zone.DisableOriginLabelValidation,
		SystemNamespace:              rt.Config().Store.Kubernetes.SystemNamespace,
		ZoneName:                     rt.Config().Multizone.Zone.Name,
	}
	defaultMutator := k8s_webhooks.DefaultingWebhookFor(mgr.GetScheme(), converter, resourceAdmissionChecker)
	mgr.GetWebhookServer().Register("/default-kuma-io-v1alpha1-mesh", defaultMutator)
	return nil
}
