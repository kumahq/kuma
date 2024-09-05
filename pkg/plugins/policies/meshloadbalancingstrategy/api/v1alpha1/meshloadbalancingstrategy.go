// +kubebuilder:object:generate=true
package v1alpha1

import (
	k8s "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"

	common_api "github.com/kumahq/kuma/api/common/v1alpha1"
)

// MeshLoadBalancingStrategy is a policy that defines load balancing configuration for
// between data planes proxies.
type MeshLoadBalancingStrategy struct {
	// TargetRef is a reference to the resource the policy takes an effect on.
	// The resource could be either a real store object or virtual resource
	// defined inplace.
	TargetRef *common_api.TargetRef `json:"targetRef,omitempty"`
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
	// LocalZone defines locality aware load balancing priorities between dataplane proxies inside a zone
	LocalZone *LocalZone `json:"localZone,omitempty"`
	// CrossZone defines locality aware load balancing priorities when dataplane proxies inside local zone
	// are unavailable
	CrossZone *CrossZone `json:"crossZone,omitempty"`
}

type LocalZone struct {
	// AffinityTags list of tags for local zone load balancing.
	AffinityTags *[]AffinityTag `json:"affinityTags,omitempty"`
}

type AffinityTag struct {
	// Key defines tag for which affinity is configured
	Key string `json:"key"`
	// Weight of the tag used for load balancing. The bigger the weight the bigger the priority.
	// Percentage of local traffic load balanced to tag is computed by dividing weight by sum of weights from all tags.
	// For example with two affinity tags first with weight 80 and second with weight 20,
	// then 80% of traffic will be redirected to the first tag, and 20% of traffic will be redirected to second one.
	// Setting weights is not mandatory. When weights are not set control plane will compute default weight based on list order.
	// Default: If you do not specify weight we will adjust them so that 90% traffic goes to first tag, 9% to next, and 1% to third and so on.
	Weight *uint32 `json:"weight,omitempty"`
}

type CrossZone struct {
	// Failover defines list of load balancing rules in order of priority
	Failover []Failover `json:"failover,omitempty"`
	// FailoverThreshold defines the percentage of live destination dataplane proxies below which load balancing to the
	// next priority starts.
	// Example: If you configure failoverThreshold to 70, and you have deployed 10 destination dataplane proxies.
	// Load balancing to next priority will start when number of live destination dataplane proxies drops below 7.
	// Default 50
	FailoverThreshold *FailoverThreshold `json:"failoverThreshold,omitempty"`
}

type Failover struct {
	// From defines the list of zones to which the rule applies
	From *FromZone `json:"from,omitempty"`
	// To defines to which zones the traffic should be load balanced
	To ToZone `json:"to"`
}

type FromZone struct {
	Zones []string `json:"zones"`
}

type ToZone struct {
	// Type defines how target zones will be picked from available zones
	Type  ToZoneType `json:"type"`
	Zones *[]string  `json:"zones,omitempty"`
}

type FailoverThreshold struct {
	Percentage intstr.IntOrString `json:"percentage"`
}

// +kubebuilder:validation:Enum=None;Only;Any;AnyExcept
type ToZoneType string

const (
	// Traffic will not be load balanced to any zone
	None ToZoneType = "None"
	// Traffic will be load balanced only to zones specified in zones list
	Only ToZoneType = "Only"
	// Traffic will be load balanced to every available zone
	Any ToZoneType = "Any"
	// Traffic will be load balanced to every available zone except these specified in zones list
	AnyExcept ToZoneType = "AnyExcept"
)

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
	// +kubebuilder:validation:Minimum=2
	ChoiceCount *uint32 `json:"choiceCount,omitempty"`
	// ActiveRequestBias refers to dynamic weights applied when hosts have varying load
	// balancing weights. A higher value here aggressively reduces the weight of endpoints
	// that are currently handling active requests. In essence, the higher the ActiveRequestBias
	// value, the more forcefully it reduces the load balancing weight of endpoints that are
	// actively serving requests.
	ActiveRequestBias *intstr.IntOrString `json:"activeRequestBias,omitempty"`
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
	// +kubebuilder:validation:MinLength=1
	Name string `json:"name"`
}

type Cookie struct {
	// The name of the cookie that will be used to obtain the hash key.
	// +kubebuilder:validation:MinLength=1
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
	// +kubebuilder:validation:MinLength=1
	Name string `json:"name"`
}

type FilterState struct {
	// The name of the Object in the per-request filterState, which is
	// an Envoy::Hashable object. If there is no data associated with the key,
	// or the stored object is not Envoy::Hashable, no hash will be produced.
	// +kubebuilder:validation:MinLength=1
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
