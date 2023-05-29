package gatewayapi

import (
	"context"
	"fmt"
	"strings"

	kube_core "k8s.io/api/core/v1"
	kube_types "k8s.io/apimachinery/pkg/types"
	kube_client "sigs.k8s.io/controller-runtime/pkg/client"
	gatewayapi "sigs.k8s.io/gateway-api/apis/v1beta1"

	common_api "github.com/kumahq/kuma/api/common/v1alpha1"
	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
	core_model "github.com/kumahq/kuma/pkg/core/resources/model"
	"github.com/kumahq/kuma/pkg/plugins/policies/meshhttproute/api/v1alpha1"
	"github.com/kumahq/kuma/pkg/plugins/runtime/k8s/controllers"
	k8s_util "github.com/kumahq/kuma/pkg/plugins/runtime/k8s/util"
	"github.com/kumahq/kuma/pkg/util/pointer"
)

type ServiceAndPorts struct {
	Name  kube_types.NamespacedName
	Ports []int32
}

func serviceAndPorts(svc *kube_core.Service, port *gatewayapi.PortNumber) ServiceAndPorts {
	var ports []int32
	for _, port := range svc.Spec.Ports {
		ports = append(ports, port.Port)
	}
	if port != nil {
		ports = []int32{int32(*port)}
	}
	return ServiceAndPorts{
		Name:  kube_client.ObjectKeyFromObject(svc),
		Ports: ports,
	}
}

func (r *HTTPRouteReconciler) gapiToMeshRouteSpecs(
	ctx context.Context, mesh string, route *gatewayapi.HTTPRoute, svcs []ServiceAndPorts,
) (map[string]core_model.ResourceSpec, error) {
	var rules []v1alpha1.Rule
	for _, rule := range route.Spec.Rules {
		kumaRule, err := r.gapiToKumaMeshRule(ctx, mesh, route, rule)
		if err != nil {
			return nil, err
		}

		rules = append(rules, kumaRule)
	}

	routes := map[string]core_model.ResourceSpec{}

	for _, svcRef := range svcs {
		// consumer route
		targetRef := common_api.TargetRef{
			Kind: common_api.MeshSubset,
			Tags: map[string]string{
				controllers.KubeNamespaceTag: route.Namespace,
			},
		}
		// producer route
		if route.Namespace == svcRef.Name.Namespace {
			targetRef = common_api.TargetRef{
				Kind: common_api.Mesh,
			}
		}

		var tos []v1alpha1.To

		for _, port := range svcRef.Ports {
			p := port

			serviceName := k8s_util.ServiceTag(
				svcRef.Name,
				&p,
			)

			tos = append(tos, v1alpha1.To{
				TargetRef: common_api.TargetRef{
					Kind: common_api.MeshService,
					Name: serviceName,
				},
				Rules: rules,
			})
		}

		routeSubName := fmt.Sprintf(
			"%s-%s-%s.%s",
			route.Name, route.Namespace, svcRef.Name.Name, svcRef.Name.Namespace,
		)
		routes[routeSubName] = &v1alpha1.MeshHTTPRoute{
			TargetRef: targetRef,
			To:        tos,
		}
	}

	return routes, nil
}

func (r *HTTPRouteReconciler) gapiToKumaMeshRule(
	ctx context.Context, mesh string, route *gatewayapi.HTTPRoute, rule gatewayapi.HTTPRouteRule,
) (v1alpha1.Rule, error) {
	var matches []v1alpha1.Match
	var filters []v1alpha1.Filter
	var backendRefs []v1alpha1.BackendRef

	for _, gapiMatch := range rule.Matches {
		match, ok := r.gapiToKumaMeshMatch(gapiMatch)
		if !ok {
			continue
			// TODO set condition
		}
		matches = append(matches, match)
	}

	for _, gapiFilter := range rule.Filters {
		filter, ok := r.gapiToKumaMeshFilter(ctx, mesh, route.Namespace, gapiFilter)
		if !ok {
			continue
			// TODO set condition
		}
		filters = append(filters, filter)
	}

	for _, gapiBackendRef := range rule.BackendRefs {
		// TODO condition
		// ReferenceGrants don't need to be taken into account for Mesh
		ref, _, err := r.uncheckedGapiToKumaRef(ctx, mesh, route.Namespace, gapiBackendRef.BackendObjectReference)
		if err != nil {
			return v1alpha1.Rule{}, err
		}

		backendRefs = append(backendRefs, v1alpha1.BackendRef{
			TargetRef: common_api.TargetRef{
				Kind: common_api.MeshService,
				Name: ref[mesh_proto.ServiceTag],
			},
			Weight: pointer.To(uint(*gapiBackendRef.Weight)),
		})
	}

	return v1alpha1.Rule{
		Matches: matches,
		Default: v1alpha1.RuleConf{
			Filters:     &filters,
			BackendRefs: &backendRefs,
		},
	}, nil
}

func (r *HTTPRouteReconciler) gapiToKumaMeshMatch(gapiMatch gatewayapi.HTTPRouteMatch) (v1alpha1.Match, bool) {
	var match v1alpha1.Match

	match.Path = &v1alpha1.PathMatch{
		Type:  v1alpha1.PathMatchType(*gapiMatch.Path.Type),
		Value: *gapiMatch.Path.Value,
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
		case gatewayapi.QueryParamMatchExact:
			param = v1alpha1.QueryParamsMatch{
				Type:  v1alpha1.ExactQueryMatch,
				Value: gapiParam.Value,
			}
		case gatewayapi.QueryParamMatchRegularExpression:
			param = v1alpha1.QueryParamsMatch{
				Type:  v1alpha1.RegularExpressionQueryMatch,
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
	case gatewayapi.FullPathHTTPPathModifier:
		return v1alpha1.PathRewrite{
			Type:            v1alpha1.ReplaceFullPathType,
			ReplaceFullPath: gapiPath.ReplaceFullPath,
		}, true
	case gatewayapi.PrefixMatchHTTPPathModifier:
		return v1alpha1.PathRewrite{
			Type:               v1alpha1.ReplacePrefixMatchType,
			ReplacePrefixMatch: gapiPath.ReplacePrefixMatch,
		}, true
	default:
		return v1alpha1.PathRewrite{}, false
	}
}

func (r *HTTPRouteReconciler) gapiToKumaMeshFilter(ctx context.Context, mesh, routeNamespace string, gapiFilter gatewayapi.HTTPRouteFilter) (v1alpha1.Filter, bool) {
	switch gapiFilter.Type {
	case gatewayapi.HTTPRouteFilterRequestHeaderModifier:
		modifier := gapiFilter.RequestHeaderModifier
		return v1alpha1.Filter{
			Type: v1alpha1.RequestHeaderModifierType,
			RequestHeaderModifier: &v1alpha1.HeaderModifier{
				Add:    fromGAPIHeaders(modifier.Add),
				Set:    fromGAPIHeaders(modifier.Set),
				Remove: modifier.Remove,
			},
		}, true
	case gatewayapi.HTTPRouteFilterResponseHeaderModifier:
		modifier := gapiFilter.ResponseHeaderModifier
		return v1alpha1.Filter{
			Type: v1alpha1.ResponseHeaderModifierType,
			ResponseHeaderModifier: &v1alpha1.HeaderModifier{
				Add:    fromGAPIHeaders(modifier.Add),
				Set:    fromGAPIHeaders(modifier.Set),
				Remove: modifier.Remove,
			},
		}, true
	case gatewayapi.HTTPRouteFilterRequestRedirect:
		redirect := gapiFilter.RequestRedirect

		var path *v1alpha1.PathRewrite
		if gapiPath := redirect.Path; gapiPath != nil {
			meshPath, ok := fromGAPIPath(*gapiPath)
			if !ok {
				return v1alpha1.Filter{}, false
			}
			path = &meshPath
		}

		return v1alpha1.Filter{
			Type: v1alpha1.RequestRedirectType,
			RequestRedirect: &v1alpha1.RequestRedirect{
				Scheme:     redirect.Scheme,
				Hostname:   (*v1alpha1.PreciseHostname)(redirect.Hostname),
				Path:       path,
				Port:       (*v1alpha1.PortNumber)(redirect.Port),
				StatusCode: redirect.StatusCode,
			},
		}, true
	case gatewayapi.HTTPRouteFilterURLRewrite:
		rewrite := gapiFilter.URLRewrite

		var path *v1alpha1.PathRewrite
		if gapiPath := rewrite.Path; gapiPath != nil {
			meshPath, ok := fromGAPIPath(*gapiPath)
			if !ok {
				return v1alpha1.Filter{}, false
			}
			path = &meshPath
		}

		return v1alpha1.Filter{
			Type: v1alpha1.URLRewriteType,
			URLRewrite: &v1alpha1.URLRewrite{
				Hostname: (*v1alpha1.PreciseHostname)(rewrite.Hostname),
				Path:     path,
			},
		}, true
	case gatewayapi.HTTPRouteFilterRequestMirror:
		mirror := gapiFilter.RequestMirror

		// TODO conditions
		ref, _, err := r.uncheckedGapiToKumaRef(ctx, mesh, routeNamespace, mirror.BackendRef)
		if err != nil {
			return v1alpha1.Filter{}, false
		}

		return v1alpha1.Filter{
			Type: v1alpha1.RequestMirrorType,
			RequestMirror: &v1alpha1.RequestMirror{
				BackendRef: common_api.TargetRef{
					Kind: common_api.MeshService,
					Name: ref[mesh_proto.ServiceTag],
				},
			},
		}, true
	default:
		return v1alpha1.Filter{}, false
	}
}
