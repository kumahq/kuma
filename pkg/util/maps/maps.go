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
