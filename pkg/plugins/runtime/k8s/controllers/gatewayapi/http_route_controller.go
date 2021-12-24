package gatewayapi

import (
	"context"
	"fmt"
	"path"
	"reflect"

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
	k8s_common "github.com/kumahq/kuma/pkg/plugins/common/k8s"
	mesh_k8s "github.com/kumahq/kuma/pkg/plugins/resources/k8s/native/api/v1alpha1"
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

const ownerLabel = "gateways.kuma.io/owner"

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

	mesh := k8s_util.MeshFor(httpRoute)

	specs, status, err := r.gapiToKumaRoutes(ctx, mesh, httpRoute)
	if err != nil {
		return kube_ctrl.Result{}, errors.Wrap(err, "error generating GatewayRoute")
	}

	if err := reconcileLabelledObjectSet(ctx, r.Client, req.NamespacedName, specs); err != nil {
		return kube_ctrl.Result{}, errors.Wrap(err, "could not CreateOrUpdate GatewayRoutes")
	}

	httpRoute.Status = r.mergeStatus(ctx, httpRoute, status)

	if err := r.Client.Status().Update(ctx, httpRoute); err != nil {
		return kube_ctrl.Result{}, errors.Wrap(err, "unable to update HTTPRoute status")
	}

	return kube_ctrl.Result{}, nil
}

type ParentStatuses map[gatewayapi.ParentRef]gatewayapi.RouteParentStatus
type ParentRoutes map[gatewayapi.ParentRef]*mesh_proto.GatewayRoute

// gapiToKumaRoutes returns some number of GatewayRoutes that should be created
// for this HTTPRoute along with any statuses to be set on the HTTPRoute.
// Only unexpected errors are returned as error.
func (r *HTTPRouteReconciler) gapiToKumaRoutes(
	ctx context.Context,
	mesh string,
	route *gatewayapi.HTTPRoute,
) (
	ParentRoutes,
	ParentStatuses,
	error,
) {
	refRoutes := map[gatewayapi.ParentRef]*mesh_proto.GatewayRoute{}

	// Convert GAPI parent refs into some number of GatewayRoutes with Kuma tag
	// matchers
	for i, ref := range route.Spec.ParentRefs {
		// TODO support listeners here
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

	status := ParentStatuses{}

	for ref, route := range refRoutes {
		if routeConf != nil {
			route.Conf = routeConf
		}

		status[ref] = gatewayapi.RouteParentStatus{
			ParentRef:      ref,
			ControllerName: controllerName,
			Conditions:     routeParentStatus,
		}
	}

	return refRoutes, status, nil
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

// reconcileLabelledObjectSet manages a set of owned objects based on labeling
// with the owner key and indexed by the service tag and listener tag.
func reconcileLabelledObjectSet(
	ctx context.Context,
	client kube_client.Client,
	owner kube_types.NamespacedName,
	routeSpecs ParentRoutes,
) error {
	// First we list which existing objects are owned by this route.
	// Then we iterate through the current refs and either create new ones or
	// update existing ones. Finally, we delete any existing objects for refs
	// which no longer exist.
	ownerLabelValue := util_k8s.K8sNamespacedNameToCoreName(owner.Name, owner.Namespace)
	labels := kube_client.MatchingLabels{
		ownerLabel: ownerLabelValue,
	}

	list := &mesh_k8s.GatewayRouteList{}
	if err := client.List(ctx, list, labels); err != nil {
		return err
	}

	type GatewayRef string

	routesForGateway := map[GatewayRef]*mesh_k8s.GatewayRoute{}

	for _, route := range list.Items {
		selectorMatch := route.Spec.GetSelectors()[0].GetMatch()
		// We know that GatewayRoutes owned by us contain exactly one selector,
		// with the serviceTagForGateway format
		tag := selectorMatch[mesh_proto.ServiceTag]

		gatewayName, err := gatewayForServiceTag(tag)
		if err != nil {
			// TODO log and continue
			return err
		}

		key := gatewayName.String()

		// We might also be binding to a specific listener
		if listener, ok := selectorMatch[mesh_proto.ListenerTag]; ok {
			key = path.Join(key, listener)
		}

		// GatewayRoutes are cluster-scoped
		routesForGateway[GatewayRef(key)] = &route
	}

	for gatewayRef, spec := range routeSpecs {
		// First we need to figure out the key for this reference
		gatewayNamespace := owner.Namespace
		if ns := gatewayRef.Namespace; ns != nil {
			gatewayNamespace = string(*ns)
		}

		gatewayKey := kube_types.NamespacedName{
			Namespace: gatewayNamespace,
			Name:      string(gatewayRef.Name),
		}.String()

		if listener := gatewayRef.SectionName; listener != nil {
			gatewayKey = path.Join(gatewayKey, string(*listener))
		}

		// We either are already maintaining a GatewayRoute for this Ref
		if route, ok := routesForGateway[GatewayRef(gatewayKey)]; ok {
			route.Spec = spec

			if err := client.Update(ctx, route); err != nil {
				return errors.Wrapf(err, "could not update GatewayRoute for Gateway %s", gatewayKey)
			}

			delete(routesForGateway, GatewayRef(gatewayKey))
			continue
		}

		// Or it's a new Ref
		route := &mesh_k8s.GatewayRoute{
			ObjectMeta: kube_meta.ObjectMeta{
				GenerateName: fmt.Sprintf("%s-", ownerLabelValue),
				Labels: map[string]string{
					ownerLabel: ownerLabelValue,
				},
			},
			Spec: spec,
		}
		if err := client.Create(ctx, route); err != nil {
			return errors.Wrapf(err, "could not create GatewayRoute for Gateway %s", gatewayKey)
		}
	}

	// Any objects left over we want to cleanup
	for _, toDelete := range routesForGateway {
		if err := client.Delete(ctx, toDelete); err != nil {
			return err
		}
	}

	return nil
}

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
