package invalid_mergeable

// TestPolicy
type TestPolicy struct {
    Conf *Conf `json:"conf,omitempty"` // OK
}

type Conf struct {
    MissingOmitEmpty *string `json:"missing_omit"` // want "mergeable field TestPolicy.Conf.MissingOmitEmpty must have 'omitempty' in JSON tag"

    InvalidList []string `json:"invalid_list,omitempty"` // want "mergeable field TestPolicy.Conf.InvalidList must be a pointer"

    InvalidListNoOmitEmpty *[]string `json:"invalid_list_no_omit"` // want "mergeable field TestPolicy.Conf.InvalidListNoOmitEmpty must have 'omitempty' in JSON tag"

    A, B int // want "field must have exactly one name"
    NestedType // want "field must have exactly one name"
}

type NestedType struct {
    C int `json:"c"` // OK
}
