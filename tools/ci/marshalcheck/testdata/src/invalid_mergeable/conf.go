package invalid_mergeable

type Conf struct {
    MissingOmitEmpty *string `json:"missing_omit"` // want "field Conf.MissingOmitEmpty mergeable field must have 'omitempty' in JSON tag"

    InvalidList []string `json:"invalid_list"` // want "field Conf.InvalidList \\(mergeable list\\) must be a pointer to a slice \\(e.g., \\*\\[\\]T\\)"

    InvalidListNoOmitEmpty *[]string `json:"invalid_list_no_omit"` // want "field Conf.InvalidListNoOmitEmpty \\(mergeable list\\) must have 'omitempty' in JSON tag"

    ValidList *[]string `json:"valid_list,omitempty"` // OK
}
