package util

import (
	"github.com/kumahq/kuma/pkg/core/resources/model"
	"github.com/kumahq/kuma/pkg/core/resources/registry"
)

func GetSupportedTypes() []string {
	var types []string
	for _, def := range registry.Global().ObjectTypes(model.HasKdsEnabled()) {
		types = append(types, string(def))
	}
	return types
}
