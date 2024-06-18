// +kubebuilder:object:generate=true
package v1alpha1

import (
	"k8s.io/apimachinery/pkg/util/intstr"

	hostnamegenerator_api "github.com/kumahq/kuma/pkg/core/resources/apis/hostnamegenerator/api/v1alpha1"
	core_mesh "github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
)

type DataplaneTags map[string]string

type Selector struct {
	DataplaneTags DataplaneTags `json:"dataplaneTags,omitempty"`
	DataplaneRef  *DataplaneRef `json:"dataplaneRef,omitempty"`
}

type DataplaneRef struct {
	Name string `json:"name,omitempty"`
}

type Port struct {
	Name       string             `json:"name,omitempty"`
	Port       uint32             `json:"port"`
	TargetPort intstr.IntOrString `json:"targetPort,omitempty"`
	// +kubebuilder:default=tcp
	AppProtocol core_mesh.Protocol `json:"appProtocol,omitempty"`
}

// MeshService
// +kuma:policy:is_policy=false
// +kuma:policy:has_status=true
// +kuma:policy:kds_flags=model.ZoneToGlobalFlag | model.GlobalToAllButOriginalZoneFlag
type MeshService struct {
	Selector Selector `json:"selector,omitempty"`
	// +patchMergeKey=port
	// +patchStrategy=merge
	// +listType=map
	// +listMapKey=port
	// +listMapKey=appProtocol
	Ports      []Port                `json:"ports,omitempty"`
	Identities []MeshServiceIdentity `json:"identities,omitempty"`
}

type VIP struct {
	IP string `json:"ip,omitempty"`
}

// +kubebuilder:validation:Enum=Ready;NotReady
type TLSStatus string

const (
	TLSReady    TLSStatus = "Ready"
	TLSNotReady TLSStatus = "NotReady"
)

type TLS struct {
	Status TLSStatus `json:"status,omitempty"`
}

type MeshServiceStatus struct {
	Addresses          []hostnamegenerator_api.Address                 `json:"addresses,omitempty"`
	VIPs               []VIP                                           `json:"vips,omitempty"`
	TLS                TLS                                             `json:"tls,omitempty"`
	HostnameGenerators []hostnamegenerator_api.HostnameGeneratorStatus `json:"hostnameGenerators,omitempty"`
}

// +kubebuilder:validation:Enum=ServiceTag
type MeshServiceIdentityType string

const (
	MeshServiceIdentityServiceTagType = "ServiceTag"
)

type MeshServiceIdentity struct {
	Type  MeshServiceIdentityType `json:"type"`
	Value string                  `json:"value"`
}
