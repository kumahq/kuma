package match

import (
	"strings"

	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
)

type hostname struct {
	Host        string
	DomainParts []string
}

func (h *hostname) wildcard() bool {
	return h.Host == mesh_proto.WildcardHostname
}

func (h *hostname) matches(name string) bool {
	n := makeHostname(name)
	return h.contains(n) || n.contains(*h)
}

func (h *hostname) contains(n hostname) bool {
	if len(h.DomainParts) > len(n.DomainParts) {
		return false
	}

	for i := 1; i <= len(h.DomainParts); i++ {
		hInd := len(h.DomainParts) - i
		nInd := len(n.DomainParts) - i
		if n.DomainParts[nInd] != h.DomainParts[hInd] {
			return false
		}
	}

	return h.wildcard() || h.Host == n.Host
}

func makeHostname(name string) hostname {
	if name == "" {
		name = mesh_proto.WildcardHostname
	}
	parts := strings.Split(name, ".")
	return hostname{Host: parts[0], DomainParts: parts[1:]}
}

func Contains(target string, test string) bool {
	targetHost := makeHostname(target)
	testHost := makeHostname(test)
	return targetHost.contains(testHost)
}

// Hostnames returns true if target is a host or domain name match for
// any of the given matches. All the hostnames are assumed to be fully
// qualified (e.g. "foo.example.com") or wildcards (e.g. "*.example.com).
//
// # Two hostnames match if
//
// 1. They are exactly equal, OR
// 2. One of them is a domain wildcard and the domain part matches.
func Hostnames(target string, matches ...string) bool {
	targetHost := makeHostname(target)

	for _, m := range matches {
		if m == target {
			return true
		}
		if targetHost.matches(m) {
			return true
		}
	}

	return false
}
