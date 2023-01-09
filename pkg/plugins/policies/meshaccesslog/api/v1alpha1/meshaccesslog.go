// +kubebuilder:object:generate=true
package v1alpha1

import (
	common_api "github.com/kumahq/kuma/api/common/v1alpha1"
)

// MeshAccessLog defines access log policies between different data plane
// proxies entities.
type MeshAccessLog struct {
	// TargetRef is a reference to the resource the policy takes an effect on.
	// The resource could be either a real store object or virtual resource
	// defined inplace.
	TargetRef common_api.TargetRef `json:"targetRef"`
	// To list makes a match between the consumed services and corresponding configurations
	To []To `json:"to,omitempty"`
	// From list makes a match between clients and corresponding configurations
	From []From `json:"from,omitempty"`
}

type To struct {
	// TargetRef is a reference to the resource that represents a group of
	// destinations.
	TargetRef common_api.TargetRef `json:"targetRef"`
	// Default is a configuration specific to the group of destinations referenced in
	// 'targetRef'
	Default Conf `json:"default,omitempty"`
}

type From struct {
	// TargetRef is a reference to the resource that represents a group of
	// clients.
	TargetRef common_api.TargetRef `json:"targetRef"`
	// Default is a configuration specific to the group of clients referenced in
	// 'targetRef'
	Default Conf `json:"default,omitempty"`
}

type Conf struct {
	Backends *[]Backend `json:"backends,omitempty"`
}

type Backend struct {
	Tcp  *TCPBackend  `json:"tcp,omitempty"`
	File *FileBackend `json:"file,omitempty"`
}

// TCPBackend defines a TCP logging backend.
type TCPBackend struct {
	// Format of access logs. Placeholders available on
	// https://www.envoyproxy.io/docs/envoy/latest/configuration/observability/access_log
	Format *Format `json:"format,omitempty"`
	// Address of the TCP logging backend
	Address string `json:"address"`
}

// FileBackend defines configuration for file based access logs
type FileBackend struct {
	// Format of access logs. Placeholders available on
	// https://www.envoyproxy.io/docs/envoy/latest/configuration/observability/access_log
	Format *Format `json:"format,omitempty"`
	// Path to a file that logs will be written to
	Path string `json:"path"`
}

type Format struct {
	Plain           *string      `json:"plain,omitempty"`
	Json            *[]JsonValue `json:"json,omitempty"`
	OmitEmptyValues *bool        `json:"omitEmptyValues,omitempty"`
}

type JsonValue struct {
	Key   string `json:"key,omitempty"`
	Value string `json:"value,omitempty"`
}
