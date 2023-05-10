package gatewayapi

import (
	"context"
	"fmt"
	"strings"

	"github.com/pkg/errors"
	kube_core "k8s.io/api/core/v1"
	kube_apierrs "k8s.io/apimachinery/pkg/api/errors"
	kube_meta "k8s.io/apimachinery/pkg/apis/meta/v1"
	kube_client "sigs.k8s.io/controller-runtime/pkg/client"
	gatewayapi "sigs.k8s.io/gateway-api/apis/v1beta1"

	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
	core_mesh "github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
	"github.com/kumahq/kuma/pkg/core/resources/store"
	mesh_k8s "github.com/kumahq/kuma/pkg/plugins/resources/k8s/native/api/v1alpha1"
	"github.com/kumahq/kuma/pkg/plugins/runtime/gateway"
	"github.com/kumahq/kuma/pkg/plugins/runtime/k8s/controllers/gatewayapi/policy"
	k8s_util "github.com/kumahq/kuma/pkg/plugins/runtime/k8s/util"
	util_proto "github.com/kumahq/kuma/pkg/util/proto"
)

func (r *HTTPRouteReconciler) gapiToKumaRule(
	ctx context.Context, mesh string, route *gatewayapi.HTTPRoute, rule gatewayapi.HTTPRouteRule,
) (*mesh_proto.MeshGatewayRoute_HttpRoute_Rule, *ResolvedRefsConditionFalse, error) {
	var backends []*mesh_proto.MeshGatewayRoute_Backend

	var condition *ResolvedRefsConditionFalse

	for _, backend := range rule.BackendRefs {
		ref := backend.BackendObjectReference

		destination, refCondition, err := r.gapiToKumaRef(ctx, mesh, route.Namespace, ref)
		if err != nil {
			return nil, condition, err
		}

		if refCondition != nil {
			condition = refCondition
		}
		backends = append(backends, &mesh_proto.MeshGatewayRoute_Backend{
			// Weight has a default of 1
			Weight:      uint32(*backend.Weight),
			Destination: destination,
		})
	}

	var matches []*mesh_proto.MeshGatewayRoute_HttpRoute_Match

	for _, match := range rule.Matches {
		kumaMatch, err := gapiToKumaMatch(match)
		if err != nil {
			return nil, nil, errors.Wrap(err, "couldn't convert match")
		}

		matches = append(matches, kumaMatch)
	}

	var filters []*mesh_proto.MeshGatewayRoute_HttpRoute_Filter

	var foundBackendlessFilter bool

	for _, filter := range rule.Filters {
		kumaFilters, filterCondition, err := r.gapiToKumaFilters(ctx, mesh, route.Namespace, filter)
		if err != nil {
			return nil, condition, err
		}

		switch filter.Type {
		case gatewayapi.HTTPRouteFilterRequestRedirect:
			foundBackendlessFilter = true
		}

		if filterCondition != nil {
			condition = filterCondition
		} else {
			filters = append(filters, kumaFilters...)
		}
	}

	var kumaRule *mesh_proto.MeshGatewayRoute_HttpRoute_Rule
	if len(backends) > 0 || foundBackendlessFilter {
		// TODO Make sure this results in a 500
		kumaRule = &mesh_proto.MeshGatewayRoute_HttpRoute_Rule{
			Matches:  matches,
			Filters:  filters,
			Backends: backends,
		}
	}
	return kumaRule, condition, nil
}

// gapiToKumaRouteConf converts the route into a route spec and returns any
// conditions that should be set on parent refs. These are the same across all
// Kuma parent refs. If a conf cannot be created, it returns a nil conf.
// It returns error only if an unexpected error has occurred. Issues related to the
// HTTPRoute spec are reflected in the Conditions.
func (r *HTTPRouteReconciler) gapiToKumaRouteConf(
	ctx context.Context, mesh string, route *gatewayapi.HTTPRoute,
) (*mesh_proto.MeshGatewayRoute_Conf, []kube_meta.Condition, error) {
	var hostnames []string

	for _, hn := range route.Spec.Hostnames {
		hostnames = append(hostnames, string(hn))
	}

	var resolvedRefFalse *ResolvedRefsConditionFalse
	var rules []*mesh_proto.MeshGatewayRoute_HttpRoute_Rule

	for _, rule := range route.Spec.Rules {
		kumaRule, ruleCondition, err := r.gapiToKumaRule(ctx, mesh, route, rule)
		if err != nil {
			return nil, nil, errors.Wrap(err, "couldn't convert HTTPRoute to Kuma GatewayRoute")
		}

		if ruleCondition != nil {
			resolvedRefFalse = ruleCondition
		}

		if kumaRule != nil {
			rules = append(rules, kumaRule)
		}
	}

	conditions := []kube_meta.Condition{
		// TODO: reflect the true state from the actual gateway of this
		// route
		{
			Type:   string(gatewayapi.RouteConditionAccepted),
			Status: kube_meta.ConditionTrue,
			Reason: string(gatewayapi.RouteReasonAccepted),
		},
	}

	resolvedRefCondition := kube_meta.Condition{
		Type:   string(gatewayapi.RouteConditionResolvedRefs),
		Status: kube_meta.ConditionTrue,
		Reason: string(gatewayapi.RouteReasonResolvedRefs),
	}
	if resolvedRefFalse != nil {
		resolvedRefCondition.Status = kube_meta.ConditionFalse
		resolvedRefCondition.Reason = resolvedRefFalse.Reason
		resolvedRefCondition.Message = resolvedRefFalse.Message
	}
	conditions = append(conditions, resolvedRefCondition)

	var routeConf *mesh_proto.MeshGatewayRoute_Conf
	if len(rules) > 0 {
		routeConf = &mesh_proto.MeshGatewayRoute_Conf{
			Route: &mesh_proto.MeshGatewayRoute_Conf_Http{
				Http: &mesh_proto.MeshGatewayRoute_HttpRoute{
					Hostnames: hostnames,
					Rules:     rules,
				},
			},
		}
	}

	return routeConf, conditions, nil
}

func k8sToKumaHeader(header gatewayapi.HTTPHeader) *mesh_proto.MeshGatewayRoute_HttpRoute_Filter_HeaderFilter_Header {
	return &mesh_proto.MeshGatewayRoute_HttpRoute_Filter_HeaderFilter_Header{
		// note that our resources disallow uppercase letters in header names
		Name:  strings.ToLower(string(header.Name)),
		Value: header.Value,
	}
}

type ResolvedRefsConditionFalse struct {
	Reason  string
	Message string
}

func (r *HTTPRouteReconciler) uncheckedGapiToKumaRef(
	ctx context.Context, mesh string, objectNamespace string, ref gatewayapi.BackendObjectReference,
) (map[string]string, *ResolvedRefsConditionFalse, error) {
	unresolvedBackendTags := map[string]string{
		mesh_proto.ServiceTag: gateway.UnresolvedBackendServiceTag,
	}

	policyRef := policy.PolicyReferenceBackend(policy.FromHTTPRouteIn(objectNamespace), ref)

	gk := policyRef.GroupKindReferredTo()
	namespacedName := policyRef.NamespacedNameReferredTo()

	switch {
	case gk.Kind == "Service" && gk.Group == "":
		// References to Services are required by GAPI to include a port
		port := int32(*ref.Port)

		svc := &kube_core.Service{}
		if err := r.Client.Get(ctx, namespacedName, svc); err != nil {
			if kube_apierrs.IsNotFound(err) {
				return unresolvedBackendTags,
					&ResolvedRefsConditionFalse{
						Reason:  string(gatewayapi.RouteReasonBackendNotFound),
						Message: fmt.Sprintf("backend reference references a non-existent Service %q", namespacedName.String()),
					},
					nil
			}
			return nil, nil, err
		}

		return map[string]string{
			mesh_proto.ServiceTag: k8s_util.ServiceTag(kube_client.ObjectKeyFromObject(svc), &port),
		}, nil, nil
	case gk.Kind == "ExternalService" && gk.Group == mesh_k8s.GroupVersion.Group:
		resource := core_mesh.NewExternalServiceResource()
		if err := r.ResourceManager.Get(ctx, resource, store.GetByKey(namespacedName.Name, mesh)); err != nil {
			if store.IsResourceNotFound(err) {
				return unresolvedBackendTags,
					&ResolvedRefsConditionFalse{
						Reason:  string(gatewayapi.RouteReasonBackendNotFound),
						Message: fmt.Sprintf("backend reference references a non-existent ExternalService %q", namespacedName.Name),
					},
					nil
			}
			return nil, nil, err
		}

		return map[string]string{
			mesh_proto.ServiceTag: resource.Spec.GetService(),
		}, nil, nil
	}

	return unresolvedBackendTags,
		&ResolvedRefsConditionFalse{
			Reason:  string(gatewayapi.RouteReasonInvalidKind),
			Message: "backend reference must be Service or externalservice.kuma.io",
		},
		nil
}

// gapiToKumaRef checks a reference and tries to resolve if it's supported by
// Kuma. It returns a condition with Reason/Message if it fails or an error for
// unexpected errors.
func (r *HTTPRouteReconciler) gapiToKumaRef(
	ctx context.Context, mesh string, objectNamespace string, ref gatewayapi.BackendObjectReference,
) (map[string]string, *ResolvedRefsConditionFalse, error) {
	unresolvedBackendTags := map[string]string{
		mesh_proto.ServiceTag: gateway.UnresolvedBackendServiceTag,
	}

	policyRef := policy.PolicyReferenceBackend(policy.FromHTTPRouteIn(objectNamespace), ref)

	gk := policyRef.GroupKindReferredTo()
	namespacedName := policyRef.NamespacedNameReferredTo()

	if permitted, err := policy.IsReferencePermitted(ctx, r.Client, policyRef); err != nil {
		return nil, nil, errors.Wrap(err, "couldn't determine if backend reference is permitted")
	} else if !permitted {
		return unresolvedBackendTags,
			&ResolvedRefsConditionFalse{
				Reason:  string(gatewayapi.RouteReasonRefNotPermitted),
				Message: fmt.Sprintf("reference to %s %q not permitted by any ReferencePolicy", gk, namespacedName),
			},
			nil
	}

	return r.uncheckedGapiToKumaRef(ctx, mesh, objectNamespace, ref)
}

func gapiToKumaMatch(match gatewayapi.HTTPRouteMatch) (*mesh_proto.MeshGatewayRoute_HttpRoute_Match, error) {
	kumaMatch := &mesh_proto.MeshGatewayRoute_HttpRoute_Match{}

	if m := match.Method; m != nil {
		if kumaMethod, ok := mesh_proto.HttpMethod_value[string(*m)]; ok && kumaMethod != int32(mesh_proto.HttpMethod_NONE) {
			kumaMatch.Method = mesh_proto.HttpMethod(kumaMethod)
		} else if *m != "" {
			return nil, fmt.Errorf("unexpected HTTP method %s", *m)
		}
	}

	if p := match.Path; p != nil {
		path := &mesh_proto.MeshGatewayRoute_HttpRoute_Match_Path{
			Value: *p.Value,
		}

		switch *p.Type {
		case gatewayapi.PathMatchExact:
			path.Match = mesh_proto.MeshGatewayRoute_HttpRoute_Match_Path_EXACT
		case gatewayapi.PathMatchPathPrefix:
			path.Match = mesh_proto.MeshGatewayRoute_HttpRoute_Match_Path_PREFIX
		case gatewayapi.PathMatchRegularExpression:
			path.Match = mesh_proto.MeshGatewayRoute_HttpRoute_Match_Path_REGEX
		}

		kumaMatch.Path = path
	}

	for _, header := range match.Headers {
		kumaHeader := &mesh_proto.MeshGatewayRoute_HttpRoute_Match_Header{
			// note that our resources disallow uppercase letters in header names
			Name:  strings.ToLower(string(header.Name)),
			Value: header.Value,
		}

		switch *header.Type {
		case gatewayapi.HeaderMatchExact:
			kumaHeader.Match = mesh_proto.MeshGatewayRoute_HttpRoute_Match_Header_EXACT
		case gatewayapi.HeaderMatchRegularExpression:
			kumaHeader.Match = mesh_proto.MeshGatewayRoute_HttpRoute_Match_Header_REGEX
		}

		kumaMatch.Headers = append(kumaMatch.Headers, kumaHeader)
	}

	for _, query := range match.QueryParams {
		kumaQuery := &mesh_proto.MeshGatewayRoute_HttpRoute_Match_Query{
			Name:  string(query.Name),
			Value: query.Value,
		}

		switch *query.Type {
		case gatewayapi.QueryParamMatchExact:
			kumaQuery.Match = mesh_proto.MeshGatewayRoute_HttpRoute_Match_Query_EXACT
		case gatewayapi.QueryParamMatchRegularExpression:
			kumaQuery.Match = mesh_proto.MeshGatewayRoute_HttpRoute_Match_Query_REGEX
		}

		kumaMatch.QueryParameters = append(kumaMatch.QueryParameters, kumaQuery)
	}

	return kumaMatch, nil
}

func pathRewriteToKuma(modifier gatewayapi.HTTPPathModifier) *mesh_proto.MeshGatewayRoute_HttpRoute_Filter_Rewrite {
	rewrite := mesh_proto.MeshGatewayRoute_HttpRoute_Filter_Rewrite{}

	switch modifier.Type {
	case gatewayapi.FullPathHTTPPathModifier:
		rewrite.Path = &mesh_proto.MeshGatewayRoute_HttpRoute_Filter_Rewrite_ReplaceFull{
			ReplaceFull: *modifier.ReplaceFullPath,
		}
	case gatewayapi.PrefixMatchHTTPPathModifier:
		rewrite.Path = &mesh_proto.MeshGatewayRoute_HttpRoute_Filter_Rewrite_ReplacePrefixMatch{
			ReplacePrefixMatch: *modifier.ReplacePrefixMatch,
		}
	}

	return &rewrite
}

func (r *HTTPRouteReconciler) gapiToKumaFilters(
	ctx context.Context, mesh string, namespace string, filter gatewayapi.HTTPRouteFilter,
) ([]*mesh_proto.MeshGatewayRoute_HttpRoute_Filter, *ResolvedRefsConditionFalse, error) {
	var kumaFilters []*mesh_proto.MeshGatewayRoute_HttpRoute_Filter

	var condition *ResolvedRefsConditionFalse

	switch filter.Type {
	case gatewayapi.HTTPRouteFilterRequestHeaderModifier:
		filter := filter.RequestHeaderModifier

		var headerFilter mesh_proto.MeshGatewayRoute_HttpRoute_Filter_HeaderFilter

		for _, set := range filter.Set {
			headerFilter.Set = append(headerFilter.Set, k8sToKumaHeader(set))
		}

		for _, add := range filter.Add {
			headerFilter.Add = append(headerFilter.Add, k8sToKumaHeader(add))
		}

		headerFilter.Remove = filter.Remove

		kumaFilters = append(kumaFilters, &mesh_proto.MeshGatewayRoute_HttpRoute_Filter{
			Filter: &mesh_proto.MeshGatewayRoute_HttpRoute_Filter_RequestHeader{
				RequestHeader: &headerFilter,
			},
		})
	case gatewayapi.HTTPRouteFilterRequestMirror:
		filter := filter.RequestMirror

		// For mirrors we skip unresolved refs
		destinationRef, refCondition, err := r.gapiToKumaRef(ctx, mesh, namespace, filter.BackendRef)
		if err != nil {
			return nil, nil, err
		}
		if refCondition != nil {
			condition = refCondition
		}

		mirror := mesh_proto.MeshGatewayRoute_HttpRoute_Filter_Mirror{
			Backend: &mesh_proto.MeshGatewayRoute_Backend{
				Destination: destinationRef,
			},
			Percentage: util_proto.Double(100),
		}

		kumaFilters = append(kumaFilters, &mesh_proto.MeshGatewayRoute_HttpRoute_Filter{
			Filter: &mesh_proto.MeshGatewayRoute_HttpRoute_Filter_Mirror_{
				Mirror: &mirror,
			},
		})
	case gatewayapi.HTTPRouteFilterRequestRedirect:
		filter := filter.RequestRedirect

		redirect := mesh_proto.MeshGatewayRoute_HttpRoute_Filter_Redirect{}

		if s := filter.Scheme; s != nil {
			redirect.Scheme = *s
		}

		if h := filter.Hostname; h != nil {
			redirect.Hostname = string(*h)
		}

		if p := filter.Port; p != nil {
			redirect.Port = uint32(*p)
		}

		if sc := filter.StatusCode; sc != nil {
			redirect.StatusCode = uint32(*sc)
		}

		if p := filter.Path; p != nil {
			redirect.Path = pathRewriteToKuma(*p)
		}

		kumaFilters = append(kumaFilters, &mesh_proto.MeshGatewayRoute_HttpRoute_Filter{
			Filter: &mesh_proto.MeshGatewayRoute_HttpRoute_Filter_Redirect_{
				Redirect: &redirect,
			},
		})
	case gatewayapi.HTTPRouteFilterURLRewrite:
		filter := filter.URLRewrite

		if filter.Hostname != nil {
			var requestHeader mesh_proto.MeshGatewayRoute_HttpRoute_Filter_HeaderFilter
			requestHeader.Set = append(requestHeader.Set, &mesh_proto.MeshGatewayRoute_HttpRoute_Filter_HeaderFilter_Header{
				Name:  "Host",
				Value: string(*filter.Hostname),
			})
			kumaFilters = append(kumaFilters, &mesh_proto.MeshGatewayRoute_HttpRoute_Filter{
				Filter: &mesh_proto.MeshGatewayRoute_HttpRoute_Filter_RequestHeader{
					RequestHeader: &requestHeader,
				},
			})
		}

		if p := filter.Path; p != nil {
			filter := mesh_proto.MeshGatewayRoute_HttpRoute_Filter{
				Filter: &mesh_proto.MeshGatewayRoute_HttpRoute_Filter_Rewrite_{
					Rewrite: pathRewriteToKuma(*p),
				},
			}
			kumaFilters = append(kumaFilters, &filter)
		}
	case gatewayapi.HTTPRouteFilterResponseHeaderModifier:
		filter := filter.ResponseHeaderModifier

		var headerFilter mesh_proto.MeshGatewayRoute_HttpRoute_Filter_HeaderFilter

		for _, set := range filter.Set {
			headerFilter.Set = append(headerFilter.Set, k8sToKumaHeader(set))
		}

		for _, add := range filter.Add {
			headerFilter.Add = append(headerFilter.Add, k8sToKumaHeader(add))
		}

		headerFilter.Remove = filter.Remove

		kumaFilters = append(kumaFilters, &mesh_proto.MeshGatewayRoute_HttpRoute_Filter{
			Filter: &mesh_proto.MeshGatewayRoute_HttpRoute_Filter_ResponseHeader{
				ResponseHeader: &headerFilter,
			},
		})
	default:
		return nil, nil, fmt.Errorf("unsupported filter type %q", filter.Type)
	}

	return kumaFilters, condition, nil
}
