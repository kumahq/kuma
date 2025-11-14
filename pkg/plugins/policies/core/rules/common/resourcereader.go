package common

import core_model "github.com/kumahq/kuma/v2/pkg/core/resources/model"

type ResourceReader interface {
	Get(resourceType core_model.ResourceType, ri core_model.ResourceIdentifier) core_model.Resource
	ListOrEmpty(resourceType core_model.ResourceType) core_model.ResourceList
}
