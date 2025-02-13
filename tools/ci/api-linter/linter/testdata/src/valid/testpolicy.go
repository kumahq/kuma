package valid

type TargetRef string

// TestPolicy
type TestPolicy struct {
    TargetRef   *TargetRef   `json:"targetRef,omitempty"`    // OK
    To          *[]To        `json:"to,omitempty"`           // OK
    From        *[]From      `json:"from,omitempty"`         // OK
    OtherStruct *OtherStruct `json:"other_struct,omitempty"` // OK
}

type To struct {
    TargetRef *TargetRef `json:"targetRef,omitempty"` // OK
    Default   *Conf      `json:"default,omitempty"`   // OK
}

type From struct {
    TargetRef *TargetRef `json:"targetRef,omitempty"` // OK
    Default   *Conf      `json:"default,omitempty"`   // OK
}

type Conf struct {
    ValidPtr  *string   `json:"valid_ptr,omitempty"`  // OK
    ValidList *[]string `json:"valid_list,omitempty"` // OK
    // +kuma:non-mergeable-struct
    NonMergeableStruct NonMergeableStruct `json:"non_mergeable_struct"` // OK

    Discriminator Discriminator `json:"discriminator"` // OK
}

type Discriminator struct {
    // +kuma:discriminator
    Type string `json:"type"` // OK

    OptionOne *OptionOne `json:"option_one,omitempty"` // OK
    OptionTwo *OptionTwo `json:"option_two,omitempty"` // OK
}

type OptionOne struct {
    OptionOneField *string `json:"option_one_field,omitempty"` // OK
}

type OptionTwo struct {
    OptionTwoField *string `json:"option_two_field,omitempty"` // OK
}

type NonMergeableStruct struct {
    RequiredIntField int `json:"required_int_field"` // OK
    RequiredStrField string `json:"required_str_field"` // OK
}

type OtherStruct struct {
    NonMergeableRequired string   `json:"non_mergeable"`           // OK
    // +kubebuilder:validation:Optional
    // +kubebuilder:default=false
    NonMergeableOptional string   `json:"non_mergeable_optional"`  // OK
    NonMergeablePtr      *string  `json:"non_mergeable,omitempty"` // OK
    NonMergeableList     []string `json:"list"`                    // OK
}
