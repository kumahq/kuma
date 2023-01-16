package v3

import (
	"strconv"
	"strings"
)

// Headers represents a set of headers
// that might include both regular and pseudo headers.
type Headers interface {
	Get(name string) (value string, exists bool)
}

// HeaderMap represents a set of regular headers.
type HeaderMap map[string]string

func (m HeaderMap) Get(name string) (string, bool) {
	value, exists := m[name]
	return value, exists
}

// HeaderFormatter represents reusable formatting logic that is
// shared by `%REQ(X?Y):Z%`, `%RESP(X?Y):Z%` and `%TRAILER(X?Y):Z%`
// command operators.
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

func (f *HeaderFormatter) GetOperandHeaders() []string {
	var headers []string
	// apparently, Envoy allows both `Header` and `AltHeader` to be empty
	if f.Header != "" {
		headers = append(headers, f.Header)
	}
	if f.AltHeader != "" {
		headers = append(headers, f.AltHeader)
	}
	return headers
}

// String returns the canonical representation of a header command operator
// arguments and max length constraint.
func (f *HeaderFormatter) String() string {
	var builder []string
	builder = append(builder, "(")
	if f.Header != "" || f.AltHeader != "" {
		builder = append(builder, f.Header)
		if f.AltHeader != "" {
			builder = append(builder, "?")
			builder = append(builder, f.AltHeader)
		}
	}
	builder = append(builder, ")")
	if f.MaxLength != 0 {
		builder = append(builder, ":")
		builder = append(builder, strconv.FormatInt(int64(f.MaxLength), 10))
	}
	return strings.Join(builder, "")
}
