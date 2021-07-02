package proto

import (
	"time"

	"google.golang.org/protobuf/types/known/timestamppb"
)

func MustTimestampProto(t time.Time) *timestamppb.Timestamp {
	return timestamppb.New(t)
}
