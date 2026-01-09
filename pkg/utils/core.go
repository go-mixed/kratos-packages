package utils

import (
	"reflect"
)

// IsNil 指针是否为nil，支持多层Ptr
func IsNil(v any) bool {
	if v == nil {
		return true
	}
	rv := reflect.ValueOf(v)
	// 循环解引用指针和接口
	for rv.Kind() == reflect.Ptr || rv.Kind() == reflect.Interface {
		if rv.IsNil() {
			return true
		}
		rv = rv.Elem() // 深入解引用
	}

	// 检查其他可 nil 类型
	switch rv.Kind() {
	case reflect.Slice, reflect.Map, reflect.Func, reflect.Chan:
		return rv.IsNil()
	default:
	}

	return false
}

// IsZero 判断是否为nil/零值，支持多层Ptr
// 能够判断map、slice、array类型是否没有元素
//
//	```
//
// var nilSlice []int
// var emptySlice = []int{}
//
// var nilMap map[string]int
// var emptyMap = map[string]int{}
//
// var arrZero = [3]int{}
// var arrNonZero = [3]int{1}
//
// var iface interface{} = &arrZero // 接口包裹指针
//
// IsZero(nilSlice)      // true (nil)
// IsZero(emptySlice)    // true (长度0)
// IsZero(nilMap)        // true (nil)
// IsZero(emptyMap)      // true (长度0)
// IsZero(arrZero)       // true (所有元素零值)
// IsZero(arrNonZero)    // false
// IsZero(iface)         // true (解引用后数组为零值)
//
//	```
func IsZero(v any) bool {
	if IsNil(v) {
		return true
	}
	rv := PtrElement(v)
	switch rv.Kind() {
	case reflect.Slice, reflect.Map:
		return rv.Len() == 0
	case reflect.Array: // 数组需要检查所有元素是否为零值
		return rv.IsZero()
	default:

	}
	return rv.IsZero()
}

// IsZeroT 判断是否为零值，T必须是可比较的类型，性能比IsZero高
// 但是无法判断map、slice、array类型是否没有元素
func IsZeroT[T comparable](v T) bool {
	var zero T
	return v == zero
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

// IsFunc 判断是否是函数类型
func IsFunc(fn any) bool {
	return reflect.TypeOf(fn).Kind() == reflect.Func
}

// IsFuncArgs 判断函数是否符合参数类型
//   - 不允许在args中传入 裸指针nil，除非是 (*someType)(nil)
//   - 普通函数：func example(a int, b string, c *someType) 那检查方式为 IsFuncArgs(example, int(0), "", &someType{})
//   - 对于变参，比如func example(a int, b string...) 那检查方式为 IsFuncArgs(example, int(0), []string{})
func IsFuncArgs(fn any, args ...any) bool {
	if !IsFunc(fn) {
		return false
	}

	fnType := reflect.TypeOf(fn)

	// 统一处理参数数量检查
	// 非变参函数：参数数量必须完全匹配
	// 变参函数：参数数量必须完全匹配（包括变参部分作为一个切片）
	if len(args) != fnType.NumIn() {
		return false
	}

	// 检查每个参数类型
	for i := 0; i < len(args); i++ {
		expectedType := fnType.In(i)

		arg := args[i]
		// 严格模式，必须类型要完全相等（可以使用 !reflect.TypeOf(arg).AssignableTo(expectedType) 表示 可以为延伸类型）
		if reflect.TypeOf(arg) != expectedType {
			return false
		}
	}

	return true
}

// Ptrs 将多个值转为指针
func Ptrs[T any](vs ...T) []*T {
	ptrs := make([]*T, len(vs))
	for i, v := range vs {
		v := v // 避免闭包引用同一个变量（Golang1.22之前）
		ptrs[i] = &v
	}
	return ptrs
}

// IsPtr 判断是否是指针类型
func IsPtr[T any](v T) bool {
	return reflect.TypeOf(v).Kind() == reflect.Ptr
}

// PtrElement 获取指针的元素
func PtrElement(v any) reflect.Value {
	rv := reflect.ValueOf(v)
	for rv.Kind() == reflect.Ptr || rv.Kind() == reflect.Interface {
		if rv.IsNil() { // 避免解引用 nil 值导致的 panic
			return rv
		}
		rv = rv.Elem()
	}
	return rv
}

// New 创建对象
//   - 如果非指针类型，返回该类型的零值（利用泛型的特性）；
//   - 如果是指针类型，返回new(T)；
//   - 如果是map、slice、chan类型，返回make后的map、slice、chan
//
// golang v1.26中自带new(Type)
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

// IsBasicType 判断是否是基础类型
func IsBasicType(typOf reflect.Type) bool {
	switch typOf.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64,
		reflect.Float32, reflect.Float64,
		reflect.Complex64, reflect.Complex128,
		reflect.String, reflect.Bool:
		return true
	default:
		return false
	}
}
