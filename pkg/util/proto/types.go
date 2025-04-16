package proto

import (
	"encoding/json"
	"reflect"
	"time"

	"google.golang.org/protobuf/types/known/durationpb"
	"google.golang.org/protobuf/types/known/structpb"
	"google.golang.org/protobuf/types/known/timestamppb"
	"google.golang.org/protobuf/types/known/wrapperspb"
)

func Now() *timestamppb.Timestamp {
	return timestamppb.Now()
}

func MustTimestampProto(t time.Time) *timestamppb.Timestamp {
	ts := timestamppb.New(t)

	if err := ts.CheckValid(); err != nil {
		panic(err.Error())
	}

	return ts
}

func MustTimestampFromProto(ts *timestamppb.Timestamp) *time.Time {
	if ts == nil {
		return nil
	}

	if err := ts.CheckValid(); err != nil {
		panic(err.Error())
	}

	t := ts.AsTime()
	return &t
}

func Bool(b bool) *wrapperspb.BoolValue {
	return &wrapperspb.BoolValue{Value: b}
}

func Bytes(b []byte) *wrapperspb.BytesValue {
	return &wrapperspb.BytesValue{Value: b}
}

func String(s string) *wrapperspb.StringValue {
	return &wrapperspb.StringValue{Value: s}
}

func UInt32(u uint32) *wrapperspb.UInt32Value {
	return &wrapperspb.UInt32Value{Value: u}
}

func UInt64(u uint64) *wrapperspb.UInt64Value {
	return &wrapperspb.UInt64Value{Value: u}
}

func Double(f float64) *wrapperspb.DoubleValue {
	return &wrapperspb.DoubleValue{Value: f}
}

func Duration(d time.Duration) *durationpb.Duration {
	return durationpb.New(d)
}

func Struct(in map[string]interface{}) (*structpb.Struct, error) {
	return structpb.NewStruct(in)
}

func MustStruct(in map[string]interface{}) *structpb.Struct {
	r, err := Struct(in)
	if err != nil {
		panic(err.Error())
	}
	return r
}

func NewValueForStruct(in interface{}) (*structpb.Value, error) {
	return structpb.NewValue(in)
}

func MustNewValueForStruct(in interface{}) *structpb.Value {
	r, err := NewValueForStruct(in)
	if err != nil {
		panic(err.Error())
	}
	return r
}

func StructToMapOfAny(value any) map[string]any {
	var v reflect.Value

	switch val := value.(type) {
	case reflect.Value:
		v = val
	default:
		v = reflect.ValueOf(value)
	}

	if v.Kind() == reflect.Pointer {
		v = v.Elem()
	}

	out := make(map[string]any)

	t := v.Type()
	for i := 0; i < v.NumField(); i++ {
		field := t.Field(i)
		if field.PkgPath != "" {
			continue
		}
		key := field.Tag.Get("json")
		if key == "" {
			key = field.Name
		}
		val := v.Field(i)

		switch val.Kind() {
		case reflect.Struct:
			out[key] = StructToMapOfAny(val)
		case reflect.Bool:
			out[key] = val.Bool()
		case reflect.String:
			out[key] = val.String()
		case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
			out[key] = float64(val.Uint())
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			out[key] = float64(val.Int())
		default:
			out[key] = val.Interface()
		}
	}

	return out
}

func StructToProtoStruct(in any) (*structpb.Struct, error) {
	return structpb.NewStruct(StructToMapOfAny(in))
}

func MustStructToProtoStruct(in any) *structpb.Struct {
	r, err := StructToProtoStruct(in)
	if err != nil {
		panic(err.Error())
	}
	return r
}

func MustFromMapOfAny[T any, U map[string]any | *structpb.Struct](in U) T {
	out, _ := FromMapOfAny[T](in)
	return out
}

func FromMapOfAny[T any, U map[string]any | *structpb.Struct](value U) (T, error) {
	var result T
	var in map[string]any

	switch val := any(value).(type) {
	case *structpb.Struct:
		in = val.AsMap()
	case map[string]any:
		in = val
	default:
		return result, nil
	}

	// encode to JSON using structpb-compatible values
	b, err := json.Marshal(convertToPlainJSON(in))
	if err != nil {
		return result, err
	}

	if err := json.Unmarshal(b, &result); err != nil {
		return result, err
	}

	return result, nil
}

func convertToPlainJSON(v any) any {
	switch v := v.(type) {
	case map[string]any:
		out := map[string]any{}
		for k, val := range v {
			out[k] = convertToPlainJSON(val)
		}
		return out
	case []any:
		for i := range v {
			v[i] = convertToPlainJSON(v[i])
		}
	}
	return v
}
