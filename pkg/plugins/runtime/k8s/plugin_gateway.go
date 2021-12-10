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
	"github.com/kumahq/kuma/pkg/plugins/resources/k8s/native/pkg/registry"
	"github.com/kumahq/kuma/pkg/plugins/runtime/k8s/containers"
	gatewayapi_controllers "github.com/kumahq/kuma/pkg/plugins/runtime/k8s/controllers/gatewayapi"
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

func addGatewayReconciler(mgr kube_ctrl.Manager, rt core_runtime.Runtime, converter k8s_common.Converter) error {
	// If we haven't registered our type, we're not reconciling gatewayapi
	// objects.
	if _, err := registry.Global().NewObject(&mesh_proto.Gateway{}); err != nil {
		var unknownTypeError *registry.UnknownTypeError
		if errors.As(err, &unknownTypeError) {
			return nil
		}
	}

	if !crdsPresent(mgr) {
		log.Info("Gateway API CRDs not registered")
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

	gatewayAPIGatewayReconciler := &gatewayapi_controllers.GatewayReconciler{
		Client:          mgr.GetClient(),
		Reader:          mgr.GetAPIReader(),
		Log:             core.Log.WithName("controllers").WithName("gatewayapi").WithName("Gateway"),
		Scheme:          mgr.GetScheme(),
		Converter:       converter,
		SystemNamespace: rt.Config().Store.Kubernetes.SystemNamespace,
		ProxyFactory:    proxyFactory,
		ResourceManager: rt.ResourceManager(),
	}
	if err := gatewayAPIGatewayReconciler.SetupWithManager(mgr); err != nil {
		return errors.Wrap(err, "could not setup Gateway API Gateway reconciler")
	}

	gatewayAPIHTTPRouteReconciler := &gatewayapi_controllers.HTTPRouteReconciler{
		Client:          mgr.GetClient(),
		Reader:          mgr.GetAPIReader(),
		Log:             core.Log.WithName("controllers").WithName("gatewayapi").WithName("HTTPRoute"),
		Scheme:          mgr.GetScheme(),
		Converter:       converter,
		SystemNamespace: rt.Config().Store.Kubernetes.SystemNamespace,
		ResourceManager: rt.ResourceManager(),
	}
	if err := gatewayAPIHTTPRouteReconciler.SetupWithManager(mgr); err != nil {
		return errors.Wrap(err, "could not setup Gateway API HTTPRoute reconciler")
	}

	return nil
}
