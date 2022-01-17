package k8s

import (
	"fmt"
	"os"

	"github.com/pkg/errors"
	"k8s.io/apimachinery/pkg/runtime/schema"
	kube_ctrl "sigs.k8s.io/controller-runtime"
	gatewayapi "sigs.k8s.io/gateway-api/apis/v1alpha2"

	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
	"github.com/kumahq/kuma/pkg/core"
	core_runtime "github.com/kumahq/kuma/pkg/core/runtime"
	k8s_common "github.com/kumahq/kuma/pkg/plugins/common/k8s"
	k8s_registry "github.com/kumahq/kuma/pkg/plugins/resources/k8s/native/pkg/registry"
	"github.com/kumahq/kuma/pkg/plugins/runtime/k8s/containers"
	controllers "github.com/kumahq/kuma/pkg/plugins/runtime/k8s/controllers"
	gatewayapi_controllers "github.com/kumahq/kuma/pkg/plugins/runtime/k8s/controllers/gatewayapi"
	k8s_webhooks "github.com/kumahq/kuma/pkg/plugins/runtime/k8s/webhooks"
)

func crdsPresent(mgr kube_ctrl.Manager) bool {
	gk := schema.GroupKind{
		Group: gatewayapi.SchemeGroupVersion.Group,
		Kind:  "Gateway",
	}

	mappings, _ := mgr.GetClient().RESTMapper().RESTMappings(
		gk,
		gatewayapi.SchemeGroupVersion.Version,
	)

	return len(mappings) > 0
}

func gatewayPresent() bool {
	// If we haven't registered our type, we're not reconciling GatewayInstance
	// or gatewayapi objects.
	if _, err := k8s_registry.Global().NewObject(&mesh_proto.Gateway{}); err != nil {
		var unknownTypeError *k8s_registry.UnknownTypeError
		if errors.As(err, &unknownTypeError) {
			return false
		}
	}

	return true
}

func addGatewayReconcilers(mgr kube_ctrl.Manager, rt core_runtime.Runtime, converter k8s_common.Converter) error {
	if !gatewayPresent() {
		return nil
	}

	cpURL := fmt.Sprintf("https://%s.%s:%d", rt.Config().Runtime.Kubernetes.ControlPlaneServiceName, rt.Config().Store.Kubernetes.SystemNamespace, rt.Config().DpServer.Port)

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

	proxyFactory := containers.DataplaneProxyFactory{
		ControlPlaneURL:    cpURL,
		ControlPlaneCACert: caCert,
		ContainerConfig:    cfg.SidecarContainer.DataplaneContainer,
		BuiltinDNS:         cfg.BuiltinDNS,
	}

	gatewayInstanceReconciler := &controllers.GatewayInstanceReconciler{
		Client:          mgr.GetClient(),
		Log:             core.Log.WithName("controllers").WithName("GatewayInstance"),
		Scheme:          mgr.GetScheme(),
		Converter:       converter,
		ProxyFactory:    proxyFactory,
		ResourceManager: rt.ResourceManager(),
	}
	if err := gatewayInstanceReconciler.SetupWithManager(mgr); err != nil {
		return errors.Wrap(err, "could not setup GatewayInstance reconciler")
	}

	if !crdsPresent(mgr) {
		log.Info("Gateway API CRDs not registered")
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
	}
	if err := gatewayAPIHTTPRouteReconciler.SetupWithManager(mgr); err != nil {
		return errors.Wrap(err, "could not setup Gateway API HTTPRoute reconciler")
	}

	return nil
}

// gatewayValidators returns all the Gateway-related validators we want to
// start.
func gatewayValidators(rt core_runtime.Runtime, converter k8s_common.Converter) []k8s_common.AdmissionValidator {
	if !gatewayPresent() {
		return nil
	}

	return []k8s_common.AdmissionValidator{
		k8s_webhooks.NewGatewayInstanceValidatorWebhook(converter, rt.ResourceManager()),
	}
}
