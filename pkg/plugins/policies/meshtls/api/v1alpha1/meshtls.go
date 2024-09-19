// +kubebuilder:object:generate=true
package v1alpha1

import (
	common_api "github.com/kumahq/kuma/api/common/v1alpha1"
	common_tls "github.com/kumahq/kuma/api/common/v1alpha1/tls"
)

// MeshTLS
// +kuma:policy:singular_display_name=Mesh TLS
type MeshTLS struct {
	// TargetRef is a reference to the resource the policy takes an effect on.
	// The resource could be either a real store object or virtual resource
	// defined in-place.
	TargetRef *common_api.TargetRef `json:"targetRef,omitempty"`
	// From list makes a match between clients and corresponding configurations
	From []From `json:"from,omitempty"`
}

type From struct {
	// TargetRef is a reference to the resource that represents a group of
	// clients.
	TargetRef common_api.TargetRef `json:"targetRef"`
	// Default is a configuration specific to the group of clients referenced in
	// 'targetRef'
	Default Conf `json:"default,omitempty"`
}

// +kubebuilder:validation:Enum=Permissive;Strict
type Mode string

const (
	ModeStrict     Mode = "Strict"
	ModePermissive Mode = "Permissive"
)

var allModes = []string{string(ModeStrict), string(ModePermissive)}

type Conf struct {
	// Version section for providing version specification.
	TlsVersion *common_tls.Version `json:"tlsVersion,omitempty"`

	// TlsCiphers section for providing ciphers specification.
	TlsCiphers common_tls.TlsCiphers `json:"tlsCiphers,omitempty"`

	// Mode defines the behavior of inbound listeners with regard to traffic encryption.
	Mode *Mode `json:"mode,omitempty"`
}
