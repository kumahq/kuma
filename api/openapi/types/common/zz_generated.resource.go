// Package types provides primitives to interact with the openapi HTTP API.
//
// Code generated by github.com/deepmap/oapi-codegen version v1.15.0 DO NOT EDIT.
package types

// Defines values for ResourceTypeDescriptionScope.
const (
	Global ResourceTypeDescriptionScope = "Global"
	Mesh   ResourceTypeDescriptionScope = "Mesh"
)

// FromRule defines model for FromRule.
type FromRule struct {
	Inbound Inbound `json:"inbound"`
	Rules   []Rule  `json:"rules"`
}

// Inbound defines model for Inbound.
type Inbound struct {
	Port int               `json:"port"`
	Tags map[string]string `json:"tags"`
}

// InspectRule defines model for InspectRule.
type InspectRule struct {
	// FromRules a set of rules for each inbound of this proxy
	FromRules *[]FromRule `json:"fromRules,omitempty"`
	ProxyRule *ProxyRule  `json:"proxyRule,omitempty"`

	// ToRules a set of rules for the outbounds of this proxy
	ToRules *[]Rule `json:"toRules,omitempty"`

	// Type the type of the policy
	Type string `json:"type"`

	// Warnings a set of warnings to show in policy matching
	Warnings *[]string `json:"warnings,omitempty"`
}

// Meta defines model for Meta.
type Meta struct {
	// Mesh the mesh this resource is part of
	Mesh string `json:"mesh"`

	// Name the name of the resource
	Name string `json:"name"`

	// Type the type of this resource
	Type string `json:"type"`
}

// PolicyDescription information about a policy
type PolicyDescription struct {
	// HasFromTargetRef indicates that this policy can be used as an inbound policy
	HasFromTargetRef bool `json:"hasFromTargetRef"`

	// HasToTargetRef indicates that this policy can be used as an outbound policy
	HasToTargetRef bool `json:"hasToTargetRef"`

	// IsTargetRef whether this policy uses targetRef matching
	IsTargetRef bool `json:"isTargetRef"`
}

// ProxyRule defines model for ProxyRule.
type ProxyRule struct {
	// Conf The actual conf generated
	Conf   interface{} `json:"conf"`
	Origin []Meta      `json:"origin"`
}

// ResourceTypeDescription Description of a resource type, this is useful for dynamically generated clients and the gui
type ResourceTypeDescription struct {
	// IncludeInDump description resources of this type should be included in dump (especially useful for moving from non-federated to federated or migrating to a new global).
	IncludeInDump bool `json:"includeInDump"`

	// Name the name of the resource type
	Name string `json:"name"`

	// Path the path to use for accessing this resource. If scope is `Global` then it will be `/<path>` otherwise it will be `/meshes/<path>`
	Path              string `json:"path"`
	PluralDisplayName string `json:"pluralDisplayName"`

	// Policy information about a policy
	Policy              *PolicyDescription           `json:"policy,omitempty"`
	ReadOnly            bool                         `json:"readOnly"`
	Scope               ResourceTypeDescriptionScope `json:"scope"`
	SingularDisplayName string                       `json:"singularDisplayName"`
}

// ResourceTypeDescriptionScope defines model for ResourceTypeDescription.Scope.
type ResourceTypeDescriptionScope string

// Rule defines model for Rule.
type Rule struct {
	// Conf The actual conf generated
	Conf     interface{}   `json:"conf"`
	Matchers []RuleMatcher `json:"matchers"`
	Origin   []Meta        `json:"origin"`
}

// RuleMatcher A matcher to select which traffic this conf applies to
type RuleMatcher struct {
	// Key the key to match against
	Key string `json:"key"`

	// Not whether we check on the absence of this key:value pair
	Not bool `json:"not"`

	// Value the value for the key to match against
	Value string `json:"value"`
}
