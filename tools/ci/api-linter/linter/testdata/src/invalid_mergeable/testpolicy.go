package invalid_mergeable

// TestPolicy
type TestPolicy struct {
    Conf *Conf `json:"conf,omitempty"` // OK
}

type Conf struct {
    MissingOmitEmpty *string `json:"missing_omit"` // want "field TestPolicy.Conf.MissingOmitEmpty does not match any allowed non-mergeable category"

    InvalidList []string `json:"invalid_list,omitempty"` // want "field TestPolicy.Conf.InvalidList does not match any allowed non-mergeable category"

    InvalidListNoOmitEmpty *[]string `json:"invalid_list_no_omit"` // want "field TestPolicy.Conf.InvalidListNoOmitEmpty does not match any allowed non-mergeable category"

    A, B int // want "field must have exactly one name"
    NestedType // want "field must have exactly one name"
}

type NestedType struct {
    C int `json:"c"` // OK
}
