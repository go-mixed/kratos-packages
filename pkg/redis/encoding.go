package redis

import (
	"encoding"
	"encoding/json"
	"fmt"
	"github.com/redis/go-redis/v9"
	"reflect"
	"strconv"
	"time"
)

type anyStruct struct {
	original any
}

var _ encoding.BinaryMarshaler = (*anyStruct)(nil)
var _ encoding.BinaryUnmarshaler = (*anyStruct)(nil)

// WrapBinaryMarshaler 将一个不支持BinaryMarshaler的类型包装成一个结构体，以便redis的Writer时MarshalBinary
func WrapBinaryMarshaler(v any) any {
	if _, ok := v.(*anyStruct); ok {
		return v
	}
	// redis 只能处理下面这些类型
	switch v.(type) {
	case string, []byte, int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64, float32, float64, bool, time.Time, time.Duration, encoding.BinaryMarshaler:
		return v
	}
	// 否则，就包装成一个结构体，然后实现 MarshalBinary 和 UnmarshalBinary 接口
	// 当v==nil时，为了防止redis的Scan报错，也会返回一个结构体
	return &anyStruct{original: v}
}

// WrapBinaryUnmarshaler 将一个不支持BinaryUnmarshaler的类型包装成一个结构体，以便redis的Reader时UnmarshalBinary
func WrapBinaryUnmarshaler(v any) any {
	// 和 WrapBinaryMarshaler 不同，nil不会被包装成结构体，以便redis的Writer报错
	if v == nil {
		return nil
	} else if _, ok := v.(*anyStruct); ok {
		return v
	}

	// redis 只能处理下面这些类型
	switch v.(type) {
	case *string, *[]byte,
		*int, *int8, *int16, *int32, *int64, *uint, *uint8, *uint16, *uint32, *uint64,
		*float32, *float64, *bool,
		*time.Time, *time.Duration, encoding.BinaryUnmarshaler:
		return v
	}
	return &anyStruct{original: v}
}

func (a *anyStruct) MarshalBinary() ([]byte, error) {
	if a.original == nil {
		return nil, nil
	}
	return json.Marshal(a.original)
}

func (a *anyStruct) UnmarshalBinary(data []byte) error {
	if a.original == nil {
		return nil
	}
	return json.Unmarshal(data, a.original)
}

// WrapMapBinaryMarshaler 包装map[string]any，以便redis的Writer时MarshalBinary
func WrapMapBinaryMarshaler(v map[string]any) any {
	_v := make(map[string]any, len(v))
	for k, val := range v {
		_v[k] = WrapBinaryMarshaler(val)
	}
	return _v
}

// ScanCmd 将data转换成actual，和Scan的区别是，ScanCmd会检查第一个参数err是否为nil
func ScanCmd(err error, data string, actual any) error {
	if err != nil {
		return err
	}

	return Scan(data, actual)
}

// Scan 将data转换成actual
func Scan(data string, actual any) error {
	if actual == nil {
		return nil
	}

	cmd := redis.StringCmd{}
	cmd.SetVal(data)
	return cmd.Scan(WrapBinaryUnmarshaler(actual))
}

// ScanStringSliceCmd 将data从[]string转换成actual，和ScanStringSlice的区别是，ScanStringSliceCmd会检查第一个参数err是否为nil
func ScanStringSliceCmd(err error, data []string, actual any) error {
	if err != nil {
		return err
	}

	return ScanStringSlice(data, actual)
}

// ScanStringSlice 将data从[]string转换成actual，actual必须是一个slice的指针
func ScanStringSlice(data []string, slice any) error {
	if slice == nil {
		return nil
	}

	v := reflect.ValueOf(slice)
	if !v.IsValid() {
		return fmt.Errorf("redis: scanSlice(nil)")
	}
	if v.Kind() != reflect.Ptr {
		return fmt.Errorf("redis: scanSlice(non-pointer %T)", slice)
	}
	v = v.Elem()
	if v.Kind() != reflect.Slice {
		return fmt.Errorf("redis: scanSlice(non-slice %T)", slice)
	}

	next := makeSliceNextElemFunc(v)
	for i, s := range data {
		elem := next()
		if err := Scan(s, elem.Addr().Interface()); err != nil {
			err = fmt.Errorf("redis: scanSlice index=%d value=%q failed: %w", i, s, err)
			return err
		}
	}

	return nil
}

// makeSliceNextElemFunc 根据slice的reflect.Value，通过反射创建下一个元素
func makeSliceNextElemFunc(v reflect.Value) func() reflect.Value {
	elemType := v.Type().Elem()

	if elemType.Kind() == reflect.Ptr {
		elemType = elemType.Elem()
		return func() reflect.Value {
			if v.Len() < v.Cap() {
				v.Set(v.Slice(0, v.Len()+1))
				elem := v.Index(v.Len() - 1)
				if elem.IsNil() {
					elem.Set(reflect.New(elemType))
				}
				return elem.Elem()
			}

			elem := reflect.New(elemType)
			v.Set(reflect.Append(v, elem))
			return elem.Elem()
		}
	}

	zero := reflect.Zero(elemType)
	return func() reflect.Value {
		if v.Len() < v.Cap() {
			v.Set(v.Slice(0, v.Len()+1))
			return v.Index(v.Len() - 1)
		}

		v.Set(reflect.Append(v, zero))
		return v.Index(v.Len() - 1)
	}
}

// toString 将v转换成[]byte，来源于redis/v8/internal/proto/writer.go
func toString(v any) ([]byte, error) {
	switch v := v.(type) {
	case nil:
		return nil, nil
	case string:
		return []byte(v), nil
	case []byte:
		return v, nil
	case int:
		return []byte(strconv.FormatInt(int64(v), 10)), nil
	case int8:
		return []byte(strconv.FormatInt(int64(v), 10)), nil
	case int16:
		return []byte(strconv.FormatInt(int64(v), 10)), nil
	case int32:
		return []byte(strconv.FormatInt(int64(v), 10)), nil
	case int64:
		return []byte(strconv.FormatInt(v, 10)), nil
	case uint:
		return []byte(strconv.FormatUint(uint64(v), 10)), nil
	case uint8:
		return []byte(strconv.FormatUint(uint64(v), 10)), nil
	case uint16:
		return []byte(strconv.FormatUint(uint64(v), 10)), nil
	case uint32:
		return []byte(strconv.FormatUint(uint64(v), 10)), nil
	case uint64:
		return []byte(strconv.FormatUint(v, 10)), nil
	case float32:
		return []byte(strconv.FormatFloat(float64(v), 'f', -1, 32)), nil
	case float64:
		return []byte(strconv.FormatFloat(v, 'f', -1, 64)), nil
	case bool:
		if v {
			return []byte("1"), nil
		}
		return []byte("0"), nil
	case time.Time:
		return []byte(v.Format(time.RFC3339Nano)), nil
	case time.Duration:
		return []byte(strconv.FormatInt(v.Nanoseconds(), 10)), nil
	case encoding.BinaryMarshaler:
		return v.MarshalBinary()
	default:
		return nil, fmt.Errorf(
			"redis: can't marshal %T (implement encoding.BinaryMarshaler)", v)
	}
}
