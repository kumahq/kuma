// +kubebuilder:object:generate=true
package v1alpha1

import (
	k8s "k8s.io/apimachinery/pkg/apis/meta/v1"

	common_api "github.com/kumahq/kuma/api/common/v1alpha1"
)

// MeshTimeout allows users to configure timeouts for communication between services in mesh
type MeshTimeout struct {
	// TargetRef is a reference to the resource the policy takes an effect on.
	// The resource could be either a real store object or virtual resource
	// defined in-place.
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
	// ConnectionTimeout specifies the amount of time proxy will wait for an TCP connection to be established.
	// Cannot be set to "0".
	// +kubebuilder:default="5s"
	ConnectionTimeout *k8s.Duration `json:"connectionTimeout,omitempty"`
	// IdleTimeout is defined as the period in which there are no bytes sent or received on connection
	// Setting this timeout to "0s" will disable it. Be cautious when disabling it because
	// it can lead to connection leaking.
	// +kubebuilder:default="1h"
	IdleTimeout *k8s.Duration `json:"idleTimeout,omitempty"`
	// Http provides configuration for HTTP specific timeouts
	Http *Http `json:"http,omitempty"`
}

type Http struct {
	// RequestTimeout The amount of time that proxy will wait for the entire request to be received.
	// The timer is activated when the request is initiated, and is disarmed when the last byte of the request is sent,
	// OR when the response is initiated. Setting this timeout to "0s" will disable it.
	// +kubebuilder:default="15s"
	RequestTimeout *k8s.Duration `json:"requestTimeout,omitempty"`
	// StreamIdleTimeout is the amount of time that proxy will allow a stream to exist with no activity.
	// Setting this timeout to "0s" will disable it.
	// +kubebuilder:default="30s"
	StreamIdleTimeout *k8s.Duration `json:"streamIdleTimeout,omitempty"`
	// MaxStreamDuration is the maximum time that a stream’s lifetime will span.
	// Setting this timeout to "0s" will disable it.
	// +kubebuilder:default="0s"
	MaxStreamDuration *k8s.Duration `json:"maxStreamDuration,omitempty"`
	// MaxConnectionDuration is the time after which a connection will be drained and/or closed,
	// starting from when it was first established. Setting this timeout to "0s" will disable it.
	// +kubebuilder:default="0s"
	MaxConnectionDuration *k8s.Duration `json:"maxConnectionDuration,omitempty"`
}
