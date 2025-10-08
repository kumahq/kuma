package generator

import (
	"fmt"
	"reflect"

	"github.com/pkg/errors"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/reflect/protoregistry"
	"google.golang.org/protobuf/types/dynamicpb"
)

func OneofWrapperTypes(parentT reflect.Type, oneofField reflect.StructField) (map[string]reflect.Type, error) {
	md, ok := protoDescFromType(parentT)
	if !ok {
		return nil, errors.Errorf("not a protobuf message: %v", parentT)
	}
	oneofTag, ok := oneofField.Tag.Lookup("protobuf_oneof")
	if !ok {
		return nil, errors.New("protobuf_oneof tag is not set")
	}
	od := md.Oneofs().ByName(protoreflect.Name(oneofTag))
	if od == nil || od.IsSynthetic() {
		return nil, errors.Errorf("oneof %q missing or synthetic (proto3 optional)", oneofTag)
	}

	out := map[string]reflect.Type{}
	fs := od.Fields()
	for i := range fs.Len() {
		fd := fs.Get(i)

		// Fresh parent each loop => no need to clear anything
		pm := newParent(parentT).ProtoReflect()

		val, err := zeroValueFor(fd) // benign value for this alternative
		if err != nil {
			return nil, fmt.Errorf("build value for %s: %w", fd.FullName(), err)
		}
		pm.Set(fd, val) // sets this oneof alternative (and only this one)

		// Read the Go oneof interface field value and capture its concrete type
		parent := pm.Interface() // proto.Message; underlying *T
		rv := reflect.ValueOf(parent).Elem()
		wrap := rv.FieldByName(oneofField.Name).Interface() // e.g. &pb.DataSource_File{}
		out[string(fd.Name())] = reflect.TypeOf(wrap).Elem()
	}
	return out, nil
}

func protoDescFromType(t reflect.Type) (protoreflect.MessageDescriptor, bool) {
	for t.Kind() == reflect.Pointer {
		t = t.Elem()
	}
	if t.Kind() != reflect.Struct {
		return nil, false
	}
	pt := reflect.PointerTo(t)
	m, ok := reflect.Zero(pt).Interface().(proto.Message)
	if !ok {
		return nil, false
	}
	return m.ProtoReflect().Descriptor(), true
}

func newParent(parentT reflect.Type) proto.Message {
	for parentT.Kind() == reflect.Pointer {
		parentT = parentT.Elem()
	}
	return reflect.New(parentT).Interface().(proto.Message)
}

func zeroValueFor(fd protoreflect.FieldDescriptor) (protoreflect.Value, error) {
	switch fd.Kind() {
	case protoreflect.BoolKind:
		return protoreflect.ValueOfBool(false), nil
	case protoreflect.Int32Kind, protoreflect.Sint32Kind, protoreflect.Sfixed32Kind:
		return protoreflect.ValueOfInt32(0), nil
	case protoreflect.Int64Kind, protoreflect.Sint64Kind, protoreflect.Sfixed64Kind:
		return protoreflect.ValueOfInt64(0), nil
	case protoreflect.Uint32Kind, protoreflect.Fixed32Kind:
		return protoreflect.ValueOfUint32(0), nil
	case protoreflect.Uint64Kind, protoreflect.Fixed64Kind:
		return protoreflect.ValueOfUint64(0), nil
	case protoreflect.FloatKind:
		return protoreflect.ValueOfFloat32(0), nil
	case protoreflect.DoubleKind:
		return protoreflect.ValueOfFloat64(0), nil
	case protoreflect.StringKind:
		return protoreflect.ValueOfString(""), nil
	case protoreflect.BytesKind:
		return protoreflect.ValueOfBytes(nil), nil
	case protoreflect.EnumKind:
		evs := fd.Enum().Values()
		var num protoreflect.EnumNumber
		if evs.Len() > 0 {
			num = evs.Get(0).Number()
		}
		return protoreflect.ValueOfEnum(num), nil
	case protoreflect.MessageKind, protoreflect.GroupKind:
		if mt, err := protoregistry.GlobalTypes.FindMessageByName(fd.Message().FullName()); err == nil {
			return protoreflect.ValueOfMessage(mt.New()), nil
		}
		return protoreflect.ValueOfMessage(dynamicpb.NewMessage(fd.Message())), nil
	default:
		return protoreflect.Value{}, fmt.Errorf("unsupported kind: %v", fd.Kind())
	}
}
