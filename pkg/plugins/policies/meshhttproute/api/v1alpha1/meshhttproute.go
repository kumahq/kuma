// +kubebuilder:object:generate=true
package v1alpha1

import (
	"encoding/json"

	"k8s.io/apimachinery/pkg/util/intstr"

	common_api "github.com/kumahq/kuma/api/common/v1alpha1"
	"github.com/kumahq/kuma/pkg/xds/cache/sha256"
)

// MeshHTTPRoute
// +kuma:policy:singular_display_name=Mesh HTTP Route
//
// This policy defines its own `GetDefault` method so that it can have the given
// structure for deserialization but still use the generic policy merging
// machinery.
//
// +kuma:policy:skip_get_default=true
type MeshHTTPRoute struct {
	// TargetRef is a reference to the resource the policy takes an effect on.
	// The resource could be either a real store object or virtual resource
	// defined inplace.
	TargetRef *common_api.TargetRef `json:"targetRef,omitempty"`

	// To matches destination services of requests and holds configuration.
	To []To `json:"to,omitempty"`
}

type To struct {
	// Hostnames is only valid when targeting MeshGateway and limits the
	// effects of the rules to requests to this hostname.
	// Given hostnames must intersect with the hostname of the listeners the
	// route attaches to.
	Hostnames []string `json:"hostnames,omitempty"`
	// TargetRef is a reference to the resource that represents a group of
	// request destinations.
	TargetRef common_api.TargetRef `json:"targetRef,omitempty"`
	// Rules contains the routing rules applies to a combination of top-level
	// targetRef and the targetRef in this entry.
	Rules []Rule `json:"rules,omitempty"`
}

type Rule struct {
	// Matches describes how to match HTTP requests this rule should be applied
	// to.
	// +kubebuilder:validation:MinItems=1
	Matches []Match `json:"matches" policyMerge:"mergeKey"`
	// Default holds routing rules that can be merged with rules from other
	// policies.
	Default RuleConf `json:"default"`
}

func HashMatches(m []Match) string {
	bytes, _ := json.Marshal(m)
	h := sha256.Hash(string(bytes))
	return h
}

type Match struct {
	Path   *PathMatch `json:"path,omitempty"`
	Method *Method    `json:"method,omitempty"`
	// QueryParams matches based on HTTP URL query parameters. Multiple matches
	// are ANDed together such that all listed matches must succeed.
	QueryParams []QueryParamsMatch       `json:"queryParams,omitempty"`
	Headers     []common_api.HeaderMatch `json:"headers,omitempty"`
}

// +kubebuilder:validation:Enum=Exact;PathPrefix;RegularExpression
type PathMatchType string

// +kubebuilder:validation:Enum=CONNECT;DELETE;GET;HEAD;OPTIONS;PATCH;POST;PUT;TRACE
type Method string

const (
	Exact             PathMatchType = "Exact"
	PathPrefix        PathMatchType = "PathPrefix"
	RegularExpression PathMatchType = "RegularExpression"
)

type PathMatch struct {
	// Exact or prefix matches must be an absolute path. A prefix matches only
	// if separated by a slash or the entire path.
	// +kubebuilder:validation:MinLength=1
	Value string        `json:"value"`
	Type  PathMatchType `json:"type"`
}

// +kubebuilder:validation:Enum=Exact;RegularExpression
type QueryParamsMatchType string

const (
	ExactQueryMatch             QueryParamsMatchType = "Exact"
	RegularExpressionQueryMatch QueryParamsMatchType = "RegularExpression"
)

type QueryParamsMatch struct {
	Type QueryParamsMatchType `json:"type"`
	// +kubebuilder:validation:MinLength=1
	Name  string `json:"name"`
	Value string `json:"value"`
}

type RuleConf struct {
	Filters     *[]Filter                `json:"filters,omitempty"`
	BackendRefs *[]common_api.BackendRef `json:"backendRefs,omitempty"`
}

// +kubebuilder:validation:Enum=RequestHeaderModifier;ResponseHeaderModifier;RequestRedirect;URLRewrite;RequestMirror
type FilterType string

const (
	RequestHeaderModifierType  FilterType = "RequestHeaderModifier"
	ResponseHeaderModifierType FilterType = "ResponseHeaderModifier"
	RequestRedirectType        FilterType = "RequestRedirect"
	URLRewriteType             FilterType = "URLRewrite"
	RequestMirrorType          FilterType = "RequestMirror"
)

type HeaderKeyValue struct {
	Name  common_api.HeaderName  `json:"name"`
	Value common_api.HeaderValue `json:"value"`
}

// Only one action is supported per header name.
// Configuration to set or add multiple values for a header must use RFC 7230
// header value formatting, separating each value with a comma.
type HeaderModifier struct {
	// +listType=map
	// +listMapKey=name
	// +kubebuilder:validation:MaxItems=16
	Set []HeaderKeyValue `json:"set,omitempty"`
	// +listType=map
	// +listMapKey=name
	// +kubebuilder:validation:MaxItems=16
	Add []HeaderKeyValue `json:"add,omitempty"`
	// +kubebuilder:validation:MaxItems=16
	Remove []string `json:"remove,omitempty"`
}

// PreciseHostname is the fully qualified domain name of a network host. This
// matches the RFC 1123 definition of a hostname with 1 notable exception that
// numeric IP addresses are not allowed.
//
// Note that as per RFC1035 and RFC1123, a *label* must consist of lower case
// alphanumeric characters or '-', and must start and end with an alphanumeric
// character. No other punctuation is allowed.
//
// +kubebuilder:validation:MinLength=1
// +kubebuilder:validation:MaxLength=253
// +kubebuilder:validation:Pattern=`^[a-z0-9]([-a-z0-9]*[a-z0-9])?(\.[a-z0-9]([-a-z0-9]*[a-z0-9])?)*$`
type PreciseHostname string

// PortNumber defines a network port.
//
// +kubebuilder:validation:Minimum=1
// +kubebuilder:validation:Maximum=65535
type PortNumber int32

type RequestRedirect struct {
	// +kubebuilder:validation:Enum=http;https
	Scheme   *string          `json:"scheme,omitempty"`
	Hostname *PreciseHostname `json:"hostname,omitempty"`
	// Path defines parameters used to modify the path of the incoming request.
	// The modified path is then used to construct the location header.
	// When empty, the request path is used as-is.
	Path *PathRewrite `json:"path,omitempty"`
	// Port is the port to be used in the value of the `Location`
	// header in the response.
	// When empty, port (if specified) of the request is used.
	Port *PortNumber `json:"port,omitempty"`
	// StatusCode is the HTTP status code to be used in response.
	//
	// +kubebuilder:default=302
	// +kubebuilder:validation:Enum=301;302;303;307;308
	StatusCode *int `json:"statusCode,omitempty"`
}

// +kubebuilder:validation:Enum=ReplaceFullPath;ReplacePrefixMatch
type PathRewriteType string

const (
	ReplaceFullPathType    PathRewriteType = "ReplaceFullPath"
	ReplacePrefixMatchType PathRewriteType = "ReplacePrefixMatch"
)

type PathRewrite struct {
	Type               PathRewriteType `json:"type"`
	ReplaceFullPath    *string         `json:"replaceFullPath,omitempty"`
	ReplacePrefixMatch *string         `json:"replacePrefixMatch,omitempty"`
}

type URLRewrite struct {
	// Hostname is the value to be used to replace the host header value during forwarding.
	Hostname *PreciseHostname `json:"hostname,omitempty"`
	// Path defines a path rewrite.
	Path *PathRewrite `json:"path,omitempty"`
	// HostToBackendHostname rewrites the hostname to the hostname of the
	// upstream host. This option is only available when targeting MeshGateways.
	HostToBackendHostname bool `json:"hostToBackendHostname,omitempty"`
}

type RequestMirror struct {
	// Percentage of requests to mirror. If not specified, all requests
	// to the target cluster will be mirrored.
	Percentage *intstr.IntOrString `json:"percentage,omitempty"`
	// TODO forbid weight
	BackendRef common_api.BackendRef `json:"backendRef"`
}

type Filter struct {
	Type                   FilterType       `json:"type"`
	RequestHeaderModifier  *HeaderModifier  `json:"requestHeaderModifier,omitempty"`
	ResponseHeaderModifier *HeaderModifier  `json:"responseHeaderModifier,omitempty"`
	RequestRedirect        *RequestRedirect `json:"requestRedirect,omitempty"`
	URLRewrite             *URLRewrite      `json:"urlRewrite,omitempty"`
	RequestMirror          *RequestMirror   `json:"requestMirror,omitempty"`
}
