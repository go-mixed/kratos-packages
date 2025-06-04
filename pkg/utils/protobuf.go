package utils

import (
	"bytes"
	"cmp"
	"github.com/pkg/errors"
	"github.com/samber/lo"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
	"reflect"
	"strings"
)

// ProtobufToMap 将protobuf对象转为map
func ProtobufToMap(protobuf proto.Message, keepNil bool) map[string]any {
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

func GetProtoBufField[P proto.Message](message P, fieldName string) protoreflect.FieldDescriptor {
	messageDesc := message.ProtoReflect().Descriptor()
	return messageDesc.Fields().ByName(protoreflect.Name(fieldName))
}

// CompareProtoBuf 比较两个protobuf消息，
// message1 < message2 返回 -1，message1 > message2 返回 1，message1 == message2 返回 0
func CompareProtoBuf[P proto.Message](message1 P, message2 P, fieldName string) int {
	field1 := GetProtoBufField(message1, fieldName)
	field2 := GetProtoBufField(message2, fieldName)

	if field1 == nil || field2 == nil { // field1 == field2
		return 0
	}

	kind := field1.Kind()
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
		// 只支持字符串、数字、布尔等类型，其它类型返回0
		return 0
	}

	val1 := message1.ProtoReflect().Get(field1)
	val2 := message2.ProtoReflect().Get(field2)

	switch kind {
	case protoreflect.StringKind:
		return strings.Compare(val1.String(), val2.String())
	case protoreflect.BytesKind:
		return bytes.Compare(val1.Bytes(), val2.Bytes())
	case protoreflect.BoolKind:
		return lo.If(val1.Bool(), 1).Else(0) - lo.If(val2.Bool(), 1).Else(0)
	case protoreflect.Int32Kind, protoreflect.Int64Kind, protoreflect.Sint32Kind, protoreflect.Sint64Kind,
		protoreflect.Fixed32Kind, protoreflect.Fixed64Kind, protoreflect.Sfixed32Kind, protoreflect.Sfixed64Kind:
		return cmp.Compare(val1.Int(), val2.Int())
	case protoreflect.Uint32Kind, protoreflect.Uint64Kind:
		return cmp.Compare(val1.Uint(), val2.Uint())
	case protoreflect.FloatKind, protoreflect.DoubleKind: // 这里在Goland中会提示Condition is always true，因为它识别到上面的if条件排除了其它类型，不用管这个问题
		return cmp.Compare(val1.Float(), val2.Float())
	}
	return 0
}

// StringToProtoEnum 将enum的string名称转换为protobuf枚举类型
func StringToProtoEnum[T protoreflect.Enum](enumVal T, enumName string) (T, error) {
	enumType := enumVal.Type()
	// 通过枚举类型的描述符查找枚举值描述符
	valueDesc := enumType.Descriptor().Values().ByName(protoreflect.Name(enumName))
	if valueDesc == nil {
		var nilT T
		return nilT, errors.New("enum value not found")
	}

	// 创建一个 protoreflect.Enum 类型的枚举值
	protoEnum := enumType.New(valueDesc.Number())
	return protoEnum.(T), nil
}
