package v1alpha1

import (
	kube_core "k8s.io/api/core/v1"
	kube_meta "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:resource:categories=kuma,scope=Cluster

// MeshGatewayConfig holds the configuration of a MeshGateway. A
// GatewayClass can refer to a MeshGatewayConfig via parametersRef.
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

	// ServiceTemplate configures the Service owned by this config.
	//
	// +optional
	ServiceTemplate MeshGatewayServiceTemplate `json:"serviceTemplate,omitempty"`

	// PodTemplate configures the Pod owned by this config.
	//
	// +optional
	PodTemplate MeshGatewayPodTemplate `json:"podTemplate,omitempty"`
}

// +k8s:deepcopy-gen=true

// MeshGatewayServiceTemplate holds configuration for a Service.
type MeshGatewayServiceTemplate struct {
	// Metadata holds metadata configuration for a Service.
	Metadata MeshGatewayObjectMetadata `json:"metadata,omitempty"`

	// Spec holds some customizable fields of a Service.
	Spec MeshGatewayServiceSpec `json:"spec,omitempty"`
}

// +k8s:deepcopy-gen=true

// MeshGatewayPodTemplate holds configuration for a Service.
type MeshGatewayPodTemplate struct {
	// Metadata holds metadata configuration for a Service.
	Metadata MeshGatewayObjectMetadata `json:"metadata,omitempty"`

	// Spec holds some customizable fields of a Pod.
	Spec MeshGatewayPodSpec `json:"spec,omitempty"`
}

// +k8s:deepcopy-gen=true

// MeshGatewayObjectMetadata holds Service metadata.
type MeshGatewayObjectMetadata struct {
	// Annotations holds annotations to be set on an object.
	Annotations map[string]string `json:"annotations,omitempty"`

	// Labels holds labels to be set on an objects.
	Labels map[string]string `json:"labels,omitempty"`
}

// +k8s:deepcopy-gen=true

// MeshGatewayServiceSpec holds customizable fields of a Service spec.
type MeshGatewayServiceSpec struct {
	// LoadBalancerIP corresponds to ServiceSpec.LoadBalancerIP.
	// +optional
	LoadBalancerIP string `json:"loadBalancerIP,omitempty"`
}

// +k8s:deepcopy-gen=true

// MeshGatewayPodSpec holds customizable fields of a Service spec.
type MeshGatewayPodSpec struct {
	// ServiceAccountName corresponds to PodSpec.ServiceAccountName.
	// +optional
	ServiceAccountName string `json:"serviceAccountName,omitempty"`

	// PodSecurityContext corresponds to PodSpec.SecurityContext
	// +optional
	PodSecurityContext PodSecurityContext `json:"securityContext,omitempty"`

	// Container corresponds to PodSpec.Container
	// +optional
	Container Container `json:"container,omitempty"`
}

// +k8s:deepcopy-gen=true

// PodSecurityContext corresponds to PodSpec.SecurityContext
type PodSecurityContext struct {
	// FSGroup corresponds to PodSpec.SecurityContext.FSGroup
	// +optional
	FSGroup *int64 `json:"fsGroup,omitempty"`
}

// +k8s:deepcopy-gen=true

// Container corresponds to PodSpec.Container
type Container struct {
	// ContainerSecurityContext corresponds to PodSpec.Container.SecurityContext
	SecurityContext SecurityContext `json:"securityContext,omitempty"`
}

// +k8s:deepcopy-gen=true

// SecurityContext corresponds to PodSpec.Container.SecurityContext
type SecurityContext struct {
	// ReadOnlyRootFilesystem corresponds to PodSpec.Container.SecurityContext.ReadOnlyRootFilesystem
	// +optional
	ReadOnlyRootFilesystem *bool `json:"readOnlyRootFilesystem,omitempty"`
}

// +k8s:deepcopy-gen=true

// MeshGatewayConfigSpec specifies the options available for a Kuma MeshGateway.
type MeshGatewayConfigSpec struct {
	MeshGatewayCommonConfig `json:",inline"`

	// CrossMesh specifies whether listeners configured by this gateway are
	// cross mesh listeners.
	CrossMesh bool `json:"crossMesh,omitempty"`

	// Tags specifies a set of Kuma tags that are included in the
	// MeshGatewayInstance and thus propagated to every Dataplane generated to
	// serve the MeshGateway.
	// These tags should include a maximum of one `kuma.io/service` tag.
	//
	// +optional
	Tags map[string]string `json:"tags,omitempty"`
}

// +k8s:deepcopy-gen=true

// MeshGatewayConfigStatus holds information about the status of the gateway
// instance.
type MeshGatewayConfigStatus struct{}

// MeshGatewayConfigList contains a list of MeshGatewayConfigs.
//
// +kubebuilder:object:root=true
type MeshGatewayConfigList struct {
	kube_meta.TypeMeta `json:",inline"`
	kube_meta.ListMeta `json:"metadata,omitempty"`
	Items              []MeshGatewayConfig `json:"items"`
}
