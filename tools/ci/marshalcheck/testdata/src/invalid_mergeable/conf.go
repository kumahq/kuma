package invalid_mergeable

type Conf struct {
    MissingOmitEmpty *string `json:"missing_omit"` // want "mergeable field Conf.MissingOmitEmpty must have 'omitempty' in JSON tag"

    InvalidList []string `json:"invalid_list,omitempty"` // want "mergeable field Conf.InvalidList must be a pointer"

    InvalidListNoOmitEmpty *[]string `json:"invalid_list_no_omit"` // want "mergeable field Conf.InvalidListNoOmitEmpty must have 'omitempty' in JSON tag"
}
