package proto

import (
	"fmt"
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

func MustTimestampFromProto(ts *tspb.Timestamp) *time.Time {
	if ts == nil {
		return nil
	}
	t, err := ptypes.Timestamp(ts)
	if err != nil {
		panic(err.Error())
	}
	return &t
}

func TimestampString(ts *tspb.Timestamp, layout string) string {
	t, err := ptypes.Timestamp(ts)
	if err != nil {
		return fmt.Sprintf("(%v)", err)
	}
	return t.Format(layout)
}
