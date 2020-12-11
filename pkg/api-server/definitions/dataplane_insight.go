package definitions

import (
	"github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
	"github.com/kumahq/kuma/pkg/core/resources/model"
)

var DataplaneInsightWsDefinition = ResourceWsDefinition{
	Name: "Dataplane Insight",
	Path: "dataplane-insights",
	ResourceFactory: func() model.Resource {
		return mesh.NewDataplaneInsightResource()
	},
	ResourceListFactory: func() model.ResourceList {
		return &mesh.DataplaneInsightResourceList{}
	},
}
