package gatewayapi

import (
	"context"
	"fmt"
	"reflect"
	"slices"
	"strconv"
	"strings"

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

	meshservice_k8s "github.com/kumahq/kuma/v3/pkg/core/resources/apis/meshservice/k8s/v1alpha1"
	"github.com/kumahq/kuma/v3/pkg/core/resources/manager"
	core_model "github.com/kumahq/kuma/v3/pkg/core/resources/model"
	meshhttproute_api "github.com/kumahq/kuma/v3/pkg/plugins/policies/meshhttproute/api/v1alpha1"
	k8s_registry "github.com/kumahq/kuma/v3/pkg/plugins/resources/k8s/native/pkg/registry"
	"github.com/kumahq/kuma/v3/pkg/plugins/runtime/k8s/controllers/gatewayapi/attachment"
	"github.com/kumahq/kuma/v3/pkg/plugins/runtime/k8s/controllers/gatewayapi/common"
	"github.com/kumahq/kuma/v3/pkg/plugins/runtime/k8s/metadata"
	k8s_util "github.com/kumahq/kuma/v3/pkg/plugins/runtime/k8s/util"
	k8s_names "github.com/kumahq/kuma/v3/pkg/util/k8s"
	"github.com/kumahq/kuma/v3/pkg/util/pointer"
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

type backendObjectReferenceDetails struct {
	Namespace string
	Group     string
	Kind      string
	Name      string
	Ref       gatewayapi.BackendObjectReference
}

func backendObjectReferenceInfo(routeNamespace string, ref gatewayapi.BackendObjectReference) (backendObjectReferenceDetails, bool) {
	namespace := routeNamespace
	if ref.Namespace != nil {
		namespace = string(*ref.Namespace)
	}

	group := ""
	if ref.Group != nil {
		group = string(*ref.Group)
	}

	kind := "Service"
	if ref.Kind != nil {
		kind = string(*ref.Kind)
	}

	switch {
	case group == kube_core.SchemeGroupVersion.Group && kind == "Service":
		return backendObjectReferenceDetails{
			Namespace: namespace,
			Group:     group,
			Kind:      kind,
			Name:      string(ref.Name),
			Ref:       ref,
		}, true
	case group == meshservice_k8s.GroupVersion.Group && kind == "MeshService":
		return backendObjectReferenceDetails{
			Namespace: namespace,
			Group:     group,
			Kind:      kind,
			Name:      string(ref.Name),
			Ref:       ref,
		}, true
	default:
		return backendObjectReferenceDetails{}, false
	}
}

func backendObjectReferencesOfRoute(route *gatewayapi.HTTPRoute) []gatewayapi.BackendObjectReference {
	var refs []gatewayapi.BackendObjectReference

	for _, rule := range route.Spec.Rules {
		for _, backendRef := range rule.BackendRefs {
			refs = append(refs, backendRef.BackendObjectReference)
		}
		for _, filter := range rule.Filters {
			if filter.Type == gatewayapi_v1.HTTPRouteFilterRequestMirror {
				refs = append(refs, filter.RequestMirror.BackendRef)
			}
		}
	}

	return refs
}

func backendObjectReferenceMatchesGrant(grant *gatewayapi.ReferenceGrant, routeNamespace string, backendRef backendObjectReferenceDetails) bool {
	fromMatches := slices.ContainsFunc(grant.Spec.From, func(from gatewayapi.ReferenceGrantFrom) bool {
		return string(from.Group) == gatewayapi.GroupVersion.Group &&
			string(from.Kind) == "HTTPRoute" &&
			string(from.Namespace) == routeNamespace
	})
	if !fromMatches {
		return false
	}

	return slices.ContainsFunc(grant.Spec.To, func(to gatewayapi.ReferenceGrantTo) bool {
		if string(to.Group) != backendRef.Group || string(to.Kind) != backendRef.Kind {
			return false
		}
		return to.Name == nil || string(*to.Name) == backendRef.Name
	})
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
				ctx, r.Log, r.TypeRegistry, r.Client, req.NamespacedName, core_model.NoMesh, &meshhttproute_api.MeshHTTPRoute{}, nil,
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

	mesh := k8s_util.MeshOfByLabel(httpRoute, &ns)

	meshRouteSpecs, conditions, err := r.gapiToKumaRoutes(ctx, httpRoute)
	if err != nil {
		return kube_ctrl.Result{}, errors.Wrap(err, "could not generate MeshHTTPRoute.kuma.io resources")
	}

	if err := common.ReconcileLabelledObject(
		ctx, r.Log, r.TypeRegistry, r.Client, req.NamespacedName, mesh, &meshhttproute_api.MeshHTTPRoute{}, meshRouteSpecs,
	); err != nil {
		reconcileErr := errors.Wrap(err, "could not reconcile owned MeshHTTPRoute.kuma.io")
		if statusErr := r.updateStatus(ctx, httpRoute, generatedRouteWriteFailureConditions(conditions, reconcileErr)); statusErr != nil {
			r.Log.Error(statusErr, "unable to update HTTPRoute status after MeshHTTPRoute reconcile failure", "name", httpRoute.Name, "namespace", httpRoute.Namespace)
		}
		return kube_ctrl.Result{}, reconcileErr
	}

	if err := r.updateStatus(ctx, httpRoute, conditions); err != nil {
		return kube_ctrl.Result{}, errors.Wrap(err, "unable to update HTTPRoute status")
	}

	return kube_ctrl.Result{}, nil
}

type ParentConditions map[gatewayapi.ParentReference][]kube_meta.Condition

const maxGeneratedMeshHTTPRouteNameLength = 253

// gapiToKumaRoutes returns some number of GatewayRoutes that should be created
// for this HTTPRoute along with any statuses to be set on the HTTPRoute.
// Only unexpected errors are returned as error.
func (r *HTTPRouteReconciler) gapiToKumaRoutes(
	ctx context.Context,
	route *gatewayapi.HTTPRoute,
) (map[string]common.OwnedObject, ParentConditions, error) {
	routes := map[string]common.OwnedObject{}

	// The conditions we accumulate for each ParentRef
	conditions := ParentConditions{}

	labels := map[string]string{
		metadata.GatewayAPIRouteCreationTimestampLabel: strconv.FormatInt(route.CreationTimestamp.UnixNano(), 10),
	}

	for i, ref := range route.Spec.ParentRefs {
		refAttachment, refKind, err := attachment.EvaluateParentRefAttachment(ref)
		if err != nil {
			return nil, nil, errors.Wrapf(err, "unable to check parent ref %d", i)
		}

		if refAttachment == attachment.Unknown {
			continue
		}

		rules, rulesConditions, err := r.gapiToMeshRules(ctx, route)
		if err != nil {
			return nil, nil, err
		}

		var parentConditions []kube_meta.Condition

		// refAttachment is always Allowed here: Unknown was handled above.
		switch refKind {
		case attachment.Service:
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

				parentConditions = append(
					parentConditions,
					kube_meta.Condition{
						Type:    string(gatewayapi.RouteConditionAccepted),
						Status:  kube_meta.ConditionFalse,
						Reason:  string(gatewayapi_v1.RouteReasonNoMatchingParent),
						Message: fmt.Sprintf("Service %q does not exist", kube_types.NamespacedName{Namespace: namespace, Name: string(ref.Name)}.String()),
					},
				)

				conditions[ref] = prepareConditions(append(parentConditions, rulesConditions...))
				continue
			}

			if parent.Spec.ClusterIP == kube_core.ClusterIPNone {
				parentConditions = append(
					parentConditions,
					kube_meta.Condition{
						Type:    string(gatewayapi.RouteConditionAccepted),
						Status:  kube_meta.ConditionFalse,
						Reason:  string(gatewayapi_v1.RouteReasonNoMatchingParent),
						Message: fmt.Sprintf("Service %q has no MeshService to attach to", kube_types.NamespacedName{Namespace: namespace, Name: string(ref.Name)}.String()),
					},
				)

				conditions[ref] = prepareConditions(append(parentConditions, rulesConditions...))
				continue
			}

			routeSubName := generatedMeshHTTPRouteName(route, "Service", parent.GetNamespace(), parent.GetName())

			meshRoute, ok := r.gapiServiceToMeshRoute(route.Namespace, rules, &parent, ref.Port, ref.SectionName)
			if !ok {
				parentConditions = append(
					parentConditions,
					kube_meta.Condition{
						Type:    string(gatewayapi.RouteConditionAccepted),
						Status:  kube_meta.ConditionFalse,
						Reason:  string(gatewayapi_v1.RouteReasonNoMatchingParent),
						Message: fmt.Sprintf("Service %q has no port matching the parentRef", kube_types.NamespacedName{Namespace: namespace, Name: string(ref.Name)}.String()),
					},
				)

				conditions[ref] = prepareConditions(append(parentConditions, rulesConditions...))
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
			if err := r.Get(ctx, kube_types.NamespacedName{
				Name:      string(ref.Name),
				Namespace: namespace,
			}, &parent); err != nil {
				if !kube_apierrs.IsNotFound(err) {
					return nil, nil, err
				}

				parentConditions = append(
					parentConditions,
					kube_meta.Condition{
						Type:    string(gatewayapi.RouteConditionAccepted),
						Status:  kube_meta.ConditionFalse,
						Reason:  string(gatewayapi_v1.RouteReasonNoMatchingParent),
						Message: fmt.Sprintf("MeshService %q does not exist", kube_types.NamespacedName{Namespace: namespace, Name: string(ref.Name)}.String()),
					},
				)

				conditions[ref] = prepareConditions(append(parentConditions, rulesConditions...))
				continue
			}

			routeSubName := generatedMeshHTTPRouteName(route, "MeshService", parent.GetNamespace(), parent.GetName())

			meshRoute, ok := r.gapiMeshServiceToMeshRoute(route.Namespace, rules, &parent, ref.Port, ref.SectionName)
			if !ok {
				parentConditions = append(
					parentConditions,
					kube_meta.Condition{
						Type:    string(gatewayapi.RouteConditionAccepted),
						Status:  kube_meta.ConditionFalse,
						Reason:  string(gatewayapi_v1.RouteReasonNoMatchingParent),
						Message: fmt.Sprintf("MeshService %q has no port matching the parentRef", kube_types.NamespacedName{Namespace: namespace, Name: string(ref.Name)}.String()),
					},
				)

				conditions[ref] = prepareConditions(append(parentConditions, rulesConditions...))
				continue
			}

			ownedNamespace := r.SystemNamespace
			if route.Namespace == parent.GetNamespace() {
				ownedNamespace = route.Namespace
			}

			storeMeshHTTPRoute(routes, routeSubName, ownedNamespace, labels, meshRoute)
		}

		conditions[ref] = prepareConditions(append(parentConditions, rulesConditions...))
	}

	return routes, conditions, nil
}

func storeMeshHTTPRoute(routes map[string]common.OwnedObject, name string, namespace string, labels map[string]string, spec core_model.ResourceSpec) {
	route, ok := spec.(*meshhttproute_api.MeshHTTPRoute)
	if !ok {
		routes[name] = common.OwnedObject{Namespace: namespace, Spec: spec, Labels: labels}
		return
	}

	existing, ok := routes[name]
	if !ok {
		routes[name] = common.OwnedObject{Namespace: namespace, Spec: spec, Labels: labels}
		return
	}

	existingRoute, ok := existing.Spec.(*meshhttproute_api.MeshHTTPRoute)
	if !ok {
		routes[name] = common.OwnedObject{Namespace: namespace, Spec: spec, Labels: labels}
		return
	}

	merged := pointer.Deref(existingRoute.To)
	for _, candidate := range pointer.Deref(route.To) {
		duplicate := false
		for _, existingTo := range merged {
			if reflect.DeepEqual(existingTo, candidate) {
				duplicate = true
				break
			}
		}
		if !duplicate {
			merged = append(merged, candidate)
		}
	}
	existingRoute.To = pointer.To(merged)
}

func generatedMeshHTTPRouteName(route *gatewayapi.HTTPRoute, parentKind, parentNamespace, parentName string) string {
	normalizedParentKind := strings.ToLower(parentKind)
	normalizedParentNamespace := strings.ToLower(parentNamespace)
	normalizedParentName := strings.ToLower(parentName)

	fullName := strings.Join([]string{
		generatedMeshHTTPRouteNameSegment("rns", route.Namespace),
		generatedMeshHTTPRouteNameSegment("rn", route.Name),
		generatedMeshHTTPRouteNameSegment("pk", normalizedParentKind),
		generatedMeshHTTPRouteNameSegment("pns", normalizedParentNamespace),
		generatedMeshHTTPRouteNameSegment("pn", normalizedParentName),
	}, "--")
	if len(fullName) <= maxGeneratedMeshHTTPRouteNameLength {
		return fullName
	}

	hasher := k8s_names.NewHasher()
	hasher.Write([]byte(fullName))
	hashSuffix := k8s_names.HashToString(hasher)
	prefix := k8s_names.EnsureMaxLength(fullName, maxGeneratedMeshHTTPRouteNameLength-len(hashSuffix)-1)
	prefix = strings.TrimRight(prefix, "-.")
	if prefix == "" {
		return hashSuffix
	}
	return fmt.Sprintf("%s-%s", prefix, hashSuffix)
}

func generatedMeshHTTPRouteNameSegment(label string, value string) string {
	return fmt.Sprintf("%s-%d-%s", label, len(value), value)
}

func generatedRouteWriteFailureConditions(existing ParentConditions, reconcileErr error) ParentConditions {
	if len(existing) == 0 {
		return nil
	}

	failed := ParentConditions{}
	for ref, existingConditions := range existing {
		conditions := append([]kube_meta.Condition{}, existingConditions...)
		conditions = append(conditions, kube_meta.Condition{
			Type:    string(gatewayapi.RouteConditionAccepted),
			Status:  kube_meta.ConditionFalse,
			Reason:  string(gatewayapi.RouteReasonPending),
			Message: fmt.Sprintf("Failed to write generated MeshHTTPRoute: %s", reconcileErr.Error()),
		})
		failed[ref] = prepareConditions(conditions)
	}

	return failed
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

// routesForMeshService returns a function that calculates which HTTPRoutes might
// be affected by changes in a MeshService.
func routesForMeshService(l logr.Logger, client kube_client.Client) kube_handler.MapFunc {
	l = l.WithName("meshservice-to-routes-mapper")

	return func(ctx context.Context, obj kube_client.Object) []kube_reconcile.Request {
		ms, ok := obj.(*meshservice_k8s.MeshService)
		if !ok {
			l.Error(nil, "unexpected error converting object to MeshService", "typ", reflect.TypeOf(obj))
			return nil
		}

		var routes gatewayapi.HTTPRouteList
		if err := client.List(ctx, &routes, kube_client.MatchingFields{
			meshServicesOfRouteField: kube_client.ObjectKeyFromObject(ms).String(),
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

// routesForReferenceGrant returns the routes that may change validity when a
// grant in the target namespace is added, updated, or deleted. Update events
// run this mapper for both the old and new grant objects.
func routesForReferenceGrant(l logr.Logger, client kube_client.Client) kube_handler.MapFunc {
	l = l.WithName("referencegrant-to-routes-mapper")

	return func(ctx context.Context, obj kube_client.Object) []kube_reconcile.Request {
		grant, ok := obj.(*gatewayapi.ReferenceGrant)
		if !ok {
			l.Error(nil, "unexpected error converting object to ReferenceGrant", "typ", reflect.TypeOf(obj))
			return nil
		}

		var routes gatewayapi.HTTPRouteList
		if err := client.List(ctx, &routes); err != nil {
			l.Error(err, "unexpected error listing HTTPRoutes")
			return nil
		}

		var requests []kube_reconcile.Request
		for i := range routes.Items {
			route := &routes.Items[i]
			for _, backendRef := range backendObjectReferencesOfRoute(route) {
				details, ok := backendObjectReferenceInfo(route.Namespace, backendRef)
				if !ok || details.Namespace != grant.Namespace || details.Namespace == route.Namespace {
					continue
				}
				if backendObjectReferenceMatchesGrant(grant, route.Namespace, details) {
					requests = append(requests, kube_reconcile.Request{
						NamespacedName: kube_client.ObjectKeyFromObject(route),
					})
					break
				}
			}
		}

		return requests
	}
}

const (
	servicesOfRouteField     = ".metadata.services"
	meshServicesOfRouteField = ".metadata.meshservices"
)

// isMeshServiceRef reports whether the given group/kind pair references a
// kuma.io MeshService.
func isMeshServiceRef(group *gatewayapi.Group, kind *gatewayapi.Kind) bool {
	return group != nil && kind != nil && string(*group) == meshservice_k8s.GroupVersion.Group && string(*kind) == "MeshService"
}

// isServiceParentRef reports whether the given group/kind pair references a
// Service that can be used as a route parent.
func isServiceParentRef(group *gatewayapi.Group, kind *gatewayapi.Kind) bool {
	if group == nil || kind == nil || string(*kind) != "Service" {
		return false
	}
	return string(*group) == kube_core.GroupName || string(*group) == gatewayapi.GroupName
}

// meshServicesOfRoute returns the namespaced names of the MeshServices
// referenced by the given HTTPRoute's parentRefs, backendRefs and
// request-mirror backendRefs, defaulting an omitted namespace to the route's.
func meshServicesOfRoute(obj kube_client.Object) []string {
	route := obj.(*gatewayapi.HTTPRoute)

	var names []string

	for _, parentRef := range route.Spec.ParentRefs {
		if !isMeshServiceRef(parentRef.Group, parentRef.Kind) {
			continue
		}
		namespace := route.Namespace
		if parentRef.Namespace != nil {
			namespace = string(*parentRef.Namespace)
		}
		names = append(
			names,
			kube_types.NamespacedName{Namespace: namespace, Name: string(parentRef.Name)}.String(),
		)
	}

	for _, backendRef := range backendObjectReferencesOfRoute(route) {
		details, ok := backendObjectReferenceInfo(route.Namespace, backendRef)
		if !ok || details.Kind != "MeshService" {
			continue
		}
		names = append(
			names,
			kube_types.NamespacedName{Namespace: details.Namespace, Name: details.Name}.String(),
		)
	}

	return names
}

// servicesOfRoute returns the namespaced names of the Services referenced by
// the given HTTPRoute's parentRefs, backendRefs and request-mirror backendRefs,
// defaulting an omitted namespace to the route's.
func servicesOfRoute(obj kube_client.Object) []string {
	route := obj.(*gatewayapi.HTTPRoute)

	var names []string

	for _, parentRef := range route.Spec.ParentRefs {
		if !isServiceParentRef(parentRef.Group, parentRef.Kind) {
			continue
		}
		namespace := route.Namespace
		if parentRef.Namespace != nil {
			namespace = string(*parentRef.Namespace)
		}
		names = append(
			names,
			kube_types.NamespacedName{Namespace: namespace, Name: string(parentRef.Name)}.String(),
		)
	}

	for _, backendRef := range backendObjectReferencesOfRoute(route) {
		details, ok := backendObjectReferenceInfo(route.Namespace, backendRef)
		if !ok || details.Kind != "Service" {
			continue
		}
		names = append(
			names,
			kube_types.NamespacedName{Namespace: details.Namespace, Name: details.Name}.String(),
		)
	}

	return names
}

func (r *HTTPRouteReconciler) SetupWithManager(mgr kube_ctrl.Manager) error {
	if err := mgr.GetFieldIndexer().IndexField(context.Background(), &gatewayapi.HTTPRoute{}, meshServicesOfRouteField, meshServicesOfRoute); err != nil {
		return err
	}
	if err := mgr.GetFieldIndexer().IndexField(context.Background(), &gatewayapi.HTTPRoute{}, servicesOfRouteField, servicesOfRoute); err != nil {
		return err
	}
	return kube_ctrl.NewControllerManagedBy(mgr).
		Named("kuma-http-route-controller").
		For(&gatewayapi.HTTPRoute{}).
		Watches(
			&kube_core.Service{},
			kube_handler.EnqueueRequestsFromMapFunc(routesForService(r.Log, r.Client)),
		).
		Watches(
			&meshservice_k8s.MeshService{},
			kube_handler.EnqueueRequestsFromMapFunc(routesForMeshService(r.Log, r.Client)),
		).
		Watches(
			&gatewayapi.ReferenceGrant{},
			kube_handler.EnqueueRequestsFromMapFunc(routesForReferenceGrant(r.Log, r.Client)),
		).
		Complete(r)
}
