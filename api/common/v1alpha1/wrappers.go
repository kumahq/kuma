package v1alpha1

type UInt32Value struct {
	// The uint32 value.
	Value uint32 `json:"value,omitempty"`
}

func (x *UInt32Value) GetValue() uint32 {
	if x != nil {
		return x.Value
	}
	return 0
}

type BoolValue struct {
	// The bool value.
	Value bool `json:"value,omitempty"`
}

func (x *BoolValue) GetValue() bool {
	if x != nil {
		return x.Value
	}
	return false
}
