package policy

import (
	"github.com/Kong/kuma/api/mesh/v1alpha1"
	"github.com/Kong/kuma/pkg/core/resources/model"
)

type ConnectionPolicy interface {
	model.Resource
	Sources() []*v1alpha1.Selector
	Destinations() []*v1alpha1.Selector
}

// ServiceName is a convenience type alias to clarify the meaning of string value.
type ServiceName = string

// RouteMap holds the most specific ConnectionPolicy for each outbound interface of a Dataplane.
type ConnectionPolicyMap map[ServiceName]ConnectionPolicy
