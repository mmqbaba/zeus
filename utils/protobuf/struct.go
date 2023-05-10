package zprotobuf

import (
	"encoding/json"
	"fmt"
	"reflect"
	"strings"
	"time"

	jsonpb "github.com/golang/protobuf/jsonpb"
	structpb "github.com/golang/protobuf/ptypes/struct"
)

func checkEmbeddedStruct(embedded, target reflect.Type) bool {
	for i := 0; i < embedded.NumField(); i++ {
		field := embedded.Field(i)
		if field.Anonymous && field.Type == target {
			return true
		}
	}
	return false
}

func callStringMethod(value reflect.Value) (string, error) {
	stringMethod := value.MethodByName("String")
	if stringMethod.IsValid() {
		results := stringMethod.Call(nil)
		if len(results) > 0 {
			return fmt.Sprint(results[0]), nil
		}
		return "", fmt.Errorf("未找到结果")
	}
	return "", fmt.Errorf("未找到 String() 方法")
}

func callTimeMarshalText(val reflect.Value) ([]byte, error) {
	method := val.MethodByName("MarshalText")
	if method.IsValid() {
		results := method.Call(nil)

		if !results[1].IsNil() {
			return nil, results[1].Interface().(error)
		}
		return results[0].Bytes(), nil
	}
	return nil, fmt.Errorf("未找到Time MarshalText() 方法")
}

// ObjMarshalToStruct objtype: map/struct; i => json string => *structpb.Struct
func ObjMarshalToStruct(i interface{}) (*structpb.Struct, error) {
	v := reflect.ValueOf(i)

	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}

	if v.Kind() == reflect.Struct || v.Kind() == reflect.Map {
		d, err := json.Marshal(i)
		if err != nil {
			return nil, err
		}
		st := &structpb.Struct{}
		err = jsonpb.UnmarshalString(string(d), st)
		if err != nil {
			return nil, err
		}
		return st, nil
	}

	return nil, fmt.Errorf("unsupported type: %T, it was not map or struct", i)
}

// ObjToStruct objtype: map/struct; use reflect
func ObjToStruct(i interface{}) (*structpb.Struct, error) {
	v := reflect.ValueOf(i)

	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}

	if v.Kind() == reflect.Struct {
		fields := make(map[string]*structpb.Value)
		typ := v.Type()

		for i := 0; i < v.NumField(); i++ {
			field := typ.Field(i)
			if field.PkgPath != "" { // Skip unexported fields
				continue
			}

			t, ok := field.Tag.Lookup("json")
			if ok && len(t) > 0 {
				arr := strings.Split(t, ",")
				if field.Anonymous && field.Type.Kind() == reflect.Struct {
					isInline := false
					for _, v := range arr {
						if v == "inline" {
							isInline = true
							break
						}
					}
					if isInline {
						val := ToValue(v.Field(i).Interface())
						for k, v := range val.GetStructValue().Fields {
							fields[k] = v
						}
						continue
					}
				}

				name := arr[0]
				if len(name) > 0 && 'A' <= name[0] && name[0] <= 'z' {
					if (v.Field(i).Kind() == reflect.Struct && checkEmbeddedStruct(field.Type, reflect.TypeOf(time.Time{}))) || field.Type == reflect.TypeOf(time.Time{}) {
						ret, err := callTimeMarshalText(v.Field(i))
						if err != nil {
							fmt.Printf("callTimeMarshalText(v.Field(i)) err: %s\n", err)
						}
						val := &structpb.Value{Kind: &structpb.Value_StringValue{StringValue: string(ret)}}
						fields[name] = val
					} else {
						val := ToValue(v.Field(i).Interface())
						if val == nil {
							val = &structpb.Value{Kind: &structpb.Value_NullValue{NullValue: 0}}
						}
						fields[name] = val
					}
				} else {
					continue
				}
			}

		}
		return &structpb.Struct{Fields: fields}, nil
	}

	if v.Kind() == reflect.Map {
		fields := make(map[string]*structpb.Value)
		iter := v.MapRange()

		for iter.Next() {
			key := iter.Key().String()
			val := ToValue(iter.Value().Interface())
			if val == nil {
				val = &structpb.Value{Kind: &structpb.Value_NullValue{NullValue: 0}}
			}
			fields[key] = val
		}
		return &structpb.Struct{Fields: fields}, nil
	}

	return nil, fmt.Errorf("unsupported type: %T, it was not map or struct", i)
}

// ToStruct converts a map[string]interface{} to a ptypes.Struct
func ToStruct(v map[string]interface{}) *structpb.Struct {
	size := len(v)
	if size == 0 {
		return nil
	}
	fields := make(map[string]*structpb.Value, size)
	for k, v := range v {
		val := ToValue(v)
		if val == nil {
			val = &structpb.Value{Kind: &structpb.Value_NullValue{NullValue: 0}}
		}
		fields[k] = val
	}
	return &structpb.Struct{
		Fields: fields,
	}
}

// ToValue converts an interface{} to a ptypes.Value
func ToValue(v interface{}) *structpb.Value {
	switch v := v.(type) {
	case nil:
		// return nil
		return &structpb.Value{Kind: &structpb.Value_NullValue{NullValue: 0}}
	case bool:
		return &structpb.Value{
			Kind: &structpb.Value_BoolValue{
				BoolValue: v,
			},
		}
	case int:
		return &structpb.Value{
			Kind: &structpb.Value_NumberValue{
				NumberValue: float64(v),
			},
		}
	case int8:
		return &structpb.Value{
			Kind: &structpb.Value_NumberValue{
				NumberValue: float64(v),
			},
		}
	case int32:
		return &structpb.Value{
			Kind: &structpb.Value_NumberValue{
				NumberValue: float64(v),
			},
		}
	case int64:
		return &structpb.Value{
			Kind: &structpb.Value_NumberValue{
				NumberValue: float64(v),
			},
		}
	case uint:
		return &structpb.Value{
			Kind: &structpb.Value_NumberValue{
				NumberValue: float64(v),
			},
		}
	case uint8:
		return &structpb.Value{
			Kind: &structpb.Value_NumberValue{
				NumberValue: float64(v),
			},
		}
	case uint32:
		return &structpb.Value{
			Kind: &structpb.Value_NumberValue{
				NumberValue: float64(v),
			},
		}
	case uint64:
		return &structpb.Value{
			Kind: &structpb.Value_NumberValue{
				NumberValue: float64(v),
			},
		}
	case float32:
		return &structpb.Value{
			Kind: &structpb.Value_NumberValue{
				NumberValue: float64(v),
			},
		}
	case float64:
		return &structpb.Value{
			Kind: &structpb.Value_NumberValue{
				NumberValue: v,
			},
		}
	case string:
		return &structpb.Value{
			Kind: &structpb.Value_StringValue{
				StringValue: v,
			},
		}
	case error:
		return &structpb.Value{
			Kind: &structpb.Value_StringValue{
				StringValue: v.Error(),
			},
		}
	default:
		// Fallback to reflection for other types
		return toValue(reflect.ValueOf(v))
	}
}

func toValue(v reflect.Value) *structpb.Value {
	switch v.Kind() {
	case reflect.Bool:
		return &structpb.Value{
			Kind: &structpb.Value_BoolValue{
				BoolValue: v.Bool(),
			},
		}
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return &structpb.Value{
			Kind: &structpb.Value_NumberValue{
				NumberValue: float64(v.Int()),
			},
		}
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		return &structpb.Value{
			Kind: &structpb.Value_NumberValue{
				NumberValue: float64(v.Uint()),
			},
		}
	case reflect.Float32, reflect.Float64:
		return &structpb.Value{
			Kind: &structpb.Value_NumberValue{
				NumberValue: v.Float(),
			},
		}
	case reflect.Ptr:
		if v.IsNil() {
			// return nil
			return &structpb.Value{Kind: &structpb.Value_NullValue{NullValue: 0}}
		}
		return toValue(reflect.Indirect(v))
	case reflect.Array, reflect.Slice:
		size := v.Len()
		if size == 0 {
			// return nil
			return &structpb.Value{Kind: &structpb.Value_NullValue{NullValue: 0}}
		}
		values := make([]*structpb.Value, size)
		for i := 0; i < size; i++ {
			values[i] = toValue(v.Index(i))
		}
		return &structpb.Value{
			Kind: &structpb.Value_ListValue{
				ListValue: &structpb.ListValue{
					Values: values,
				},
			},
		}
	case reflect.Struct:
		t := v.Type()
		size := v.NumField()
		if size == 0 {
			// return nil
			return &structpb.Value{Kind: &structpb.Value_NullValue{NullValue: 0}}
		}

		if checkEmbeddedStruct(t, reflect.TypeOf(time.Time{})) || t == reflect.TypeOf(time.Time{}) {
			// if checkEmbeddedStruct(t, reflect.TypeOf(time.Time{})) {
			ret, err := callTimeMarshalText(v)
			if err != nil {
				fmt.Printf("callTimeMarshalText(v) err: %s\n", err)
				return &structpb.Value{Kind: &structpb.Value_NullValue{NullValue: 0}}
			}
			return &structpb.Value{Kind: &structpb.Value_StringValue{StringValue: string(ret)}}
		}

		fields := make(map[string]*structpb.Value, size)
		for i := 0; i < size; i++ {
			// n := t.Field(i).Name
			// tag := t.Field(i).Tag
			// fmt.Println(n, tag)
			tj, ok := t.Field(i).Tag.Lookup("json")
			if ok && len(tj) > 0 {
				arr := strings.Split(tj, ",")
				if t.Field(i).Anonymous && t.Field(i).Type.Kind() == reflect.Struct {
					isInline := false
					for _, v := range arr {
						if v == "inline" {
							isInline = true
							break
						}
					}
					if isInline {
						val := ToValue(v.Field(i).Interface())
						for k, v := range val.GetStructValue().Fields {
							fields[k] = v
						}
						continue
					}
				}
				name := arr[0]
				// name := strings.Split(val, ",")[0]
				// Better way?
				if len(name) > 0 && 'A' <= name[0] && name[0] <= 'z' {
					fields[name] = toValue(v.Field(i))
				}
			}
		}
		if len(fields) == 0 {
			// return nil
			return &structpb.Value{Kind: &structpb.Value_NullValue{NullValue: 0}}
		}
		return &structpb.Value{
			Kind: &structpb.Value_StructValue{
				StructValue: &structpb.Struct{
					Fields: fields,
				},
			},
		}
	case reflect.Map:
		keys := v.MapKeys()
		if len(keys) == 0 {
			// return nil
			return &structpb.Value{Kind: &structpb.Value_NullValue{NullValue: 0}}
		}
		fields := make(map[string]*structpb.Value, len(keys))
		for _, k := range keys {
			if k.Kind() == reflect.String {
				fields[k.String()] = toValue(v.MapIndex(k))
			}
		}
		if len(fields) == 0 {
			// return nil
			return &structpb.Value{Kind: &structpb.Value_NullValue{NullValue: 0}}
		}
		return &structpb.Value{
			Kind: &structpb.Value_StructValue{
				StructValue: &structpb.Struct{
					Fields: fields,
				},
			},
		}
	case reflect.Interface:
		return toValue(v.Elem())
	default:
		// Last resort
		return &structpb.Value{
			Kind: &structpb.Value_StringValue{
				StringValue: fmt.Sprint(v),
			},
		}
	}
}

// DecodeToMap converts a pb.Struct to a map from strings to Go types.
// DecodeToMap panics if s is invalid.
func DecodeToMap(s *structpb.Struct) map[string]interface{} {
	if s == nil {
		return nil
	}
	m := map[string]interface{}{}
	for k, v := range s.Fields {
		m[k] = decodeValue(v)
	}
	return m
}

func decodeValue(v *structpb.Value) interface{} {
	switch k := v.Kind.(type) {
	case *structpb.Value_NullValue:
		return nil
	case *structpb.Value_NumberValue:
		return k.NumberValue
	case *structpb.Value_StringValue:
		return k.StringValue
	case *structpb.Value_BoolValue:
		return k.BoolValue
	case *structpb.Value_StructValue:
		return DecodeToMap(k.StructValue)
	case *structpb.Value_ListValue:
		s := make([]interface{}, len(k.ListValue.Values))
		for i, e := range k.ListValue.Values {
			s[i] = decodeValue(e)
		}
		return s
	default:
		panic("protostruct: unknown kind")
	}
}
