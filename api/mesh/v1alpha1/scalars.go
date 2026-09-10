package v1alpha1

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
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

// The range protobuf's Timestamp allowed: 0001-01-01T00:00:00Z through
// 9999-12-31T23:59:59.999999999Z.
const (
	minTimestampSeconds = -62135596800
	maxTimestampSeconds = 253402300800
)

// Time carries what protobuf held in a Timestamp.
type Time struct {
	time.Time `json:"-"`
}

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

// CheckValid reports the range protobuf's Timestamp accepted, so a caller that used to
// be handed an error for an out of range instant still is.
func (t *Time) CheckValid() error {
	if t == nil {
		return fmt.Errorf("invalid nil Timestamp")
	}
	seconds := t.Time.Unix()
	if seconds < minTimestampSeconds || seconds >= maxTimestampSeconds {
		return fmt.Errorf("timestamp (%v) is out of range", t.Time)
	}
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

// Duration carries what protobuf held in a Duration. Protobuf writes it as the number of
// seconds with up to nine fractional digits and a trailing "s", trimmed to zero, three,
// six or nine digits the way a timestamp is.
type Duration struct {
	time.Duration `json:"-"`
}

func NewDuration(d time.Duration) *Duration {
	return &Duration{Duration: d}
}

func (d Duration) MarshalJSON() ([]byte, error) {
	seconds := int64(d.Duration / time.Second)
	nanos := int64(d.Duration % time.Second)

	sign := ""
	if seconds < 0 || nanos < 0 {
		sign = "-"
		seconds, nanos = -seconds, -nanos
	}

	switch {
	case nanos == 0:
		return json.Marshal(fmt.Sprintf("%s%ds", sign, seconds))
	case nanos%1e6 == 0:
		return json.Marshal(fmt.Sprintf("%s%d.%03ds", sign, seconds, nanos/1e6))
	case nanos%1e3 == 0:
		return json.Marshal(fmt.Sprintf("%s%d.%06ds", sign, seconds, nanos/1e3))
	default:
		return json.Marshal(fmt.Sprintf("%s%d.%09ds", sign, seconds, nanos))
	}
}

func (d *Duration) UnmarshalJSON(data []byte) error {
	var raw string
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}
	if raw == "" {
		d.Duration = 0
		return nil
	}
	// jsonpb accepted a bare number as well as the seconds suffix, and stored dataplanes
	// carry both spellings.
	seconds, err := strconv.ParseFloat(strings.TrimSuffix(raw, "s"), 64)
	if err != nil {
		return fmt.Errorf("duration %q: %w", raw, err)
	}
	d.Duration = time.Duration(seconds * float64(time.Second))
	return nil
}

func (d *Duration) DeepCopyInto(out *Duration) {
	*out = *d
}

func (d *Duration) DeepCopy() *Duration {
	if d == nil {
		return nil
	}
	out := new(Duration)
	d.DeepCopyInto(out)
	return out
}

// AsDuration is nil safe for callers that read a duration that was never set.
func (d *Duration) AsDuration() time.Duration {
	if d == nil {
		return 0
	}
	return d.Duration
}

// UInt32Value keeps the distinction protobuf's UInt32Value drew between a field never set
// and one holding zero: the first is absent from the document, the second is a plain 0.
type UInt32Value struct {
	Value uint32
}

func NewUInt32(value uint32) *UInt32Value {
	return &UInt32Value{Value: value}
}

func (u *UInt32Value) GetValue() uint32 {
	if u == nil {
		return 0
	}
	return u.Value
}

func (u UInt32Value) MarshalJSON() ([]byte, error) {
	return json.Marshal(u.Value)
}

func (u *UInt32Value) UnmarshalJSON(data []byte) error {
	return json.Unmarshal(data, &u.Value)
}

func (u *UInt32Value) DeepCopyInto(out *UInt32Value) {
	*out = *u
}

func (u *UInt32Value) DeepCopy() *UInt32Value {
	if u == nil {
		return nil
	}
	out := new(UInt32Value)
	u.DeepCopyInto(out)
	return out
}
