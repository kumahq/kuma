package invalid_nonmergeable

type OtherStruct struct {
    // Non-Mergeable, User Optional, With Default
    // Must not be a pointer, must have "default", must not have "omitempty"
    // +kubebuilder:default="some-value"
    UserOptionalWithDefault *string `json:"user_optional_with_default"` // want "field OtherStruct.UserOptionalWithDefault does not match any allowed non-mergeable category"

    // Non-Mergeable, User Required
    // Must not have "omitempty", must not have "default", must not be a pointer
    UserRequiredPointer *string `json:"user_required,omitempty"` // want "field OtherStruct.UserRequiredPointer does not match any allowed non-mergeable category"

    // Non-Mergeable, User Required
    // Must not have "omitempty", must not have "default", must not be a pointer
    // +kubebuilder:default="some-value"
    UserRequiredDefault string `json:"user_required,omitempty"` // want "field OtherStruct.UserRequiredDefault does not match any allowed non-mergeable category"
}
