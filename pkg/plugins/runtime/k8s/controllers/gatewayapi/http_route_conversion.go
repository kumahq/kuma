package gatewayapi

import (
	"context"
	"fmt"

	"github.com/pkg/errors"
	kube_core "k8s.io/api/core/v1"
	kube_apierrs "k8s.io/apimachinery/pkg/api/errors"
	kube_meta "k8s.io/apimachinery/pkg/apis/meta/v1"
	gatewayapi "sigs.k8s.io/gateway-api/apis/v1alpha2"

	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
	core_mesh "github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
	"github.com/kumahq/kuma/pkg/core/resources/store"
	mesh_k8s "github.com/kumahq/kuma/pkg/plugins/resources/k8s/native/api/v1alpha1"
	"github.com/kumahq/kuma/pkg/plugins/runtime/k8s/controllers/gatewayapi/policy"
	k8s_util "github.com/kumahq/kuma/pkg/plugins/runtime/k8s/util"
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

	var rules []*mesh_proto.GatewayRoute_HttpRoute_Rule

	for _, rule := range route.Spec.Rules {
		kumaRule, condition, err := r.gapiToKumaRule(ctx, mesh, route, rule)
		if err != nil {
			return nil, nil, errors.Wrap(err, "couldn't convert HTTPRoute to Kuma GatewayRoute")
		}
		if condition != nil {
			return nil, []kube_meta.Condition{*condition}, nil
		}

		rules = append(rules, &kumaRule)
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
		{
			Type:   string(gatewayapi.ConditionRouteResolvedRefs),
			Status: kube_meta.ConditionTrue,
			Reason: string(gatewayapi.ConditionRouteResolvedRefs),
		},
		// TODO: reflect the true state from the actual gateway of this
		// route
		{
			Type:   string(gatewayapi.ConditionRouteAccepted),
			Status: kube_meta.ConditionTrue,
			Reason: string(gatewayapi.ConditionRouteAccepted),
		},
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
	policyRef := policy.PolicyReferenceBackend(policy.FromHTTPRouteIn(objectNamespace), ref)

	gk := policyRef.GroupKindReferredTo()
	namespacedName := policyRef.NamespacedNameReferredTo()

	if permitted, err := policy.IsReferencePermitted(ctx, r.Client, policyRef); err != nil {
		return nil, nil, errors.Wrap(err, "couldn't determine if backend reference is permitted")
	} else if !permitted {
		return nil,
			&kube_meta.Condition{
				Type:    string(gatewayapi.ConditionRouteResolvedRefs),
				Status:  kube_meta.ConditionFalse,
				Reason:  RefNotPermitted,
				Message: fmt.Sprintf("reference to %s %q not permitted by any ReferencePolicy", gk, namespacedName),
			},
			nil
	}

	switch {
	case gk.Kind == "Service" && gk.Group == "":
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
		port := int32(*ref.Port)

		svc := &kube_core.Service{}
		if err := r.Client.Get(ctx, namespacedName, svc); err != nil {
			if kube_apierrs.IsNotFound(err) {
				return nil,
					&kube_meta.Condition{
						Type:    string(gatewayapi.ConditionRouteResolvedRefs),
						Status:  kube_meta.ConditionFalse,
						Reason:  ObjectNotFound,
						Message: fmt.Sprintf("backend reference references a non-existent Service %q", namespacedName.String()),
					},
					nil
			}
			return nil, nil, err
		}

		return map[string]string{
			mesh_proto.ServiceTag: k8s_util.ServiceTagFor(svc, &port),
		}, nil, nil
	case gk.Kind == "ExternalService" && gk.Group == mesh_k8s.GroupVersion.Group:
		resource := core_mesh.NewExternalServiceResource()
		if err := r.ResourceManager.Get(ctx, resource, store.GetByKey(namespacedName.Name, mesh)); err != nil {
			if store.IsResourceNotFound(err) {
				return nil,
					&kube_meta.Condition{
						Type:    string(gatewayapi.ConditionRouteResolvedRefs),
						Status:  kube_meta.ConditionFalse,
						Reason:  ObjectNotFound,
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
		if kumaMethod, ok := mesh_proto.HttpMethod_value[string(*m)]; ok && kumaMethod != int32(mesh_proto.HttpMethod_NONE) {
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

		var requestHeader mesh_proto.GatewayRoute_HttpRoute_Filter_RequestHeader

		for _, set := range filter.Set {
			requestHeader.Set = append(requestHeader.Set, k8sToKumaHeader(set))
		}

		for _, add := range filter.Add {
			requestHeader.Add = append(requestHeader.Add, k8sToKumaHeader(add))
		}

		requestHeader.Remove = filter.Remove

		kumaFilter.Filter = &mesh_proto.GatewayRoute_HttpRoute_Filter_RequestHeader_{
			RequestHeader: &requestHeader,
		}
	case gatewayapi.HTTPRouteFilterRequestMirror:
		filter := filter.RequestMirror

		destinationRef, condition, err := r.gapiToKumaRef(ctx, mesh, namespace, filter.BackendRef)
		if err != nil || condition != nil {
			return nil, condition, err
		}

		mirror := mesh_proto.GatewayRoute_HttpRoute_Filter_Mirror{
			Backend: &mesh_proto.GatewayRoute_Backend{
				Destination: destinationRef,
			},
			Percentage: util_proto.Double(100),
		}

		kumaFilter.Filter = &mesh_proto.GatewayRoute_HttpRoute_Filter_Mirror_{
			Mirror: &mirror,
		}
	case gatewayapi.HTTPRouteFilterRequestRedirect:
		filter := filter.RequestRedirect

		redirect := mesh_proto.GatewayRoute_HttpRoute_Filter_Redirect{}

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

		kumaFilter.Filter = &mesh_proto.GatewayRoute_HttpRoute_Filter_Redirect_{
			Redirect: &redirect,
		}
	default:
		return nil, nil, fmt.Errorf("unsupported filter type %q", filter.Type)
	}

	return &kumaFilter, nil, nil
}
