package v1alpha1

// DataSource defines the source of bytes to use. Protobuf modeled the alternatives as a
// oneof and wrote the set one as a plain top level key, which is the shape kept here.
// A DataSource is built to call the loader and is never stored.
type DataSource struct {
	// Secret is the name of the secret holding the bytes.
	Secret *string `json:"secret,omitempty"`
	// File is the path to read the bytes from. It is deprecated, prefer another
	// source of data.
	File *string `json:"file,omitempty"`
	// Inline holds the bytes themselves.
	Inline *BytesValue `json:"inline,omitempty"`
	// InlineString holds the bytes as a string.
	InlineString *string `json:"inlineString,omitempty"`
}

func (ds *DataSource) GetSecret() string {
	if ds == nil || ds.Secret == nil {
		return ""
	}
	return *ds.Secret
}

func (ds *DataSource) GetFile() string {
	if ds == nil || ds.File == nil {
		return ""
	}
	return *ds.File
}

func (ds *DataSource) GetInline() *BytesValue {
	if ds == nil {
		return nil
	}
	return ds.Inline
}

func (ds *DataSource) GetInlineString() string {
	if ds == nil || ds.InlineString == nil {
		return ""
	}
	return *ds.InlineString
}

// HasSecret and the calls beside it replace the type switch the oneof wrapper needed.
func (ds *DataSource) HasSecret() bool { return ds != nil && ds.Secret != nil }

func (ds *DataSource) HasFile() bool { return ds != nil && ds.File != nil }

func (ds *DataSource) HasInline() bool { return ds != nil && ds.Inline != nil }

func (ds *DataSource) HasInlineString() bool { return ds != nil && ds.InlineString != nil }

// IsSet reports whether any alternative was chosen, which the nil oneof wrapper used to say.
func (ds *DataSource) IsSet() bool {
	return ds.HasSecret() || ds.HasFile() || ds.HasInline() || ds.HasInlineString()
}
