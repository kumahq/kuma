package common

import core_model "github.com/kumahq/kuma/pkg/core/resources/model"

// Cast attempts to cast a slice of core_model.Resource to a slice of a specific type T.
// It returns the casted slice and a boolean indicating whether the cast was successful.
// If any element in the slice cannot be cast to the specified type, the function returns nil and false.
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
