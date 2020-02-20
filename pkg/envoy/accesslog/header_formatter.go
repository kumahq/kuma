package accesslog

type Headers interface {
	Get(name string) (value string, exists bool)
}

type HeaderMap map[string]string

func (m HeaderMap) Get(name string) (value string, exists bool) {
	value, exists = m[name]
	return
}

type HeaderFormatter struct {
	Header    string
	AltHeader string
	MaxLength int
}

func (f *HeaderFormatter) Format(headers Headers) (string, error) {
	value, exists := "", false
	// apparently, Envoy allows both `Header` and `AltHeader` to be empty
	if f.Header != "" {
		value, exists = headers.Get(f.Header)
	}
	if !exists && f.AltHeader != "" {
		value, _ = headers.Get(f.AltHeader)
	}
	if f.MaxLength > 0 && len(value) > f.MaxLength {
		return value[:f.MaxLength], nil
	}
	return value, nil
}

func (f *HeaderFormatter) AppendTo(headers []string) []string {
	// apparently, Envoy allows both `Header` and `AltHeader` to be empty
	if f.Header != "" && !stringSet(headers).Contains(f.Header) {
		headers = append(headers, f.Header)
	}
	if f.AltHeader != "" && !stringSet(headers).Contains(f.AltHeader) {
		headers = append(headers, f.AltHeader)
	}
	return headers
}
