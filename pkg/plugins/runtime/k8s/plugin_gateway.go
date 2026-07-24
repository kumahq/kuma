package k8s

import (
	"context"
	"fmt"

	"github.com/pkg/errors"
	kube_apierrs "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime/schema"
	kube_ctrl "sigs.k8s.io/controller-runtime"
	kube_client "sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	gatewayapi_v1 "sigs.k8s.io/gateway-api/apis/v1"
	gatewayapi "sigs.k8s.io/gateway-api/apis/v1beta1"

	config_core "github.com/kumahq/kuma/v3/pkg/config/core"
	"github.com/kumahq/kuma/v3/pkg/core"
	core_runtime "github.com/kumahq/kuma/v3/pkg/core/runtime"
	k8s_registry "github.com/kumahq/kuma/v3/pkg/plugins/resources/k8s/native/pkg/registry"
	gatewayapi_controllers "github.com/kumahq/kuma/v3/pkg/plugins/runtime/k8s/controllers/gatewayapi"
	"github.com/kumahq/kuma/v3/pkg/plugins/runtime/k8s/controllers/gatewayapi/common"
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

func addGatewayAPIReconcilers(mgr kube_ctrl.Manager, rt core_runtime.Runtime) error {
	if rt.Config().Mode == config_core.Global {
		return nil
	}

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
