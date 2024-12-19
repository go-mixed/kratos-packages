package utils

import (
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
	"slices"
	"sort"
	"strings"
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

// SortProtobufList 对Protobuf List进行排序，
//
//   - orderField: protobuf文件中Message的字段，比如：Message{int32 xx = 1;}中的字段xx
//   - orderType: "asc" or "desc"
func SortProtobufList[P proto.Message](protoMessageList []P, orderField string, orderType string) {
	orderType = strings.ToLower(orderType)
	if orderType != "desc" && orderType != "asc" {
		orderType = "asc"
	}

	var message P
	field := GetProtoBufField(message, orderField)
	if field == nil { // 字段不存在，不做排序
		return
	}
	kind := field.Kind()
	if kind != protoreflect.StringKind &&
		kind != protoreflect.BytesKind &&
		kind != protoreflect.Int32Kind &&
		kind != protoreflect.Int64Kind &&
		kind != protoreflect.Uint32Kind &&
		kind != protoreflect.Uint64Kind &&
		kind != protoreflect.FloatKind &&
		kind != protoreflect.DoubleKind &&
		kind != protoreflect.BoolKind &&
		kind != protoreflect.Sint32Kind &&
		kind != protoreflect.Sint64Kind &&
		kind != protoreflect.Fixed32Kind &&
		kind != protoreflect.Fixed64Kind &&
		kind != protoreflect.Sfixed32Kind &&
		kind != protoreflect.Sfixed64Kind {
		// 只支持字符串、数字、布尔等类型
		return
	}

	sort.Slice(protoMessageList, func(i, j int) bool {
		cmp := CompareProtoBuf(protoMessageList[i], protoMessageList[j], orderField)
		if orderType == "asc" { // 升序时，i < j 返回true
			return cmp < 0
		}
		return cmp > 0
	})
}

// MultipleSortStableFunc 多字段稳定排序排序（如果a,b值相同，则不会改变a,b的顺序）
//   - 类似于 Order By a asc, b asc, c asc
//   - 注意：cmp返回非0时，后面的cmp不会执行
func MultipleSortStableFunc[S ~[]E, E any](list S, cmp ...func(a, b E) int) {
	slices.SortStableFunc(list, func(a, b E) int {
		for _, c := range cmp {
			if val := c(a, b); val != 0 {
				return val
			}
		}
		return 0
	})
}
