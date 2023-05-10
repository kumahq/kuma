package gatewayapi

import (
	"context"
	"fmt"

	"github.com/go-logr/logr"
	"github.com/pkg/errors"
	"golang.org/x/exp/slices"
	kube_core "k8s.io/api/core/v1"
	kube_apierrs "k8s.io/apimachinery/pkg/api/errors"
	kube_apimeta "k8s.io/apimachinery/pkg/api/meta"
	kube_meta "k8s.io/apimachinery/pkg/apis/meta/v1"
	kube_runtime "k8s.io/apimachinery/pkg/runtime"
	kube_types "k8s.io/apimachinery/pkg/types"
	kube_ctrl "sigs.k8s.io/controller-runtime"
	kube_client "sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	kube_handler "sigs.k8s.io/controller-runtime/pkg/handler"
	kube_reconcile "sigs.k8s.io/controller-runtime/pkg/reconcile"
	kube_source "sigs.k8s.io/controller-runtime/pkg/source"
	gatewayapi "sigs.k8s.io/gateway-api/apis/v1beta1"

	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
	"github.com/kumahq/kuma/pkg/core/resources/manager"
	core_model "github.com/kumahq/kuma/pkg/core/resources/model"
	meshhttproute_api "github.com/kumahq/kuma/pkg/plugins/policies/meshhttproute/api/v1alpha1"
	meshhttproute_k8s "github.com/kumahq/kuma/pkg/plugins/policies/meshhttproute/k8s/v1alpha1"
	k8s_registry "github.com/kumahq/kuma/pkg/plugins/resources/k8s/native/pkg/registry"
	"github.com/kumahq/kuma/pkg/plugins/runtime/k8s/controllers/gatewayapi/attachment"
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

// Reconcile handles transforming a gateway-api HTTPRoute into a Kuma
// GatewayRoute and managing the status of the gateway-api objects.
func (r *HTTPRouteReconciler) Reconcile(ctx context.Context, req kube_ctrl.Request) (kube_ctrl.Result, error) {
	r.Log.V(1).Info("reconcile", "req", req)
	httpRoute := &gatewayapi.HTTPRoute{}
	if err := r.Get(ctx, req.NamespacedName, httpRoute); err != nil {
		if kube_apierrs.IsNotFound(err) {
			// We don't know the mesh, but we don't need it to delete our
			// object.
			err := common.ReconcileLabelledObject(ctx, r.Log, r.TypeRegistry, r.Client, req.NamespacedName, core_model.NoMesh, &mesh_proto.MeshGatewayRoute{}, nil)
			return kube_ctrl.Result{}, errors.Wrap(err, "could not delete owned GatewayRoute.kuma.io")
		}

		return kube_ctrl.Result{}, err
	}

	ns := kube_core.Namespace{}
	if err := r.Client.Get(ctx, kube_types.NamespacedName{Name: httpRoute.Namespace}, &ns); err != nil {
		return kube_ctrl.Result{}, errors.Wrap(err, "unable to get Namespace of HTTPRoute")
	}

	mesh := k8s_util.MeshOfByAnnotation(httpRoute, &ns)

	gatewayRouteSpec, meshRouteSpecs, conditions, err := r.gapiToKumaRoutes(ctx, mesh, httpRoute)
	if err != nil {
		return kube_ctrl.Result{}, errors.Wrap(err, "error generating GatewayRoute.kuma.io")
	}

	if gatewayRouteSpec != nil {
		if err := common.ReconcileLabelledObject(ctx, r.Log, r.TypeRegistry, r.Client, req.NamespacedName, mesh, &mesh_proto.MeshGatewayRoute{}, gatewayRouteSpec); err != nil {
			return kube_ctrl.Result{}, errors.Wrap(err, "could not reconcile owned GatewayRoute.kuma.io")
		}
	}

	for subName, meshRouteSpec := range meshRouteSpecs {
		spec := meshRouteSpec
		route := meshhttproute_k8s.MeshHTTPRoute{
			ObjectMeta: kube_meta.ObjectMeta{
				Name: fmt.Sprintf("%s-%s", httpRoute.Name, subName), Namespace: r.SystemNamespace,
			},
		}
		if _, err := controllerutil.CreateOrUpdate(ctx, r.Client, &route, func() error {
			route.Spec = &spec
			return nil
		}); err != nil {
			return kube_ctrl.Result{}, err
		}
	}

	if err := r.updateStatus(ctx, httpRoute, conditions); err != nil {
		return kube_ctrl.Result{}, errors.Wrap(err, "unable to update HTTPRoute status")
	}

	return kube_ctrl.Result{}, nil
}

type ParentConditions map[gatewayapi.ParentReference][]kube_meta.Condition

// gapiToKumaRoutes returns some number of GatewayRoutes that should be created
// for this HTTPRoute along with any statuses to be set on the HTTPRoute.
// Only unexpected errors are returned as error.
func (r *HTTPRouteReconciler) gapiToKumaRoutes(
	ctx context.Context,
	mesh string,
	route *gatewayapi.HTTPRoute,
) (
	*mesh_proto.MeshGatewayRoute,
	map[string]meshhttproute_api.MeshHTTPRoute,
	ParentConditions,
	error,
) {
	routeNs := kube_core.Namespace{}
	if err := r.Client.Get(ctx, kube_types.NamespacedName{Name: route.Namespace}, &routeNs); err != nil {
		if kube_apierrs.IsNotFound(err) {
			return nil, nil, nil, nil
		} else {
			return nil, nil, nil, err
		}
	}

	routeConf, routeConditions, err := r.gapiToKumaRouteConf(ctx, mesh, route)
	if err != nil {
		return nil, nil, nil, err
	}

	var services []ServiceAndPorts
	var selectors []*mesh_proto.Selector

	// The conditions we accumulate for each ParentRef
	conditions := ParentConditions{}

	// Convert GAPI parent refs into selectors
	for i, ref := range route.Spec.ParentRefs {
		refAttachment, err := attachment.EvaluateParentRefAttachment(ctx, r.Client, route.Spec.Hostnames, &routeNs, ref)
		if err != nil {
			return nil, nil, nil, errors.Wrapf(err, "unable to check parent ref %d", i)
		}

		refConditions := slices.Clone(routeConditions)

		switch refAttachment {
		case attachment.Unknown:
			// We don't care about this ref
			continue
		case attachment.Allowed:
			switch {
			case *ref.Kind == "Gateway" && *ref.Group == gatewayapi.GroupName:
				selectors = append(
					selectors,
					&mesh_proto.Selector{
						Match: tagsForRef(route, ref),
					},
				)
			case *ref.Kind == "Service" && (*ref.Group == kube_core.GroupName || *ref.Group == gatewayapi.GroupName):
				namespace := route.Namespace
				if ref.Namespace != nil {
					namespace = string(*ref.Namespace)
				}
				namespacedName := kube_types.NamespacedName{Name: string(ref.Name), Namespace: namespace}
				var svc kube_core.Service
				if err := r.Client.Get(ctx, namespacedName, &svc); err != nil {
					if !kube_apierrs.IsNotFound(err) {
						return nil, nil, nil, err
					}
					continue // TODO what does the spec say? does NoMatchingParent apply?
				}
				services = append(services, serviceAndPorts(&svc, ref.Port))
			}
		default:
			var reason string
			switch refAttachment {
			case attachment.NotPermitted:
				reason = string(gatewayapi.RouteReasonNotAllowedByListeners)
			case attachment.NoHostnameIntersection:
				reason = string(gatewayapi.RouteReasonNoMatchingListenerHostname)
			case attachment.NoMatchingParent:
				reason = string(gatewayapi.RouteReasonNoMatchingParent)
			}

			if !kube_apimeta.IsStatusConditionFalse(refConditions, string(gatewayapi.RouteConditionAccepted)) {
				kube_apimeta.SetStatusCondition(&refConditions, kube_meta.Condition{
					Type:   string(gatewayapi.RouteConditionAccepted),
					Status: kube_meta.ConditionFalse,
					Reason: reason,
				})
			}
		}

		conditions[ref] = refConditions
	}

	meshRoutes, err := r.gapiToMeshRouteSpecs(ctx, mesh, route, services)
	if err != nil {
		return nil, nil, nil, err
	}

	var kumaRoute *mesh_proto.MeshGatewayRoute

	if routeConf != nil && len(selectors) > 0 {
		// We can only build MeshGatewayRoute if any attachment has matched, and we've got selectors
		kumaRoute = &mesh_proto.MeshGatewayRoute{
			Conf:      routeConf,
			Selectors: selectors,
		}
	}

	return kumaRoute, meshRoutes, conditions, nil
}

func tagsForRef(referrer kube_client.Object, ref gatewayapi.ParentReference) map[string]string {
	refNamespace := referrer.GetNamespace()
	if ns := ref.Namespace; ns != nil {
		refNamespace = string(*ns)
	}

	match := common.ServiceTagForGateway(kube_types.NamespacedName{Namespace: refNamespace, Name: string(ref.Name)})

	if ref.SectionName != nil {
		match[mesh_proto.ListenerTag] = string(*ref.SectionName)
	}

	return match
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
		for i := range routes.Items {
			route := routes.Items[i]
			for _, parentRef := range route.Spec.ParentRefs {
				if common.ParentRefMatchesGateway(route.Namespace, parentRef, gateway) {
					requests = append(requests, kube_reconcile.Request{
						NamespacedName: kube_client.ObjectKeyFromObject(&route),
					})
				}
			}
		}

		return requests
	}
}

// routesForGrant returns a function that calculates which HTTPRoutes might
// be affected by changes in a ReferenceGrant so they can be reconciled.
func routesForGrant(l logr.Logger, client kube_client.Client) kube_handler.MapFunc {
	l = l.WithName("routesForGrant")

	return func(obj kube_client.Object) []kube_reconcile.Request {
		grant, ok := obj.(*gatewayapi.ReferenceGrant)
		if !ok {
			l.Error(nil, "unexpected error converting to be mapped %T object to GatewayGrant", obj)
			return nil
		}

		var namespaces []gatewayapi.Namespace
		for _, from := range grant.Spec.From {
			if from.Group == gatewayapi.Group(gatewayapi.GroupVersion.Group) && from.Kind == common.HTTPRouteKind {
				namespaces = append(namespaces, from.Namespace)
			}
		}

		var requests []kube_reconcile.Request

		for _, namespace := range namespaces {
			routes := &gatewayapi.HTTPRouteList{}
			if err := client.List(context.Background(), routes, kube_client.InNamespace(namespace)); err != nil {
				l.Error(err, "unexpected error listing HTTPRoutes")
				return nil
			}

			for i := range routes.Items {
				requests = append(requests, kube_reconcile.Request{
					NamespacedName: kube_client.ObjectKeyFromObject(&routes.Items[i]),
				})
			}
		}

		return requests
	}
}

func (r *HTTPRouteReconciler) SetupWithManager(mgr kube_ctrl.Manager) error {
	return kube_ctrl.NewControllerManagedBy(mgr).
		For(&gatewayapi.HTTPRoute{}).
		Watches(
			&kube_source.Kind{Type: &gatewayapi.Gateway{}},
			kube_handler.EnqueueRequestsFromMapFunc(routesForGateway(r.Log, r.Client)),
		).
		Watches(
			&kube_source.Kind{Type: &gatewayapi.ReferenceGrant{}},
			kube_handler.EnqueueRequestsFromMapFunc(routesForGrant(r.Log, r.Client)),
		).
		Complete(r)
}
