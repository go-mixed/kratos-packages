package utils

import (
	"reflect"
	"strconv"
	"strings"
)

// FindReflectValue finds a nested value by a dot-separated path.
//
// Struct fields, map keys, and slice or array indexes can be mixed in the
// same path. Struct fields use their Go field names, map keys use their
// textual representation, and slice or array indexes are zero-based integers.
//
// Example:
//
//	type User struct {
//		Name string
//	}
//	data := map[string]any{
//		"Users": []User{{Name: "Ada"}},
//	}
//	value, ok := FindReflectValue(reflect.ValueOf(data), "Users.0.Name")
//	if ok {
//		name := value.String() // "Ada"
//		_ = name
//	}
func FindReflectValue(value reflect.Value, path string) (reflect.Value, bool) {
	if path == "" {
		value = UnwrapReflectValue(value)
		return value, value.IsValid()
	}

	for _, part := range strings.Split(path, ".") {
		if part == "" {
			return reflect.Value{}, false
		}

		value = UnwrapReflectValue(value)
		if !value.IsValid() {
			return reflect.Value{}, false
		}

		switch value.Kind() {
		case reflect.Struct:
			var found bool
			value, found = FindStructField(value, part, nil)
			if !found {
				return reflect.Value{}, false
			}
		case reflect.Map:
			key, ok := reflectMapKey(part, value.Type().Key())
			if !ok {
				return reflect.Value{}, false
			}
			value = value.MapIndex(key)
			if !value.IsValid() {
				return reflect.Value{}, false
			}
		case reflect.Array, reflect.Slice:
			index, err := strconv.Atoi(part)
			if err != nil || index < 0 || index >= value.Len() {
				return reflect.Value{}, false
			}
			value = value.Index(index)
		default:
			return reflect.Value{}, false
		}
	}

	value = UnwrapReflectValue(value)
	return value, value.IsValid()
}

// reflectMapKey converts one path segment to a supported map key type.
func reflectMapKey(value string, keyType reflect.Type) (reflect.Value, bool) {
	key := reflect.New(keyType).Elem()
	switch keyType.Kind() {
	case reflect.String:
		key.SetString(value)
	case reflect.Bool:
		parsed, err := strconv.ParseBool(value)
		if err != nil {
			return reflect.Value{}, false
		}
		key.SetBool(parsed)
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		parsed, err := strconv.ParseInt(value, 10, keyType.Bits())
		if err != nil {
			return reflect.Value{}, false
		}
		key.SetInt(parsed)
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		parsed, err := strconv.ParseUint(value, 10, keyType.Bits())
		if err != nil {
			return reflect.Value{}, false
		}
		key.SetUint(parsed)
	default:
		return reflect.Value{}, false
	}
	return key, true
}

// FindStructField finds an exported struct field by Go field name.
//
// Direct fields take precedence over promoted fields. If the field is not
// present directly, anonymous fields are searched recursively, including
// anonymous pointer and interface fields. The seen map prevents recursive
// anonymous types from causing an infinite loop; callers can pass nil when
// starting a new lookup.
//
// Example:
//
//	type Profile struct {
//		Name string
//	}
//	type User struct {
//		Profile
//	}
//
//	user := User{Profile: Profile{Name: "Ada"}}
//	field, ok := FindStructField(reflect.ValueOf(user), "Name", nil)
//	if ok {
//		name := field.String() // "Ada"
//		_ = name
//	}
func FindStructField(value reflect.Value, name string, seen map[reflect.Type]bool) (reflect.Value, bool) {
	value = UnwrapReflectValue(value)
	if !value.IsValid() || value.Kind() != reflect.Struct {
		return reflect.Value{}, false
	}
	if seen == nil {
		seen = make(map[reflect.Type]bool)
	}
	if seen[value.Type()] {
		return reflect.Value{}, false
	}
	seen[value.Type()] = true

	typeInfo := value.Type()
	// Search direct fields first. FieldByName also considers promoted fields,
	// which could bypass the recursive nil-safe traversal below.
	for i := 0; i < value.NumField(); i++ {
		field := typeInfo.Field(i)
		if field.Name == name && field.PkgPath == "" {
			result := value.Field(i)
			if result.CanInterface() {
				return result, true
			}
		}
	}

	// Search anonymous fields recursively to support promoted fields while
	// safely handling nil embedded pointers and interfaces.
	for i := 0; i < value.NumField(); i++ {
		fieldInfo := typeInfo.Field(i)
		if !fieldInfo.Anonymous {
			continue
		}
		field, ok := FindStructField(value.Field(i), name, seen)
		if ok {
			return field, true
		}
	}
	return reflect.Value{}, false
}

// UnwrapReflectValue recursively dereferences interface and pointer values.
// It returns an invalid reflect.Value when a nil interface or pointer is
// encountered, which lets callers handle nil values without panicking.
func UnwrapReflectValue(value reflect.Value) reflect.Value {
	for value.IsValid() && (value.Kind() == reflect.Interface || value.Kind() == reflect.Pointer) {
		if value.IsNil() {
			return reflect.Value{}
		}
		value = value.Elem()
	}
	return value
}
