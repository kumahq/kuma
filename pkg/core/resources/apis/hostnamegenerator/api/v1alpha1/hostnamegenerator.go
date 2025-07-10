// +kubebuilder:object:generate=true
package v1alpha1

import (
	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	kube_meta "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type LabelSelector struct {
	// +kuma:nolint
	MatchLabels map[string]string `json:"matchLabels,omitempty"`
}

type Selector struct {
	MeshService          *LabelSelector `json:"meshService,omitempty"`
	MeshExternalService  *LabelSelector `json:"meshExternalService,omitempty"`
	MeshMultiZoneService *LabelSelector `json:"meshMultiZoneService,omitempty"`
}

func (s LabelSelector) Matches(labels map[string]string) bool {
	for tag, matchValue := range s.MatchLabels {
		labelValue, exist := labels[tag]
		if !exist {
			return false
		}
		if matchValue != labelValue {
			return false
		}
	}
	return true
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
	// +kuma:nolint
	Template string `json:"template,omitempty"`
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

const (
	GeneratedCondition string = "Generated"
)

const (
	GeneratedReason     string = "Generated"
	TemplateErrorReason string = "TemplateError"
	CollisionReason     string = "Collision"
)

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
	Conditions []Condition `json:"conditions,omitempty" patchStrategy:"merge" patchMergeKey:"type"`
}

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
