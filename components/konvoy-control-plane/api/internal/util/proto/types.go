package proto

import (
	"time"

	"github.com/gogo/protobuf/types"
)

func MustTimestampProto(t time.Time) *types.Timestamp {
	ts, err := types.TimestampProto(t)
	if err != nil {
		panic(err.Error())
	}
	return ts
}
