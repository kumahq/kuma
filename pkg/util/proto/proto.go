package proto

import (
	"bytes"
	"encoding/json"
	"fmt"

	"github.com/ghodss/yaml"
	"github.com/golang/protobuf/jsonpb"
	"github.com/golang/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/types/known/durationpb"
	"google.golang.org/protobuf/types/known/structpb"
)

func FromYAML(content []byte, pb proto.Message) error {
	json, err := yaml.YAMLToJSON(content)
	if err != nil {
		return err
	}
	return FromJSON(json, pb)
}

func ToYAML(pb proto.Message) ([]byte, error) {
	marshaler := &jsonpb.Marshaler{}
	json, err := marshaler.MarshalToString(pb)
	if err != nil {
		return nil, err
	}
	return yaml.JSONToYAML([]byte(json))
}

func ToJSON(pb proto.Message) ([]byte, error) {
	var buf bytes.Buffer
	marshaler := &jsonpb.Marshaler{}
	if err := marshaler.Marshal(&buf, pb); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func FromJSON(content []byte, out proto.Message) error {
	unmarshaler := &jsonpb.Unmarshaler{AllowUnknownFields: true}
	return unmarshaler.Unmarshal(bytes.NewReader(content), out)
}

func ToMap(pb proto.Message) (map[string]interface{}, error) {
	content, err := ToJSON(pb)
	if err != nil {
		return nil, err
	}
	obj := make(map[string]interface{})
	if err := json.Unmarshal(content, &obj); err != nil {
		return nil, err
	}
	return obj, nil
}

func FromMap(in map[string]interface{}, out proto.Message) error {
	content, err := json.Marshal(in)
	if err != nil {
		return err
	}
	return FromJSON(content, out)
}

// Converts loosely typed Struct to strongly typed Message
func ToTyped(protoStruct *structpb.Struct, message proto.Message) error {
	if protoStruct == nil {
		return nil
	}
	configBytes, err := ToJSON(protoStruct)
	if err != nil {
		return err
	}
	if err := FromJSON(configBytes, message); err != nil {
		return err
	}
	return nil
}

// Converts loosely typed Struct to strongly typed Message
func ToStruct(message proto.Message) (*structpb.Struct, error) {
	configBytes, err := ToJSON(message)
	if err != nil {
		return nil, err
	}
	str := &structpb.Struct{}
	if err := FromJSON(configBytes, str); err != nil {
		return nil, err
	}
	return str, nil
}

func MustToStruct(message proto.Message) *structpb.Struct {
	str, err := ToStruct(message)
	if err != nil {
		panic(err)
	}
	return str
}

type MergeFunction func(dst, src protoreflect.Message)
type mergeOptions struct {
	customMergeFn map[protoreflect.FullName]MergeFunction
}
type OptionFn func(options mergeOptions) mergeOptions

func MergeFunctionOptionFn(name protoreflect.FullName, function MergeFunction) OptionFn {
	return func(options mergeOptions) mergeOptions {
		options.customMergeFn[name] = function
		return options
	}
}

// ReplaceMergeFn instead of merging all subfields one by one, takes src and set it to dest
var ReplaceMergeFn MergeFunction = func(dst, src protoreflect.Message) {
	dst.Range(func(fd protoreflect.FieldDescriptor, v protoreflect.Value) bool {
		dst.Clear(fd)
		return true
	})
	src.Range(func(fd protoreflect.FieldDescriptor, v protoreflect.Value) bool {
		dst.Set(fd, v)
		return true
	})
}

func MergeForKuma(dst, src proto.Message) {
	duration := &durationpb.Duration{}
	Merge(dst, src, MergeFunctionOptionFn(duration.ProtoReflect().Descriptor().FullName(), ReplaceMergeFn))
}

// Merge Code of proto.Merge with modifications to support custom types
func Merge(dst, src proto.Message, opts ...OptionFn) {
	mo := mergeOptions{customMergeFn: map[protoreflect.FullName]MergeFunction{}}
	for _, opt := range opts {
		mo = opt(mo)
	}
	mo.mergeMessage(proto.MessageReflect(dst), proto.MessageReflect(src))
}

func (o mergeOptions) mergeMessage(dst, src protoreflect.Message) {
	// The regular proto.mergeMessage would have a fast path method option here.
	// As we want to have exceptions we always use the slow path.
	if !dst.IsValid() {
		panic(fmt.Sprintf("cannot merge into invalid %v message", dst.Descriptor().FullName()))
	}

	src.Range(func(fd protoreflect.FieldDescriptor, v protoreflect.Value) bool {
		switch {
		case fd.IsList():
			o.mergeList(dst.Mutable(fd).List(), v.List(), fd)
		case fd.IsMap():
			o.mergeMap(dst.Mutable(fd).Map(), v.Map(), fd.MapValue())
		case fd.Message() != nil:
			mergeFn, exists := o.customMergeFn[fd.Message().FullName()]
			if exists {
				mergeFn(dst.Mutable(fd).Message(), v.Message())
			} else {
				o.mergeMessage(dst.Mutable(fd).Message(), v.Message())
			}
		case fd.Kind() == protoreflect.BytesKind:
			dst.Set(fd, o.cloneBytes(v))
		default:
			dst.Set(fd, v)
		}
		return true
	})

	if len(src.GetUnknown()) > 0 {
		dst.SetUnknown(append(dst.GetUnknown(), src.GetUnknown()...))
	}
}

func (o mergeOptions) mergeList(dst, src protoreflect.List, fd protoreflect.FieldDescriptor) {
	// Merge semantics appends to the end of the existing list.
	for i, n := 0, src.Len(); i < n; i++ {
		switch v := src.Get(i); {
		case fd.Message() != nil:
			dstv := dst.NewElement()
			o.mergeMessage(dstv.Message(), v.Message())
			dst.Append(dstv)
		case fd.Kind() == protoreflect.BytesKind:
			dst.Append(o.cloneBytes(v))
		default:
			dst.Append(v)
		}
	}
}

func (o mergeOptions) mergeMap(dst, src protoreflect.Map, fd protoreflect.FieldDescriptor) {
	// Merge semantics replaces, rather than merges into existing entries.
	src.Range(func(k protoreflect.MapKey, v protoreflect.Value) bool {
		switch {
		case fd.Message() != nil:
			dstv := dst.NewValue()
			o.mergeMessage(dstv.Message(), v.Message())
			dst.Set(k, dstv)
		case fd.Kind() == protoreflect.BytesKind:
			dst.Set(k, o.cloneBytes(v))
		default:
			dst.Set(k, v)
		}
		return true
	})
}

func (o mergeOptions) cloneBytes(v protoreflect.Value) protoreflect.Value {
	return protoreflect.ValueOfBytes(append([]byte{}, v.Bytes()...))
}
