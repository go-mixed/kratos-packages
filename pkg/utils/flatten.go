package utils

import (
	"fmt"
	"reflect"
)

// Flatten 将结构体递归展开为 1维 map[string]any。对于多层级的结构，使用 parent + separator + fieldName 为key
func Flatten(v any, separator string) map[string]any {
	valOf := reflect.ValueOf(v)

	return flatten(valOf, "", separator)
}

func flatten(valOf reflect.Value, keyPrefix string, separator string) map[string]any {
	for {
		if valOf.Kind() == reflect.Ptr && !valOf.IsNil() {
			valOf = valOf.Elem()
		} else {
			break
		}
	}

	newKey := func(fieldName string) string {
		if keyPrefix == "" {
			return fieldName
		}
		return keyPrefix + separator + fieldName
	}

	typOf := valOf.Type()
	var result = make(map[string]any)
	kv := make(map[string]any)

	switch valOf.Kind() {
	case reflect.Struct:
		for i := 0; i < valOf.NumField(); i++ {
			field := typOf.Field(i)
			fieldName := field.Name
			fieldVal := valOf.Field(i)

			if !field.IsExported() {
				continue
			}

			if field.Anonymous {
				kv = flatten(fieldVal, keyPrefix, separator)
				for k, v := range kv {
					result[k] = v
				}
			} else {
				// 尝试递归 flatten 子结构体
				kv = flatten(fieldVal, newKey(fieldName), separator)
				if kv == nil { // 没有子字段，直接赋值
					result[newKey(fieldName)] = fieldVal.Interface()
				} else {
					for k, v := range kv {
						result[k] = v
					}
				}
			}
		}

	case reflect.Slice, reflect.Array:
		// 基础类型, 不递归 flatten
		if IsBasicType(typOf.Elem()) {
			return nil
		}

		for i := 0; i < valOf.Len(); i++ {
			fieldVal := valOf.Index(i)
			fieldName := fmt.Sprintf("%d", i)
			kv = flatten(fieldVal, newKey(fieldName), separator)
			if len(kv) == 0 { // 没有子字段，直接赋值
				result[newKey(fieldName)] = fieldVal.Interface()
			} else {
				for k, v := range kv {
					result[k] = v
				}
			}
		}

	case reflect.Map:
		// 基础类型, 才遍历map
		if !IsBasicType(typOf.Key()) {
			return nil
		}

		for _, key := range valOf.MapKeys() {
			fieldVal := valOf.MapIndex(key)
			fieldName := fmt.Sprintf("%v", key.Interface())
			kv = flatten(fieldVal, newKey(fieldName), separator)
			if len(kv) == 0 { // 没有子字段，直接赋值
				result[newKey(fieldName)] = fieldVal.Interface()
			} else {
				for k, v := range kv {
					result[k] = v
				}
			}
		}

	default:
		return nil
	}

	return result
}
