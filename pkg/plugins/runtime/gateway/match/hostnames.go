package match

import "strings"

type hostname struct {
	Host   string
	Domain string
}

func (h *hostname) wildcard() bool {
	return h.Host == "*"
}

func (h *hostname) matches(name string) bool {
	n := makeHostname(name)

	if h.wildcard() || n.wildcard() {
		return h.Domain == n.Domain
	}

	return h.Host == n.Host && h.Domain == n.Domain
}

func makeHostname(name string) hostname {
	parts := strings.SplitN(name, ".", 2)
	return hostname{Host: parts[0], Domain: parts[1]}
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
