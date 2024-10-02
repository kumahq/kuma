package v1alpha1

import (
	kube_core "k8s.io/api/core/v1"
	kube_meta "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:resource:categories=kuma,scope=Namespaced

// MeshGatewayInstance represents a managed instance of a dataplane proxy for a Kuma
// Gateway.
type MeshGatewayInstance struct {
	kube_meta.TypeMeta   `json:",inline"`
	kube_meta.ObjectMeta `json:"metadata,omitempty"`

	Spec   MeshGatewayInstanceSpec   `json:"spec,omitempty"`
	Status MeshGatewayInstanceStatus `json:"status,omitempty"`
}

// +k8s:deepcopy-gen=true

// MeshGatewayInstanceSpec specifies the options available for a GatewayDataplane.
type MeshGatewayInstanceSpec struct {
	MeshGatewayCommonConfig `json:",inline"`

	// Tags specifies the Kuma tags that are propagated to the managed
	// dataplane proxies. These tags should not include `kuma.io/service` tag
	// since is auto-generated, and should match exactly one Gateway
	// resource.
	//
	// +optional
	// +kubebuilder:validation:MinLen=1
	Tags map[string]string `json:"tags,omitempty"`
}

// +k8s:deepcopy-gen=true

// MeshGatewayInstanceStatus holds information about the status of the gateway
// instance.
type MeshGatewayInstanceStatus struct {
	// LoadBalancer contains the current status of the load-balancer,
	// if one is present.
	//
	// +optional
	LoadBalancer *kube_core.LoadBalancerStatus `json:"loadBalancer,omitempty"`

	// Conditions is an array of gateway instance conditions.
	//
	// +optional
	// +patchMergeKey=type
	// +patchStrategy=merge
	// +listType=map
	// +listMapKey=type
	Conditions []kube_meta.Condition `json:"conditions,omitempty" patchStrategy:"merge" patchMergeKey:"type"`
}

const (
	GatewayInstanceReady string = "Ready"

	GatewayInstanceNoGatewayMatched       = "NoGatewayMatched"
	GatewayInstanceDeploymentNotAvailable = "DeploymentNotAvailable"

	GatewayInstanceAddressNotReady = "LoadBalancerAddressNotReady"
)

// MeshGatewayInstanceList contains a list of GatewayInstances.
//
// +kubebuilder:object:root=true
type MeshGatewayInstanceList struct {
	kube_meta.TypeMeta `json:",inline"`
	kube_meta.ListMeta `json:"metadata,omitempty"`
	Items              []MeshGatewayInstance `json:"items"`
}
