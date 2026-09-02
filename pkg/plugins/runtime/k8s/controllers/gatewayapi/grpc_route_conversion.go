package gatewayapi

import (
	"context"
	"fmt"
	"regexp"
	"strings"

	"github.com/shopspring/decimal"
	kube_core "k8s.io/api/core/v1"
	kube_apierrs "k8s.io/apimachinery/pkg/api/errors"
	kube_apimeta "k8s.io/apimachinery/pkg/api/meta"
	kube_meta "k8s.io/apimachinery/pkg/apis/meta/v1"
	kube_schema "k8s.io/apimachinery/pkg/runtime/schema"
	kube_types "k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/intstr"
	kube_client "sigs.k8s.io/controller-runtime/pkg/client"
	gatewayapi "sigs.k8s.io/gateway-api/apis/v1"
	gatewayapi_beta "sigs.k8s.io/gateway-api/apis/v1beta1"

	common_api "github.com/kumahq/kuma/v3/api/common/v1alpha1"
	mesh_proto "github.com/kumahq/kuma/v3/api/mesh/v1alpha1"
	meshservice_api "github.com/kumahq/kuma/v3/pkg/core/resources/apis/meshservice/api/v1alpha1"
	meshservice_k8s "github.com/kumahq/kuma/v3/pkg/core/resources/apis/meshservice/k8s/v1alpha1"
	core_model "github.com/kumahq/kuma/v3/pkg/core/resources/model"
	"github.com/kumahq/kuma/v3/pkg/plugins/policies/meshhttproute/api/v1alpha1"
	"github.com/kumahq/kuma/v3/pkg/util/pointer"
)

func (r *GRPCRouteReconciler) gapiGRPCToKumaMeshRules(
	ctx context.Context,
	route *gatewayapi.GRPCRoute,
) ([]v1alpha1.Rule, []kube_meta.Condition, error) {
	var rules []v1alpha1.Rule
	var conditions []kube_meta.Condition

	for _, rule := range route.Spec.Rules {
		kumaRule, ruleConditions, err := r.gapiGRPCToKumaMeshRule(ctx, route, rule)
		if err != nil {
			return nil, nil, err
		}

		for _, condition := range ruleConditions {
			if kube_apimeta.FindStatusCondition(conditions, condition.Type) == nil {
				kube_apimeta.SetStatusCondition(&conditions, condition)
			}
		}

		rules = append(rules, kumaRule)
	}

	return rules, conditions, nil
}

func (r *GRPCRouteReconciler) gapiServiceToMeshRoute(
	routeNamespace string,
	rules []v1alpha1.Rule,
	parent *kube_core.Service,
	parentPort *gatewayapi.PortNumber,
	parentSectionName *gatewayapi.SectionName,
) (core_model.ResourceSpec, bool) {
	targetRef := common_api.TopLevelTargetRef{
		Kind: common_api.TopLevelTargetRefKindDataplane,
		Labels: &map[string]string{
			mesh_proto.KubeNamespaceTag: routeNamespace,
		},
	}

	if routeNamespace == parent.GetNamespace() {
		targetRef = common_api.TopLevelTargetRef{Kind: common_api.TopLevelTargetRefKindMesh}
	}

	var tos []v1alpha1.To
	for _, port := range parent.Spec.Ports {
		if parentPort != nil && port.Port != *parentPort {
			continue
		}
		sectionName := port.Name
		if sectionName == "" {
			sectionName = fmt.Sprintf("%d", port.Port)
		}
		if parentSectionName != nil && sectionName != string(*parentSectionName) {
			continue
		}
		tos = append(tos, v1alpha1.To{
			TargetRef: common_api.OutboundTargetRef{
				Kind: common_api.OutboundTargetRefKindMeshService,
				Labels: &map[string]string{
					mesh_proto.DisplayName:      parent.GetName(),
					mesh_proto.KubeNamespaceTag: parent.GetNamespace(),
				},
				SectionName: pointer.To(sectionName),
			},
			Rules: rules,
		})
	}

	if len(tos) == 0 && (parentPort != nil || parentSectionName != nil) {
		return nil, false
	}

	return &v1alpha1.MeshHTTPRoute{
		TargetRef: &targetRef,
		To:        &tos,
	}, true
}

func (r *GRPCRouteReconciler) gapiMeshServiceToMeshRoute(
	routeNamespace string,
	rules []v1alpha1.Rule,
	parent *meshservice_k8s.MeshService,
	parentPort *gatewayapi.PortNumber,
	parentSectionName *gatewayapi.SectionName,
) (core_model.ResourceSpec, bool) {
	targetRef := common_api.TopLevelTargetRef{
		Kind: common_api.TopLevelTargetRefKindDataplane,
		Labels: &map[string]string{
			mesh_proto.KubeNamespaceTag: routeNamespace,
		},
	}

	if routeNamespace == parent.GetNamespace() {
		targetRef = common_api.TopLevelTargetRef{Kind: common_api.TopLevelTargetRefKindMesh}
	}

	var ports []meshservice_api.Port
	if parent.Spec != nil {
		for _, port := range parent.Spec.Ports {
			if parentPort != nil && port.Port != *parentPort {
				continue
			}
			if parentSectionName != nil && port.GetName() != string(*parentSectionName) {
				continue
			}
			ports = append(ports, port)
		}
	}
	if len(ports) == 0 && (parentPort != nil || parentSectionName != nil) {
		return nil, false
	}

	var tos []v1alpha1.To
	labels := meshServiceRefLabels(parent)
	for _, port := range ports {
		tos = append(tos, v1alpha1.To{
			TargetRef: common_api.OutboundTargetRef{
				Kind:        common_api.OutboundTargetRefKindMeshService,
				Labels:      &labels,
				SectionName: pointer.To(port.GetName()),
			},
			Rules: rules,
		})
	}

	return &v1alpha1.MeshHTTPRoute{
		TargetRef: &targetRef,
		To:        &tos,
	}, true
}

func (r *GRPCRouteReconciler) gapiGRPCToKumaMeshRule(
	ctx context.Context,
	route *gatewayapi.GRPCRoute,
	rule gatewayapi.GRPCRouteRule,
) (v1alpha1.Rule, []kube_meta.Condition, error) {
	var conditions []kube_meta.Condition
	var matches []v1alpha1.Match
	var filters []v1alpha1.Filter
	var backendRefs []v1alpha1.BackendRef

	for _, gapiMatch := range rule.Matches {
		match, ok := r.gapiGRPCToKumaMeshMatch(gapiMatch)
		if !ok {
			continue
		}
		matches = append(matches, match)
	}
	if len(matches) == 0 {
		matches = []v1alpha1.Match{{
			Path: &v1alpha1.PathMatch{Type: v1alpha1.PathPrefix, Value: "/"},
		}}
	}

	for _, gapiFilter := range rule.Filters {
		filter, filterConditions, ok := r.gapiGRPCToKumaMeshFilter(ctx, route.Namespace, gapiFilter)
		if !ok {
			continue
		}

		for _, condition := range filterConditions {
			if kube_apimeta.FindStatusCondition(conditions, condition.Type) == nil {
				kube_apimeta.SetStatusCondition(&conditions, condition)
			}
		}

		if len(filterConditions) == 0 {
			filters = append(filters, filter)
		}
	}

	for _, gapiBackendRef := range rule.BackendRefs {
		ref, refCondition, err := r.uncheckedGapiToKumaRef(ctx, route.Namespace, gapiBackendRef.BackendObjectReference)
		if err != nil {
			return v1alpha1.Rule{}, nil, err
		}

		refCondition.AddIfFalseAndNotPresent(&conditions)
		if refCondition.preventsBackendTarget() {
			continue
		}

		weight := uint(1)
		if gapiBackendRef.Weight != nil {
			weight = uint(*gapiBackendRef.Weight)
		}
		backendRefs = append(backendRefs, v1alpha1.BackendRef{
			TargetRef: ref,
			Weight:    pointer.To(weight),
		})
	}

	return v1alpha1.Rule{
		Matches: matches,
		Default: v1alpha1.RuleConf{
			Filters:     &filters,
			BackendRefs: &backendRefs,
		},
	}, conditions, nil
}

func (r *GRPCRouteReconciler) gapiGRPCToKumaMeshMatch(gapiMatch gatewayapi.GRPCRouteMatch) (v1alpha1.Match, bool) {
	var match v1alpha1.Match

	if gapiMatch.Method != nil {
		path, ok := grpcMethodMatchToPathMatch(*gapiMatch.Method)
		if !ok {
			return v1alpha1.Match{}, false
		}
		match.Path = path
	}

	for _, gapiHeader := range gapiMatch.Headers {
		headerType := gatewayapi.GRPCHeaderMatchExact
		if gapiHeader.Type != nil {
			headerType = *gapiHeader.Type
		}

		header := common_api.HeaderMatch{
			Type:  pointer.To(common_api.HeaderMatchType(headerType)),
			Name:  common_api.HeaderName(strings.ToLower(string(gapiHeader.Name))),
			Value: common_api.HeaderValue(gapiHeader.Value),
		}
		match.Headers = pointer.To(append(pointer.Deref(match.Headers), header))
	}

	return match, true
}

func grpcMethodMatchToPathMatch(methodMatch gatewayapi.GRPCMethodMatch) (*v1alpha1.PathMatch, bool) {
	matchType := gatewayapi.GRPCMethodMatchExact
	if methodMatch.Type != nil {
		matchType = *methodMatch.Type
	}

	service := pointer.Deref(methodMatch.Service)
	method := pointer.Deref(methodMatch.Method)

	switch matchType {
	case gatewayapi.GRPCMethodMatchExact:
		switch {
		case service != "" && method != "":
			return &v1alpha1.PathMatch{Type: v1alpha1.Exact, Value: fmt.Sprintf("/%s/%s", service, method)}, true
		case service != "":
			return &v1alpha1.PathMatch{Type: v1alpha1.PathPrefix, Value: fmt.Sprintf("/%s", service)}, true
		case method != "":
			return &v1alpha1.PathMatch{Type: v1alpha1.RegularExpression, Value: fmt.Sprintf("^/[^/]+/%s$", regexp.QuoteMeta(method))}, true
		default:
			return nil, false
		}
	case gatewayapi.GRPCMethodMatchRegularExpression:
		servicePattern := "[^/]+"
		if service != "" {
			servicePattern = service
		}
		methodPattern := "[^/]+"
		if method != "" {
			methodPattern = method
		}
		return &v1alpha1.PathMatch{Type: v1alpha1.RegularExpression, Value: fmt.Sprintf("^/(?:%s)/(?:%s)$", servicePattern, methodPattern)}, true
	default:
		return nil, false
	}
}

func grpcMirrorPercentage(mirror gatewayapi.HTTPRequestMirrorFilter) *intstr.IntOrString {
	if mirror.Percent != nil {
		return pointer.To(intstr.FromInt32(*mirror.Percent))
	}
	if mirror.Fraction == nil {
		return nil
	}

	denominator := int32(100)
	if mirror.Fraction.Denominator != nil {
		denominator = *mirror.Fraction.Denominator
	}
	percentage := decimal.NewFromInt(int64(mirror.Fraction.Numerator)).
		Mul(decimal.NewFromInt(100)).
		Div(decimal.NewFromInt(int64(denominator)))

	return pointer.To(intstr.FromString(percentage.String()))
}

func (r *GRPCRouteReconciler) gapiGRPCToKumaMeshFilter(
	ctx context.Context,
	routeNamespace string,
	gapiFilter gatewayapi.GRPCRouteFilter,
) (v1alpha1.Filter, []kube_meta.Condition, bool) {
	switch gapiFilter.Type {
	case gatewayapi.GRPCRouteFilterRequestHeaderModifier:
		modifier := gapiFilter.RequestHeaderModifier
		return v1alpha1.Filter{
			Type: v1alpha1.RequestHeaderModifierType,
			RequestHeaderModifier: &v1alpha1.HeaderModifier{
				Add:    pointer.To(fromGAPIHeaders(modifier.Add)),
				Set:    pointer.To(fromGAPIHeaders(modifier.Set)),
				Remove: pointer.To(modifier.Remove),
			},
		}, nil, true
	case gatewayapi.GRPCRouteFilterResponseHeaderModifier:
		modifier := gapiFilter.ResponseHeaderModifier
		return v1alpha1.Filter{
			Type: v1alpha1.ResponseHeaderModifierType,
			ResponseHeaderModifier: &v1alpha1.HeaderModifier{
				Add:    pointer.To(fromGAPIHeaders(modifier.Add)),
				Set:    pointer.To(fromGAPIHeaders(modifier.Set)),
				Remove: pointer.To(modifier.Remove),
			},
		}, nil, true
	case gatewayapi.GRPCRouteFilterRequestMirror:
		mirror := gapiFilter.RequestMirror

		ref, refCondition, err := r.uncheckedGapiToKumaRef(ctx, routeNamespace, mirror.BackendRef)
		if err != nil {
			return v1alpha1.Filter{}, nil, false
		}

		var conditions []kube_meta.Condition
		refCondition.AddIfFalseAndNotPresent(&conditions)

		return v1alpha1.Filter{
			Type: v1alpha1.RequestMirrorType,
			RequestMirror: &v1alpha1.RequestMirror{
				BackendRef: common_api.BackendRef{TargetRef: ref},
				Percentage: grpcMirrorPercentage(*mirror),
			},
		}, conditions, true
	default:
		return v1alpha1.Filter{}, []kube_meta.Condition{{
			Type:    string(gatewayapi.RouteConditionAccepted),
			Status:  kube_meta.ConditionFalse,
			Reason:  string(gatewayapi.RouteReasonUnsupportedValue),
			Message: fmt.Sprintf("GRPCRoute filter type %q is not supported", gapiFilter.Type),
		}}, true
	}
}

func (r *GRPCRouteReconciler) uncheckedGapiToKumaRef(
	ctx context.Context, objectNamespace string, ref gatewayapi.BackendObjectReference,
) (common_api.TargetRef, *ResolvedRefsConditionFalse, error) {
	details, ok := backendObjectReferenceInfo(objectNamespace, ref)
	refNamespace := objectNamespace
	if ok {
		refNamespace = details.Namespace
	}

	unresolvedTargetRef := common_api.TargetRef{
		Kind: common_api.MeshService,
		Labels: &map[string]string{
			mesh_proto.DisplayName:      string(ref.Name),
			mesh_proto.KubeNamespaceTag: refNamespace,
		},
	}
	if !ok {
		return unresolvedTargetRef,
			&ResolvedRefsConditionFalse{
				Reason:  string(gatewayapi.RouteReasonInvalidKind),
				Message: "backend reference must be Service or MeshService",
			},
			nil
	}

	namespacedName := kube_types.NamespacedName{Namespace: details.Namespace, Name: details.Name}
	gk := kube_schema.GroupKind{Kind: details.Kind, Group: details.Group}

	if details.Namespace != objectNamespace {
		allowed, err := r.referenceGrantAllowsBackendRef(ctx, sourceRouteKindGRPCRoute, objectNamespace, details)
		if err != nil {
			return common_api.TargetRef{}, nil, err
		}
		if !allowed {
			return common_api.TargetRef{},
				&ResolvedRefsConditionFalse{
					Reason:  string(gatewayapi.RouteReasonRefNotPermitted),
					Message: fmt.Sprintf("backend reference to %s %q is not permitted by any ReferenceGrant", gk.String(), namespacedName.String()),
				},
				nil
		}
	}

	if gk.Kind == "Service" && gk.Group == "" {
		if ref.Port == nil {
			return unresolvedTargetRef,
				&ResolvedRefsConditionFalse{
					Reason:  string(gatewayapi.RouteReasonInvalidKind),
					Message: "backend reference to Service must include a port",
				},
				nil
		}
		port := *ref.Port

		svc := &kube_core.Service{}
		if err := r.Get(ctx, namespacedName, svc); err != nil {
			if kube_apierrs.IsNotFound(err) {
				return unresolvedTargetRef,
					&ResolvedRefsConditionFalse{
						Reason:  string(gatewayapi.RouteReasonBackendNotFound),
						Message: fmt.Sprintf("backend reference references a non-existent Service %q", namespacedName.String()),
					},
					nil
			}
			return common_api.TargetRef{}, nil, err
		}

		sectionName := fmt.Sprintf("%d", port)
		portFound := false
		for _, svcPort := range svc.Spec.Ports {
			if svcPort.Port == port {
				portFound = true
				if svcPort.Name != "" {
					sectionName = svcPort.Name
				}
				break
			}
		}

		targetRef := common_api.TargetRef{
			Kind: common_api.MeshService,
			Labels: &map[string]string{
				mesh_proto.DisplayName:      svc.GetName(),
				mesh_proto.KubeNamespaceTag: svc.GetNamespace(),
			},
			SectionName: pointer.To(sectionName),
		}

		if !portFound {
			return targetRef,
				&ResolvedRefsConditionFalse{
					Reason:  string(gatewayapi.RouteReasonBackendNotFound),
					Message: fmt.Sprintf("Service %q does not have a port %d", namespacedName.String(), port),
				},
				nil
		}

		return targetRef, nil, nil
	}

	if gk.Kind == "MeshService" && gk.Group == meshservice_k8s.GroupVersion.Group {
		ms := &meshservice_k8s.MeshService{}
		if err := r.Get(ctx, namespacedName, ms); err != nil {
			if kube_apierrs.IsNotFound(err) {
				return unresolvedTargetRef,
					&ResolvedRefsConditionFalse{
						Reason:  string(gatewayapi.RouteReasonBackendNotFound),
						Message: fmt.Sprintf("backend reference references a non-existent MeshService %q", namespacedName.String()),
					},
					nil
			}
			return common_api.TargetRef{}, nil, err
		}

		labels := meshServiceRefLabels(ms)
		if ref.Port == nil {
			return common_api.TargetRef{
				Kind:   common_api.MeshService,
				Labels: &labels,
			}, nil, nil
		}

		port := *ref.Port
		sectionName := fmt.Sprintf("%d", port)
		portFound := false
		if ms.Spec != nil {
			for _, msPort := range ms.Spec.Ports {
				if msPort.Port == port {
					portFound = true
					sectionName = msPort.GetName()
					break
				}
			}
		}
		if !portFound {
			return common_api.TargetRef{
					Kind:        common_api.MeshService,
					Labels:      &labels,
					SectionName: pointer.To(sectionName),
				},
				&ResolvedRefsConditionFalse{
					Reason:  string(gatewayapi.RouteReasonBackendNotFound),
					Message: fmt.Sprintf("MeshService %q does not have a port %d", namespacedName.String(), port),
				},
				nil
		}

		return common_api.TargetRef{
			Kind:        common_api.MeshService,
			Labels:      &labels,
			SectionName: pointer.To(sectionName),
		}, nil, nil
	}

	return unresolvedTargetRef, nil, nil
}

func (r *GRPCRouteReconciler) referenceGrantAllowsBackendRef(
	ctx context.Context,
	sourceRouteKind string,
	routeNamespace string,
	backendRef backendObjectReferenceDetails,
) (bool, error) {
	var grants gatewayapi_beta.ReferenceGrantList
	if err := r.List(ctx, &grants, kube_client.InNamespace(backendRef.Namespace)); err != nil {
		return false, err
	}

	for i := range grants.Items {
		if backendObjectReferenceMatchesGrant(&grants.Items[i], sourceRouteKind, routeNamespace, backendRef) {
			return true, nil
		}
	}

	return false, nil
}
