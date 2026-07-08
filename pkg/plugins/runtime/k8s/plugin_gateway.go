package k8s

import (
	"context"
	"fmt"
	"os"

	"github.com/pkg/errors"
	kube_apierrs "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/discovery"
	kube_ctrl "sigs.k8s.io/controller-runtime"
	kube_client "sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	gatewayapi_v1 "sigs.k8s.io/gateway-api/apis/v1"
	gatewayapi "sigs.k8s.io/gateway-api/apis/v1beta1"

	mesh_proto "github.com/kumahq/kuma/v3/api/mesh/v1alpha1"
	config_core "github.com/kumahq/kuma/v3/pkg/config/core"
	"github.com/kumahq/kuma/v3/pkg/core"
	core_runtime "github.com/kumahq/kuma/v3/pkg/core/runtime"
	k8s_common "github.com/kumahq/kuma/v3/pkg/plugins/common/k8s"
	k8s_registry "github.com/kumahq/kuma/v3/pkg/plugins/resources/k8s/native/pkg/registry"
	"github.com/kumahq/kuma/v3/pkg/plugins/runtime/k8s/containers"
	"github.com/kumahq/kuma/v3/pkg/plugins/runtime/k8s/controllers"
	gatewayapi_controllers "github.com/kumahq/kuma/v3/pkg/plugins/runtime/k8s/controllers/gatewayapi"
	"github.com/kumahq/kuma/v3/pkg/plugins/runtime/k8s/controllers/gatewayapi/common"
	k8s_webhooks "github.com/kumahq/kuma/v3/pkg/plugins/runtime/k8s/webhooks"
)

var requiredGatewayCRDs = map[string]string{
	"HTTPRoute": gatewayCRDNameWithGroupKindAndVersion("HTTPRoute"),
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
		if !gatewayAPICRDPresent(mgr, kind) {
			missing = append(missing, fullName)
		}
	}

	return len(missing) == 0, missing
}

func gatewayAPICRDPresent(mgr kube_ctrl.Manager, kind string) bool {
	gk := schema.GroupKind{
		Group: gatewayapi.GroupVersion.Group,
		Kind:  kind,
	}

	mappings, _ := mgr.GetClient().RESTMapper().RESTMappings(
		gk,
		gatewayapi.GroupVersion.Version,
	)

	return len(mappings) > 0
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
		cpURL, caCert, rt.Config().GetEnvoyAdminPort(), rt.Config().GetEnvoyReadinessPort(),
		cfg.SidecarContainer.DataplaneContainer, cfg.BuiltinDNS,
		false, rt.Config().BootstrapServer.Params.EnvoyAdminUnixSocket, false, false, 0,
		cfg.UnifiedResourceNamingEnabled, cfg.OtelPipeEnabled, cfg.Spire.Enabled,
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

	if err := addGatewayAPIReconcilers(mgr, rt); err != nil {
		return err
	}

	return nil
}

func addGatewayAPIReconcilers(mgr kube_ctrl.Manager, rt core_runtime.Runtime) error {
	if gatewayAPICRDPresent(mgr, "GatewayClass") {
		if err := cleanupLegacyGatewayClassFinalizers(
			context.Background(),
			mgr.GetAPIReader(),
			mgr.GetClient(),
		); err != nil {
			return errors.Wrap(err, "could not clean up legacy GatewayClass finalizers")
		}
	}

	if ok, missingGatewayCRDs := gatewayAPICRDsPresent(mgr); !ok {
		log.Info("[WARNING] GatewayAPI CRDs are not registered. Disabling support", "missing", missingGatewayCRDs)
		return nil
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

	return nil
}

// cleanupLegacyGatewayClassFinalizers runs before mgr.Start(), so the manager's
// cached client cannot serve reads yet; list via the direct API reader and
// only use the cached client for the (uncached) write path.
func cleanupLegacyGatewayClassFinalizers(ctx context.Context, reader kube_client.Reader, client kube_client.Client) error {
	classes := &gatewayapi.GatewayClassList{}
	if err := reader.List(ctx, classes); err != nil {
		if kube_apierrs.IsForbidden(err) {
			log.Info(
				"[WARNING] unable to clean up legacy GatewayClass finalizers because RBAC is missing; remove gateway.networking.k8s.io/gatewayclasses finalizers manually if GatewayClass objects are stuck deleting",
				"err", err,
			)
			return nil
		}
		return err
	}

	for i := range classes.Items {
		class := &classes.Items[i]
		if class.Spec.ControllerName != common.ControllerName {
			continue
		}
		if !controllerutil.ContainsFinalizer(class, gatewayapi_v1.GatewayClassFinalizerGatewaysExist) {
			continue
		}

		updated := class.DeepCopy()
		controllerutil.RemoveFinalizer(updated, gatewayapi_v1.GatewayClassFinalizerGatewaysExist)
		if err := client.Patch(ctx, updated, kube_client.MergeFrom(class)); err != nil {
			if kube_apierrs.IsForbidden(err) {
				log.Info(
					"[WARNING] unable to remove legacy GatewayClass finalizer because RBAC is missing; remove gateway.networking.k8s.io/gatewayclasses finalizers manually if GatewayClass objects are stuck deleting",
					"gatewayClass", class.Name,
					"err", err,
				)
				continue
			}
			return err
		}
		log.Info(
			"removed legacy GatewayClass finalizer for removed built-in Gateway API support",
			"gatewayClass", class.Name,
			"finalizer", gatewayapi_v1.GatewayClassFinalizerGatewaysExist,
		)
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
