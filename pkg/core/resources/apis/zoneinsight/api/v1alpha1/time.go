package v1alpha1

import (
	"encoding/json"
	"time"
)

// Protobuf's JSON mapping writes the fractional second in zero, three, six or nine
// digits, never fewer, so a timestamp ending in zeroes keeps them. Go's own ".999"
// formats trim every trailing zero, which would rewrite a stored timestamp into
// different bytes for the same instant.
const (
	timeFormatSeconds = "2006-01-02T15:04:05Z"
	timeFormatMillis  = "2006-01-02T15:04:05.000Z"
	timeFormatMicros  = "2006-01-02T15:04:05.000000Z"
	timeFormatNanos   = "2006-01-02T15:04:05.000000000Z"
)

func NewTime(t time.Time) *Time {
	return &Time{Time: t}
}

func (t Time) MarshalJSON() ([]byte, error) {
	utc := t.Time.UTC()

	format := timeFormatNanos

	switch nanos := utc.Nanosecond(); {
	case nanos == 0:
		format = timeFormatSeconds
	case nanos%1e6 == 0:
		format = timeFormatMillis
	case nanos%1e3 == 0:
		format = timeFormatMicros
	}

	return json.Marshal(utc.Format(format))
}

func (t *Time) UnmarshalJSON(data []byte) error {
	var raw string
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}
	if raw == "" {
		t.Time = time.Time{}
		return nil
	}
	parsed, err := time.Parse(time.RFC3339Nano, raw)
	if err != nil {
		return err
	}
	t.Time = parsed.UTC()
	return nil
}

func (t *Time) DeepCopyInto(out *Time) {
	*out = *t
}

func (t *Time) DeepCopy() *Time {
	if t == nil {
		return nil
	}
	out := new(Time)
	t.DeepCopyInto(out)
	return out
}

// AsTime is nil safe so callers can render a timestamp that was never set.
func (t *Time) AsTime() *time.Time {
	if t == nil {
		return nil
	}
	return &t.Time
}

// OrZero is nil safe for callers that compare instants rather than render them.
func (t *Time) OrZero() time.Time {
	if t == nil {
		return time.Time{}
	}
	return t.Time
}
