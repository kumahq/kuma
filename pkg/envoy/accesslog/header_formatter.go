package accesslog

import (
	"strings"
)

type Headers interface {
	Get(name string) (value string, exists bool)
}

type HeaderMap map[string]string

func (m HeaderMap) Get(name string) (value string, exists bool) {
	value, exists = m[strings.ToLower(name)] // Envoy keeps all headers in lower case
	return
}

type HeaderFormatter struct {
	Header    string
	AltHeader string
	MaxLength int
}

func (f *HeaderFormatter) Format(headers Headers) (string, error) {
	value, exists := headers.Get(f.Header)
	if !exists && f.AltHeader != "" {
		value, _ = headers.Get(f.AltHeader)
	}
	if f.MaxLength > 0 && len(value) > f.MaxLength {
		return value[:f.MaxLength], nil
	}
	return value, nil
}

func (f *HeaderFormatter) AppendTo(headers []string) []string {
	if f.Header != "" {
		headers = append(headers, f.Header)
	}
	if f.AltHeader != "" {
		headers = append(headers, f.AltHeader)
	}
	return headers
}
