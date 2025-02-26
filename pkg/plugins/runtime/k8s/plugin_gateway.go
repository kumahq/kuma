package k8s

import (
	"fmt"
	"os"

	"github.com/pkg/errors"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/discovery"
	kube_ctrl "sigs.k8s.io/controller-runtime"
	gatewayapi "sigs.k8s.io/gateway-api/apis/v1beta1"

	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
	config_core "github.com/kumahq/kuma/pkg/config/core"
	"github.com/kumahq/kuma/pkg/core"
	core_runtime "github.com/kumahq/kuma/pkg/core/runtime"
	k8s_common "github.com/kumahq/kuma/pkg/plugins/common/k8s"
	k8s_registry "github.com/kumahq/kuma/pkg/plugins/resources/k8s/native/pkg/registry"
	"github.com/kumahq/kuma/pkg/plugins/runtime/k8s/containers"
	"github.com/kumahq/kuma/pkg/plugins/runtime/k8s/controllers"
	gatewayapi_controllers "github.com/kumahq/kuma/pkg/plugins/runtime/k8s/controllers/gatewayapi"
	k8s_webhooks "github.com/kumahq/kuma/pkg/plugins/runtime/k8s/webhooks"
	util_maps "github.com/kumahq/kuma/pkg/util/maps"
)

var requiredGatewayCRDs = map[string]string{
	"GatewayClass":   gatewayCRDNameWithGroupKindAndVersion("GatewayClass"),
	"Gateway":        gatewayCRDNameWithGroupKindAndVersion("Gateway"),
	"HTTPRoute":      gatewayCRDNameWithGroupKindAndVersion("HTTPRoute"),
	"ReferenceGrant": gatewayCRDNameWithGroupKindAndVersion("ReferenceGrant"),
}

func gatewayCRDNameWithGroupKindAndVersion(name string) string {
	return fmt.Sprintf(
		"%s.%s/%s",
		name,
		gatewayapi.GroupVersion.Group,
		gatewayapi.GroupVersion.Version,
	)
}

func gatewayAPICRDsPresent(mgr kube_ctrl.Manager) (bool, []string) {
	var missing []string

	for kind, fullName := range requiredGatewayCRDs {
		gk := schema.GroupKind{
			Group: gatewayapi.GroupVersion.Group,
			Kind:  kind,
		}

		mappings, _ := mgr.GetClient().RESTMapper().RESTMappings(
			gk,
			gatewayapi.GroupVersion.Version,
		)

		if len(mappings) == 0 {
			missing = append(missing, fullName)
		}
	}

	return len(missing) == 0, missing
}

func meshGatewayCRDsPresent() bool {
	// If we haven't registered our type, we're not reconciling MeshGatewayInstance
	// or gatewayapi objects.
	if _, err := k8s_registry.Global().NewObject(&mesh_proto.MeshGateway{}); err != nil {
		var unknownTypeError *k8s_registry.UnknownTypeError
		if errors.As(err, &unknownTypeError) {
			return false
		}
	}

	return true
}

func addGatewayReconcilers(mgr kube_ctrl.Manager, rt core_runtime.Runtime, converter k8s_common.Converter) error {
	cpURL := fmt.Sprintf("https://%s.%s:%d", rt.Config().Runtime.Kubernetes.ControlPlaneServiceName, rt.Config().Store.Kubernetes.SystemNamespace, rt.Config().DpServer.Port)

	if rt.Config().Mode == config_core.Global {
		return nil
	}
	// TODO don't use injector config
	cfg := rt.Config().Runtime.Kubernetes.Injector

	var caCert string
	if cfg.CaCertFile != "" {
		bytes, err := os.ReadFile(cfg.CaCertFile)
		if err != nil {
			return errors.Wrapf(err, "could not read provided CA cert file %s", cfg.CaCertFile)
		}
		caCert = string(bytes)
	}

	proxyFactory := containers.NewDataplaneProxyFactory(
		cpURL, caCert, rt.Config().GetEnvoyAdminPort(), cfg.SidecarContainer.DataplaneContainer, cfg.BuiltinDNS,
		false, false, false, 0,
	)

	kubeConfig := mgr.GetConfig()

	discClient, err := discovery.NewDiscoveryClientForConfig(kubeConfig)
	if err != nil {
		return err
	}

	k8sVersion, err := discClient.ServerVersion()
	if err != nil {
		return err
	}

	gatewayInstanceReconciler := &controllers.GatewayInstanceReconciler{
		K8sVersion:      k8sVersion,
		Client:          mgr.GetClient(),
		Log:             core.Log.WithName("controllers").WithName("MeshGatewayInstance"),
		Scheme:          mgr.GetScheme(),
		Converter:       converter,
		ProxyFactory:    proxyFactory,
		ResourceManager: rt.ResourceManager(),
	}
	if err := gatewayInstanceReconciler.SetupWithManager(mgr); err != nil {
		return errors.Wrap(err, "could not setup MeshGatewayInstance reconciler")
	}

	if err := addGatewayAPIReconcilers(mgr, rt, proxyFactory); err != nil {
		return err
	}

	return nil
}

func addGatewayAPIReconcilers(mgr kube_ctrl.Manager, rt core_runtime.Runtime, proxyFactory *containers.DataplaneProxyFactory) error {
	if ok, missingGatewayCRDs := gatewayAPICRDsPresent(mgr); !ok {
		if len(requiredGatewayCRDs) != len(missingGatewayCRDs) {
			// Logging this as error as in such case there is possibility that user is expecting
			// Gateway API support to work, but might be unaware that some (not all) CRDs are
			// missing. Such scenario might occur when old version of CRDs is installed with
			// missing ReferenceGrant.
			log.Error(
				errors.New("only subset of required GatewayAPI CRDs registered"),
				"disabling support for GatewayAPI",
				"required", util_maps.AllValues(requiredGatewayCRDs),
				"missing", missingGatewayCRDs,
			)
		} else {
			log.Info("[WARNING] GatewayAPI CRDs are not registered. Disabling support")
		}

		return nil
	}

	gatewayAPIGatewayClassReconciler := &gatewayapi_controllers.GatewayClassReconciler{
		Client: mgr.GetClient(),
		Log:    core.Log.WithName("controllers").WithName("gatewayapi").WithName("GatewayClass"),
	}
	if err := gatewayAPIGatewayClassReconciler.SetupWithManager(mgr); err != nil {
		return errors.Wrap(err, "could not setup Gateway API GatewayClass reconciler")
	}

	gatewayAPIGatewayReconciler := &gatewayapi_controllers.GatewayReconciler{
		Client:          mgr.GetClient(),
		Log:             core.Log.WithName("controllers").WithName("gatewayapi").WithName("Gateway"),
		Scheme:          mgr.GetScheme(),
		TypeRegistry:    k8s_registry.Global(),
		SystemNamespace: rt.Config().Store.Kubernetes.SystemNamespace,
		ProxyFactory:    proxyFactory,
		ResourceManager: rt.ResourceManager(),
	}
	if err := gatewayAPIGatewayReconciler.SetupWithManager(mgr); err != nil {
		return errors.Wrap(err, "could not setup Gateway API Gateway reconciler")
	}

	gatewayAPIHTTPRouteReconciler := &gatewayapi_controllers.HTTPRouteReconciler{
		Client:          mgr.GetClient(),
		Log:             core.Log.WithName("controllers").WithName("gatewayapi").WithName("HTTPRoute"),
		Scheme:          mgr.GetScheme(),
		TypeRegistry:    k8s_registry.Global(),
		SystemNamespace: rt.Config().Store.Kubernetes.SystemNamespace,
		ResourceManager: rt.ResourceManager(),
		Zone:            rt.Config().Multizone.Zone.Name,
	}
	if err := gatewayAPIHTTPRouteReconciler.SetupWithManager(mgr); err != nil {
		return errors.Wrap(err, "could not setup Gateway API HTTPRoute reconciler")
	}

	secretController := &gatewayapi_controllers.SecretController{
		Log:                                  core.Log.WithName("controllers").WithName("secret"),
		Client:                               mgr.GetClient(),
		SystemNamespace:                      rt.Config().Store.Kubernetes.SystemNamespace,
		SupportGatewaySecretsInAllNamespaces: rt.Config().Runtime.Kubernetes.SupportGatewaySecretsInAllNamespaces,
	}
	if err := secretController.SetupWithManager(mgr); err != nil {
		return errors.Wrap(err, "could not setup Secret reconciler")
	}

	return nil
}

// gatewayValidators returns all the Gateway-related validators we want to
// start.
func gatewayValidators(rt core_runtime.Runtime, converter k8s_common.Converter) []k8s_common.AdmissionValidator {
	if !meshGatewayCRDsPresent() {
		return nil
	}

	return []k8s_common.AdmissionValidator{
		k8s_webhooks.NewGatewayInstanceValidatorWebhook(converter, rt.ResourceManager(), rt.Config().Mode),
	}
}
