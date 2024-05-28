// +kubebuilder:object:generate=true
package v1alpha1

import (
	"k8s.io/apimachinery/pkg/util/intstr"

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
	Port       uint32             `json:"port"`
	TargetPort intstr.IntOrString `json:"targetPort,omitempty"`
	// +kubebuilder:default=tcp
	Protocol core_mesh.Protocol `json:"protocol,omitempty"`
}

// MeshService
// +kuma:policy:is_policy=false
// +kuma:policy:has_status=true
type MeshService struct {
	Selector Selector `json:"selector,omitempty"`
	// +patchMergeKey=port
	// +patchStrategy=merge
	// +listType=map
	// +listMapKey=port
	// +listMapKey=protocol
	Ports []Port `json:"ports,omitempty"`
}

type Address struct {
	Hostname string `json:"hostname,omitempty"`
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
	Addresses []Address `json:"addresses,omitempty"`
	VIPs      []VIP     `json:"vips,omitempty"`
	TLS       TLS       `json:"tls,omitempty"`
}
