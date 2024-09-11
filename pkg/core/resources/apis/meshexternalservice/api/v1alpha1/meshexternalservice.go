// +kubebuilder:object:generate=true
package v1alpha1

import (
	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"

	"github.com/kumahq/kuma/api/common/v1alpha1"
	common_tls "github.com/kumahq/kuma/api/common/v1alpha1/tls"
	hostnamegenerator_api "github.com/kumahq/kuma/pkg/core/resources/apis/hostnamegenerator/api/v1alpha1"
	core_mesh "github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
)

// MeshExternalService
// +kuma:policy:is_policy=false
// +kuma:policy:allowed_on_system_namespace_only=true
// +kuma:policy:has_status=true
type MeshExternalService struct {
	// Match defines traffic that should be routed through the sidecar.
	Match Match `json:"match"`
	// Extension struct for a plugin configuration, in the presence of an extension `endpoints` and `tls` are not required anymore - it's up to the extension to validate them independently.
	Extension *Extension `json:"extension,omitempty"`
	// Endpoints defines a list of destinations to send traffic to.
	Endpoints []Endpoint `json:"endpoints,omitempty"`
	// Tls provides a TLS configuration when proxy is resposible for a TLS origination
	Tls *Tls `json:"tls,omitempty"`
}

// +kubebuilder:validation:Enum=HostnameGenerator
type MatchType string

const (
	HostnameGeneratorType MatchType = "HostnameGenerator"
)

type Match struct {
	// Type of the match, only `HostnameGenerator` is available at the moment.
	// +kubebuilder:default=HostnameGenerator
	Type *MatchType `json:"type,omitempty"`
	// Port defines a port to which a user does request.
	Port Port `json:"port"`
	// Protocol defines a protocol of the communication. Possible values: `tcp`, `grpc`, `http`, `http2`.
	// +kubebuilder:default=tcp
	// +kubebuilder:validation:Enum=tcp;grpc;http;http2
	Protocol core_mesh.Protocol `json:"protocol,omitempty"`
}

type Extension struct {
	// Type of the extension.
	Type string `json:"type"`
	// Config freeform configuration for the extension.
	Config *apiextensionsv1.JSON `json:"config"`
}

// +kubebuilder:validation:Minimum=1
// +kubebuilder:validation:Maximum=65535
type Port int

type Endpoint struct {
	// Address defines an address to which a user want to send a request. Is possible to provide `domain`, `ip`.
	// +kubebuilder:example="127.0.0.1"
	// +kubebuilder:example="example.com"
	// +kubebuilder:validation:MinLength=1
	Address string `json:"address"`
	// Port of the endpoint
	Port *Port `json:"port,omitempty"`
}

type Tls struct {
	// Enabled defines if proxy should originate TLS.
	// +kubebuilder:default=false
	Enabled bool `json:"enabled,omitempty"`
	// Version section for providing version specification.
	Version *common_tls.Version `json:"version,omitempty"`
	// AllowRenegotiation defines if TLS sessions will allow renegotiation.
	// Setting this to true is not recommended for security reasons.
	// +kubebuilder:default=false
	AllowRenegotiation bool `json:"allowRenegotiation,omitempty"`
	// Verification section for providing TLS verification details.
	Verification *Verification `json:"verification,omitempty"`
}

// +kubebuilder:validation:Enum=SkipSAN;SkipCA;Secured;SkipAll
type VerificationMode string

const (
	TLSVerificationSkipSAN VerificationMode = "SkipSAN"
	TLSVerificationSkipCA  VerificationMode = "SkipCA"
	TLSVerificationSkipAll VerificationMode = "SkipAll"
	TLSVerificationSecured VerificationMode = "Secured"
)

type Verification struct {
	// Mode defines if proxy should skip verification, one of `SkipSAN`, `SkipCA`, `Secured`, `SkipAll`. Default `Secured`.
	// +kubebuilder:default=Secured
	Mode *VerificationMode `json:"mode,omitempty"`
	// ServerName overrides the default Server Name Indicator set by Kuma.
	ServerName *string `json:"serverName,omitempty"`
	// SubjectAltNames list of names to verify in the certificate.
	SubjectAltNames *[]SANMatch `json:"subjectAltNames,omitempty"`
	// CaCert defines a certificate of CA.
	CaCert *v1alpha1.DataSource `json:"caCert,omitempty"`
	// ClientCert defines a certificate of a client.
	ClientCert *v1alpha1.DataSource `json:"clientCert,omitempty"`
	// ClientKey defines a client private key.
	ClientKey *v1alpha1.DataSource `json:"clientKey,omitempty"`
}

// +kubebuilder:validation:Enum=Exact;Prefix
type SANMatchType string

const (
	SANMatchExact  SANMatchType = "Exact"
	SANMatchPrefix SANMatchType = "Prefix"
)

type SANMatch struct {
	// Type specifies matching type, one of `Exact`, `Prefix`. Default: `Exact`
	// +kubebuilder:default=Exact
	Type SANMatchType `json:"type,omitempty"`
	// Value to match.
	Value string `json:"value"`
}

type MeshExternalServiceStatus struct {
	// Vip section for allocated IP
	VIP VIP `json:"vip,omitempty"`
	// Addresses section for generated domains
	Addresses          []hostnamegenerator_api.Address                 `json:"addresses,omitempty"`
	HostnameGenerators []hostnamegenerator_api.HostnameGeneratorStatus `json:"hostnameGenerators,omitempty"`
}

// +kubebuilder:validation:Enum=Kuma
type StatusType string

type VIP struct {
	// Value allocated IP for a provided domain with `HostnameGenerator` type in a match section.
	IP string `json:"ip,omitempty"`
}
