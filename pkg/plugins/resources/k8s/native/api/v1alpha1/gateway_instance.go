package v1alpha1

import (
	kube_core "k8s.io/api/core/v1"
	kube_meta "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status

// GatewayInstance represents a managed instance of a dataplane proxy for a Kuma
// Gateway.
type GatewayInstance struct {
	kube_meta.TypeMeta   `json:",inline"`
	kube_meta.ObjectMeta `json:"metadata,omitempty"`

	Spec   GatewayInstanceSpec   `json:"spec,omitempty"`
	Status GatewayInstanceStatus `json:"status,omitempty"`
}

// +k8s:deepcopy-gen=true

// GatewayInstanceSpec specifies the options available for a GatewayDataplane.
type GatewayInstanceSpec struct {
	// Tags specifies the Kuma tags that are propagated to the managed
	// dataplane proxies. These tags should include exactly one
	// `kuma.io/service` tag, and should match exactly one Gateway
	// resource.
	//
	// +required
	// +kubebuilder:validation:MinLen=1
	Tags map[string]string `json:"tags,omitempty"`

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

// GatewayInstanceStatus holds information about the status of the gateway
// instance.
type GatewayInstanceStatus struct {
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

	GatewayInstanceDeploymentNotAvailable = "DeploymentNotAvailable"

	GatewayInstanceAddressNotReady = "LoadBalancerAddressNotReady"
)

// GatewayInstanceList contains a list of GatewayInstances.
//
// +kubebuilder:object:root=true
type GatewayInstanceList struct {
	kube_meta.TypeMeta `json:",inline"`
	kube_meta.ListMeta `json:"metadata,omitempty"`
	Items              []GatewayInstance `json:"items"`
}
