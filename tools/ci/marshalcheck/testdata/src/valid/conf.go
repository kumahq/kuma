package valid

type Conf struct {
    ValidPtr  *string `json:"valid_ptr,omitempty"`   // OK
    ValidList []string `json:"valid_list"`           // OK
}

type OtherStruct struct {
    NonMergeablePtr *string `json:"non_mergeable,omitempty"` // OK
    NonMergeableList []string `json:"list"`                 // OK
}
