package accesslog

type HeaderFormatter struct {
	Header    string
	AltHeader string
	MaxLength int
}

func (f *HeaderFormatter) Format(headers map[string]string) (string, error) {
	value, exist := headers[f.Header]
	if !exist && f.AltHeader != "" {
		value = headers[f.AltHeader]
	}
	if f.MaxLength > 0 && len(value) > f.MaxLength {
		return value[:f.MaxLength], nil
	}
	return value, nil
}

func (f *HeaderFormatter) AppendTo(headers []string) []string {
	headers = append(headers, f.Header)
	if f.AltHeader != "" {
		headers = append(headers, f.AltHeader)
	}
	return headers
}
