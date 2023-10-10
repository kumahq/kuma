package filters

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

// Resource return a store filter depending on the resource. We take a descriptor so that we can do advance filtering options
// For example we could make a filter that works on top level targetRef by looking at the descriptor info
func Resource(resDescriptor core_model.ResourceTypeDescriptor) func(request *restful.Request) (store.ListFilterFunc, error) {
	switch resDescriptor.Name {
	case mesh.DataplaneType:
		return func(request *restful.Request) (store.ListFilterFunc, error) {
			gatewayFilter, err := gatewayModeFilterFromParameter(request)
			if err != nil {
				return nil, err
			}

			tags := parseTags(request.QueryParameters("tag"))

			return func(rs core_model.Resource) bool {
				dataplane, ok := rs.(*mesh.DataplaneResource)
				if !ok { // Sometimes this is going to return insights for example which will not match
					return true
				}
				if !gatewayFilter(dataplane.Spec.GetNetworking().GetGateway()) {
					return false
				}

				if !dataplane.Spec.MatchTagsFuzzy(tags) {
					return false
				}

				return true
			}, nil
		}
	case mesh.ExternalServiceType:
		return func(request *restful.Request) (store.ListFilterFunc, error) {
			tags := parseTags(request.QueryParameters("tag"))

			return func(rs core_model.Resource) bool {
				return rs.(*mesh.ExternalServiceResource).Spec.MatchTagsFuzzy(tags)
			}, nil
		}
	default:
		return func(request *restful.Request) (store.ListFilterFunc, error) {
			return nil, nil
		}
	}
}

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
