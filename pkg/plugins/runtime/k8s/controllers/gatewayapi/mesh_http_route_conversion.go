package gatewayapi

import (
	"context"
	"strings"

	"golang.org/x/exp/slices"
	kube_core "k8s.io/api/core/v1"
	kube_apimeta "k8s.io/apimachinery/pkg/api/meta"
	kube_meta "k8s.io/apimachinery/pkg/apis/meta/v1"
	kube_client "sigs.k8s.io/controller-runtime/pkg/client"
	gatewayapi_v1 "sigs.k8s.io/gateway-api/apis/v1"
	gatewayapi "sigs.k8s.io/gateway-api/apis/v1beta1"

	common_api "github.com/kumahq/kuma/api/common/v1alpha1"
	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
	core_model "github.com/kumahq/kuma/pkg/core/resources/model"
	"github.com/kumahq/kuma/pkg/plugins/policies/meshhttproute/api/v1alpha1"
	"github.com/kumahq/kuma/pkg/plugins/runtime/k8s/controllers/gatewayapi/attachment"
	k8s_util "github.com/kumahq/kuma/pkg/plugins/runtime/k8s/util"
	"github.com/kumahq/kuma/pkg/util/pointer"
)

func (r *HTTPRouteReconciler) gapiToMeshRules(
	ctx context.Context,
	mesh string,
	route *gatewayapi.HTTPRoute,
	parentRefAttachmentKind attachment.Kind,
) ([]v1alpha1.Rule, []kube_meta.Condition, error) {
	var rules []v1alpha1.Rule
	var conditions []kube_meta.Condition

	for _, rule := range route.Spec.Rules {
		kumaRule, ruleConditions, err := r.gapiToKumaMeshRule(ctx, mesh, route, rule, parentRefAttachmentKind)
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

	return rules, prepareConditions(conditions), nil
}

func (r *HTTPRouteReconciler) gapiServiceToMeshRoute(
	route *gatewayapi.HTTPRoute,
	rules []v1alpha1.Rule,
	svc *kube_core.Service,
	refPort *gatewayapi_v1.PortNumber,
) core_model.ResourceSpec {
	// consumer route
	targetRef := common_api.TargetRef{
		Kind: common_api.MeshSubset,
		Tags: map[string]string{
			mesh_proto.KubeNamespaceTag: route.Namespace,
		},
	}

	// producer route
	if route.Namespace == svc.GetNamespace() {
		targetRef = common_api.TargetRef{
			Kind: common_api.Mesh,
		}
	}

	var tos []v1alpha1.To

	var ports []int32
	if refPort != nil {
		ports = []int32{int32(*refPort)}
	} else {
		for _, port := range svc.Spec.Ports {
			ports = append(ports, port.Port)
		}
	}

	for _, port := range ports {
		serviceName := k8s_util.ServiceTag(
			kube_client.ObjectKeyFromObject(svc),
			pointer.To(port),
		)

		tos = append(tos, v1alpha1.To{
			TargetRef: common_api.TargetRef{
				Kind: common_api.MeshService,
				Name: serviceName,
			},
			Rules: rules,
		})
	}

	return &v1alpha1.MeshHTTPRoute{
		TargetRef: targetRef,
		To:        tos,
	}
}

// During conversion of GatewayAPI's `backendRefs` to our own `backendRefs`, we need
// to consider the scope of the referenced resource.
//   - Kubernetes Services: These have cluster-wide scope. When referencing a Kubernetes
//     Service, our `backendRef` should be restricted to the mesh service within the same
//     zone.
//   - Future Support: In the future, we might support Multi-Cluster Services (MCS)
//     (https://gateway-api.sigs.k8s.io/geps/gep-1748/). This will allow targeting
//     services across multiple zones.
func kumaLocalTargetRef(name, zone string) common_api.TargetRef {
	return common_api.TargetRef{
		Kind: common_api.MeshServiceSubset,
		Name: name,
		Tags: map[string]string{mesh_proto.ZoneTag: zone},
	}
}

func (r *HTTPRouteReconciler) gapiToKumaMeshRule(
	ctx context.Context,
	mesh string,
	route *gatewayapi.HTTPRoute,
	rule gatewayapi.HTTPRouteRule,
	parentRefAttachmentKind attachment.Kind,
) (v1alpha1.Rule, []kube_meta.Condition, error) {
	var conditions []kube_meta.Condition

	var matches []v1alpha1.Match
	var filters []v1alpha1.Filter
	var backendRefs []common_api.BackendRef

	for _, gapiMatch := range rule.Matches {
		match, ok := r.gapiToKumaMeshMatch(gapiMatch)
		if !ok {
			continue
			// TODO set condition
		}
		matches = append(matches, match)
	}

	for _, gapiFilter := range rule.Filters {
		filter, filterConditions, ok := r.gapiToKumaMeshFilter(ctx, mesh, route.Namespace, gapiFilter, parentRefAttachmentKind)
		if !ok {
			// TODO use err
			continue
		}

		filterConditions = slices.DeleteFunc(
			filterConditions,
			func(cond kube_meta.Condition) bool {
				return cond.Type == string(gatewayapi.RouteConditionResolvedRefs) &&
					cond.Reason == string(gatewayapi.RouteReasonRefNotPermitted)
			},
		)

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
		ref, refCondition, err := r.gapiToKumaRef(ctx, mesh, route.Namespace, gapiBackendRef.BackendObjectReference, parentRefAttachmentKind)
		if err != nil {
			return v1alpha1.Rule{}, nil, err
		}

		refCondition.AddIfFalseAndNotPresent(&conditions)

		backendRefs = append(backendRefs, common_api.BackendRef{
			TargetRef: kumaLocalTargetRef(ref[mesh_proto.ServiceTag], r.Zone),
			Weight:    pointer.To(uint(*gapiBackendRef.Weight)),
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

func (r *HTTPRouteReconciler) gapiToKumaMeshMatch(gapiMatch gatewayapi.HTTPRouteMatch) (v1alpha1.Match, bool) {
	var match v1alpha1.Match

	match.Path = &v1alpha1.PathMatch{
		Type:  v1alpha1.PathMatchType(*gapiMatch.Path.Type),
		Value: *gapiMatch.Path.Value,
	}

	// Matches based on a URL path prefix split by `/`. Matching is
	// case-sensitive and done on a path element by element basis. A
	// path element refers to the list of labels in the path split by
	// the `/` separator. When specified, a trailing `/` is ignored.
	//
	// For example, the paths `/abc`, `/abc/`, and `/abc/def` would all match
	// the prefix `/abc`, but the path `/abcd` would not.
	//
	// ref. https://github.com/kubernetes-sigs/gateway-api/blob/50091d071226d4ab2dbdb115ae65e27cf3fd5b85/apis/v1/httproute_types.go#L357-L367
	//
	// Necessary as MehHTTPRoute validator won't allow value with trailing `/`
	if match.Path.Type == v1alpha1.PathPrefix && match.Path.Value != "/" {
		match.Path.Value = strings.TrimSuffix(match.Path.Value, "/")
	}

	for _, gapiHeader := range gapiMatch.Headers {
		header := common_api.HeaderMatch{
			Type: pointer.To(common_api.HeaderMatchType(*gapiHeader.Type)),
			// note that our resources disallow uppercase letters in header names
			Name:  common_api.HeaderName(strings.ToLower(string(gapiHeader.Name))),
			Value: common_api.HeaderValue(gapiHeader.Value),
		}
		match.Headers = append(match.Headers, header)
	}

	for _, gapiParam := range gapiMatch.QueryParams {
		var param v1alpha1.QueryParamsMatch
		switch *gapiParam.Type {
		case gatewayapi_v1.QueryParamMatchExact:
			param = v1alpha1.QueryParamsMatch{
				Type:  v1alpha1.ExactQueryMatch,
				Name:  string(gapiParam.Name),
				Value: gapiParam.Value,
			}
		case gatewayapi_v1.QueryParamMatchRegularExpression:
			param = v1alpha1.QueryParamsMatch{
				Type:  v1alpha1.RegularExpressionQueryMatch,
				Name:  string(gapiParam.Name),
				Value: gapiParam.Value,
			}
		default:
			return v1alpha1.Match{}, false
		}
		match.QueryParams = append(match.QueryParams, param)
	}

	if gapiMatch.Method != nil {
		match.Method = (*v1alpha1.Method)(gapiMatch.Method)
	}

	return match, true
}

func fromGAPIHeaders(gapiHeaders []gatewayapi.HTTPHeader) []v1alpha1.HeaderKeyValue {
	var headers []v1alpha1.HeaderKeyValue
	for _, header := range gapiHeaders {
		headers = append(headers, v1alpha1.HeaderKeyValue{
			// note that our resources disallow uppercase letters in header names
			Name:  common_api.HeaderName(strings.ToLower(string(header.Name))),
			Value: common_api.HeaderValue(header.Value),
		})
	}
	return headers
}

func fromGAPIPath(gapiPath gatewayapi.HTTPPathModifier) (v1alpha1.PathRewrite, bool) {
	switch gapiPath.Type {
	case gatewayapi_v1.FullPathHTTPPathModifier:
		return v1alpha1.PathRewrite{
			Type:            v1alpha1.ReplaceFullPathType,
			ReplaceFullPath: gapiPath.ReplaceFullPath,
		}, true
	case gatewayapi_v1.PrefixMatchHTTPPathModifier:
		return v1alpha1.PathRewrite{
			Type:               v1alpha1.ReplacePrefixMatchType,
			ReplacePrefixMatch: gapiPath.ReplacePrefixMatch,
		}, true
	default:
		return v1alpha1.PathRewrite{}, false
	}
}

func (r *HTTPRouteReconciler) gapiToKumaMeshFilter(
	ctx context.Context,
	mesh, routeNamespace string,
	gapiFilter gatewayapi.HTTPRouteFilter,
	refAttachmentKind attachment.Kind,
) (v1alpha1.Filter, []kube_meta.Condition, bool) {
	switch gapiFilter.Type {
	case gatewayapi_v1.HTTPRouteFilterRequestHeaderModifier:
		modifier := gapiFilter.RequestHeaderModifier
		return v1alpha1.Filter{
			Type: v1alpha1.RequestHeaderModifierType,
			RequestHeaderModifier: &v1alpha1.HeaderModifier{
				Add:    fromGAPIHeaders(modifier.Add),
				Set:    fromGAPIHeaders(modifier.Set),
				Remove: modifier.Remove,
			},
		}, nil, true
	case gatewayapi_v1.HTTPRouteFilterResponseHeaderModifier:
		modifier := gapiFilter.ResponseHeaderModifier
		return v1alpha1.Filter{
			Type: v1alpha1.ResponseHeaderModifierType,
			ResponseHeaderModifier: &v1alpha1.HeaderModifier{
				Add:    fromGAPIHeaders(modifier.Add),
				Set:    fromGAPIHeaders(modifier.Set),
				Remove: modifier.Remove,
			},
		}, nil, true
	case gatewayapi_v1.HTTPRouteFilterRequestRedirect:
		redirect := gapiFilter.RequestRedirect

		var path *v1alpha1.PathRewrite
		if gapiPath := redirect.Path; gapiPath != nil {
			meshPath, ok := fromGAPIPath(*gapiPath)
			if !ok {
				return v1alpha1.Filter{}, nil, false
			}
			path = &meshPath
		}

		port := (*v1alpha1.PortNumber)(redirect.Port)
		if redirect.Scheme != nil && redirect.Port == nil {
			// See https://github.com/kubernetes-sigs/gateway-api/pull/1880
			// this would have been a breaking change for MeshGateway, so handle
			// it here.
			switch *redirect.Scheme {
			case "http":
				port = (*v1alpha1.PortNumber)(pointer.To(int32(80)))
			case "https":
				port = (*v1alpha1.PortNumber)(pointer.To(int32(443)))
			}
		}

		return v1alpha1.Filter{
			Type: v1alpha1.RequestRedirectType,
			RequestRedirect: &v1alpha1.RequestRedirect{
				Scheme:     redirect.Scheme,
				Hostname:   (*v1alpha1.PreciseHostname)(redirect.Hostname),
				Path:       path,
				Port:       port,
				StatusCode: redirect.StatusCode,
			},
		}, nil, true
	case gatewayapi_v1.HTTPRouteFilterURLRewrite:
		rewrite := gapiFilter.URLRewrite

		var path *v1alpha1.PathRewrite
		if gapiPath := rewrite.Path; gapiPath != nil {
			meshPath, ok := fromGAPIPath(*gapiPath)
			if !ok {
				return v1alpha1.Filter{}, nil, false
			}
			path = &meshPath
		}

		return v1alpha1.Filter{
			Type: v1alpha1.URLRewriteType,
			URLRewrite: &v1alpha1.URLRewrite{
				Hostname: (*v1alpha1.PreciseHostname)(rewrite.Hostname),
				Path:     path,
			},
		}, nil, true
	case gatewayapi_v1.HTTPRouteFilterRequestMirror:
		mirror := gapiFilter.RequestMirror

		ref, refCondition, err := r.gapiToKumaRef(ctx, mesh, routeNamespace, mirror.BackendRef, refAttachmentKind)
		if err != nil {
			return v1alpha1.Filter{}, nil, false
		}

		var conditions []kube_meta.Condition
		refCondition.AddIfFalseAndNotPresent(&conditions)

		return v1alpha1.Filter{
			Type: v1alpha1.RequestMirrorType,
			RequestMirror: &v1alpha1.RequestMirror{
				BackendRef: common_api.BackendRef{
					TargetRef: kumaLocalTargetRef(ref[mesh_proto.ServiceTag], r.Zone),
				},
			},
		}, conditions, true
	default:
		return v1alpha1.Filter{}, nil, false
	}
}
