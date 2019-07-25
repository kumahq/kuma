package definitions

import "github.com/Kong/konvoy/components/konvoy-control-plane/pkg/core/resources/model"

type ResourceWsDefinition struct {
	Name                string
	Path                string
	ResourceFactory     func() model.Resource
	ResourceListFactory func() model.ResourceList
}
