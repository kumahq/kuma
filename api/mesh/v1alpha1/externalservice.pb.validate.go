// Code generated by protoc-gen-validate. DO NOT EDIT.
// source: mesh/v1alpha1/externalservice.proto

package v1alpha1

import (
	"bytes"
	"errors"
	"fmt"
	"net"
	"net/mail"
	"net/url"
	"regexp"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/golang/protobuf/ptypes"
)

// ensure the imports are used
var (
	_ = bytes.MinRead
	_ = errors.New("")
	_ = fmt.Print
	_ = utf8.UTFMax
	_ = (*regexp.Regexp)(nil)
	_ = (*strings.Reader)(nil)
	_ = net.IPv4len
	_ = time.Duration(0)
	_ = (*url.URL)(nil)
	_ = (*mail.Address)(nil)
	_ = ptypes.DynamicAny{}
)

// define the regex for a UUID once up-front
var _externalservice_uuidPattern = regexp.MustCompile("^[0-9a-fA-F]{8}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{12}$")

// Validate checks the field values on ExternalService with the rules defined
// in the proto definition for this message. If any rules are violated, an
// error is returned.
func (m *ExternalService) Validate() error {
	if m == nil {
		return nil
	}

	if v, ok := interface{}(m.GetNetworking()).(interface{ Validate() error }); ok {
		if err := v.Validate(); err != nil {
			return ExternalServiceValidationError{
				field:  "Networking",
				reason: "embedded message failed validation",
				cause:  err,
			}
		}
	}

	if len(m.GetTags()) < 1 {
		return ExternalServiceValidationError{
			field:  "Tags",
			reason: "value must contain at least 1 pair(s)",
		}
	}

	return nil
}

// ExternalServiceValidationError is the validation error returned by
// ExternalService.Validate if the designated constraints aren't met.
type ExternalServiceValidationError struct {
	field  string
	reason string
	cause  error
	key    bool
}

// Field function returns field value.
func (e ExternalServiceValidationError) Field() string { return e.field }

// Reason function returns reason value.
func (e ExternalServiceValidationError) Reason() string { return e.reason }

// Cause function returns cause value.
func (e ExternalServiceValidationError) Cause() error { return e.cause }

// Key function returns key value.
func (e ExternalServiceValidationError) Key() bool { return e.key }

// ErrorName returns error name.
func (e ExternalServiceValidationError) ErrorName() string { return "ExternalServiceValidationError" }

// Error satisfies the builtin error interface
func (e ExternalServiceValidationError) Error() string {
	cause := ""
	if e.cause != nil {
		cause = fmt.Sprintf(" | caused by: %v", e.cause)
	}

	key := ""
	if e.key {
		key = "key for "
	}

	return fmt.Sprintf(
		"invalid %sExternalService.%s: %s%s",
		key,
		e.field,
		e.reason,
		cause)
}

var _ error = ExternalServiceValidationError{}

var _ interface {
	Field() string
	Reason() string
	Key() bool
	Cause() error
	ErrorName() string
} = ExternalServiceValidationError{}

// Validate checks the field values on ExternalService_Networking with the
// rules defined in the proto definition for this message. If any rules are
// violated, an error is returned.
func (m *ExternalService_Networking) Validate() error {
	if m == nil {
		return nil
	}

	// no validation rules for Address

	if v, ok := interface{}(m.GetTls()).(interface{ Validate() error }); ok {
		if err := v.Validate(); err != nil {
			return ExternalService_NetworkingValidationError{
				field:  "Tls",
				reason: "embedded message failed validation",
				cause:  err,
			}
		}
	}

	return nil
}

// ExternalService_NetworkingValidationError is the validation error returned
// by ExternalService_Networking.Validate if the designated constraints aren't met.
type ExternalService_NetworkingValidationError struct {
	field  string
	reason string
	cause  error
	key    bool
}

// Field function returns field value.
func (e ExternalService_NetworkingValidationError) Field() string { return e.field }

// Reason function returns reason value.
func (e ExternalService_NetworkingValidationError) Reason() string { return e.reason }

// Cause function returns cause value.
func (e ExternalService_NetworkingValidationError) Cause() error { return e.cause }

// Key function returns key value.
func (e ExternalService_NetworkingValidationError) Key() bool { return e.key }

// ErrorName returns error name.
func (e ExternalService_NetworkingValidationError) ErrorName() string {
	return "ExternalService_NetworkingValidationError"
}

// Error satisfies the builtin error interface
func (e ExternalService_NetworkingValidationError) Error() string {
	cause := ""
	if e.cause != nil {
		cause = fmt.Sprintf(" | caused by: %v", e.cause)
	}

	key := ""
	if e.key {
		key = "key for "
	}

	return fmt.Sprintf(
		"invalid %sExternalService_Networking.%s: %s%s",
		key,
		e.field,
		e.reason,
		cause)
}

var _ error = ExternalService_NetworkingValidationError{}

var _ interface {
	Field() string
	Reason() string
	Key() bool
	Cause() error
	ErrorName() string
} = ExternalService_NetworkingValidationError{}

// Validate checks the field values on ExternalService_Networking_TLS with the
// rules defined in the proto definition for this message. If any rules are
// violated, an error is returned.
func (m *ExternalService_Networking_TLS) Validate() error {
	if m == nil {
		return nil
	}

	// no validation rules for Enabled

	return nil
}

// ExternalService_Networking_TLSValidationError is the validation error
// returned by ExternalService_Networking_TLS.Validate if the designated
// constraints aren't met.
type ExternalService_Networking_TLSValidationError struct {
	field  string
	reason string
	cause  error
	key    bool
}

// Field function returns field value.
func (e ExternalService_Networking_TLSValidationError) Field() string { return e.field }

// Reason function returns reason value.
func (e ExternalService_Networking_TLSValidationError) Reason() string { return e.reason }

// Cause function returns cause value.
func (e ExternalService_Networking_TLSValidationError) Cause() error { return e.cause }

// Key function returns key value.
func (e ExternalService_Networking_TLSValidationError) Key() bool { return e.key }

// ErrorName returns error name.
func (e ExternalService_Networking_TLSValidationError) ErrorName() string {
	return "ExternalService_Networking_TLSValidationError"
}

// Error satisfies the builtin error interface
func (e ExternalService_Networking_TLSValidationError) Error() string {
	cause := ""
	if e.cause != nil {
		cause = fmt.Sprintf(" | caused by: %v", e.cause)
	}

	key := ""
	if e.key {
		key = "key for "
	}

	return fmt.Sprintf(
		"invalid %sExternalService_Networking_TLS.%s: %s%s",
		key,
		e.field,
		e.reason,
		cause)
}

var _ error = ExternalService_Networking_TLSValidationError{}

var _ interface {
	Field() string
	Reason() string
	Key() bool
	Cause() error
	ErrorName() string
} = ExternalService_Networking_TLSValidationError{}
