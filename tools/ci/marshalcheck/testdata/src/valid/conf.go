package valid

type TargetRef string

// TestPolicy
type TestPolicy struct {
    TargetRef *TargetRef `json:"targetRef,omitempty"` // OK
    To *[]To `json:"to,omitempty"` // OK
    From *[]From `json:"from,omitempty"` // OK
}

type To struct {
    TargetRef *TargetRef `json:"targetRef,omitempty"` // OK
    Default *Conf `json:"default,omitempty"` // OK
}

type From struct {
    TargetRef *TargetRef `json:"targetRef,omitempty"` // OK
    Default *Conf `json:"default,omitempty"` // OK
}


type Conf struct {
    ValidPtr  *string `json:"valid_ptr,omitempty"`    // OK
    ValidList *[]string `json:"valid_list,omitempty"` // OK
}

type OtherStruct struct {
    NonMergeablePtr *string `json:"non_mergeable,omitempty"` // OK
    NonMergeableList []string `json:"list"`                 // OK
}
