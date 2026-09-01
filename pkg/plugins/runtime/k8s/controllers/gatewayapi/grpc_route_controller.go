package gatewayapi

import (
	"context"
	"fmt"
	"strconv"

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
	gatewayapi "sigs.k8s.io/gateway-api/apis/v1"
	gatewayapi_beta "sigs.k8s.io/gateway-api/apis/v1beta1"

	meshservice_k8s "github.com/kumahq/kuma/v3/pkg/core/resources/apis/meshservice/k8s/v1alpha1"
	"github.com/kumahq/kuma/v3/pkg/core/resources/manager"
	core_model "github.com/kumahq/kuma/v3/pkg/core/resources/model"
	meshhttproute_api "github.com/kumahq/kuma/v3/pkg/plugins/policies/meshhttproute/api/v1alpha1"
	k8s_registry "github.com/kumahq/kuma/v3/pkg/plugins/resources/k8s/native/pkg/registry"
	"github.com/kumahq/kuma/v3/pkg/plugins/runtime/k8s/controllers/gatewayapi/attachment"
	"github.com/kumahq/kuma/v3/pkg/plugins/runtime/k8s/controllers/gatewayapi/common"
	"github.com/kumahq/kuma/v3/pkg/plugins/runtime/k8s/metadata"
	k8s_util "github.com/kumahq/kuma/v3/pkg/plugins/runtime/k8s/util"
)

type GRPCRouteReconciler struct {
	kube_client.Client
	Log logr.Logger

	Scheme          *kube_runtime.Scheme
	TypeRegistry    k8s_registry.TypeRegistry
	SystemNamespace string
	ResourceManager manager.ResourceManager
	Zone            string
}

func backendObjectReferencesOfGRPCRoute(route *gatewayapi.GRPCRoute) []gatewayapi.BackendObjectReference {
	var refs []gatewayapi.BackendObjectReference

	for _, rule := range route.Spec.Rules {
		for _, backendRef := range rule.BackendRefs {
			refs = append(refs, backendRef.BackendObjectReference)
		}
		for _, filter := range rule.Filters {
			if filter.Type == gatewayapi.GRPCRouteFilterRequestMirror {
				refs = append(refs, filter.RequestMirror.BackendRef)
			}
		}
	}

	return refs
}

func (r *GRPCRouteReconciler) Reconcile(ctx context.Context, req kube_ctrl.Request) (kube_ctrl.Result, error) {
	r.Log.V(1).Info("reconcile", "req", req)
	grpcRoute := &gatewayapi.GRPCRoute{}
	if err := r.Get(ctx, req.NamespacedName, grpcRoute); err != nil {
		if kube_apierrs.IsNotFound(err) {
			if err := common.ReconcileLabelledObject(
				ctx, r.Log, r.TypeRegistry, r.Client, sourceRouteKindGRPCRoute, req.NamespacedName, core_model.NoMesh, &meshhttproute_api.MeshHTTPRoute{}, nil,
			); err != nil {
				return kube_ctrl.Result{}, errors.Wrap(err, "could not delete owned MeshHTTPRoute.kuma.io")
			}
			return kube_ctrl.Result{}, nil
		}

		return kube_ctrl.Result{}, err
	}

	ns := kube_core.Namespace{}
	if err := r.Get(ctx, kube_types.NamespacedName{Name: grpcRoute.Namespace}, &ns); err != nil {
		return kube_ctrl.Result{}, errors.Wrap(err, "unable to get Namespace of GRPCRoute")
	}

	mesh := k8s_util.MeshOfByLabel(grpcRoute, &ns)
	meshRouteSpecs, conditions, err := r.gapiGRPCToKumaRoutes(ctx, grpcRoute)
	if err != nil {
		return kube_ctrl.Result{}, errors.Wrap(err, "could not generate MeshHTTPRoute.kuma.io resources")
	}

	if err := common.ReconcileLabelledObject(
		ctx, r.Log, r.TypeRegistry, r.Client, sourceRouteKindGRPCRoute, req.NamespacedName, mesh, &meshhttproute_api.MeshHTTPRoute{}, meshRouteSpecs,
	); err != nil {
		reconcileErr := errors.Wrap(err, "could not reconcile owned MeshHTTPRoute.kuma.io")
		if statusErr := r.updateGRPCRouteStatus(ctx, grpcRoute, generatedRouteWriteFailureConditions(conditions, reconcileErr)); statusErr != nil {
			r.Log.Error(statusErr, "unable to update GRPCRoute status after MeshHTTPRoute reconcile failure", "name", grpcRoute.Name, "namespace", grpcRoute.Namespace)
		}
		return kube_ctrl.Result{}, reconcileErr
	}

	if err := r.updateGRPCRouteStatus(ctx, grpcRoute, conditions); err != nil {
		return kube_ctrl.Result{}, errors.Wrap(err, "unable to update GRPCRoute status")
	}

	return kube_ctrl.Result{}, nil
}

func (r *GRPCRouteReconciler) gapiGRPCToKumaRoutes(
	ctx context.Context,
	route *gatewayapi.GRPCRoute,
) (map[string]common.OwnedObject, ParentConditions, error) {
	routes := map[string]common.OwnedObject{}
	conditions := ParentConditions{}
	labels := map[string]string{
		metadata.GatewayAPIRouteCreationTimestampLabel: strconv.FormatInt(route.CreationTimestamp.UnixNano(), 10),
	}

	for i, ref := range route.Spec.ParentRefs {
		refAttachment, refKind, err := attachment.EvaluateParentRefAttachment(gatewayapi_beta.ParentReference(ref))
		if err != nil {
			return nil, nil, errors.Wrapf(err, "unable to check parent ref %d", i)
		}
		if refAttachment == attachment.Unknown {
			continue
		}

		rules, rulesConditions, err := r.gapiGRPCToKumaMeshRules(ctx, route)
		if err != nil {
			return nil, nil, err
		}

		var parentConditions []kube_meta.Condition

		switch refKind {
		case attachment.Service:
			namespace := route.Namespace
			if ref.Namespace != nil {
				namespace = string(*ref.Namespace)
			}

			var parent kube_core.Service
			if err := r.Get(ctx, kube_types.NamespacedName{Name: string(ref.Name), Namespace: namespace}, &parent); err != nil {
				if !kube_apierrs.IsNotFound(err) {
					return nil, nil, err
				}
				parentConditions = append(parentConditions, kube_meta.Condition{
					Type:    string(gatewayapi.RouteConditionAccepted),
					Status:  kube_meta.ConditionFalse,
					Reason:  string(gatewayapi.RouteReasonNoMatchingParent),
					Message: fmt.Sprintf("Service %q does not exist", kube_types.NamespacedName{Namespace: namespace, Name: string(ref.Name)}.String()),
				})
				conditions[gatewayapi_beta.ParentReference(ref)] = prepareConditions(append(parentConditions, rulesConditions...))
				continue
			}

			if parent.Spec.ClusterIP == kube_core.ClusterIPNone {
				parentConditions = append(parentConditions, kube_meta.Condition{
					Type:    string(gatewayapi.RouteConditionAccepted),
					Status:  kube_meta.ConditionFalse,
					Reason:  string(gatewayapi.RouteReasonNoMatchingParent),
					Message: fmt.Sprintf("Service %q has no MeshService to attach to", kube_types.NamespacedName{Namespace: namespace, Name: string(ref.Name)}.String()),
				})
				conditions[gatewayapi_beta.ParentReference(ref)] = prepareConditions(append(parentConditions, rulesConditions...))
				continue
			}

			routeSubName := generatedMeshHTTPRouteName(sourceRouteKindGRPCRoute, route.Namespace, route.Name, "Service", parent.GetNamespace(), parent.GetName())
			meshRoute, ok := r.gapiServiceToMeshRoute(route.Namespace, rules, &parent, ref.Port, ref.SectionName)
			if !ok {
				parentConditions = append(parentConditions, kube_meta.Condition{
					Type:    string(gatewayapi.RouteConditionAccepted),
					Status:  kube_meta.ConditionFalse,
					Reason:  string(gatewayapi.RouteReasonNoMatchingParent),
					Message: fmt.Sprintf("Service %q has no port matching the parentRef", kube_types.NamespacedName{Namespace: namespace, Name: string(ref.Name)}.String()),
				})
				conditions[gatewayapi_beta.ParentReference(ref)] = prepareConditions(append(parentConditions, rulesConditions...))
				continue
			}
			if hasAcceptedFalse(rulesConditions) {
				conditions[gatewayapi_beta.ParentReference(ref)] = prepareConditions(append(parentConditions, rulesConditions...))
				continue
			}

			ownedNamespace := r.SystemNamespace
			if route.Namespace == parent.GetNamespace() {
				ownedNamespace = route.Namespace
			}
			storeMeshHTTPRoute(routes, routeSubName, ownedNamespace, labels, meshRoute)
		case attachment.MeshService:
			namespace := route.Namespace
			if ref.Namespace != nil {
				namespace = string(*ref.Namespace)
			}

			var parent meshservice_k8s.MeshService
			if err := r.Get(ctx, kube_types.NamespacedName{Name: string(ref.Name), Namespace: namespace}, &parent); err != nil {
				if !kube_apierrs.IsNotFound(err) {
					return nil, nil, err
				}
				parentConditions = append(parentConditions, kube_meta.Condition{
					Type:    string(gatewayapi.RouteConditionAccepted),
					Status:  kube_meta.ConditionFalse,
					Reason:  string(gatewayapi.RouteReasonNoMatchingParent),
					Message: fmt.Sprintf("MeshService %q does not exist", kube_types.NamespacedName{Namespace: namespace, Name: string(ref.Name)}.String()),
				})
				conditions[gatewayapi_beta.ParentReference(ref)] = prepareConditions(append(parentConditions, rulesConditions...))
				continue
			}

			routeSubName := generatedMeshHTTPRouteName(sourceRouteKindGRPCRoute, route.Namespace, route.Name, "MeshService", parent.GetNamespace(), parent.GetName())
			meshRoute, ok := r.gapiMeshServiceToMeshRoute(route.Namespace, rules, &parent, ref.Port, ref.SectionName)
			if !ok {
				parentConditions = append(parentConditions, kube_meta.Condition{
					Type:    string(gatewayapi.RouteConditionAccepted),
					Status:  kube_meta.ConditionFalse,
					Reason:  string(gatewayapi.RouteReasonNoMatchingParent),
					Message: fmt.Sprintf("MeshService %q has no port matching the parentRef", kube_types.NamespacedName{Namespace: namespace, Name: string(ref.Name)}.String()),
				})
				conditions[gatewayapi_beta.ParentReference(ref)] = prepareConditions(append(parentConditions, rulesConditions...))
				continue
			}
			if hasAcceptedFalse(rulesConditions) {
				conditions[gatewayapi_beta.ParentReference(ref)] = prepareConditions(append(parentConditions, rulesConditions...))
				continue
			}

			ownedNamespace := r.SystemNamespace
			if route.Namespace == parent.GetNamespace() {
				ownedNamespace = route.Namespace
			}
			storeMeshHTTPRoute(routes, routeSubName, ownedNamespace, labels, meshRoute)
		}

		conditions[gatewayapi_beta.ParentReference(ref)] = prepareConditions(append(parentConditions, rulesConditions...))
	}

	return routes, conditions, nil
}

func meshServicesOfGRPCRoute(obj kube_client.Object) []string {
	route := obj.(*gatewayapi.GRPCRoute)
	var names []string

	for _, parentRef := range route.Spec.ParentRefs {
		if parentRef.Group == nil || parentRef.Kind == nil || string(*parentRef.Group) != meshservice_k8s.GroupVersion.Group || string(*parentRef.Kind) != "MeshService" {
			continue
		}
		namespace := route.Namespace
		if parentRef.Namespace != nil {
			namespace = string(*parentRef.Namespace)
		}
		names = append(names, kube_types.NamespacedName{Namespace: namespace, Name: string(parentRef.Name)}.String())
	}

	for _, backendRef := range backendObjectReferencesOfGRPCRoute(route) {
		details, ok := backendObjectReferenceInfo(route.Namespace, gatewayapi_beta.BackendObjectReference(backendRef))
		if !ok || details.Kind != "MeshService" {
			continue
		}
		names = append(names, kube_types.NamespacedName{Namespace: details.Namespace, Name: details.Name}.String())
	}

	return names
}

func servicesOfGRPCRoute(obj kube_client.Object) []string {
	route := obj.(*gatewayapi.GRPCRoute)
	var names []string

	for _, parentRef := range route.Spec.ParentRefs {
		if parentRef.Group == nil || parentRef.Kind == nil || string(*parentRef.Kind) != "Service" {
			continue
		}
		if string(*parentRef.Group) != kube_core.GroupName && string(*parentRef.Group) != gatewayapi.GroupName {
			continue
		}
		namespace := route.Namespace
		if parentRef.Namespace != nil {
			namespace = string(*parentRef.Namespace)
		}
		names = append(names, kube_types.NamespacedName{Namespace: namespace, Name: string(parentRef.Name)}.String())
	}

	for _, backendRef := range backendObjectReferencesOfGRPCRoute(route) {
		details, ok := backendObjectReferenceInfo(route.Namespace, gatewayapi_beta.BackendObjectReference(backendRef))
		if !ok || details.Kind != "Service" {
			continue
		}
		names = append(names, kube_types.NamespacedName{Namespace: details.Namespace, Name: details.Name}.String())
	}

	return names
}

func grpcRoutesForService(l logr.Logger, client kube_client.Client) kube_handler.MapFunc {
	l = l.WithName("service-to-grpcroutes-mapper")

	return func(ctx context.Context, obj kube_client.Object) []kube_reconcile.Request {
		svc, ok := obj.(*kube_core.Service)
		if !ok {
			l.Error(nil, "unexpected error converting object to Service", "typ", fmt.Sprintf("%T", obj))
			return nil
		}

		var routes gatewayapi.GRPCRouteList
		if err := client.List(ctx, &routes, kube_client.MatchingFields{servicesOfRouteField: kube_client.ObjectKeyFromObject(svc).String()}); err != nil {
			l.Error(err, "unexpected error listing GRPCRoutes")
			return nil
		}

		var requests []kube_reconcile.Request
		for i := range routes.Items {
			requests = append(requests, kube_reconcile.Request{NamespacedName: kube_client.ObjectKeyFromObject(&routes.Items[i])})
		}
		return requests
	}
}

func grpcRoutesForMeshService(l logr.Logger, client kube_client.Client) kube_handler.MapFunc {
	l = l.WithName("meshservice-to-grpcroutes-mapper")

	return func(ctx context.Context, obj kube_client.Object) []kube_reconcile.Request {
		ms, ok := obj.(*meshservice_k8s.MeshService)
		if !ok {
			l.Error(nil, "unexpected error converting object to MeshService", "typ", fmt.Sprintf("%T", obj))
			return nil
		}

		var routes gatewayapi.GRPCRouteList
		if err := client.List(ctx, &routes, kube_client.MatchingFields{meshServicesOfRouteField: kube_client.ObjectKeyFromObject(ms).String()}); err != nil {
			l.Error(err, "unexpected error listing GRPCRoutes")
			return nil
		}

		var requests []kube_reconcile.Request
		for i := range routes.Items {
			requests = append(requests, kube_reconcile.Request{NamespacedName: kube_client.ObjectKeyFromObject(&routes.Items[i])})
		}
		return requests
	}
}

func grpcRoutesForReferenceGrant(l logr.Logger, client kube_client.Client) kube_handler.MapFunc {
	l = l.WithName("referencegrant-to-grpcroutes-mapper")

	return func(ctx context.Context, obj kube_client.Object) []kube_reconcile.Request {
		grant, ok := obj.(*gatewayapi_beta.ReferenceGrant)
		if !ok {
			l.Error(nil, "unexpected error converting object to ReferenceGrant", "typ", fmt.Sprintf("%T", obj))
			return nil
		}

		var routes gatewayapi.GRPCRouteList
		if err := client.List(ctx, &routes); err != nil {
			l.Error(err, "unexpected error listing GRPCRoutes")
			return nil
		}

		var requests []kube_reconcile.Request
		for i := range routes.Items {
			route := &routes.Items[i]
			for _, backendRef := range backendObjectReferencesOfGRPCRoute(route) {
				details, ok := backendObjectReferenceInfo(route.Namespace, gatewayapi_beta.BackendObjectReference(backendRef))
				if !ok || details.Namespace != grant.Namespace || details.Namespace == route.Namespace {
					continue
				}
				if backendObjectReferenceMatchesGrant(grant, sourceRouteKindGRPCRoute, route.Namespace, details) {
					requests = append(requests, kube_reconcile.Request{NamespacedName: kube_client.ObjectKeyFromObject(route)})
					break
				}
			}
		}

		return requests
	}
}

func (r *GRPCRouteReconciler) SetupWithManager(mgr kube_ctrl.Manager) error {
	if err := mgr.GetFieldIndexer().IndexField(context.Background(), &gatewayapi.GRPCRoute{}, meshServicesOfRouteField, meshServicesOfGRPCRoute); err != nil {
		return err
	}
	if err := mgr.GetFieldIndexer().IndexField(context.Background(), &gatewayapi.GRPCRoute{}, servicesOfRouteField, servicesOfGRPCRoute); err != nil {
		return err
	}

	return kube_ctrl.NewControllerManagedBy(mgr).
		Named("kuma-grpc-route-controller").
		For(&gatewayapi.GRPCRoute{}).
		Watches(&kube_core.Service{}, kube_handler.EnqueueRequestsFromMapFunc(grpcRoutesForService(r.Log, r.Client))).
		Watches(&meshservice_k8s.MeshService{}, kube_handler.EnqueueRequestsFromMapFunc(grpcRoutesForMeshService(r.Log, r.Client))).
		Watches(&gatewayapi_beta.ReferenceGrant{}, kube_handler.EnqueueRequestsFromMapFunc(grpcRoutesForReferenceGrant(r.Log, r.Client))).
		Complete(r)
}
