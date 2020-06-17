// Code generated by protoc-gen-validate. DO NOT EDIT.
// source: mesh/v1alpha1/dataplane.proto

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
var _dataplane_uuidPattern = regexp.MustCompile("^[0-9a-fA-F]{8}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{12}$")

// Validate checks the field values on Dataplane with the rules defined in the
// proto definition for this message. If any rules are violated, an error is returned.
func (m *Dataplane) Validate() error {
	if m == nil {
		return nil
	}

	if v, ok := interface{}(m.GetNetworking()).(interface{ Validate() error }); ok {
		if err := v.Validate(); err != nil {
			return DataplaneValidationError{
				field:  "Networking",
				reason: "embedded message failed validation",
				cause:  err,
			}
		}
	}

	if v, ok := interface{}(m.GetMetrics()).(interface{ Validate() error }); ok {
		if err := v.Validate(); err != nil {
			return DataplaneValidationError{
				field:  "Metrics",
				reason: "embedded message failed validation",
				cause:  err,
			}
		}
	}

	return nil
}

// DataplaneValidationError is the validation error returned by
// Dataplane.Validate if the designated constraints aren't met.
type DataplaneValidationError struct {
	field  string
	reason string
	cause  error
	key    bool
}

// Field function returns field value.
func (e DataplaneValidationError) Field() string { return e.field }

// Reason function returns reason value.
func (e DataplaneValidationError) Reason() string { return e.reason }

// Cause function returns cause value.
func (e DataplaneValidationError) Cause() error { return e.cause }

// Key function returns key value.
func (e DataplaneValidationError) Key() bool { return e.key }

// ErrorName returns error name.
func (e DataplaneValidationError) ErrorName() string { return "DataplaneValidationError" }

// Error satisfies the builtin error interface
func (e DataplaneValidationError) Error() string {
	cause := ""
	if e.cause != nil {
		cause = fmt.Sprintf(" | caused by: %v", e.cause)
	}

	key := ""
	if e.key {
		key = "key for "
	}

	return fmt.Sprintf(
		"invalid %sDataplane.%s: %s%s",
		key,
		e.field,
		e.reason,
		cause)
}

var _ error = DataplaneValidationError{}

var _ interface {
	Field() string
	Reason() string
	Key() bool
	Cause() error
	ErrorName() string
} = DataplaneValidationError{}

// Validate checks the field values on Dataplane_Networking with the rules
// defined in the proto definition for this message. If any rules are
// violated, an error is returned.
func (m *Dataplane_Networking) Validate() error {
	if m == nil {
		return nil
	}

	if v, ok := interface{}(m.GetIngress()).(interface{ Validate() error }); ok {
		if err := v.Validate(); err != nil {
			return Dataplane_NetworkingValidationError{
				field:  "Ingress",
				reason: "embedded message failed validation",
				cause:  err,
			}
		}
	}

	// no validation rules for Address

	if v, ok := interface{}(m.GetGateway()).(interface{ Validate() error }); ok {
		if err := v.Validate(); err != nil {
			return Dataplane_NetworkingValidationError{
				field:  "Gateway",
				reason: "embedded message failed validation",
				cause:  err,
			}
		}
	}

	for idx, item := range m.GetInbound() {
		_, _ = idx, item

		if v, ok := interface{}(item).(interface{ Validate() error }); ok {
			if err := v.Validate(); err != nil {
				return Dataplane_NetworkingValidationError{
					field:  fmt.Sprintf("Inbound[%v]", idx),
					reason: "embedded message failed validation",
					cause:  err,
				}
			}
		}

	}

	for idx, item := range m.GetOutbound() {
		_, _ = idx, item

		if v, ok := interface{}(item).(interface{ Validate() error }); ok {
			if err := v.Validate(); err != nil {
				return Dataplane_NetworkingValidationError{
					field:  fmt.Sprintf("Outbound[%v]", idx),
					reason: "embedded message failed validation",
					cause:  err,
				}
			}
		}

	}

	if v, ok := interface{}(m.GetTransparentProxying()).(interface{ Validate() error }); ok {
		if err := v.Validate(); err != nil {
			return Dataplane_NetworkingValidationError{
				field:  "TransparentProxying",
				reason: "embedded message failed validation",
				cause:  err,
			}
		}
	}

	return nil
}

// Dataplane_NetworkingValidationError is the validation error returned by
// Dataplane_Networking.Validate if the designated constraints aren't met.
type Dataplane_NetworkingValidationError struct {
	field  string
	reason string
	cause  error
	key    bool
}

// Field function returns field value.
func (e Dataplane_NetworkingValidationError) Field() string { return e.field }

// Reason function returns reason value.
func (e Dataplane_NetworkingValidationError) Reason() string { return e.reason }

// Cause function returns cause value.
func (e Dataplane_NetworkingValidationError) Cause() error { return e.cause }

// Key function returns key value.
func (e Dataplane_NetworkingValidationError) Key() bool { return e.key }

// ErrorName returns error name.
func (e Dataplane_NetworkingValidationError) ErrorName() string {
	return "Dataplane_NetworkingValidationError"
}

// Error satisfies the builtin error interface
func (e Dataplane_NetworkingValidationError) Error() string {
	cause := ""
	if e.cause != nil {
		cause = fmt.Sprintf(" | caused by: %v", e.cause)
	}

	key := ""
	if e.key {
		key = "key for "
	}

	return fmt.Sprintf(
		"invalid %sDataplane_Networking.%s: %s%s",
		key,
		e.field,
		e.reason,
		cause)
}

var _ error = Dataplane_NetworkingValidationError{}

var _ interface {
	Field() string
	Reason() string
	Key() bool
	Cause() error
	ErrorName() string
} = Dataplane_NetworkingValidationError{}

// Validate checks the field values on Dataplane_Networking_Ingress with the
// rules defined in the proto definition for this message. If any rules are
// violated, an error is returned.
func (m *Dataplane_Networking_Ingress) Validate() error {
	if m == nil {
		return nil
	}

	for idx, item := range m.GetAvailableServices() {
		_, _ = idx, item

		if v, ok := interface{}(item).(interface{ Validate() error }); ok {
			if err := v.Validate(); err != nil {
				return Dataplane_Networking_IngressValidationError{
					field:  fmt.Sprintf("AvailableServices[%v]", idx),
					reason: "embedded message failed validation",
					cause:  err,
				}
			}
		}

	}

	return nil
}

// Dataplane_Networking_IngressValidationError is the validation error returned
// by Dataplane_Networking_Ingress.Validate if the designated constraints
// aren't met.
type Dataplane_Networking_IngressValidationError struct {
	field  string
	reason string
	cause  error
	key    bool
}

// Field function returns field value.
func (e Dataplane_Networking_IngressValidationError) Field() string { return e.field }

// Reason function returns reason value.
func (e Dataplane_Networking_IngressValidationError) Reason() string { return e.reason }

// Cause function returns cause value.
func (e Dataplane_Networking_IngressValidationError) Cause() error { return e.cause }

// Key function returns key value.
func (e Dataplane_Networking_IngressValidationError) Key() bool { return e.key }

// ErrorName returns error name.
func (e Dataplane_Networking_IngressValidationError) ErrorName() string {
	return "Dataplane_Networking_IngressValidationError"
}

// Error satisfies the builtin error interface
func (e Dataplane_Networking_IngressValidationError) Error() string {
	cause := ""
	if e.cause != nil {
		cause = fmt.Sprintf(" | caused by: %v", e.cause)
	}

	key := ""
	if e.key {
		key = "key for "
	}

	return fmt.Sprintf(
		"invalid %sDataplane_Networking_Ingress.%s: %s%s",
		key,
		e.field,
		e.reason,
		cause)
}

var _ error = Dataplane_Networking_IngressValidationError{}

var _ interface {
	Field() string
	Reason() string
	Key() bool
	Cause() error
	ErrorName() string
} = Dataplane_Networking_IngressValidationError{}

// Validate checks the field values on Dataplane_Networking_Inbound with the
// rules defined in the proto definition for this message. If any rules are
// violated, an error is returned.
func (m *Dataplane_Networking_Inbound) Validate() error {
	if m == nil {
		return nil
	}

	// no validation rules for Interface

	// no validation rules for Port

	// no validation rules for ServicePort

	// no validation rules for Address

	if len(m.GetTags()) < 1 {
		return Dataplane_Networking_InboundValidationError{
			field:  "Tags",
			reason: "value must contain at least 1 pair(s)",
		}
	}

	return nil
}

// Dataplane_Networking_InboundValidationError is the validation error returned
// by Dataplane_Networking_Inbound.Validate if the designated constraints
// aren't met.
type Dataplane_Networking_InboundValidationError struct {
	field  string
	reason string
	cause  error
	key    bool
}

// Field function returns field value.
func (e Dataplane_Networking_InboundValidationError) Field() string { return e.field }

// Reason function returns reason value.
func (e Dataplane_Networking_InboundValidationError) Reason() string { return e.reason }

// Cause function returns cause value.
func (e Dataplane_Networking_InboundValidationError) Cause() error { return e.cause }

// Key function returns key value.
func (e Dataplane_Networking_InboundValidationError) Key() bool { return e.key }

// ErrorName returns error name.
func (e Dataplane_Networking_InboundValidationError) ErrorName() string {
	return "Dataplane_Networking_InboundValidationError"
}

// Error satisfies the builtin error interface
func (e Dataplane_Networking_InboundValidationError) Error() string {
	cause := ""
	if e.cause != nil {
		cause = fmt.Sprintf(" | caused by: %v", e.cause)
	}

	key := ""
	if e.key {
		key = "key for "
	}

	return fmt.Sprintf(
		"invalid %sDataplane_Networking_Inbound.%s: %s%s",
		key,
		e.field,
		e.reason,
		cause)
}

var _ error = Dataplane_Networking_InboundValidationError{}

var _ interface {
	Field() string
	Reason() string
	Key() bool
	Cause() error
	ErrorName() string
} = Dataplane_Networking_InboundValidationError{}

// Validate checks the field values on Dataplane_Networking_Outbound with the
// rules defined in the proto definition for this message. If any rules are
// violated, an error is returned.
func (m *Dataplane_Networking_Outbound) Validate() error {
	if m == nil {
		return nil
	}

	// no validation rules for Interface

	// no validation rules for Address

	// no validation rules for Port

	if err := m._validateHostname(m.GetService()); err != nil {
		return Dataplane_Networking_OutboundValidationError{
			field:  "Service",
			reason: "value must be a valid hostname",
			cause:  err,
		}
	}

	// no validation rules for Tags

	return nil
}

func (m *Dataplane_Networking_Outbound) _validateHostname(host string) error {
	s := strings.ToLower(strings.TrimSuffix(host, "."))

	if len(host) > 253 {
		return errors.New("hostname cannot exceed 253 characters")
	}

	for _, part := range strings.Split(s, ".") {
		if l := len(part); l == 0 || l > 63 {
			return errors.New("hostname part must be non-empty and cannot exceed 63 characters")
		}

		if part[0] == '-' {
			return errors.New("hostname parts cannot begin with hyphens")
		}

		if part[len(part)-1] == '-' {
			return errors.New("hostname parts cannot end with hyphens")
		}

		for _, r := range part {
			if (r < 'a' || r > 'z') && (r < '0' || r > '9') && r != '-' {
				return fmt.Errorf("hostname parts can only contain alphanumeric characters or hyphens, got %q", string(r))
			}
		}
	}

	return nil
}

// Dataplane_Networking_OutboundValidationError is the validation error
// returned by Dataplane_Networking_Outbound.Validate if the designated
// constraints aren't met.
type Dataplane_Networking_OutboundValidationError struct {
	field  string
	reason string
	cause  error
	key    bool
}

// Field function returns field value.
func (e Dataplane_Networking_OutboundValidationError) Field() string { return e.field }

// Reason function returns reason value.
func (e Dataplane_Networking_OutboundValidationError) Reason() string { return e.reason }

// Cause function returns cause value.
func (e Dataplane_Networking_OutboundValidationError) Cause() error { return e.cause }

// Key function returns key value.
func (e Dataplane_Networking_OutboundValidationError) Key() bool { return e.key }

// ErrorName returns error name.
func (e Dataplane_Networking_OutboundValidationError) ErrorName() string {
	return "Dataplane_Networking_OutboundValidationError"
}

// Error satisfies the builtin error interface
func (e Dataplane_Networking_OutboundValidationError) Error() string {
	cause := ""
	if e.cause != nil {
		cause = fmt.Sprintf(" | caused by: %v", e.cause)
	}

	key := ""
	if e.key {
		key = "key for "
	}

	return fmt.Sprintf(
		"invalid %sDataplane_Networking_Outbound.%s: %s%s",
		key,
		e.field,
		e.reason,
		cause)
}

var _ error = Dataplane_Networking_OutboundValidationError{}

var _ interface {
	Field() string
	Reason() string
	Key() bool
	Cause() error
	ErrorName() string
} = Dataplane_Networking_OutboundValidationError{}

// Validate checks the field values on Dataplane_Networking_Gateway with the
// rules defined in the proto definition for this message. If any rules are
// violated, an error is returned.
func (m *Dataplane_Networking_Gateway) Validate() error {
	if m == nil {
		return nil
	}

	if len(m.GetTags()) < 1 {
		return Dataplane_Networking_GatewayValidationError{
			field:  "Tags",
			reason: "value must contain at least 1 pair(s)",
		}
	}

	return nil
}

// Dataplane_Networking_GatewayValidationError is the validation error returned
// by Dataplane_Networking_Gateway.Validate if the designated constraints
// aren't met.
type Dataplane_Networking_GatewayValidationError struct {
	field  string
	reason string
	cause  error
	key    bool
}

// Field function returns field value.
func (e Dataplane_Networking_GatewayValidationError) Field() string { return e.field }

// Reason function returns reason value.
func (e Dataplane_Networking_GatewayValidationError) Reason() string { return e.reason }

// Cause function returns cause value.
func (e Dataplane_Networking_GatewayValidationError) Cause() error { return e.cause }

// Key function returns key value.
func (e Dataplane_Networking_GatewayValidationError) Key() bool { return e.key }

// ErrorName returns error name.
func (e Dataplane_Networking_GatewayValidationError) ErrorName() string {
	return "Dataplane_Networking_GatewayValidationError"
}

// Error satisfies the builtin error interface
func (e Dataplane_Networking_GatewayValidationError) Error() string {
	cause := ""
	if e.cause != nil {
		cause = fmt.Sprintf(" | caused by: %v", e.cause)
	}

	key := ""
	if e.key {
		key = "key for "
	}

	return fmt.Sprintf(
		"invalid %sDataplane_Networking_Gateway.%s: %s%s",
		key,
		e.field,
		e.reason,
		cause)
}

var _ error = Dataplane_Networking_GatewayValidationError{}

var _ interface {
	Field() string
	Reason() string
	Key() bool
	Cause() error
	ErrorName() string
} = Dataplane_Networking_GatewayValidationError{}

// Validate checks the field values on Dataplane_Networking_TransparentProxying
// with the rules defined in the proto definition for this message. If any
// rules are violated, an error is returned.
func (m *Dataplane_Networking_TransparentProxying) Validate() error {
	if m == nil {
		return nil
	}

	if m.GetRedirectPort() > 65535 {
		return Dataplane_Networking_TransparentProxyingValidationError{
			field:  "RedirectPort",
			reason: "value must be less than or equal to 65535",
		}
	}

	return nil
}

// Dataplane_Networking_TransparentProxyingValidationError is the validation
// error returned by Dataplane_Networking_TransparentProxying.Validate if the
// designated constraints aren't met.
type Dataplane_Networking_TransparentProxyingValidationError struct {
	field  string
	reason string
	cause  error
	key    bool
}

// Field function returns field value.
func (e Dataplane_Networking_TransparentProxyingValidationError) Field() string { return e.field }

// Reason function returns reason value.
func (e Dataplane_Networking_TransparentProxyingValidationError) Reason() string { return e.reason }

// Cause function returns cause value.
func (e Dataplane_Networking_TransparentProxyingValidationError) Cause() error { return e.cause }

// Key function returns key value.
func (e Dataplane_Networking_TransparentProxyingValidationError) Key() bool { return e.key }

// ErrorName returns error name.
func (e Dataplane_Networking_TransparentProxyingValidationError) ErrorName() string {
	return "Dataplane_Networking_TransparentProxyingValidationError"
}

// Error satisfies the builtin error interface
func (e Dataplane_Networking_TransparentProxyingValidationError) Error() string {
	cause := ""
	if e.cause != nil {
		cause = fmt.Sprintf(" | caused by: %v", e.cause)
	}

	key := ""
	if e.key {
		key = "key for "
	}

	return fmt.Sprintf(
		"invalid %sDataplane_Networking_TransparentProxying.%s: %s%s",
		key,
		e.field,
		e.reason,
		cause)
}

var _ error = Dataplane_Networking_TransparentProxyingValidationError{}

var _ interface {
	Field() string
	Reason() string
	Key() bool
	Cause() error
	ErrorName() string
} = Dataplane_Networking_TransparentProxyingValidationError{}

// Validate checks the field values on
// Dataplane_Networking_Ingress_AvailableService with the rules defined in the
// proto definition for this message. If any rules are violated, an error is returned.
func (m *Dataplane_Networking_Ingress_AvailableService) Validate() error {
	if m == nil {
		return nil
	}

	// no validation rules for Tags

	return nil
}

// Dataplane_Networking_Ingress_AvailableServiceValidationError is the
// validation error returned by
// Dataplane_Networking_Ingress_AvailableService.Validate if the designated
// constraints aren't met.
type Dataplane_Networking_Ingress_AvailableServiceValidationError struct {
	field  string
	reason string
	cause  error
	key    bool
}

// Field function returns field value.
func (e Dataplane_Networking_Ingress_AvailableServiceValidationError) Field() string { return e.field }

// Reason function returns reason value.
func (e Dataplane_Networking_Ingress_AvailableServiceValidationError) Reason() string {
	return e.reason
}

// Cause function returns cause value.
func (e Dataplane_Networking_Ingress_AvailableServiceValidationError) Cause() error { return e.cause }

// Key function returns key value.
func (e Dataplane_Networking_Ingress_AvailableServiceValidationError) Key() bool { return e.key }

// ErrorName returns error name.
func (e Dataplane_Networking_Ingress_AvailableServiceValidationError) ErrorName() string {
	return "Dataplane_Networking_Ingress_AvailableServiceValidationError"
}

// Error satisfies the builtin error interface
func (e Dataplane_Networking_Ingress_AvailableServiceValidationError) Error() string {
	cause := ""
	if e.cause != nil {
		cause = fmt.Sprintf(" | caused by: %v", e.cause)
	}

	key := ""
	if e.key {
		key = "key for "
	}

	return fmt.Sprintf(
		"invalid %sDataplane_Networking_Ingress_AvailableService.%s: %s%s",
		key,
		e.field,
		e.reason,
		cause)
}

var _ error = Dataplane_Networking_Ingress_AvailableServiceValidationError{}

var _ interface {
	Field() string
	Reason() string
	Key() bool
	Cause() error
	ErrorName() string
} = Dataplane_Networking_Ingress_AvailableServiceValidationError{}
