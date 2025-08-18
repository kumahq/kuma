package gateway

import (
	"strings"

	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
	core_meta "github.com/kumahq/kuma/pkg/core/metadata"
	core_mesh "github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
	"github.com/kumahq/kuma/pkg/plugins/runtime/gateway/match"
	"github.com/kumahq/kuma/pkg/plugins/runtime/gateway/route"
)

func filterGatewayRoutes(in []*core_mesh.MeshGatewayRouteResource, accept func(resource *core_mesh.MeshGatewayRouteResource) bool) []*core_mesh.MeshGatewayRouteResource {
	routes := make([]*core_mesh.MeshGatewayRouteResource, 0, len(in))

	for _, r := range in {
		if accept(r) {
			routes = append(routes, r)
		}
	}

	return routes
}

func GenerateEnvoyRouteEntries(host GatewayHost) []route.Entry {
	gatewayRoutes := filterGatewayRoutes(host.Routes, func(route *core_mesh.MeshGatewayRouteResource) bool {
		// Wildcard virtual host accepts all routes.
		if host.Hostname == mesh_proto.WildcardHostname {
			return true
		}

		// If the route has no hostnames, it matches all virtualhosts.
		names := route.Spec.GetConf().GetHttp().GetHostnames()
		if len(names) == 0 {
			return true
		}

		// Otherwise, match the virtualhost name to the route names.
		return match.Hostnames(host.Hostname, names...)
	})

	if len(gatewayRoutes) == 0 {
		return nil
	}

	var entries []route.Entry

	// Index the routes by their path. There are typically multiple
	// routes per path with additional matching criteria.
	exactEntries := map[string][]route.Entry{}
	prefixEntries := map[string][]route.Entry{}

	for _, route := range gatewayRoutes {
		for _, rule := range route.Spec.GetConf().GetHttp().GetRules() {
			entry := makeHttpRouteEntry(route.GetMeta().GetName(), rule)

			// The rule matches if any of the matches is successful (it has OR
			// semantics). That means that we have to duplicate the route table
			// entry for each repeated match so that the rule can match any of
			// the criteria.
			for _, m := range rule.GetMatches() {
				routeEntry := entry // Shallow copy.
				routeEntry.Match = makeRouteMatch(m)

				switch {
				case routeEntry.Match.ExactPath != "":
					exactEntries[routeEntry.Match.ExactPath] = append(exactEntries[routeEntry.Match.ExactPath], routeEntry)
				case routeEntry.Match.PrefixPath != "":
					prefixEntries[routeEntry.Match.PrefixPath] = append(prefixEntries[routeEntry.Match.PrefixPath], routeEntry)
				default:
					entries = append(entries, routeEntry)
				}
			}
		}
		for _, rule := range route.Spec.GetConf().GetTcp().GetRules() {
			entries = append(entries,
				makeTcpRouteEntry(route.GetMeta().GetName(), rule),
			)
		}
	}

	return HandlePrefixMatchesAndPopulatePolicies(host, exactEntries, prefixEntries, entries)
}

func HandlePrefixMatchesAndPopulatePolicies(host GatewayHost, exactEntries, prefixEntries map[string][]route.Entry, entries []route.Entry) []route.Entry {
	// The Kubernetes Ingress and Gateway APIs define prefix matching
	// to match in terms of path components, so we follow suit here.
	// Envoy path prefix matching is byte-wise, so we need to do some
	// transformations. Unless there is already an exact match for the
	// path in question, we expand each prefix path to both a prefix and
	// an exact path, duplicating the route.
	for path, pathEntries := range prefixEntries {
		exactPath := strings.TrimRight(path, "/")

		_, hasExactMatch := exactEntries[exactPath]

		for _, e := range pathEntries {
			var exactPathRewrite *string
			if rw := e.Rewrite; rw != nil && rw.ReplacePrefixMatch != nil {
				rewrite := strings.TrimRight(*rw.ReplacePrefixMatch, "/")
				exactPathRewrite = &rewrite
			}
			var exactPathRedirectPathRewrite *string
			if redir := e.Action.Redirect; redir != nil {
				if rw := redir.PathRewrite; rw != nil && rw.ReplacePrefixMatch != nil {
					rewrite := strings.TrimRight(*rw.ReplacePrefixMatch, "/")
					exactPathRedirectPathRewrite = &rewrite
				}
			}

			// Make sure the prefix has a trailing '/' so that it only matches
			// complete path components.
			e.Match.PrefixPath = exactPath + "/"

			// We need to make sure the prefix replacement
			// _also_ gets a trailing slash
			if exactPathRewrite != nil {
				replace := *exactPathRewrite + "/"
				e.Rewrite.ReplacePrefixMatch = &replace
			}
			if exactPathRedirectPathRewrite != nil {
				replace := *exactPathRedirectPathRewrite + "/"
				e.Action.Redirect.PathRewrite.ReplacePrefixMatch = &replace
			}
			entries = append(entries, e)

			// Duplicate the route to an exact match only if there
			// isn't already an exact match for this path.
			if !hasExactMatch {
				exactMatch := e
				exactMatch.Match.PrefixPath = ""
				exactMatch.Match.ExactPath = exactPath
				if exactPath == "" {
					exactMatch.Match.ExactPath = "/"
				}

				// We need to make sure this prefix replacement
				// does _not_ get a trailing slash (unless it is "/")
				if exactPathRewrite != nil {
					path := *exactPathRewrite
					if path == "" {
						path = "/"
					}
					exactMatch.Rewrite = &route.Rewrite{
						ReplaceFullPath: &path,
					}
				}
				if exactPathRedirectPathRewrite != nil {
					path := *exactPathRedirectPathRewrite
					if path == "" {
						path = "/"
					}
					redir := *exactMatch.Action.Redirect
					redir.PathRewrite = &route.Rewrite{
						ReplaceFullPath: &path,
					}
					exactMatch.Action.Redirect = &redir
				}
				exactEntries[exactPath] = append(exactEntries[exactPath], exactMatch)
			}
		}
	}

	for _, pathEntries := range exactEntries {
		entries = append(entries, pathEntries...)
	}

	return PopulatePolicies(host, entries)
}

func makeTcpRouteEntry(name string, rule *mesh_proto.MeshGatewayRoute_TcpRoute_Rule) route.Entry {
	entry := route.Entry{
		Route: name,
	}

	for _, b := range rule.GetBackends() {
		target := route.Destination{
			Destination:   b.GetDestination(),
			Weight:        b.GetWeight(),
			Policies:      nil,
			RouteProtocol: core_meta.ProtocolTCP,
		}

		entry.Action.Forward = append(entry.Action.Forward, target)
	}

	return entry
}

func makeHttpRouteEntry(name string, rule *mesh_proto.MeshGatewayRoute_HttpRoute_Rule) route.Entry {
	entry := route.Entry{
		Route: name,
	}

	for _, b := range rule.GetBackends() {
		target := route.Destination{
			Destination:   b.GetDestination(),
			Weight:        b.GetWeight(),
			Policies:      nil,
			RouteProtocol: core_meta.ProtocolHTTP,
		}

		entry.Action.Forward = append(entry.Action.Forward, target)
	}

	for _, f := range rule.GetFilters() {
		if r := f.GetRedirect(); r != nil {
			redirection := &route.Redirection{
				Status:     r.GetStatusCode(),
				Scheme:     r.GetScheme(),
				Host:       r.GetHostname(),
				Port:       r.GetPort(),
				StripQuery: true,
			}
			if p := r.GetPath(); p != nil {
				rewrite := &route.Rewrite{}
				switch t := p.GetPath().(type) {
				case *mesh_proto.MeshGatewayRoute_HttpRoute_Filter_Rewrite_ReplaceFull:
					rewrite.ReplaceFullPath = &t.ReplaceFull
				case *mesh_proto.MeshGatewayRoute_HttpRoute_Filter_Rewrite_ReplacePrefixMatch:
					rewrite.ReplacePrefixMatch = &t.ReplacePrefixMatch
				}
				redirection.PathRewrite = rewrite
			}
			entry.Action.Redirect = redirection
		} else if m := f.GetMirror(); m != nil {
			entry.Mirror = &route.Mirror{
				Percentage: m.GetPercentage().GetValue(),
				Forward: route.Destination{
					Destination: m.Backend.GetDestination(),
				},
			}
		} else if h := f.GetRequestHeader(); h != nil {
			if entry.RequestHeaders == nil {
				entry.RequestHeaders = &route.Headers{}
			}

			for _, s := range h.GetSet() {
				entry.RequestHeaders.Replace = append(
					entry.RequestHeaders.Replace, route.Pair(s.GetName(), s.GetValue()))
			}

			for _, s := range h.GetAdd() {
				entry.RequestHeaders.Append = append(
					entry.RequestHeaders.Append, route.Pair(s.GetName(), s.GetValue()))
			}

			entry.RequestHeaders.Delete = append(
				entry.RequestHeaders.Delete, h.GetRemove()...)
		} else if h := f.GetResponseHeader(); h != nil {
			if entry.ResponseHeaders == nil {
				entry.ResponseHeaders = &route.Headers{}
			}

			for _, s := range h.GetSet() {
				entry.ResponseHeaders.Replace = append(
					entry.ResponseHeaders.Replace, route.Pair(s.GetName(), s.GetValue()))
			}

			for _, s := range h.GetAdd() {
				entry.ResponseHeaders.Append = append(
					entry.ResponseHeaders.Append, route.Pair(s.GetName(), s.GetValue()))
			}

			entry.ResponseHeaders.Delete = append(
				entry.ResponseHeaders.Delete, h.GetRemove()...)
		} else if r := f.GetRewrite(); r != nil {
			rewrite := route.Rewrite{}

			if p := r.GetPath(); p != nil {
				switch t := p.(type) {
				case *mesh_proto.MeshGatewayRoute_HttpRoute_Filter_Rewrite_ReplaceFull:
					rewrite.ReplaceFullPath = &t.ReplaceFull
				case *mesh_proto.MeshGatewayRoute_HttpRoute_Filter_Rewrite_ReplacePrefixMatch:
					rewrite.ReplacePrefixMatch = &t.ReplacePrefixMatch
				}
			}

			if r.GetHostToBackendHostname() {
				rewrite.HostToBackendHostname = true
			}

			entry.Rewrite = &rewrite
		}
	}

	return entry
}

func makeRouteMatch(ruleMatch *mesh_proto.MeshGatewayRoute_HttpRoute_Match) route.Match {
	match := route.Match{}

	if p := ruleMatch.GetPath(); p != nil {
		switch p.GetMatch() {
		case mesh_proto.MeshGatewayRoute_HttpRoute_Match_Path_EXACT:
			match.ExactPath = p.GetValue()
		case mesh_proto.MeshGatewayRoute_HttpRoute_Match_Path_PREFIX:
			match.PrefixPath = p.GetValue()
		case mesh_proto.MeshGatewayRoute_HttpRoute_Match_Path_REGEX:
			match.RegexPath = p.GetValue()
		}
	} else {
		// Envoy routes require a path match, so if the route
		// didn't specify, we match any path so that the additional
		// match criteria will be applied.
		match.PrefixPath = "/"
	}

	if m := ruleMatch.GetMethod(); m != mesh_proto.HttpMethod_NONE {
		names := map[mesh_proto.HttpMethod]string{
			mesh_proto.HttpMethod_CONNECT: "CONNECT",
			mesh_proto.HttpMethod_DELETE:  "DELETE",
			mesh_proto.HttpMethod_GET:     "GET",
			mesh_proto.HttpMethod_HEAD:    "HEAD",
			mesh_proto.HttpMethod_OPTIONS: "OPTIONS",
			mesh_proto.HttpMethod_PATCH:   "PATCH",
			mesh_proto.HttpMethod_POST:    "POST",
			mesh_proto.HttpMethod_PUT:     "PUT",
			mesh_proto.HttpMethod_TRACE:   "TRACE",
		}

		match.Method = names[m]
	}

	for _, h := range ruleMatch.GetHeaders() {
		switch h.GetMatch() {
		case mesh_proto.MeshGatewayRoute_HttpRoute_Match_Header_EXACT:
			match.ExactHeader = append(
				match.ExactHeader, route.Pair(h.GetName(), h.GetValue()))
		case mesh_proto.MeshGatewayRoute_HttpRoute_Match_Header_REGEX:
			match.RegexHeader = append(
				match.RegexHeader, route.Pair(h.GetName(), h.GetValue()))
		case mesh_proto.MeshGatewayRoute_HttpRoute_Match_Header_ABSENT:
			match.AbsentHeader = append(match.AbsentHeader, h.Name)
		case mesh_proto.MeshGatewayRoute_HttpRoute_Match_Header_PRESENT:
			match.PresentHeader = append(match.PresentHeader, h.Name)
		}
	}

	for _, q := range ruleMatch.GetQueryParameters() {
		switch q.GetMatch() {
		case mesh_proto.MeshGatewayRoute_HttpRoute_Match_Query_EXACT:
			match.ExactQuery = append(
				match.ExactQuery, route.Pair(q.GetName(), q.GetValue()))
		case mesh_proto.MeshGatewayRoute_HttpRoute_Match_Query_REGEX:
			match.RegexQuery = append(
				match.RegexQuery, route.Pair(q.GetName(), q.GetValue()))
		}
	}

	return match
}
