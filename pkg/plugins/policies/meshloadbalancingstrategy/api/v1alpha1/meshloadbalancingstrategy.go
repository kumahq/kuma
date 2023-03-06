// +kubebuilder:object:generate=true
package v1alpha1

import (
	k8s "k8s.io/apimachinery/pkg/apis/meta/v1"

	common_api "github.com/kumahq/kuma/api/common/v1alpha1"
)

// MeshLoadBalancingStrategy is a policy that defines load balancing configuration for
// between data planes proxies.
type MeshLoadBalancingStrategy struct {
	// TargetRef is a reference to the resource the policy takes an effect on.
	// The resource could be either a real store object or virtual resource
	// defined inplace.
	TargetRef common_api.TargetRef `json:"targetRef"`
	// To list makes a match between the consumed services and corresponding configurations
	To []To `json:"to,omitempty"`
}

type To struct {
	// TargetRef is a reference to the resource that represents a group of
	// destinations.
	TargetRef common_api.TargetRef `json:"targetRef"`
	// Default is a configuration specific to the group of destinations referenced in
	// 'targetRef'
	Default Conf `json:"default,omitempty"`
}

type Conf struct {
	// LocalityAwareness contains configuration for locality aware load balancing.
	LocalityAwareness *LocalityAwareness `json:"localityAwareness,omitempty"`
	// LoadBalancer allows to specify load balancing algorithm.
	LoadBalancer *LoadBalancer `json:"loadBalancer,omitempty"`
}

type LocalityAwareness struct {
	// Disabled allows to disable locality-aware load balancing.
	// When disabled requests are distributed across all endpoints regardless of locality.
	Disabled *bool `json:"disabled,omitempty"`
}

// +kubebuilder:validation:Enum=RoundRobin;LeastRequest;RingHash;Random;Maglev
type LoadBalancerType string

const (
	RoundRobinType   LoadBalancerType = "RoundRobin"
	LeastRequestType LoadBalancerType = "LeastRequest"
	RingHashType     LoadBalancerType = "RingHash"
	RandomType       LoadBalancerType = "Random"
	MaglevType       LoadBalancerType = "Maglev"
)

type LoadBalancer struct {
	Type LoadBalancerType `json:"type"`

	// RoundRobin is a load balancing algorithm that distributes requests
	// across available upstream hosts in round-robin order.
	RoundRobin *RoundRobin `json:"roundRobin,omitempty"`

	// LeastRequest selects N random available hosts as specified in 'choiceCount' (2 by default)
	// and picks the host which has the fewest active requests
	LeastRequest *LeastRequest `json:"leastRequest,omitempty"`

	// RingHash  implements consistent hashing to upstream hosts. Each host is mapped
	// onto a circle (the “ring”) by hashing its address; each request is then routed
	// to a host by hashing some property of the request, and finding the nearest
	// corresponding host clockwise around the ring.
	RingHash *RingHash `json:"ringHash,omitempty"`

	// Random selects a random available host. The random load balancer generally
	// performs better than round-robin if no health checking policy is configured.
	// Random selection avoids bias towards the host in the set that comes after a failed host.
	Random *Random `json:"random,omitempty"`

	// Maglev implements consistent hashing to upstream hosts. Maglev can be used as
	// a drop in replacement for the ring hash load balancer any place in which
	// consistent hashing is desired.
	Maglev *Maglev `json:"maglev,omitempty"`
}

type RoundRobin struct{}

type LeastRequest struct {
	// ChoiceCount is the number of random healthy hosts from which the host with
	// the fewest active requests will be chosen. Defaults to 2 so that Envoy performs
	// two-choice selection if the field is not set.
	ChoiceCount *uint32 `json:"choiceCount,omitempty"`
}

// +kubebuilder:validation:Enum=XXHash;MurmurHash2
type HashFunctionType string

const (
	XXHashType      HashFunctionType = "XXHash"
	MurmurHash2Type HashFunctionType = "MurmurHash2"
)

type RingHash struct {
	// HashFunction is a function used to hash hosts onto the ketama ring.
	// The value defaults to XX_HASH. Available values – XX_HASH, MURMUR_HASH_2.
	HashFunction *HashFunctionType `json:"hashFunction,omitempty"`

	// Minimum hash ring size. The larger the ring is (that is,
	// the more hashes there are for each provided host) the better the request distribution
	// will reflect the desired weights. Defaults to 1024 entries, and limited to 8M entries.
	// +kubebuilder:validation:Minimum=1
	// +kubebuilder:validation:Maximum=8000000
	MinRingSize *uint32 `json:"minRingSize,omitempty"`

	// Maximum hash ring size. Defaults to 8M entries, and limited to 8M entries,
	// but can be lowered to further constrain resource use.
	// +kubebuilder:validation:Minimum=1
	// +kubebuilder:validation:Maximum=8000000
	MaxRingSize *uint32 `json:"maxRingSize,omitempty"`

	// HashPolicies specify a list of request/connection properties that are used to calculate a hash.
	// These hash policies are executed in the specified order. If a hash policy has the “terminal” attribute
	// set to true, and there is already a hash generated, the hash is returned immediately,
	// ignoring the rest of the hash policy list.
	HashPolicies *[]HashPolicy `json:"hashPolicies,omitempty"`
}

// +kubebuilder:validation:Enum=Header;Cookie;SourceIP;QueryParameter;FilterState
type HashPolicyType string

const (
	HeaderType         HashPolicyType = "Header"
	CookieType         HashPolicyType = "Cookie"
	ConnectionType     HashPolicyType = "Connection"
	QueryParameterType HashPolicyType = "QueryParameter"
	FilterStateType    HashPolicyType = "FilterState"
)

type HashPolicy struct {
	Type HashPolicyType `json:"type"`

	// Terminal is a flag that short-circuits the hash computing. This field provides
	// a ‘fallback’ style of configuration: “if a terminal policy doesn’t work, fallback
	// to rest of the policy list”, it saves time when the terminal policy works.
	// If true, and there is already a hash computed, ignore rest of the list of hash polices.
	Terminal *bool `json:"terminal,omitempty"`

	Header         *Header         `json:"header,omitempty"`
	Cookie         *Cookie         `json:"cookie,omitempty"`
	Connection     *Connection     `json:"connection,omitempty"`
	QueryParameter *QueryParameter `json:"queryParameter,omitempty"`
	FilterState    *FilterState    `json:"filterState,omitempty"`
}

type Header struct {
	// The name of the request header that will be used to obtain the hash key.
	Name string `json:"name"`
}

type Cookie struct {
	// The name of the cookie that will be used to obtain the hash key.
	Name string `json:"name"`
	// If specified, a cookie with the TTL will be generated if the cookie is not present.
	TTL *k8s.Duration `json:"ttl,omitempty"`
	// The name of the path for the cookie.
	Path *string `json:"path,omitempty"`
}

type Connection struct {
	// Hash on source IP address.
	SourceIP *bool `json:"sourceIP,omitempty"`
}

type QueryParameter struct {
	// The name of the URL query parameter that will be used to obtain the hash key.
	// If the parameter is not present, no hash will be produced. Query parameter names
	// are case-sensitive.
	Name string `json:"name"`
}

type FilterState struct {
	// The name of the Object in the per-request filterState, which is
	// an Envoy::Hashable object. If there is no data associated with the key,
	// or the stored object is not Envoy::Hashable, no hash will be produced.
	Key string `json:"key"`
}

type Random struct{}

type Maglev struct {
	// The table size for Maglev hashing. Maglev aims for “minimal disruption”
	// rather than an absolute guarantee. Minimal disruption means that when
	// the set of upstream hosts change, a connection will likely be sent
	// to the same upstream as it was before. Increasing the table size reduces
	// the amount of disruption. The table size must be prime number limited to 5000011.
	// If it is not specified, the default is 65537.
	// +kubebuilder:validation:Minimum=1
	// +kubebuilder:validation:Maximum=5000011
	TableSize *uint32 `json:"tableSize,omitempty"`

	// HashPolicies specify a list of request/connection properties that are used to calculate a hash.
	// These hash policies are executed in the specified order. If a hash policy has the “terminal” attribute
	// set to true, and there is already a hash generated, the hash is returned immediately,
	// ignoring the rest of the hash policy list.
	HashPolicies *[]HashPolicy `json:"hashPolicies,omitempty"`
}
