package proto

import (
	"github.com/gogo/protobuf/types"
	"time"
)

func MustTimestampProto(t time.Time) *types.Timestamp {
	ts, err := types.TimestampProto(t)
	if err != nil {
		panic(err.Error())
	}
	return ts
}
