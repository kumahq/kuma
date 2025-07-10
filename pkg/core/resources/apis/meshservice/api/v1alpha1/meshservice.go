// +kubebuilder:object:generate=true
package v1alpha1

import (
	"k8s.io/apimachinery/pkg/util/intstr"

	hostnamegenerator_api "github.com/kumahq/kuma/pkg/core/resources/apis/hostnamegenerator/api/v1alpha1"
	core_mesh "github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
)

type Selector struct {
	DataplaneTags *map[string]string `json:"dataplaneTags,omitempty"`
	DataplaneRef  *DataplaneRef      `json:"dataplaneRef,omitempty"`
}

type DataplaneRef struct {
	// +kuma:comment It should be required but MeshService doesn't have any validation https://github.com/kumahq/kuma/issues/13814 so adding validation here would be a breaking change
	// +kuma:nolint
	Name string `json:"name,omitempty"`
}

type Port struct {
	Name *string `json:"name,omitempty"`
	Port int32   `json:"port"`
	TargetPort *intstr.IntOrString `json:"targetPort,omitempty"`
	// +kuma:comment It should be without omitempty but MeshService doesn't have any validation https://github.com/kumahq/kuma/issues/13814 so if it was ever persisted empty this would cause a nack
	// +kubebuilder:default=tcp
	// +kuma:nolint
	AppProtocol core_mesh.Protocol `json:"appProtocol,omitempty"`
}

const maxNameLength = 63

// MeshService
// +kuma:policy:is_policy=false
// +kuma:policy:has_status=true
// +kuma:policy:kds_flags=model.ZoneToGlobalFlag | model.SyncedAcrossZonesFlag
// +kuma:policy:is_referenceable_in_to=true
// +kuma:policy:short_name=msvc
// +kuma:policy:is_destination=true
// +kubebuilder:printcolumn:JSONPath=".status.addresses[0].hostname",name=Hostname,type=string
type MeshService struct {
	// State of MeshService. Available if there is at least one healthy endpoint. Otherwise, Unavailable.
	// It's used for cross zone communication to check if we should send traffic to it, when MeshService is aggregated into MeshMultiZoneService.
	// +kubebuilder:default=Unavailable
	// +kuma:comment It should be required but MeshService doesn't have any validation https://github.com/kumahq/kuma/issues/13814 so adding validation here would be a breaking change
	// +kuma:nolint
	State State `json:"state,omitempty"`
	// +kuma:comment It should be required but MeshService doesn't have any validation https://github.com/kumahq/kuma/issues/13814 so adding validation here would be a breaking change
	// +kuma:nolint
	Selector Selector `json:"selector,omitempty"`
	// +patchMergeKey=port
	// +patchStrategy=merge
	// +listType=map
	// +listMapKey=port
	// +listMapKey=appProtocol
	// +kuma:comment It should be required but MeshService doesn't have any validation https://github.com/kumahq/kuma/issues/13814 so adding validation here would be a breaking change
	// +kuma:nolint
	Ports []Port `json:"ports,omitempty"`
	Identities *[]MeshServiceIdentity `json:"identities,omitempty"`
}

type VIP struct {
	IP string `json:"ip,omitempty"`
}

// +kubebuilder:validation:Enum=Available;Unavailable
type State string

const (
	StateAvailable   State = "Available"
	StateUnavailable State = "Unavailable"
)

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
	// Data plane proxies statistics selected by this MeshService.
	DataplaneProxies DataplaneProxies `json:"dataplaneProxies,omitempty"`
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

type DataplaneProxies struct {
	// Number of data plane proxies connected to the zone control plane
	Connected int `json:"connected,omitempty"`
	// Number of data plane proxies with all healthy inbounds selected by this MeshService.
	Healthy int `json:"healthy,omitempty"`
	// Total number of data plane proxies.
	Total int `json:"total,omitempty"`
}
