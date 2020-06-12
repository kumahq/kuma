package definitions

import "github.com/Kong/kuma/pkg/core/resources/model"

type ResourceWsDefinition struct {
	Name                string
	Path                string
	ResourceFactory     func() model.Resource
	ResourceListFactory func() model.ResourceList
	ReadOnly            bool
}
