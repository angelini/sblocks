package maps

import (
	"golang.org/x/exp/constraints"
	"golang.org/x/exp/slices"
)

type OrderedComparable interface {
	constraints.Ordered
	comparable
}

func SortedValues[K OrderedComparable, V any](m map[K]V) []V {
	keys := make([]K, 0, len(m))
	for key := range m {
		keys = append(keys, key)
	}

	slices.Sort(keys)

	values := make([]V, 0, len(m))
	for _, key := range keys {
		values = append(values, m[key])
	}

	return values
}

func SortedKeys[K OrderedComparable, V any](m map[K]V) []K {
	keys := make([]K, 0, len(m))
	for key := range m {
		keys = append(keys, key)
	}

	slices.Sort(keys)
	return keys
}
