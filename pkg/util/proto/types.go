package proto

import (
	"time"

	"google.golang.org/protobuf/types/known/timestamppb"
)

func MustTimestampProto(t time.Time) *timestamppb.Timestamp {
	return timestamppb.New(t)
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
