package gatewayapi

import (
	"errors"
	"fmt"

	gatewayapi "sigs.k8s.io/gateway-api/apis/v1alpha2"

	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
	util_proto "github.com/kumahq/kuma/pkg/util/proto"
)

func k8sToKumaHeader(header gatewayapi.HTTPHeader) *mesh_proto.GatewayRoute_HttpRoute_Filter_RequestHeader_Header {
	return &mesh_proto.GatewayRoute_HttpRoute_Filter_RequestHeader_Header{
		Name:  string(header.Name),
		Value: header.Value,
	}
}

func gapiToKumaRef(objectNamespace string, ref gatewayapi.BackendObjectReference) (map[string]string, error) {
	// References to Services are required by GAPI to include a port
	// TODO remove when https://github.com/kubernetes-sigs/gateway-api/pull/944
	// is in master
	if ref.Port == nil {
		return nil, errors.New("backend reference must include port")
	}

	if *ref.Kind != "Service" {
		return nil, errors.New("backend reference must be a Service") // TODO setappropriate status on gateway
	}

	namespace := objectNamespace
	if ref.Namespace != nil {
		namespace = string(*ref.Namespace)
	}

	return map[string]string{
		mesh_proto.ServiceTag: fmt.Sprintf("%s_%s_svc_%d", ref.Name, namespace, *ref.Port),
	}, nil
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

func gapiToKumaFilter(namespace string, filter gatewayapi.HTTPRouteFilter) (*mesh_proto.GatewayRoute_HttpRoute_Filter, error) {
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

		destinationRef, err := gapiToKumaRef(namespace, filter.BackendRef)
		if err != nil {
			return nil, err
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
		return nil, fmt.Errorf("unsupported filter type %v", filter.Type)
	}

	return &kumaFilter, nil
}
