package gatewayapi

import (
	"context"
	"reflect"

	"github.com/go-logr/logr"
	"github.com/pkg/errors"
	kube_apierrs "k8s.io/apimachinery/pkg/api/errors"
	kube_runtime "k8s.io/apimachinery/pkg/runtime"
	kube_types "k8s.io/apimachinery/pkg/types"
	kube_ctrl "sigs.k8s.io/controller-runtime"
	kube_client "sigs.k8s.io/controller-runtime/pkg/client"
	gatewayapi "sigs.k8s.io/gateway-api/apis/v1alpha2"

	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
	core_mesh "github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
	"github.com/kumahq/kuma/pkg/core/resources/manager"
	"github.com/kumahq/kuma/pkg/core/resources/model"
	k8s_common "github.com/kumahq/kuma/pkg/plugins/common/k8s"
	k8s_util "github.com/kumahq/kuma/pkg/plugins/runtime/k8s/util"
	util_k8s "github.com/kumahq/kuma/pkg/util/k8s"
)

// HTTPRouteReconciler reconciles a GatewayAPI object into Kuma-native objects
type HTTPRouteReconciler struct {
	kube_client.Client
	Log logr.Logger

	Scheme          *kube_runtime.Scheme
	Converter       k8s_common.Converter
	SystemNamespace string
	ResourceManager manager.ResourceManager
}

const (
	ObjectTypeUnknownOrInvalid = "ObjectTypeUnknownOrInvalid"
	ObjectNotFound             = "ObjectNotFound"
	RefInvalid                 = "RefInvalid"
)

// Reconcile handles transforming a gateway-api HTTPRoute into a Kuma
// GatewayRoute and managing the status of the gateway-api objects.
func (r *HTTPRouteReconciler) Reconcile(ctx context.Context, req kube_ctrl.Request) (kube_ctrl.Result, error) {
	httpRoute := &gatewayapi.HTTPRoute{}
	if err := r.Get(ctx, req.NamespacedName, httpRoute); err != nil {
		if kube_apierrs.IsNotFound(err) {
			return kube_ctrl.Result{}, nil
		}

		return kube_ctrl.Result{}, err
	}

	coreName := util_k8s.K8sNamespacedNameToCoreName(httpRoute.Name, httpRoute.Namespace)
	mesh := k8s_util.MeshFor(httpRoute)

	resource := core_mesh.NewGatewayRouteResource()

	specs, status, err := r.gapiToKumaRoutes(ctx, mesh, httpRoute)
	if err != nil {
		return kube_ctrl.Result{}, errors.Wrap(err, "error generating GatewayRoute")
	}

	for _, spec := range specs {
		if err := manager.Upsert(r.ResourceManager, model.ResourceKey{Mesh: mesh, Name: coreName}, resource, func(resource model.Resource) error {
			return resource.SetSpec(spec)
		}); err != nil {
			return kube_ctrl.Result{}, errors.Wrap(err, "could not upsert GatewayRoute")
		}
	}

	httpRoute.Status = r.mergeStatus(ctx, httpRoute, status)

	if err := r.Client.Status().Update(ctx, httpRoute); err != nil {
		return kube_ctrl.Result{}, errors.Wrap(err, "unable to update HTTPRoute status")
	}

	return kube_ctrl.Result{}, nil
}

// gapiToKumaRoutes returns some number of GatewayRoutes that should be created
// for this HTTPRoute along with any statuses to be set on the HTTPRoute.
// Only unexpected errors are returned as error.
func (r *HTTPRouteReconciler) gapiToKumaRoutes(
	ctx context.Context, mesh string, route *gatewayapi.HTTPRoute,
) ([]*mesh_proto.GatewayRoute, ParentStatuses, error) {
	refRoutes := map[gatewayapi.ParentRef]*mesh_proto.GatewayRoute{}

	// Convert GAPI parent refs into some number of GatewayRoutes with Kuma tag
	// matchers
	for i, ref := range route.Spec.ParentRefs {
		if handle, err := r.shouldHandleParentRef(ctx, route, ref); err != nil {
			return nil, nil, errors.Wrapf(err, "unable to check parent ref %d", i)
		} else if !handle {
			continue
		}

		var selectors []*mesh_proto.Selector

		refNamespace := route.Namespace
		if ns := ref.Namespace; ns != nil {
			refNamespace = string(*ns)
		}

		match := serviceTagForGateway(kube_types.NamespacedName{Namespace: refNamespace, Name: string(ref.Name)})

		if ref.SectionName != nil {
			match[mesh_proto.ListenerTag] = string(*ref.SectionName)
		}

		selectors = append(selectors, &mesh_proto.Selector{
			Match: match,
		})

		refRoutes[ref] = &mesh_proto.GatewayRoute{Selectors: selectors}
	}

	routeConf, routeParentStatus, err := r.gapiToKumaRouteConf(ctx, mesh, route)
	if err != nil {
		return nil, nil, err
	}

	var routes []*mesh_proto.GatewayRoute
	status := ParentStatuses{}

	for ref, route := range refRoutes {
		if routeConf != nil {
			route.Conf = routeConf
			routes = append(routes, route)
		}

		status[ref] = gatewayapi.RouteParentStatus{
			ParentRef:      ref,
			ControllerName: controllerName,
			Conditions:     routeParentStatus,
		}
	}

	return routes, status, nil
}

func (r *HTTPRouteReconciler) shouldHandleParentRef(
	ctx context.Context, route kube_client.Object, ref gatewayapi.ParentRef,
) (bool, error) {
	namespace := route.GetNamespace()
	if ns := ref.Namespace; ns != nil {
		namespace = string(*ns)
	}

	gateway := &gatewayapi.Gateway{}

	if *ref.Group != gatewayapi.Group(gatewayapi.GroupName) || *ref.Kind != gatewayapi.Kind(reflect.TypeOf(gateway).Elem().Name()) {
		return false, nil
	}

	if err := r.Client.Get(ctx, kube_types.NamespacedName{Namespace: namespace, Name: string(ref.Name)}, gateway); err != nil {
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

type ParentStatuses map[gatewayapi.ParentRef]gatewayapi.RouteParentStatus

func (r *HTTPRouteReconciler) mergeStatus(ctx context.Context, route *gatewayapi.HTTPRoute, statuses ParentStatuses) gatewayapi.HTTPRouteStatus {
	mergedStatuses := ParentStatuses{}

	// Keep statuses that don't belong to us
	for _, status := range route.Status.Parents {
		if status.ControllerName != controllerName {
			mergedStatuses[status.ParentRef] = status
		}
	}

	for ref, status := range statuses {
		mergedStatuses[ref] = status
	}

	routeStatus := gatewayapi.HTTPRouteStatus{}

	for _, status := range mergedStatuses {
		routeStatus.Parents = append(routeStatus.Parents, status)
	}

	return routeStatus
}

func (r *HTTPRouteReconciler) SetupWithManager(mgr kube_ctrl.Manager) error {
	return kube_ctrl.NewControllerManagedBy(mgr).
		For(&gatewayapi.HTTPRoute{}).
		Complete(r)
}
