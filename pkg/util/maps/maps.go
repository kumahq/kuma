package maps

import (
	"cmp"
	"maps"
	"slices"
)

func SortedKeys[M ~map[K]V, K cmp.Ordered, V any](m M) []K {
	keys := maps.Keys(m)
	return slices.Sorted(keys)
}

func AllKeys[M ~map[K]V, K comparable, V any](m M) []K {
	keys := maps.Keys(m)
	return slices.Collect(keys)
}

func AllValues[M ~map[K]V, K comparable, V any](m M) []V {
	values := maps.Values(m)
	return slices.Collect(values)
}

func MapValues[K comparable, V any, R any](input map[K]V, f func(K, V) R) map[K]R {
	output := make(map[K]R, len(input))
	for k, v := range input {
		output[k] = f(k, v)
	}
	return output
}

func FlatMapKV[K comparable, V any, R any](m map[K]V, f func(K, V) []R) []R {
	var result []R
	for k, v := range m {
		result = append(result, f(k, v)...)
	}
	return result
}
