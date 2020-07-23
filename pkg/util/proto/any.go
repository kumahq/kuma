package proto

import (
	"reflect"
	"strings"

	"github.com/golang/protobuf/proto"
	"github.com/golang/protobuf/ptypes/any"
	"github.com/pkg/errors"
)

const googleApis = "type.googleapis.com/"

// When saving Snapshot in SnapshotCache we generate version based on proto.Equal()
// Therefore we need deterministic way of marshalling Any which is part of the Protobuf on which we execute Equal()
//
// Based on proto.MarshalAny
func MarshalAnyDeterministic(pb proto.Message) (*any.Any, error) {
	bytes := make([]byte, 0, proto.Size(pb))
	buffer := proto.NewBuffer(bytes)
	buffer.SetDeterministic(true)
	if err := buffer.Marshal(pb); err != nil {
		return nil, err
	}
	return &any.Any{TypeUrl: googleApis + proto.MessageName(pb), Value: buffer.Bytes()}, nil
}

// MergeAnys merges two Any messages of the same type. We cannot just use proto#Merge on Any directly because values are encoded in byte slices.
// Instead we have to unmarshal types, merge them and marshal again.
func MergeAnys(dst *any.Any, src *any.Any) (*any.Any, error) {
	if src == nil {
		return dst, nil
	}
	if dst == nil {
		return src, nil
	}
	if src.TypeUrl != dst.TypeUrl {
		return nil, errors.Errorf("type URL of dst %q is different than src %q", dst.TypeUrl, src.TypeUrl)
	}

	msgTypeName := strings.ReplaceAll(dst.TypeUrl, googleApis, "") // TypeURL in Any contains type.googleapis.com/ prefix, but in Proto registry it does not have this prefix.
	msgType := proto.MessageType(msgTypeName).Elem()

	dstMsg := reflect.New(msgType).Interface().(proto.Message)
	if err := proto.Unmarshal(dst.Value, dstMsg); err != nil {
		return nil, err
	}

	srcMsg := reflect.New(msgType).Interface().(proto.Message)
	if err := proto.Unmarshal(src.Value, srcMsg); err != nil {
		return nil, err
	}

	proto.Merge(dstMsg, srcMsg)
	return MarshalAnyDeterministic(dstMsg)
}
