package gatewayapi

import (
	"context"

	"github.com/go-logr/logr"
	"github.com/pkg/errors"
	kube_core "k8s.io/api/core/v1"
	kube_apierrs "k8s.io/apimachinery/pkg/api/errors"
	kube_meta "k8s.io/apimachinery/pkg/apis/meta/v1"
	kube_runtime "k8s.io/apimachinery/pkg/runtime"
	kube_types "k8s.io/apimachinery/pkg/types"
	kube_ctrl "sigs.k8s.io/controller-runtime"
	kube_client "sigs.k8s.io/controller-runtime/pkg/client"
	kube_handler "sigs.k8s.io/controller-runtime/pkg/handler"
	kube_reconcile "sigs.k8s.io/controller-runtime/pkg/reconcile"
	kube_source "sigs.k8s.io/controller-runtime/pkg/source"
	gatewayapi "sigs.k8s.io/gateway-api/apis/v1alpha2"

	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
	"github.com/kumahq/kuma/pkg/core/resources/manager"
	core_model "github.com/kumahq/kuma/pkg/core/resources/model"
	k8s_registry "github.com/kumahq/kuma/pkg/plugins/resources/k8s/native/pkg/registry"
	"github.com/kumahq/kuma/pkg/plugins/runtime/k8s/controllers/gatewayapi/common"
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

// Reconcile handles transforming a gateway-api HTTPRoute into a Kuma
// GatewayRoute and managing the status of the gateway-api objects.
func (r *HTTPRouteReconciler) Reconcile(ctx context.Context, req kube_ctrl.Request) (kube_ctrl.Result, error) {
	httpRoute := &gatewayapi.HTTPRoute{}
	if err := r.Get(ctx, req.NamespacedName, httpRoute); err != nil {
		if kube_apierrs.IsNotFound(err) {
			// We don't know the mesh, but we don't need it to delete our
			// object.
			err := common.ReconcileLabelledObject(ctx, r.TypeRegistry, r.Client, req.NamespacedName, core_model.NoMesh, &mesh_proto.GatewayRoute{}, nil)
			return kube_ctrl.Result{}, errors.Wrap(err, "could not delete owned GatewayRoute.kuma.io")
		}

		return kube_ctrl.Result{}, err
	}

	ns := kube_core.Namespace{}
	if err := r.Client.Get(ctx, kube_types.NamespacedName{Name: httpRoute.Namespace}, &ns); err != nil {
		return kube_ctrl.Result{}, errors.Wrap(err, "unable to get Namespace of HTTPRoute")
	}

	mesh := k8s_util.MeshOf(httpRoute, &ns)

	spec, conditions, err := r.gapiToKumaRoutes(ctx, mesh, httpRoute)
	if err != nil {
		return kube_ctrl.Result{}, errors.Wrap(err, "error generating GatewayRoute.kuma.io")
	}

	if err := common.ReconcileLabelledObject(ctx, r.TypeRegistry, r.Client, req.NamespacedName, mesh, &mesh_proto.GatewayRoute{}, spec); err != nil {
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

		match := common.ServiceTagForGateway(kube_types.NamespacedName{Namespace: refNamespace, Name: string(ref.Name)})

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

	class, err := common.GetGatewayClass(ctx, r.Client, gateway.Spec.GatewayClassName)
	if err != nil {
		return false, err
	}

	return class.Spec.ControllerName == common.ControllerName, nil
}

// routesForGateway returns a function that calculates which routes might
// be affected by changes in a Gateway so they can be reconciled.
func routesForGateway(l logr.Logger, client kube_client.Client) kube_handler.MapFunc {
	l = l.WithName("routesForGateway")

	return func(obj kube_client.Object) []kube_reconcile.Request {
		gateway, ok := obj.(*gatewayapi.Gateway)
		if !ok {
			l.Error(nil, "unexpected error converting to be mapped %T object to Gateway", obj)
			return nil
		}

		var routes gatewayapi.HTTPRouteList
		if err := client.List(context.Background(), &routes); err != nil {
			l.Error(err, "unexpected error listing HTTPRoutes in cluster")
			return nil
		}

		var requests []kube_reconcile.Request
		for _, route := range routes.Items {
			for _, parentRef := range route.Spec.ParentRefs {
				if parentRefMatchesGateway(route.Namespace, parentRef, gateway) {
					requests = append(requests, kube_reconcile.Request{
						NamespacedName: kube_client.ObjectKeyFromObject(&route),
					})
				}
			}
		}

		return requests
	}
}

// parentRefMatchesGateway checks whether a ref could potentially attach to at least one
// listener of Gateway.
func parentRefMatchesGateway(routeNamespace string, parentRef gatewayapi.ParentRef, gateway *gatewayapi.Gateway) bool {
	referencedNamespace := routeNamespace
	if parentRef.Namespace != nil {
		referencedNamespace = string(*parentRef.Namespace)
	}

	// We're looking at all HTTPRoutes, at some point one may
	// reference a non-Gateway object.
	// We don't care whether a specific listener is referenced
	return *parentRef.Group == gatewayapi.GroupName &&
		*parentRef.Kind == "Gateway" &&
		referencedNamespace == gateway.Namespace &&
		string(parentRef.Name) == gateway.Name
}

func (r *HTTPRouteReconciler) SetupWithManager(mgr kube_ctrl.Manager) error {
	return kube_ctrl.NewControllerManagedBy(mgr).
		For(&gatewayapi.HTTPRoute{}).
		Watches(
			&kube_source.Kind{Type: &gatewayapi.Gateway{}},
			kube_handler.EnqueueRequestsFromMapFunc(routesForGateway(r.Log, r.Client)),
		).
		Complete(r)
}
