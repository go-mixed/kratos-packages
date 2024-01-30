package utils

import (
	"encoding/json"
	"fmt"
	"github.com/samber/lo"
	"net/url"
	"reflect"
	"strconv"
	"strings"
)

type Stringer interface {
	String() string
}

// ToString 将任意类型转为字符串, 标量或标量的子类型可以直接转, 其它转为json的字符串
// 注意: type ABC string 这种类型会走到default分支, 为了减少反射带来的性能负担, 对已知可以强转的类型, 可以自行强转: string(abc)
// otherTypeAsJson: 无法识别的type转换为json 不然会返回空字符串
func ToString(v any, otherTypeAsJson bool) string {
	// 先用 type assert检查, 支持标量, 速度更快
	switch v.(type) {
	case []rune:
		return string(v.([]rune))
	case []byte:
		return string(v.([]byte))
	case string:
		return v.(string)
	case bool:
		if v.(bool) {
			return "true"
		} else {
			return "false"
		}
	case int, int8, int16, int32, int64,
		uint, uint8, uint16, uint32, uint64,
		float32, float64, complex64, complex128:
		return fmt.Sprintf("%v", v)
	case Stringer:
		return v.(Stringer).String()
	case error:
		return v.(error).Error()
	default:
		// 针对 type ABC string 这种需要使用typeof.kind检查
		switch reflect.TypeOf(v).Kind() {
		case reflect.Bool:
			if reflect.ValueOf(v).Bool() {
				return "true"
			} else {
				return "false"
			}
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
			reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64,
			reflect.Float32, reflect.Float64, reflect.Complex64, reflect.Complex128,
			reflect.String:
			return fmt.Sprintf("%v", v)
		default: // 均不符合 则使用json来处理
			if otherTypeAsJson {
				j, _ := json.Marshal(v)
				return string(j)
			} else {
				return ""
			}
		}
	}
}

// AnyToUrlValues 将map/struct/slice/string转化为url.Values。
//
//	如果是string，则会尝试使用url.ParseQuery进行转换
//	如果是map，则会将key/value转换为string，tag传入""
//	如果是struct，则会将struct的字段的tag名（比如json）作为key。注意：匿名字段会展开；子Struct一律会转换为json
func AnyToUrlValues(data any, tag string) url.Values {
	var result url.Values = url.Values{}

	if data == nil {
		return result
	}

	vOf := reflect.ValueOf(data)
	if vOf.Kind() == reflect.Ptr && vOf.IsNil() {
		return result
	}

	if vOf.Kind() == reflect.Ptr {
		vOf = vOf.Elem()
	}

	switch vOf.Kind() {
	case reflect.String:
		result, _ = url.ParseQuery(vOf.String())
	case reflect.Map:
		for _, kOf := range vOf.MapKeys() {
			k := fmt.Sprintf("%v", kOf.Interface())
			result.Set(k, ToString(vOf.MapIndex(kOf).Interface(), true))
		}
	case reflect.Struct:
		tOf := vOf.Type()
		for i := 0; i < tOf.NumField(); i++ {
			field := tOf.Field(i)
			// 如果是私有字段，则跳过
			if !field.IsExported() {
				continue
			}
			// 获取tag的值，如果没有tag，则使用字段名
			name := field.Name
			if tag != "" {
				tagName, ok := field.Tag.Lookup(tag)
				if ok && tagName != "-" && tagName != "_" {
					segments := strings.Split(tagName, ",")
					name = segments[0]
				}
			}

			// 如果是匿名字段，则展开
			if field.Anonymous {
				c := AnyToUrlValues(vOf.Field(i).Interface(), tag)
				for k := range c {
					result.Set(k, c.Get(k))
				}
				continue
			}

			result.Set(name, ToString(vOf.Field(i).Interface(), true))
		}
	case reflect.Slice:
		for i := 0; i < vOf.Len(); i++ {
			result.Set(strconv.Itoa(i), ToString(vOf.Index(i).Interface(), true))
		}
	}

	return result
}

// InterfacesToStrings []any to []string
func InterfacesToStrings(data []any) []string {
	return lo.Map(data, func(val any, _ int) string {
		switch v := val.(type) {
		case string:
			return v
		case []byte:
			return string(v)
		}
		return fmt.Sprintf("%v", val)
	})
}

// StringsToInterfaces []string to []any
func StringsToInterfaces(data []string) []any {
	return lo.Map(data, func(val string, _ int) any {
		return val
	})
}
