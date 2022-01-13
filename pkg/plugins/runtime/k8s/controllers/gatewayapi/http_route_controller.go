package gatewayapi

import (
	"context"

	"github.com/go-logr/logr"
	"github.com/pkg/errors"
	kube_apierrs "k8s.io/apimachinery/pkg/api/errors"
	kube_meta "k8s.io/apimachinery/pkg/apis/meta/v1"
	kube_runtime "k8s.io/apimachinery/pkg/runtime"
	kube_types "k8s.io/apimachinery/pkg/types"
	kube_ctrl "sigs.k8s.io/controller-runtime"
	kube_client "sigs.k8s.io/controller-runtime/pkg/client"
	gatewayapi "sigs.k8s.io/gateway-api/apis/v1alpha2"

	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
	"github.com/kumahq/kuma/pkg/core/resources/manager"
	k8s_registry "github.com/kumahq/kuma/pkg/plugins/resources/k8s/native/pkg/registry"
	k8s_util "github.com/kumahq/kuma/pkg/plugins/runtime/k8s/util"
)

// HTTPRouteReconciler reconciles a GatewayAPI object into Kuma-native objects
type HTTPRouteReconciler struct {
	kube_client.Client
	Log logr.Logger

	Scheme          *kube_runtime.Scheme
	TypeRegistry    k8s_registry.TypeRegistry
	SystemNamespace string
	ResourceManager manager.ResourceManager
}

const (
	ObjectTypeUnknownOrInvalid = "ObjectTypeUnknownOrInvalid"
	ObjectNotFound             = "ObjectNotFound"
	RefInvalid                 = "RefInvalid"
	RefNotPermitted            = "RefNotPermitted"
)

const ownerLabel = "gateways.kuma.io/gateway.networking.k8s.io-owner"

// Reconcile handles transforming a gateway-api HTTPRoute into a Kuma
// GatewayRoute and managing the status of the gateway-api objects.
func (r *HTTPRouteReconciler) Reconcile(ctx context.Context, req kube_ctrl.Request) (kube_ctrl.Result, error) {
	httpRoute := &gatewayapi.HTTPRoute{}
	if err := r.Get(ctx, req.NamespacedName, httpRoute); err != nil {
		if kube_apierrs.IsNotFound(err) {
			err := reconcileLabelledObject(ctx, r.TypeRegistry, r.Client, req.NamespacedName, &mesh_proto.GatewayRoute{}, nil)
			return kube_ctrl.Result{}, errors.Wrap(err, "could not delete owned GatewayRoute.kuma.io")
		}

		return kube_ctrl.Result{}, err
	}
	mesh := k8s_util.MeshFor(httpRoute)

	spec, conditions, err := r.gapiToKumaRoutes(ctx, mesh, httpRoute)
	if err != nil {
		return kube_ctrl.Result{}, errors.Wrap(err, "error generating GatewayRoute.kuma.io")
	}

	if err := reconcileLabelledObject(ctx, r.TypeRegistry, r.Client, req.NamespacedName, &mesh_proto.GatewayRoute{}, spec); err != nil {
		return kube_ctrl.Result{}, errors.Wrap(err, "could not reconcile owned GatewayRoute.kuma.io")
	}

	if err := r.updateStatus(ctx, httpRoute, conditions); err != nil {
		return kube_ctrl.Result{}, errors.Wrap(err, "unable to update HTTPRoute status")
	}

	return kube_ctrl.Result{}, nil
}

type ParentConditions map[gatewayapi.ParentRef][]kube_meta.Condition

// gapiToKumaRoutes returns some number of GatewayRoutes that should be created
// for this HTTPRoute along with any statuses to be set on the HTTPRoute.
// Only unexpected errors are returned as error.
func (r *HTTPRouteReconciler) gapiToKumaRoutes(
	ctx context.Context,
	mesh string,
	route *gatewayapi.HTTPRoute,
) (
	*mesh_proto.GatewayRoute,
	ParentConditions,
	error,
) {
	var refRoutes []gatewayapi.ParentRef

	kumaRoute := &mesh_proto.GatewayRoute{}

	// Convert GAPI parent refs into selectors
	for i, ref := range route.Spec.ParentRefs {
		if handle, err := r.shouldHandleParentRef(ctx, route, ref); err != nil {
			return nil, nil, errors.Wrapf(err, "unable to check parent ref %d", i)
		} else if !handle {
			continue
		}
		refRoutes = append(refRoutes, ref)

		refNamespace := route.Namespace
		if ns := ref.Namespace; ns != nil {
			refNamespace = string(*ns)
		}

		match := serviceTagForGateway(kube_types.NamespacedName{Namespace: refNamespace, Name: string(ref.Name)})

		if ref.SectionName != nil {
			match[mesh_proto.ListenerTag] = string(*ref.SectionName)
		}

		kumaRoute.Selectors = append(kumaRoute.Selectors, &mesh_proto.Selector{
			Match: match,
		})
	}

	routeConf, routeConditions, err := r.gapiToKumaRouteConf(ctx, mesh, route)
	if err != nil {
		return nil, nil, err
	}

	kumaRoute.Conf = routeConf

	conditions := ParentConditions{}

	for _, ref := range refRoutes {
		conditions[ref] = routeConditions
	}

	return kumaRoute, conditions, nil
}

func (r *HTTPRouteReconciler) shouldHandleParentRef(
	ctx context.Context, route kube_client.Object, ref gatewayapi.ParentRef,
) (bool, error) {
	name := string(ref.Name)
	// Group and Kind both have default values
	group := string(*ref.Group)
	kind := string(*ref.Kind)

	namespace := route.GetNamespace()
	if ns := ref.Namespace; ns != nil {
		namespace = string(*ns)
	}

	gateway := &gatewayapi.Gateway{}
	gatewayKind := "Gateway"

	if group != gatewayapi.GroupName || kind != gatewayKind {
		return false, nil
	}

	if err := r.Client.Get(ctx, kube_types.NamespacedName{Namespace: namespace, Name: name}, gateway); err != nil {
		if kube_apierrs.IsNotFound(err) {
			return false, nil
		} else {
			return false, err
		}
	}

	class, err := getGatewayClass(ctx, r.Client, gateway.Spec.GatewayClassName)
	if err != nil {
		return false, err
	}

	return class.Spec.ControllerName == controllerName, nil
}

func (r *HTTPRouteReconciler) SetupWithManager(mgr kube_ctrl.Manager) error {
	return kube_ctrl.NewControllerManagedBy(mgr).
		For(&gatewayapi.HTTPRoute{}).
		Complete(r)
}
