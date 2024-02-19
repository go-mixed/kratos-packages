package utils

import (
	"google.golang.org/protobuf/reflect/protoreflect"
	"reflect"
	"strings"
)

type IProtobuf interface {
	ProtoReflect() protoreflect.Message
}

// ProtobufToMap 将protobuf对象转为map
func ProtobufToMap(protobuf IProtobuf, keepNil bool) map[string]any {
	if protobuf == nil {
		return nil
	}

	vOf := reflect.ValueOf(protobuf).Elem()
	tOf := vOf.Type()
	results := make(map[string]any)

	for i := 0; i < tOf.NumField(); i++ {
		field := tOf.Field(i)
		// 如果是私有字段、匿名字段，则跳过
		if !field.IsExported() || field.Anonymous {
			continue
		}

		name := field.Name
		tagName, ok := field.Tag.Lookup("json")
		if ok && tagName != "-" && tagName != "_" {
			segments := strings.Split(tagName, ",")
			name = segments[0]
		}

		vfOf := vOf.Field(i)
		// Protobuf中没有Channel、Func、Interface、UnsafePointer类型
		// 为指针类型的一般是optional修饰的字段
		if vfOf.Kind() == reflect.Ptr {
			if vfOf.IsNil() && !keepNil { // 不保留空指针
				continue
			}
			vfOf = vfOf.Elem()
		} else if vfOf.Kind() == reflect.Slice || vfOf.Kind() == reflect.Map {
			if vfOf.Len() == 0 && !keepNil {
				continue
			}
		}

		results[name] = vfOf.Interface()

	}
	return results
}
