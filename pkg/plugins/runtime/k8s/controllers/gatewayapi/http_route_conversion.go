package gatewayapi

import (
	"context"
	"fmt"

	"github.com/pkg/errors"
	kube_meta "k8s.io/apimachinery/pkg/apis/meta/v1"
	gatewayapi "sigs.k8s.io/gateway-api/apis/v1alpha2"

	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
	core_mesh "github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
	"github.com/kumahq/kuma/pkg/core/resources/store"
	util_proto "github.com/kumahq/kuma/pkg/util/proto"
)

func (r *HTTPRouteReconciler) gapiToKumaRule(
	ctx context.Context, mesh string, route *gatewayapi.HTTPRoute, rule gatewayapi.HTTPRouteRule,
) (mesh_proto.GatewayRoute_HttpRoute_Rule, *kube_meta.Condition, error) {
	var backends []*mesh_proto.GatewayRoute_Backend

	for _, backend := range rule.BackendRefs {
		ref := backend.BackendObjectReference

		destination, condition, err := r.gapiToKumaRef(ctx, mesh, route.Namespace, ref)
		if err != nil || condition != nil {
			return mesh_proto.GatewayRoute_HttpRoute_Rule{}, condition, err
		}

		backends = append(backends, &mesh_proto.GatewayRoute_Backend{
			// Weight has a default of 1
			Weight:      uint32(*backend.Weight),
			Destination: destination,
		})
	}

	var matches []*mesh_proto.GatewayRoute_HttpRoute_Match

	for _, match := range rule.Matches {
		kumaMatch, err := gapiToKumaMatch(match)
		if err != nil {
			return mesh_proto.GatewayRoute_HttpRoute_Rule{}, nil, errors.Wrap(err, "couldn't convert match")
		}

		matches = append(matches, kumaMatch)
	}

	var filters []*mesh_proto.GatewayRoute_HttpRoute_Filter

	for _, filter := range rule.Filters {
		kumaFilter, condition, err := r.gapiToKumaFilter(ctx, mesh, route.Namespace, filter)
		if err != nil || condition != nil {
			return mesh_proto.GatewayRoute_HttpRoute_Rule{}, condition, err
		}

		filters = append(filters, kumaFilter)
	}

	return mesh_proto.GatewayRoute_HttpRoute_Rule{
		Matches:  matches,
		Filters:  filters,
		Backends: backends,
	}, nil, nil
}

// gapiToKumaRouteConf converts the route into a route spec and returns any
// conditions that should be set on parent refs. These are the same across all
// Kuma parent refs. If a conf cannot be created, it returns a nil conf.
// It returns error only if an unexpected error has occurred. Issues related to the
// HTTPRoute spec are reflected in the Conditions.
func (r *HTTPRouteReconciler) gapiToKumaRouteConf(
	ctx context.Context, mesh string, route *gatewayapi.HTTPRoute,
) (*mesh_proto.GatewayRoute_Conf, []kube_meta.Condition, error) {
	var hostnames []string

	for _, hn := range route.Spec.Hostnames {
		hostnames = append(hostnames, string(hn))
	}

	var blockerCondition *kube_meta.Condition

	var rules []*mesh_proto.GatewayRoute_HttpRoute_Rule

	for _, rule := range route.Spec.Rules {
		kumaRule, condition, err := r.gapiToKumaRule(ctx, mesh, route, rule)
		if err != nil {
			return nil, nil, errors.Wrap(err, "couldn't convert HTTPRoute to Kuma GatewayRoute")
		}
		if condition != nil {
			blockerCondition = condition
			break
		}

		rules = append(rules, &kumaRule)
	}

	if blockerCondition != nil {
		// We complete the rest of the condition
		blockerCondition.LastTransitionTime = kube_meta.Now()
		blockerCondition.ObservedGeneration = route.GetGeneration()

		return nil, []kube_meta.Condition{
			*blockerCondition,
			routeCondition(
				route,
				gatewayapi.ConditionRouteAccepted,
				kube_meta.ConditionFalse,
				"ConversionFailed",
				fmt.Sprintf("Prevented by %s Condition", blockerCondition.Type),
			),
		}, nil
	}

	routeConf := mesh_proto.GatewayRoute_Conf{
		Route: &mesh_proto.GatewayRoute_Conf_Http{
			Http: &mesh_proto.GatewayRoute_HttpRoute{
				Hostnames: hostnames,
				Rules:     rules,
			},
		},
	}

	conditions := []kube_meta.Condition{
		routeCondition(route, gatewayapi.ConditionRouteResolvedRefs, kube_meta.ConditionTrue, string(gatewayapi.ConditionRouteResolvedRefs)),
		// TODO: reflect the true state from the actual gateway of this
		// route
		routeCondition(route, gatewayapi.ConditionRouteAccepted, kube_meta.ConditionTrue, string(gatewayapi.ConditionRouteAccepted)),
	}

	return &routeConf, conditions, nil
}

func k8sToKumaHeader(header gatewayapi.HTTPHeader) *mesh_proto.GatewayRoute_HttpRoute_Filter_RequestHeader_Header {
	return &mesh_proto.GatewayRoute_HttpRoute_Filter_RequestHeader_Header{
		Name:  string(header.Name),
		Value: header.Value,
	}
}

// gapiToKumaRef checks a reference and tries to resolve if it's supported by
// Kuma. It returns a condition with Reason/Message if it fails or an error for
// unexpected errors.
func (r *HTTPRouteReconciler) gapiToKumaRef(
	ctx context.Context, mesh string, objectNamespace string, ref gatewayapi.BackendObjectReference,
) (map[string]string, *kube_meta.Condition, error) {
	switch *ref.Kind {
	case "Service":
		if *ref.Group != "" {
			break
		}

		// References to Services are required by GAPI to include a port
		// TODO remove when https://github.com/kubernetes-sigs/gateway-api/pull/944
		// is released
		if ref.Port == nil {
			return nil,
				&kube_meta.Condition{
					Type:    string(gatewayapi.ConditionRouteResolvedRefs),
					Status:  kube_meta.ConditionFalse,
					Reason:  RefInvalid,
					Message: "backend reference must include port",
				},
				nil
		}
		// TODO resolve it

		namespace := objectNamespace
		if ref.Namespace != nil {
			namespace = string(*ref.Namespace)
		}

		return map[string]string{
			mesh_proto.ServiceTag: fmt.Sprintf("%s_%s_svc_%d", ref.Name, namespace, *ref.Port),
		}, nil, nil
	case "ExternalService":
		if *ref.Group != "kuma.io" {
			break
		}

		name := string(ref.Name)

		resource := core_mesh.NewExternalServiceResource()
		if err := r.ResourceManager.Get(ctx, resource, store.GetByKey(name, mesh)); err != nil {
			// TODO this shouldn't be a fatal error
			if store.IsResourceNotFound(err) {
				return nil,
					&kube_meta.Condition{
						Type:    string(gatewayapi.ConditionRouteResolvedRefs),
						Status:  kube_meta.ConditionFalse,
						Reason:  ObjectNotFound,
						Message: fmt.Sprintf("backend reference references a non-existent ExternalService %s", name),
					},
					nil
			}
			return nil, nil, err
		}

		service := resource.Spec.GetService()

		return map[string]string{
			mesh_proto.ServiceTag: service,
		}, nil, nil
	}

	return nil,
		&kube_meta.Condition{
			Type:    string(gatewayapi.ConditionRouteResolvedRefs),
			Status:  kube_meta.ConditionFalse,
			Reason:  ObjectTypeUnknownOrInvalid,
			Message: "backend reference must be Service or externalservice.kuma.io",
		},
		nil
}

func gapiToKumaMatch(match gatewayapi.HTTPRouteMatch) (*mesh_proto.GatewayRoute_HttpRoute_Match, error) {
	kumaMatch := &mesh_proto.GatewayRoute_HttpRoute_Match{}

	if m := match.Method; m != nil {
		if kumaMethod, ok := mesh_proto.HttpMethod_value[string(*m)]; ok {
			kumaMatch.Method = mesh_proto.HttpMethod(kumaMethod)
		} else if *m != "" {
			return nil, fmt.Errorf("unexpected HTTP method %s", *m)
		}
	}

	if p := match.Path; p != nil {
		path := &mesh_proto.GatewayRoute_HttpRoute_Match_Path{
			Value: *p.Value,
		}

		switch *p.Type {
		case gatewayapi.PathMatchExact:
			path.Match = mesh_proto.GatewayRoute_HttpRoute_Match_Path_EXACT
		case gatewayapi.PathMatchPathPrefix:
			path.Match = mesh_proto.GatewayRoute_HttpRoute_Match_Path_PREFIX
		case gatewayapi.PathMatchRegularExpression:
			path.Match = mesh_proto.GatewayRoute_HttpRoute_Match_Path_REGEX
		}

		kumaMatch.Path = path
	}

	for _, header := range match.Headers {
		kumaHeader := &mesh_proto.GatewayRoute_HttpRoute_Match_Header{
			Name:  string(header.Name),
			Value: header.Value,
		}

		switch *header.Type {
		case gatewayapi.HeaderMatchExact:
			kumaHeader.Match = mesh_proto.GatewayRoute_HttpRoute_Match_Header_EXACT
		case gatewayapi.HeaderMatchRegularExpression:
			kumaHeader.Match = mesh_proto.GatewayRoute_HttpRoute_Match_Header_REGEX
		}

		kumaMatch.Headers = append(kumaMatch.Headers, kumaHeader)
	}

	for _, query := range match.QueryParams {
		kumaQuery := &mesh_proto.GatewayRoute_HttpRoute_Match_Query{
			Name:  query.Name,
			Value: query.Value,
		}

		switch *query.Type {
		case gatewayapi.QueryParamMatchExact:
			kumaQuery.Match = mesh_proto.GatewayRoute_HttpRoute_Match_Query_EXACT
		case gatewayapi.QueryParamMatchRegularExpression:
			kumaQuery.Match = mesh_proto.GatewayRoute_HttpRoute_Match_Query_REGEX
		}

		kumaMatch.QueryParameters = append(kumaMatch.QueryParameters, kumaQuery)
	}

	return kumaMatch, nil
}

func (r *HTTPRouteReconciler) gapiToKumaFilter(
	ctx context.Context, mesh string, namespace string, filter gatewayapi.HTTPRouteFilter,
) (*mesh_proto.GatewayRoute_HttpRoute_Filter, *kube_meta.Condition, error) {
	var kumaFilter mesh_proto.GatewayRoute_HttpRoute_Filter

	switch filter.Type {
	case gatewayapi.HTTPRouteFilterRequestHeaderModifier:
		filter := filter.RequestHeaderModifier

		var kumaInnerFilter mesh_proto.GatewayRoute_HttpRoute_Filter_RequestHeader

		for _, set := range filter.Set {
			kumaInnerFilter.Set = append(kumaInnerFilter.Set, k8sToKumaHeader(set))
		}

		for _, add := range filter.Add {
			kumaInnerFilter.Add = append(kumaInnerFilter.Add, k8sToKumaHeader(add))
		}

		kumaInnerFilter.Remove = filter.Remove

		kumaFilter.Filter = &mesh_proto.GatewayRoute_HttpRoute_Filter_RequestHeader_{
			RequestHeader: &kumaInnerFilter,
		}
	case gatewayapi.HTTPRouteFilterRequestMirror:
		filter := filter.RequestMirror

		destinationRef, condition, err := r.gapiToKumaRef(ctx, mesh, namespace, filter.BackendRef)
		if err != nil || condition != nil {
			return nil, condition, err
		}

		kumaInnerFilter := mesh_proto.GatewayRoute_HttpRoute_Filter_Mirror{
			Backend: &mesh_proto.GatewayRoute_Backend{
				Destination: destinationRef,
			},
			Percentage: util_proto.Double(100),
		}

		kumaFilter.Filter = &mesh_proto.GatewayRoute_HttpRoute_Filter_Mirror_{
			Mirror: &kumaInnerFilter,
		}
	case gatewayapi.HTTPRouteFilterRequestRedirect:
		filter := filter.RequestRedirect

		kumaInnerFilter := mesh_proto.GatewayRoute_HttpRoute_Filter_Redirect{}

		if s := filter.Scheme; s != nil {
			kumaInnerFilter.Scheme = *s
		}

		if h := filter.Hostname; h != nil {
			kumaInnerFilter.Hostname = string(*h)
		}

		if p := filter.Port; p != nil {
			kumaInnerFilter.Port = uint32(*p)
		}

		if sc := filter.StatusCode; sc != nil {
			kumaInnerFilter.StatusCode = uint32(*sc)
		}

		kumaFilter.Filter = &mesh_proto.GatewayRoute_HttpRoute_Filter_Redirect_{
			Redirect: &kumaInnerFilter,
		}
	default:
		return nil, nil, fmt.Errorf("unsupported filter type %v", filter.Type)
	}

	return &kumaFilter, nil, nil
}
