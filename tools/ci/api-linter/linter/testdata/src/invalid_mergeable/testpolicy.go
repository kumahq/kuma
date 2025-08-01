package invalid_mergeable

// TestPolicy
type TestPolicy struct {
    Default *Conf `json:"conf,omitempty"` // OK
}

type Conf struct {
    MissingOmitEmpty *string `json:"missing_omit"` // want "mergeable field TestPolicy.Default.MissingOmitEmpty must have 'omitempty' in JSON tag"

    InvalidList []string `json:"invalid_list,omitempty"` // want "mergeable field TestPolicy.Default.InvalidList must be a pointer"

    InvalidListNoOmitEmpty *[]string `json:"invalid_list_no_omit"` // want "mergeable field TestPolicy.Default.InvalidListNoOmitEmpty must have 'omitempty' in JSON tag"

    A, B int // want "field must have exactly one name"
    NestedType // want "field must have exactly one name"

    NestedTypeWithoutPointer NestedTypeWithoutPointer `json:"nested_type_with_invalid_pointer,omitempty"` // want "mergeable field TestPolicy.Default.NestedTypeWithoutPointer must be a pointer"
}

type NestedType struct {
    C int `json:"c"` // OK
}

type NestedTypeWithoutPointer struct {
    D *int `json:"d,omitempty"` // OK
}
