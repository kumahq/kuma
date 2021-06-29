package proto

import (
	"time"

	tspb "github.com/golang/protobuf/ptypes/timestamp"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func MustTimestampProto(t time.Time) *tspb.Timestamp {
	return timestamppb.New(t)
}

func MustTimestampFromProto(ts *tspb.Timestamp) *time.Time {
	if ts == nil {
		return nil
	}
	if err := ts.CheckValid(); err != nil {
		panic(err.Error())
	}
	t := ts.AsTime()
	return &t
}
