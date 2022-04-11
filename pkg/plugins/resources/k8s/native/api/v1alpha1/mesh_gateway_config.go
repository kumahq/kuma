package v1alpha1

import (
	kube_core "k8s.io/api/core/v1"
	kube_meta "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:resource:scope=Cluster

// MeshGatewayConfig represents the config of a MeshGateway.
type MeshGatewayConfig struct {
	kube_meta.TypeMeta   `json:",inline"`
	kube_meta.ObjectMeta `json:"metadata,omitempty"`

	Spec   MeshGatewayConfigSpec   `json:"spec,omitempty"`
	Status MeshGatewayConfigStatus `json:"status,omitempty"`
}

// +k8s:deepcopy-gen=true

// MeshGatewayCommonConfig represents the configuration in common for both
// Kuma-managed MeshGateways and Gateway API-managed MeshGateways
type MeshGatewayCommonConfig struct {
	// Replicas is the number of dataplane proxy replicas to create. For
	// now this is a fixed number, but in the future it could be
	// automatically scaled based on metrics.
	//
	// +optional
	// +kubebuilder:default=1
	// +kubebuilder:validation:Minimum=1
	Replicas int32 `json:"replicas,omitempty"`

	// ServiceType specifies the type of managed Service that will be
	// created to expose the dataplane proxies to traffic from outside
	// the cluster. The ports to expose will be taken from the matching Gateway
	// resource. If there is no matching Gateway, the managed Service will
	// be deleted.
	//
	// +optional
	// +kubebuilder:default=LoadBalancer
	// +kubebuilder:validation:Enum=LoadBalancer;ClusterIP;NodePort
	ServiceType kube_core.ServiceType `json:"serviceType,omitempty"`

	// Resources specifies the compute resources for the proxy container.
	// The default can be set in the control plane config.
	//
	// +optional
	Resources *kube_core.ResourceRequirements `json:"resources,omitempty"`
}

// +k8s:deepcopy-gen=true

// MeshGatewayConfigSpec specifies the options available for a builtin gateway Dataplane.
type MeshGatewayConfigSpec struct {
	MeshGatewayCommonConfig `json:",inline"`

	// Tags specifies the Kuma tags that are propagated to the managed
	// dataplane proxies. These tags should include a maximum of one
	// `kuma.io/service` tag.
	//
	// +optional
	Tags map[string]string `json:"tags,omitempty"`
}

// +k8s:deepcopy-gen=true

// MeshGatewayConfigStatus holds information about the status of the gateway
// instance.
type MeshGatewayConfigStatus struct {
}

// MeshGatewayConfigList contains a list of MeshGatewayConfigs.
//
// +kubebuilder:object:root=true
type MeshGatewayConfigList struct {
	kube_meta.TypeMeta `json:",inline"`
	kube_meta.ListMeta `json:"metadata,omitempty"`
	Items              []MeshGatewayConfig `json:"items"`
}
