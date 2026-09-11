package invalid_nonmergeable

// TestPolicy
type TestPolicy struct {
    OtherStruct *OtherStruct `json:"other_struct,omitempty"` // OK
}

type OtherStruct struct {
    // Non-Mergeable, User Optional, With Default
    // Must not be a pointer, must have "default", must have "optional", must not have "omitempty"
    // +kubebuilder:validation:Optional
    // +kubebuilder:default="some-value"
    UserOptionalWithDefault *string `json:"user_optional_with_default"` // want "field TestPolicy.OtherStruct.UserOptionalWithDefault does not match any allowed non-mergeable category"

    // Non-Mergeable, User Optional, With Default
    // Must not be a pointer, must have "default", must have "optional", must not have "omitempty"
    // +kubebuilder:default="some-value"
    UserOptionalWithDefaultMissingOptional string `json:"user_optional_with_default_missing_optional"` // want "field TestPolicy.OtherStruct.UserOptionalWithDefaultMissingOptional does not match any allowed non-mergeable category"

    // Non-Mergeable, User Required
    // Must not have "omitempty", must not have "default", must not be a pointer
    UserRequiredOmitEmpty string `json:"user_required,omitempty"` // want "field TestPolicy.OtherStruct.UserRequiredOmitEmpty does not match any allowed non-mergeable category"

    // Non-Mergeable, User Required
    // Must not have "omitempty", must not have "default", must not be a pointer
    // +kubebuilder:validation:Optional
    // +kubebuilder:default="some-value"
    UserRequiredDefault string `json:"user_required,omitempty"` // want "field TestPolicy.OtherStruct.UserRequiredDefault does not match any allowed non-mergeable category"
}
