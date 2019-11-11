package proto

import (
	"time"

	"github.com/golang/protobuf/ptypes"
	tspb "github.com/golang/protobuf/ptypes/timestamp"
)

func MustTimestampProto(t time.Time) *tspb.Timestamp {
	ts, err := ptypes.TimestampProto(t)
	if err != nil {
		panic(err.Error())
	}
	return ts
}
