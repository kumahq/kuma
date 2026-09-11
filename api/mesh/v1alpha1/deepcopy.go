package v1alpha1

import (
	"bytes"
	"encoding/json"
)

// deepCopy round trips through the JSON form, which is the encoding these specs are
// defined by, so a copy cannot drift from what would be stored.
func deepCopy[T any](in *T) *T {
	if in == nil {
		return nil
	}
	encoded, err := json.Marshal(in)
	if err != nil {
		return nil
	}
	out := new(T)
	if err := json.Unmarshal(encoded, out); err != nil {
		return nil
	}
	return out
}

func (d *Dataplane) DeepCopy() *Dataplane {
	return deepCopy(d)
}

func (i *DataplaneInsight) DeepCopy() *DataplaneInsight {
	return deepCopy(i)
}

func (o *DataplaneOverview) DeepCopy() *DataplaneOverview {
	return deepCopy(o)
}

func (o *DataplaneInsight_OpenTelemetry) DeepCopy() *DataplaneInsight_OpenTelemetry {
	return deepCopy(o)
}

func (v *Version) DeepCopy() *Version {
	return deepCopy(v)
}

// equal compares the JSON form, so a nil map and an empty one compare equal the way
// proto.Equal treated them.
func equal[T any](left, right *T) bool {
	leftBytes, err := json.Marshal(left)
	if err != nil {
		return false
	}
	rightBytes, err := json.Marshal(right)
	if err != nil {
		return false
	}
	return bytes.Equal(leftBytes, rightBytes)
}

func (o *DataplaneInsight_OpenTelemetry) Equal(other *DataplaneInsight_OpenTelemetry) bool {
	return equal(o, other)
}

func (d *Dataplane) Equal(other *Dataplane) bool {
	return equal(d, other)
}

func (i *DataplaneInsight) Equal(other *DataplaneInsight) bool {
	return equal(i, other)
}

func (s *DiscoverySubscription) Equal(other *DiscoverySubscription) bool {
	return equal(s, other)
}

func (s *DiscoverySubscription) DeepCopy() *DiscoverySubscription {
	return deepCopy(s)
}
