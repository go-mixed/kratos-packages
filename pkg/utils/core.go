package utils

import (
	"reflect"
	"sync"
)

// IsNil 指针是否为nil
func IsNil(v any) bool {
	if v == nil {
		return true
	}
	vOf := reflect.ValueOf(v)
	return vOf.Kind() == reflect.Ptr && vOf.IsNil()
}

// IsZero 判断是否为零值
func IsZero(v any) bool {
	if v == nil {
		return true
	}
	return reflect.ValueOf(v).IsZero()
}

// IsZeroT 判断是否为零值，T必须是可比较的类型，性能比IsZero高
func IsZeroT[T comparable](v T) bool {
	var zero T
	return v == zero
}

// If 条件判断，类似三元运算符
func If[T any](condition bool, trueValue, falseValue T) T {
	if condition {
		return trueValue
	}
	return falseValue
}

// IfFunc 条件判断，类似三元运算符，但是trueValue和falseValue是函数，返回值为这两个函数的返回值
func IfFunc[T any](condition bool, trueFunc, falseFunc func() T) T {
	if condition {
		return trueFunc()
	}
	if falseFunc == nil {
		var zero T
		return zero
	}
	return falseFunc()
}

func MapGet[T any](m map[string]any, key string, defaultValue T) T {
	if m == nil {
		return defaultValue
	}

	if v, ok := m[key]; ok {
		return v.(T)
	}

	return defaultValue
}

func SyncMapGet[T any](m *sync.Map, key string, defaultValue T) T {
	if m == nil {
		return defaultValue
	}

	if v, ok := m.Load(key); ok {
		return v.(T)
	}

	return defaultValue
}

// GetClassName 获取对象的类名
func GetClassName(v any) string {
	if v == nil {
		return ""
	}

	typeOf := reflect.TypeOf(v)
	for {
		if typeOf.Kind() == reflect.Ptr {
			typeOf = typeOf.Elem()
		} else {
			break
		}
	}

	return typeOf.PkgPath() + "." + typeOf.Name()
}

// Ptr 将一个值转为指针
func Ptr[T any](v T) *T {
	return &v
}

// IsPtr 判断是否是指针类型
func IsPtr[T any](v T) bool {
	return reflect.TypeOf(v).Kind() == reflect.Ptr
}

// PtrValue 获取指针的值
func PtrValue(v any) any {
	if v == nil {
		return nil
	} else if vOf := reflect.ValueOf(v); vOf.Kind() != reflect.Ptr || vOf.IsNil() {
		return v
	}

	return reflect.ValueOf(v).Elem().Interface()
}

// New 创建对象
//   - 如果非指针类型，返回该类型的零值（利用泛型的特性）；
//   - 如果是指针类型，返回new(T)；
//   - 如果是map、slice、chan类型，返回make后的map、slice、chan
func New[T any]() T {
	var v T
	typeOf := reflect.TypeOf(v)

	switch typeOf.Kind() {
	case reflect.Ptr:
		elemPtr := reflect.New(typeOf.Elem())
		return elemPtr.Interface().(T)
	case reflect.Map: // map需要make
		return reflect.MakeMap(typeOf).Interface().(T)
	case reflect.Slice: // slice需要make
		return reflect.MakeSlice(typeOf, 0, 0).Interface().(T)
	case reflect.Chan: // chan需要make
		return reflect.MakeChan(typeOf, 0).Interface().(T)
	}

	// 其它类型使用泛型的特性返回零值即可
	return v
}
