package gatewayapi

import (
	"context"
	"fmt"
	"reflect"

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
	gatewayapi_v1 "sigs.k8s.io/gateway-api/apis/v1"
	gatewayapi "sigs.k8s.io/gateway-api/apis/v1beta1"

	mesh_proto "github.com/kumahq/kuma/v3/api/mesh/v1alpha1"
	"github.com/kumahq/kuma/v3/pkg/core/resources/manager"
	core_model "github.com/kumahq/kuma/v3/pkg/core/resources/model"
	meshhttproute_api "github.com/kumahq/kuma/v3/pkg/plugins/policies/meshhttproute/api/v1alpha1"
	k8s_registry "github.com/kumahq/kuma/v3/pkg/plugins/resources/k8s/native/pkg/registry"
	"github.com/kumahq/kuma/v3/pkg/plugins/runtime/k8s/controllers/gatewayapi/attachment"
	"github.com/kumahq/kuma/v3/pkg/plugins/runtime/k8s/controllers/gatewayapi/common"
	k8s_util "github.com/kumahq/kuma/v3/pkg/plugins/runtime/k8s/util"
)

// HTTPRouteReconciler reconciles a GatewayAPI object into Kuma-native objects
type HTTPRouteReconciler struct {
	kube_client.Client
	Log logr.Logger

	Scheme          *kube_runtime.Scheme
	TypeRegistry    k8s_registry.TypeRegistry
	SystemNamespace string
	ResourceManager manager.ResourceManager
	Zone            string
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
			if err := common.ReconcileLabelledObject(
				ctx, r.Log, r.TypeRegistry, r.Client, req.NamespacedName, core_model.NoMesh, &mesh_proto.MeshGatewayRoute{}, "", nil,
			); err != nil {
				return kube_ctrl.Result{}, errors.Wrap(err, "could not delete owned GatewayRoute.kuma.io")
			}
			if err := common.ReconcileLabelledObject(
				ctx, r.Log, r.TypeRegistry, r.Client, req.NamespacedName, core_model.NoMesh, &meshhttproute_api.MeshHTTPRoute{}, r.SystemNamespace, nil,
			); err != nil {
				return kube_ctrl.Result{}, errors.Wrap(err, "could not delete owned MeshHTTPRoute.kuma.io")
			}

			return kube_ctrl.Result{}, nil
		}

		return kube_ctrl.Result{}, err
	}

	ns := kube_core.Namespace{}
	if err := r.Get(ctx, kube_types.NamespacedName{Name: httpRoute.Namespace}, &ns); err != nil {
		return kube_ctrl.Result{}, errors.Wrap(err, "unable to get Namespace of HTTPRoute")
	}

	mesh := k8s_util.MeshOfByLabelOrAnnotation(r.Log, httpRoute, &ns)

	meshRouteSpecs, conditions, err := r.gapiToKumaRoutes(ctx, mesh, httpRoute)
	if err != nil {
		return kube_ctrl.Result{}, errors.Wrap(err, "could not generate MeshHTTPRoute.kuma.io resources")
	}

	// After upgrading Kuma to version 2.7.x, MeshGatewayRoutes are no longer used internally.
	// This code (reconcilliation with empty list of owned resources) ensures that any existing
	// MeshGatewayRoutes will be deleted. This is a safe operation because MeshHTTPRoutes have
	// replaced MeshGatewayRoutes, and this replacement doesn't introduce any changes to the xDS
	// configuration for MeshGateway. Therefore, there won't be any disruptions in traffic flow.
	if err := common.ReconcileLabelledObject(
		ctx, r.Log, r.TypeRegistry, r.Client, req.NamespacedName, mesh, &mesh_proto.MeshGatewayRoute{}, "", nil,
	); err != nil {
		return kube_ctrl.Result{}, errors.Wrap(err, "could not delete owned GatewayRoute.kuma.io")
	}

	if err := common.ReconcileLabelledObject(
		ctx, r.Log, r.TypeRegistry, r.Client, req.NamespacedName, mesh, &meshhttproute_api.MeshHTTPRoute{}, r.SystemNamespace, meshRouteSpecs,
	); err != nil {
		return kube_ctrl.Result{}, errors.Wrap(err, "could not reconcile owned MeshHTTPRoute.kuma.io")
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
) (map[string]core_model.ResourceSpec, ParentConditions, error) {
	routes := map[string]core_model.ResourceSpec{}

	// The conditions we accumulate for each ParentRef
	conditions := ParentConditions{}

	for i, ref := range route.Spec.ParentRefs {
		refAttachment, refKind, err := attachment.EvaluateParentRefAttachment(ref)
		if err != nil {
			return nil, nil, errors.Wrapf(err, "unable to check parent ref %d", i)
		}

		if refAttachment == attachment.Unknown {
			// We don't care about this ref for route generation, but keeping
			// it in the conditions map tells status merging to preserve any
			// status we previously wrote for it.
			conditions[ref] = nil
			continue
		}

		rules, rulesConditions, err := r.gapiToMeshRules(ctx, mesh, route)
		if err != nil {
			return nil, nil, err
		}

		// refAttachment is always Allowed here: Unknown was handled above, and
		// Service refs (the only remaining Kind) always attach.
		if refKind == attachment.Service {
			namespace := route.Namespace
			if ref.Namespace != nil {
				namespace = string(*ref.Namespace)
			}

			var parent kube_core.Service
			if err := r.Get(ctx, kube_types.NamespacedName{
				Name:      string(ref.Name),
				Namespace: namespace,
			}, &parent); err != nil {
				if !kube_apierrs.IsNotFound(err) {
					return nil, nil, err
				}
				continue // TODO what does the spec say? does NoMatchingParent apply?
			}

			routeSubName := fmt.Sprintf(
				"%s-%s-%s.%s",
				route.Name,
				route.Namespace,
				parent.GetName(),
				parent.GetNamespace(),
			)

			routes[routeSubName] = r.gapiServiceToMeshRoute(route.Namespace, rules, &parent, ref.Port)
		}

		conditions[ref] = rulesConditions
	}

	return routes, conditions, nil
}

// routesForService returns a function that calculates which HTTPRoutes might
// be affected by changes in a Service.
func routesForService(l logr.Logger, client kube_client.Client) kube_handler.MapFunc {
	l = l.WithName("service-to-routes-mapper")

	return func(ctx context.Context, obj kube_client.Object) []kube_reconcile.Request {
		svc, ok := obj.(*kube_core.Service)
		if !ok {
			l.Error(nil, "unexpected error converting object to Service", "typ", reflect.TypeOf(obj))
			return nil
		}

		var routes gatewayapi.HTTPRouteList
		if err := client.List(ctx, &routes, kube_client.MatchingFields{
			servicesOfRouteField: kube_client.ObjectKeyFromObject(svc).String(),
		}); err != nil {
			l.Error(err, "unexpected error listing HTTPRoutes")
			return nil
		}

		var requests []kube_reconcile.Request
		for i := range routes.Items {
			requests = append(requests, kube_reconcile.Request{
				NamespacedName: kube_client.ObjectKeyFromObject(&routes.Items[i]),
			})
		}
		return requests
	}
}

const (
	servicesOfRouteField = ".metadata.services"
)

func (r *HTTPRouteReconciler) SetupWithManager(mgr kube_ctrl.Manager) error {
	if err := mgr.GetFieldIndexer().IndexField(context.Background(), &gatewayapi.HTTPRoute{}, servicesOfRouteField, func(obj kube_client.Object) []string {
		route := obj.(*gatewayapi.HTTPRoute)

		var names []string

		for _, rule := range route.Spec.Rules {
			var allBackendRefs []gatewayapi.BackendObjectReference
			for _, backendRef := range rule.BackendRefs {
				allBackendRefs = append(allBackendRefs, backendRef.BackendObjectReference)
			}
			for _, filter := range rule.Filters {
				if filter.Type == gatewayapi_v1.HTTPRouteFilterRequestMirror {
					allBackendRefs = append(allBackendRefs, filter.RequestMirror.BackendRef)
				}
			}
			for _, backendRef := range allBackendRefs {
				if string(*backendRef.Group) != kube_core.SchemeGroupVersion.Group || *backendRef.Kind != "Service" {
					continue
				}

				namespace := route.Namespace
				if backendRef.Namespace != nil {
					namespace = string(*backendRef.Namespace)
				}
				names = append(
					names,
					kube_types.NamespacedName{Namespace: namespace, Name: string(backendRef.Name)}.String(),
				)
			}
		}

		return names
	}); err != nil {
		return err
	}
	return kube_ctrl.NewControllerManagedBy(mgr).
		Named("kuma-http-route-controller").
		For(&gatewayapi.HTTPRoute{}).
		Watches(
			&kube_core.Service{},
			kube_handler.EnqueueRequestsFromMapFunc(routesForService(r.Log, r.Client)),
		).
		Complete(r)
}
