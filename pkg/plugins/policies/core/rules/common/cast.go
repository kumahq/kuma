package common

import core_model "github.com/kumahq/kuma/pkg/core/resources/model"

func Cast[T any](rs []core_model.Resource) ([]T, bool) {
	var rv []T
	for _, r := range rs {
		if casted, ok := r.GetSpec().(T); !ok {
			return nil, false
		} else {
			rv = append(rv, casted)
		}
	}
	return rv, true
}
