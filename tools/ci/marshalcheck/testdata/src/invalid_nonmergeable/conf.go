package invalid_nonmergeable

type OtherStruct struct {
    MissingPtr string `json:"missing_ptr"` // want "field OtherStruct.MissingPtr must be a pointer with 'omitempty' JSON tag or be inside a slice/array \\(nonmergeable\\)"
}
