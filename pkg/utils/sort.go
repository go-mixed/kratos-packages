package utils

import (
	"sort"
)

// SortSliceBy sorts a slice of any type by a slice of comparable, the order of which is given by the second slice
//
//		Example: SortSliceBy(
//		  []struct{{Name: "Alice"}, {Name: "Bob"}, {Name: "Unknown"}, {Name: "Zack"}},
//		  []string{"Bob", "Zack", "Alice"},
//		  func(i int) string { return originalList[i].Name }
//		)
//	 Result: []struct{{Name: "Bob"}, {Name: "Zack"}, {Name: "Alice"}, {Name: "Unknown"}}
func SortSliceBy[S any, T comparable](originalList S, sortedBy []T, fn func(i int) T) {
	mapIndices := make(map[T]int, len(sortedBy))
	for i, v := range sortedBy {
		mapIndices[v] = i
	}
	sort.Slice(originalList, func(i, j int) bool {
		val1 := fn(i)
		val2 := fn(j)
		index1, ok1 := mapIndices[val1]
		index2, ok2 := mapIndices[val2]
		if ok1 && ok2 {
			return index1 < index2
		} else if ok1 {
			return true
		}

		return false
	})
}
