package v1alpha1

import (
	"encoding/json"
	"time"
)

// timeFormat matches what protobuf's JSON mapping emits for a Timestamp: RFC 3339 in
// UTC, with the fractional part trimmed of trailing zeroes.
const timeFormat = "2006-01-02T15:04:05.999999999Z"

func NewTime(t time.Time) *Time {
	return &Time{Time: t}
}

func (t Time) MarshalJSON() ([]byte, error) {
	return json.Marshal(t.Time.UTC().Format(timeFormat))
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
