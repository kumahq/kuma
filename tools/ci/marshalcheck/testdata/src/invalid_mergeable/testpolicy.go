package invalid_mergeable

// TestPolicy
type TestPolicy struct {
    Conf *Conf `json:"conf,omitempty"` // OK
}

type Conf struct {
    MissingOmitEmpty *string `json:"missing_omit"` // want "field TestPolicy.Conf.MissingOmitEmpty does not match any allowed non-mergeable category"

    InvalidList []string `json:"invalid_list,omitempty"` // want "field TestPolicy.Conf.InvalidList does not match any allowed non-mergeable category"

    InvalidListNoOmitEmpty *[]string `json:"invalid_list_no_omit"` // want "field TestPolicy.Conf.InvalidListNoOmitEmpty does not match any allowed non-mergeable category"
}
