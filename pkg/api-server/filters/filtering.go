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

type FilterOp string

const (
	FilterOpEq FilterOp = "eq"
)

type filterEntry struct {
	Op    FilterOp
	Key   string
	Value string
}

// labelFilter returns a store filter that filters resources based on their labels.
// the end goal is to support https://kong-aip.netlify.app/aip/160/ which is a standard for filtering resources
// but atm we only support eq operation with exact match
func labelFilter(request *restful.Request) (store.ListFilterFunc, error) {
	verr := &validators.ValidationError{}
	filters := make([]filterEntry, 0)
	for k, v := range request.Request.URL.Query() {
		if !strings.HasPrefix(k, "filter[") {
			continue
		}
		if !strings.HasPrefix(k, "filter[labels.") {
			verr.AddViolationAt(
				validators.RootedAt(request.SelectedRoutePath()).Field(k), "filters are only supported on labels")
			continue
		}
		closingBracket := strings.Index(k, "]")
		key := k[len("filter[labels."):closingBracket]
		if closingBracket != len(k)-1 {
			verr.AddViolationAt(
				validators.RootedAt(request.SelectedRoutePath()).Field(k), "advanced filters are not supported")
			continue
		}
		op := FilterOpEq
		filters = append(filters, filterEntry{
			Op:    op,
			Key:   key,
			Value: v[len(v)-1],
		})
	}
	if verr.HasViolations() {
		return nil, verr
	}
	if len(filters) == 0 {
		return nil, nil
	}
	return func(rs core_model.Resource) bool {
		labels := rs.GetMeta().GetLabels()
		for _, filter := range filters {
			v, ok := labels[filter.Key]
			if filter.Op == FilterOpEq {
				if !ok || v != filter.Value {
					return false
				}
			}
			if !ok || v != filter.Value {
				return false
			}
		}
		return true
	}, nil
}

// Resource return a store filter depending on the resource. We take a descriptor so that we can do advance filtering options
// For example we could make a filter that works on top level targetRef by looking at the descriptor info
func Resource(resDescriptor core_model.ResourceTypeDescriptor) func(request *restful.Request) (store.ListFilterFunc, error) {
	return func(request *restful.Request) (store.ListFilterFunc, error) {
		genericFilter, err := labelFilter(request)
		if err != nil {
			return nil, err
		}
		switch resDescriptor.Name {
		case mesh.DataplaneType:
			gatewayFilter, err := gatewayModeFilterFromParameter(request)
			if err != nil {
				return nil, err
			}

			tags := parseTags(request.QueryParameters("tag"))

			return func(rs core_model.Resource) bool {
				if genericFilter != nil && !genericFilter(rs) {
					return false
				}
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
		case mesh.ExternalServiceType:
			tags := parseTags(request.QueryParameters("tag"))

			return func(rs core_model.Resource) bool {
				if genericFilter != nil && !genericFilter(rs) {
					return false
				}
				return rs.(*mesh.ExternalServiceResource).Spec.MatchTagsFuzzy(tags)
			}, nil
		default:
			return genericFilter, nil
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
