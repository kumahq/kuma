// +kubebuilder:object:generate=true
package v1alpha1

import (
	common_api "github.com/kumahq/kuma/api/common/v1alpha1"
	common_tls "github.com/kumahq/kuma/api/common/v1alpha1/tls"
)

// MeshTLS
// +kuma:policy:is_policy=true
// +kuma:policy:plural=MeshTLSes
type MeshTLS struct {
	// TargetRef is a reference to the resource the policy takes an effect on.
	// The resource could be either a real store object or virtual resource
	// defined in-place.
	TargetRef common_api.TargetRef `json:"targetRef"`
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

// +kubebuilder:validation:Enum=ECDHE-ECDSA-AES128-GCM-SHA256;ECDHE-ECDSA-AES256-GCM-SHA384;ECDHE-ECDSA-CHACHA20-POLY1305;ECDHE-RSA-AES128-GCM-SHA256;ECDHE-RSA-AES256-GCM-SHA384;ECDHE-RSA-CHACHA20-POLY1305
type TlsCipher string

const (
	EcdheEcdsaAes128GcmSha256  TlsCipher = "ECDHE-ECDSA-AES128-GCM-SHA256"
	EcdheEcdsaAes256GcmSha384  TlsCipher = "ECDHE-ECDSA-AES256-GCM-SHA384"
	EcdheEcdsaChacha20Poly1305 TlsCipher = "ECDHE-ECDSA-CHACHA20-POLY1305"
	EcdheRsaAes128GcmSha256    TlsCipher = "ECDHE-RSA-AES128-GCM-SHA256"
	EcdheRsaAes256GcmSha384    TlsCipher = "ECDHE-RSA-AES256-GCM-SHA384"
	EcdheRsaChacha20Poly1305   TlsCipher = "ECDHE-RSA-CHACHA20-POLY1305"
)

var allCiphers = []string{
	string(EcdheEcdsaAes128GcmSha256),
	string(EcdheEcdsaAes256GcmSha384),
	string(EcdheEcdsaChacha20Poly1305),
	string(EcdheRsaAes128GcmSha256),
	string(EcdheRsaAes256GcmSha384),
	string(EcdheRsaChacha20Poly1305),
}

type TlsCiphers []TlsCipher

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
	TlsCiphers TlsCiphers `json:"tlsCiphers,omitempty"`

	// Mode defines the behavior of inbound listeners with regard to traffic encryption.
	Mode *Mode `json:"mode,omitempty"`
}
