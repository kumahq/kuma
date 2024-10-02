package route

import (
	core_mesh "github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
	"github.com/kumahq/kuma/pkg/core/resources/model"
	"github.com/kumahq/kuma/pkg/xds/envoy/tags"
)

// Table stores a collection of routing Entries, aka. a routing table.
//
// A route table is a simplified model of a collection of Envoy routes.
// Although extracting a model is duplicative of the actual Envoy
// model, this approach makes it possible to support different APIs
// that combine to make a coherent routing table.
//
// The different elements of the routing table are deliberately defined
// without any behavior or internal consistency checks. The API that
// generates these elements should apply validation to ensure that it
// generates route table elements that make sense and accurately capture
// the right semantics.
type Table struct {
	Entries []Entry
}

// Entry is a single routing element. Incoming requests are matched by Match
// and dispatched according to the Action. Other optional field specify
// additional processing.
type Entry struct {
	Route  string
	Name   string
	Match  Match
	Action Action

	// Mirror specifies whether to mirror matching traffic.
	Mirror *Mirror

	// RequestHeaders specifies transformations on the HTTP
	// request headers.
	RequestHeaders *Headers

	// ResponseHeaders specifies transformations on the HTTP
	// response headers.
	ResponseHeaders *Headers

	Rewrite *Rewrite
}

// KeyValue is a generic pairing of key and value strings. Route table
// elements generally use this in preference to maps so that input ordering
// is preserved and output does not change based on map iteration order.
type KeyValue struct {
	Key   string
	Value string
}

// Pair combines key and value into a KeyValue.
func Pair(key string, value string) KeyValue {
	return KeyValue{
		Key:   key,
		Value: value,
	}
}

// Match describes how to match a HTTP request.
type Match struct {
	ExactPath  string
	PrefixPath string
	RegexPath  string

	Method string

	ExactHeader   []KeyValue // name -> value
	RegexHeader   []KeyValue // name -> regex
	AbsentHeader  []string
	PresentHeader []string
	PrefixHeader  []KeyValue

	ExactQuery []KeyValue // param -> value
	RegexQuery []KeyValue // param -> regex
}

func (m Match) numHeaderMatches() int {
	return len(m.ExactHeader) + len(m.RegexHeader) + len(m.AbsentHeader) + len(m.PresentHeader)
}

func (m Match) numQueryParamMatches() int {
	return len(m.ExactQuery) + len(m.RegexQuery)
}

// Action describes how a HTTP request should be dispatched.
type Action struct {
	Forward  []Destination
	Redirect *Redirection
	Respond  struct{} // TODO(jpeach) add DirectResponseAction support
}

// Redirection is an action that responds to a HTTP request with a HTTP
// redirect response.
type Redirection struct {
	Status      uint32 // HTTP status code.
	Scheme      string // URL scheme (optional).
	Host        string // URL host (optional).
	Port        uint32 // URL port (optional).
	PathRewrite *Rewrite

	StripQuery bool // Whether to strip the query string.
}

// Destination is a forwarding target (aka Cluster).
type Destination struct {
	Destination tags.Tags
	BackendRef  *model.ResolvedBackendRef

	Weight        uint32
	RouteProtocol core_mesh.Protocol

	// Name is the globally unique name for this destination instance.
	// It takes into account not only the service that it targets, but
	// also the configuration context.
	Name string

	// Kuma connection policies for traffic forwarded to
	// this destination.
	Policies map[model.ResourceType]model.Resource
}

// Headers is a set of operations to perform on HTTP message headers.
type Headers struct {
	// Append adds a value to a HTTP header field.
	Append []KeyValue
	// Replace adds a value to a HTTP header field, removing all other
	// values for that field.
	Replace []KeyValue
	// Delete deletes a HTTP header field.
	Delete []string
}

type Rewrite struct {
	ReplaceFullPath *string

	ReplacePrefixMatch *string

	// HostToBackendHostname indicates that during forwarding, the host header
	// should be swapped with the hostname of the upstream host chosen by the
	// Envoy's cluster manager.
	HostToBackendHostname bool

	ReplaceHostname *string
}

// Mirror specifies a traffic mirroring operation.
type Mirror struct {
	Forward    Destination
	Percentage float64
}
