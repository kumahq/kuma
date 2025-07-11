package valid

type TargetRef string

// TestPolicy
type TestPolicy struct {
    TargetRef   *TargetRef   `json:"targetRef,omitempty"`    // OK
    To          *[]To        `json:"to,omitempty"`           // OK
}

type To struct {
    TargetRef *TargetRef `json:"targetRef,omitempty"` // OK
    Default   *Conf      `json:"default,omitempty"`   // OK
}

type Conf struct {
    Uint32Pointer *uint32 `json:"invalid_type_ptr,omitempty"` // want "field TestPolicy.To\\[\\].Default.Uint32Pointer must be either int32 or int64"
    // +kuma:non-mergeable-struct
    Some *Some `json:"some,omitempty"` // OK
}

type Some struct{
    Uint32        uint32  `json:"invalid_type"`               // want "field TestPolicy.To\\[\\].Default.Some.Uint32 must be either int32 or int64"
}
