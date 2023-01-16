package api_server

import (
	"reflect"
	"strings"

	"github.com/emicklei/go-restful/v3"

	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
	"github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
	core_model "github.com/kumahq/kuma/pkg/core/resources/model"
	"github.com/kumahq/kuma/pkg/core/resources/store"
	"github.com/kumahq/kuma/pkg/core/validators"
)

type DpFilter func(*mesh_proto.Dataplane_Networking_Gateway) bool

func gatewayModeFilterFromParameter(request *restful.Request) (DpFilter, error) {
	mode := strings.ToLower(request.QueryParameter("gateway"))
	if mode != "" && mode != "true" && mode != "false" && mode != "builtin" && mode != "delegated" {
		verr := validators.ValidationError{}
		verr.AddViolationAt(
			validators.RootedAt(request.SelectedRoutePath()).Field("gateway"),
			"shoud use `true`, `false`, 'builtin' or 'delegated' instead of "+mode)
		return nil, &verr
	}

	isnil := func(a interface{}) bool {
		return a == nil || reflect.ValueOf(a).IsNil()
	}
	switch mode {
	case "true":
		return func(a *mesh_proto.Dataplane_Networking_Gateway) bool {
			return !isnil(a)
		}, nil
	case "false":
		return func(a *mesh_proto.Dataplane_Networking_Gateway) bool {
			return isnil(a)
		}, nil
	case "builtin":
		return func(a *mesh_proto.Dataplane_Networking_Gateway) bool {
			return !isnil(a) && a.Type == mesh_proto.Dataplane_Networking_Gateway_BUILTIN
		}, nil
	case "delegated":
		return func(a *mesh_proto.Dataplane_Networking_Gateway) bool {
			return !isnil(a) && a.Type == mesh_proto.Dataplane_Networking_Gateway_DELEGATED
		}, nil
	default:
		return func(a *mesh_proto.Dataplane_Networking_Gateway) bool {
			return true
		}, nil
	}
}

func genFilter(request *restful.Request) (store.ListFilterFunc, error) {
	gatewayFilter, err := gatewayModeFilterFromParameter(request)
	if err != nil {
		return nil, err
	}

	tags := parseTags(request.QueryParameters("tag"))

	return func(rs core_model.Resource) bool {
		dataplane := rs.(*mesh.DataplaneResource)
		if !gatewayFilter(dataplane.Spec.GetNetworking().GetGateway()) {
			return false
		}

		if !dataplane.Spec.MatchTagsFuzzy(tags) {
			return false
		}

		return true
	}, nil
}

// Tags should be passed in form of ?tag=service:mobile&tag=version:v1
func parseTags(queryParamValues []string) map[string]string {
	tags := make(map[string]string)
	for _, value := range queryParamValues {
		// ":" are valid in tag value so only stop at the first separator
		tagKv := strings.SplitN(value, ":", 2)
		if len(tagKv) != 2 {
			// ignore invalid formatted tags
			continue
		}
		tags[tagKv[0]] = tagKv[1]
	}
	return tags
}
