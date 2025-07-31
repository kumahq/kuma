// +kubebuilder:object:generate=true
package v1alpha1

import (
	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"

	common_api "github.com/kumahq/kuma/api/common/v1alpha1"
)

type Selector struct {
	MeshService          *common_api.LabelSelector `json:"meshService,omitempty"`
	MeshExternalService  *common_api.LabelSelector `json:"meshExternalService,omitempty"`
	MeshMultiZoneService *common_api.LabelSelector `json:"meshMultiZoneService,omitempty"`
}

// HostnameGenerator
// +kuma:policy:is_policy=false
// +kuma:policy:allowed_on_system_namespace_only=true
// +kuma:policy:scope=Global
// hostname generators to not get synced across zones
// +kuma:policy:kds_flags=model.GlobalToZonesFlag | model.ZoneToGlobalFlag
type HostnameGenerator struct {
	// +kuma:nolint
	Selector Selector `json:"selector,omitempty"`
	Template string   `json:"template"`
	// Extension struct for a plugin configuration
	Extension *Extension `json:"extension,omitempty"`
}

type Extension struct {
	// Type of the extension.
	Type string `json:"type"`
	// Config freeform configuration for the extension.
	Config *apiextensionsv1.JSON `json:"config,omitempty"`
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

type HostnameGeneratorRef struct {
	CoreName string `json:"coreName"`
}

type HostnameGeneratorStatus struct {
	HostnameGeneratorRef HostnameGeneratorRef `json:"hostnameGeneratorRef"`

	// Conditions is an array of hostname generator conditions.
	//
	// +optional
	// +patchMergeKey=type
	// +patchStrategy=merge
	// +listType=map
	// +listMapKey=type
	Conditions []common_api.Condition `json:"conditions,omitempty" patchStrategy:"merge" patchMergeKey:"type"`
}
