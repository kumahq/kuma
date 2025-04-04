package kri

import core_model "github.com/kumahq/kuma/pkg/core/resources/model"

type ResourceReader interface {
	Get(Identifier) core_model.Resource
	ListOrEmpty(core_model.ResourceType) core_model.ResourceList
}
