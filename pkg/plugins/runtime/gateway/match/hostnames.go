package match

import (
	"fmt"
	"slices"
	"sort"
	"strings"

	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
)

type Hostname struct {
	Host        string
	DomainParts []string
}

func (h Hostname) String() string {
	out := h.Host
	if len(h.DomainParts) > 0 {
		out = fmt.Sprintf("%s.%s", h.Host, strings.Join(h.DomainParts, "."))
	}
	return out
}

func (h *Hostname) wildcard() bool {
	return h.Host == mesh_proto.WildcardHostname
}

func (h *Hostname) matches(name string) bool {
	n := makeHostname(name)
	return h.contains(n) || n.contains(*h)
}

func (h *Hostname) contains(n Hostname) bool {
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

func makeHostname(name string) Hostname {
	if name == "" {
		name = mesh_proto.WildcardHostname
	}
	parts := strings.Split(name, ".")
	return Hostname{Host: parts[0], DomainParts: parts[1:]}
}

func Contains(target, test string) bool {
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

type DecByInclusion []Hostname

type hostnameKey[E any] struct {
	hostname Hostname
	e        E
}

func SortHostnamesOn[S ~[]E, E any](s S, f func(a E) string) []E {
	var keys []hostnameKey[E]
	for _, e := range s {
		keys = append(keys, hostnameKey[E]{
			hostname: makeHostname(f(e)),
			e:        e,
		})
	}
	slices.SortStableFunc(keys, func(i, j hostnameKey[E]) int {
		if IsMoreExact(i.hostname, j.hostname) {
			return -1
		} else if IsMoreExact(j.hostname, i.hostname) {
			return 1
		}
		return 0
	})
	var out []E
	for _, key := range keys {
		out = append(out, key.e)
	}
	return out
}

func SortHostnamesByExactnessDec(hs []string) []string {
	var hostnames DecByInclusion
	for _, h := range hs {
		hostnames = append(hostnames, makeHostname(h))
	}
	sort.Stable(hostnames)

	var out []string
	for _, hostname := range hostnames {
		out = append(out, hostname.String())
	}

	return out
}

func (b DecByInclusion) Len() int { return len(b) }

func (b DecByInclusion) Less(i, j int) bool {
	return IsMoreExact(b[i], b[j])
}

// example.com < *.com
// *.example.com < *.com
func IsMoreExact(a, b Hostname) bool {
	if !a.wildcard() && b.wildcard() {
		return true
	}
	if a.wildcard() && !b.wildcard() {
		return false
	}
	if a.wildcard() && b.wildcard() {
		if len(a.DomainParts) > len(b.DomainParts) {
			return true
		} else if len(a.DomainParts) < len(b.DomainParts) {
			return false
		}
	}
	return a.String() < b.String()
}

func (b DecByInclusion) Swap(i, j int) { b[i], b[j] = b[j], b[i] }
