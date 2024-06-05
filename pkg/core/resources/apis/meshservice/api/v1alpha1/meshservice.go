// +kubebuilder:object:generate=true
package v1alpha1

import (
	kube_meta "k8s.io/apimachinery/pkg/apis/meta/v1"
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
// +kuma:policy:kds_flags=model.ZoneToGlobalFlag | model.GlobalToAllButOriginalZoneFlag
type MeshService struct {
	Selector Selector `json:"selector,omitempty"`
	// +patchMergeKey=port
	// +patchStrategy=merge
	// +listType=map
	// +listMapKey=port
	// +listMapKey=protocol
	Ports []Port `json:"ports,omitempty"`
}

type Origin string

const (
	OriginGenerator  Origin = "HostnameGenerator"
	OriginKubernetes Origin = "Kubernetes"
)

type Address struct {
	Hostname             string               `json:"hostname,omitempty"`
	Origin               Origin               `json:"origin,omitempty"`
	HostnameGeneratorRef HostnameGeneratorRef `json:"hostnameGeneratorRef,omitempty"`
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

type HostnameGeneratorRef struct {
	CoreName string `json:"name"`
}

const (
	GeneratedCondition string = "Generated"
)

const (
	GeneratedReason     string = "Generated"
	TemplateErrorReason string = "TemplateError"
	CollisionReason     string = "Collision"
)

type Condition struct {
	// type of condition in CamelCase or in foo.example.com/CamelCase.
	// ---
	// Many .condition.type values are consistent across resources like Available, but because arbitrary conditions can be
	// useful (see .node.status.conditions), the ability to deconflict is important.
	// The regex it matches is (dns1123SubdomainFmt/)?(qualifiedNameFmt)
	// +required
	// +kubebuilder:validation:Required
	// +kubebuilder:validation:Pattern=`^([a-z0-9]([-a-z0-9]*[a-z0-9])?(\.[a-z0-9]([-a-z0-9]*[a-z0-9])?)*/)?(([A-Za-z0-9][-A-Za-z0-9_.]*)?[A-Za-z0-9])$`
	// +kubebuilder:validation:MaxLength=316
	Type string `json:"type"`
	// status of the condition, one of True, False, Unknown.
	// +required
	// +kubebuilder:validation:Required
	// +kubebuilder:validation:Enum=True;False;Unknown
	Status kube_meta.ConditionStatus `json:"status"`
	// reason contains a programmatic identifier indicating the reason for the condition's last transition.
	// Producers of specific condition types may define expected values and meanings for this field,
	// and whether the values are considered a guaranteed API.
	// The value should be a CamelCase string.
	// This field may not be empty.
	// +required
	// +kubebuilder:validation:Required
	// +kubebuilder:validation:MaxLength=1024
	// +kubebuilder:validation:MinLength=1
	// +kubebuilder:validation:Pattern=`^[A-Za-z]([A-Za-z0-9_,:]*[A-Za-z0-9_])?$`
	Reason string `json:"reason"`
	// message is a human readable message indicating details about the transition.
	// This may be an empty string.
	// +required
	// +kubebuilder:validation:Required
	// +kubebuilder:validation:MaxLength=32768
	Message string `json:"message"`
}

type HostnameGeneratorStatus struct {
	HostnameGeneratorRef HostnameGeneratorRef `json:"hostnameGeneratorRef"`

	// Conditions is an array of gateway instance conditions.
	//
	// +optional
	// +patchMergeKey=type
	// +patchStrategy=merge
	// +listType=map
	// +listMapKey=type
	Conditions []Condition `json:"conditions,omitempty" patchStrategy:"merge" patchMergeKey:"type"`
}

type MeshServiceStatus struct {
	Addresses          []Address                 `json:"addresses,omitempty"`
	VIPs               []VIP                     `json:"vips,omitempty"`
	TLS                TLS                       `json:"tls,omitempty"`
	HostnameGenerators []HostnameGeneratorStatus `json:"hostnameGenerators,omitempty"`
}
