package v1alpha1

import "encoding/json"

// BytesValue keeps the distinction protobuf's BytesValue drew between a field that was
// never set and one holding no bytes: the first is absent from the document, the second
// is an empty string. Protobuf wrote the wrapper as a bare base64 scalar rather than an
// object, which is what encoding/json does with a []byte.
type BytesValue struct {
	Value []byte
}

func (b *BytesValue) GetValue() []byte {
	if b == nil {
		return nil
	}
	return b.Value
}

func (b BytesValue) MarshalJSON() ([]byte, error) {
	return json.Marshal(b.Value)
}

func (b *BytesValue) UnmarshalJSON(data []byte) error {
	return json.Unmarshal(data, &b.Value)
}

func Bytes(value []byte) *BytesValue {
	return &BytesValue{Value: value}
}
