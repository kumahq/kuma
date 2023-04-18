package gatewayapi

import (
	"context"
	"fmt"

	"github.com/pkg/errors"
	kube_core "k8s.io/api/core/v1"
	kube_types "k8s.io/apimachinery/pkg/types"
	gatewayapi "sigs.k8s.io/gateway-api/apis/v1beta1"

	common_api "github.com/kumahq/kuma/api/common/v1alpha1"
	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
	"github.com/kumahq/kuma/pkg/plugins/policies/meshhttproute/api/v1alpha1"
	k8s_util "github.com/kumahq/kuma/pkg/plugins/runtime/k8s/util"
	"github.com/kumahq/kuma/pkg/util/pointer"
)

func (r *HTTPRouteReconciler) gapiToMeshRouteSpecs(
	ctx context.Context, mesh string, route *gatewayapi.HTTPRoute, svcs []kube_types.NamespacedName,
) (map[string]v1alpha1.MeshHTTPRoute, error) {
	routes := map[string]v1alpha1.MeshHTTPRoute{}

	for _, svcRef := range svcs {
		svc := kube_core.Service{}
		if err := r.Client.Get(ctx, svcRef, &svc); err != nil {
			return nil, errors.Wrap(err, "unable to get Service")
		}
		var ports []int32
		for _, port := range svc.Spec.Ports {
			ports = append(ports, port.Port)
		}
		for _, port := range ports {
			p := port

			rules, err := r.gapiToMeshRouteRules(ctx, mesh, route)
			if err != nil {
				return nil, err
			}
			serviceName := k8s_util.ServiceTag(
				svcRef,
				&p,
			)

			to := []v1alpha1.To{{
				TargetRef: common_api.TargetRef{
					Kind: common_api.MeshService,
					Name: serviceName,
				},
				Rules: rules,
			}}
			routeSubName := fmt.Sprintf("%s-%s-%d", svc.Name, svc.Namespace, port)
			routes[routeSubName] = v1alpha1.MeshHTTPRoute{
				TargetRef: common_api.TargetRef{
					Kind: common_api.Mesh,
				},
				To: to,
			}
		}
	}

	return routes, nil
}

func (r *HTTPRouteReconciler) gapiToMeshRouteRules(
	ctx context.Context, mesh string, route *gatewayapi.HTTPRoute,
) ([]v1alpha1.Rule, error) {
	var rules []v1alpha1.Rule
	for _, rule := range route.Spec.Rules {
		kumaRule, err := r.gapiToKumaMeshRule(ctx, mesh, route, rule)
		if err != nil {
			return nil, err
		}

		rules = append(rules, kumaRule)
	}

	return rules, nil
}

func (r *HTTPRouteReconciler) gapiToKumaMeshRule(
	ctx context.Context, mesh string, route *gatewayapi.HTTPRoute, rule gatewayapi.HTTPRouteRule,
) (v1alpha1.Rule, error) {
	var matches []v1alpha1.Match
	var backendRefs []v1alpha1.BackendRef

	for _, gapiMatch := range rule.Matches {
		matches = append(matches, r.gapiToKumaMeshMatch(gapiMatch))
	}

	for _, gapiBackendRef := range rule.BackendRefs {
		// TODO condition
		ref, _, err := r.gapiToKumaRef(ctx, mesh, route.Namespace, gapiBackendRef.BackendObjectReference)
		if err != nil {
			return v1alpha1.Rule{}, err
		}

		backendRefs = append(backendRefs, v1alpha1.BackendRef{
			TargetRef: common_api.TargetRef{
				Kind: common_api.MeshService,
				Name: ref[mesh_proto.ServiceTag],
			},
		})
	}

	return v1alpha1.Rule{
		Matches: matches,
		Default: v1alpha1.RuleConf{
			BackendRefs: &backendRefs,
		},
	}, nil
}

func (r *HTTPRouteReconciler) gapiToKumaMeshMatch(gapiMatch gatewayapi.HTTPRouteMatch) v1alpha1.Match {
	var match v1alpha1.Match

	match.Path = &v1alpha1.PathMatch{
		Type:  v1alpha1.PathMatchType(*gapiMatch.Path.Type),
		Value: *gapiMatch.Path.Value,
	}

	for _, gapiHeader := range gapiMatch.Headers {
		header := common_api.HeaderMatch{
			Type:  pointer.To(common_api.HeaderMatchType(*gapiHeader.Type)),
			Name:  common_api.HeaderName(gapiHeader.Name),
			Value: common_api.HeaderValue(gapiHeader.Value),
		}
		match.Headers = append(match.Headers, header)
	}

	return match
}
