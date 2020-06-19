package proto

import (
	"github.com/golang/protobuf/proto"
	"github.com/golang/protobuf/ptypes/any"
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
