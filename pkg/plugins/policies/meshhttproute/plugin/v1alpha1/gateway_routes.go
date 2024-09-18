package v1alpha1

import (
	"slices"
	"strings"

	"golang.org/x/exp/maps"
	"k8s.io/apimachinery/pkg/util/intstr"

	common_api "github.com/kumahq/kuma/api/common/v1alpha1"
	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
	core_mesh "github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
	"github.com/kumahq/kuma/pkg/core/resources/model"
	"github.com/kumahq/kuma/pkg/plugins/policies/core/rules"
	meshroute_gateway "github.com/kumahq/kuma/pkg/plugins/policies/core/xds/meshroute/gateway"
	api "github.com/kumahq/kuma/pkg/plugins/policies/meshhttproute/api/v1alpha1"
	plugin_gateway "github.com/kumahq/kuma/pkg/plugins/runtime/gateway"
	"github.com/kumahq/kuma/pkg/plugins/runtime/gateway/match"
	"github.com/kumahq/kuma/pkg/plugins/runtime/gateway/metadata"
	"github.com/kumahq/kuma/pkg/plugins/runtime/gateway/route"
	"github.com/kumahq/kuma/pkg/util/pointer"
	xds_context "github.com/kumahq/kuma/pkg/xds/context"
	"github.com/kumahq/kuma/pkg/xds/envoy/tags"
)

type ruleByHostname struct {
	Rule     ToRouteRule
	Hostname string
}

func sortRulesToHosts(
	meshLocalResources xds_context.ResourceMap,
	rawRules rules.GatewayRules,
	address string,
	port uint32,
	protocol mesh_proto.MeshGateway_Listener_Protocol,
	sublisteners []meshroute_gateway.Sublistener,
	resolver model.LabelResourceIdentifierResolver,
) []plugin_gateway.GatewayListenerHostname {
	hostInfosByHostname := map[string]plugin_gateway.GatewayListenerHostname{}

	// Iterate over the listeners in order
	sublistenersByHostname := map[string]meshroute_gateway.Sublistener{}
	for _, sublistener := range sublisteners {
		sublistenersByHostname[sublistener.Hostname] = sublistener
	}

	// Keep track of which hostnames we've seen at the listener level
	// so we can track which matches can't be matched at a different listener
	// Very important for the HTTP listeners where every virtual host ends up
	// under the same route config
	var observedHostnames []string

	for _, hostname := range match.SortHostnamesByExactnessDec(maps.Keys(sublistenersByHostname)) {
		hostnameTag := sublistenersByHostname[hostname]
		inboundListener := rules.NewInboundListenerHostname(
			address,
			port,
			hostnameTag.Hostname,
		)
		rawRules, ok := rawRules.ToRules.ByListenerAndHostname[inboundListener]
		if !ok {
			continue
		}
		var ruleHostnames []string
		rulesByHostname := map[string][]ruleByHostname{}
		for _, rawRule := range rawRules {
			conf := rawRule.Conf.(api.PolicyDefault)

			backendRefOrigin := map[common_api.MatchesHash]model.ResourceMeta{}
			for hash := range rawRule.BackendRefOriginIndex {
				if origin, ok := rawRule.GetBackendRefOrigin(hash); ok {
					backendRefOrigin[hash] = origin
				}
			}
			rule := ToRouteRule{
				Subset:           rawRule.Subset,
				Rules:            conf.Rules,
				Hostnames:        conf.Hostnames,
				Origins:          rawRule.Origin,
				BackendRefOrigin: backendRefOrigin,
			}
			hostnames := rule.Hostnames
			if len(rule.Hostnames) == 0 {
				hostnames = []string{"*"}
			}
			for _, hostname := range hostnames {
				accRule, ok := rulesByHostname[hostname]
				if !ok {
					ruleHostnames = append(ruleHostnames, hostname)
				}
				rulesByHostname[hostname] = append(accRule, ruleByHostname{
					Rule:     rule,
					Hostname: hostname,
				})
			}
		}

		// We need to find the set of hostnames that are contained by the given
		// listener hostname
		var listenerSpecificHostnameMatches []string
		listenerSpecificHostnameMatchSet := map[string]struct{}{}
		for _, ruleHostname := range ruleHostnames {
			if !match.Hostnames(hostnameTag.Hostname, ruleHostname) {
				continue
			}
			hostnameMatch := ruleHostname
			if match.Contains(ruleHostname, hostnameTag.Hostname) {
				hostnameMatch = hostnameTag.Hostname
			}
			if _, ok := listenerSpecificHostnameMatchSet[hostnameMatch]; !ok {
				listenerSpecificHostnameMatches = append(listenerSpecificHostnameMatches, hostnameMatch)
				listenerSpecificHostnameMatchSet[hostnameMatch] = struct{}{}
			}
		}

		for _, hostnameMatch := range listenerSpecificHostnameMatches {
			// Find all rules that match this hostname
			var rules []ruleByHostname
			for ruleHostname, hostnameRules := range rulesByHostname {
				if !match.Hostnames(hostnameMatch, ruleHostname) {
					continue
				}
				// If the rules hostname is more specific than the hostname
				// match, they already have their own match
				if _, ok := listenerSpecificHostnameMatchSet[ruleHostname]; ok && !match.Contains(ruleHostname, hostnameMatch) {
					continue
				}
				rules = append(rules, hostnameRules...)
			}
			// As mentioned above we shouldn't add rules if this hostname match
			// can't match because of a listener hostname
			var hostnameMatchUnmatchable bool
			for _, observedHostname := range observedHostnames {
				hostnameMatchUnmatchable = hostnameMatchUnmatchable || match.Contains(observedHostname, hostnameMatch)
			}
			if hostnameMatchUnmatchable {
				continue
			}
			// Create an info for every hostname match
			// We may end up duplicating info more than once so we copy it here
			host := plugin_gateway.GatewayHost{
				Hostname: hostnameMatch,
				Routes:   nil,
				Policies: map[model.ResourceType][]match.RankedPolicy{},
				Tags:     hostnameTag.Tags,
			}
			for _, t := range plugin_gateway.ConnectionPolicyTypes {
				matches := match.ConnectionPoliciesBySource(
					host.Tags,
					match.ToConnectionPolicies(meshLocalResources[t]))
				host.Policies[t] = matches
			}
			hostInfo := plugin_gateway.GatewayHostInfo{
				Host: host,
			}
			hostInfo.AppendEntries(generateEnvoyRouteEntries(host, rules, resolver))

			meshroute_gateway.AddToListenerByHostname(
				hostInfosByHostname,
				protocol,
				hostnameTag.Hostname,
				hostnameTag.TLS,
				hostInfo,
			)
		}
		observedHostnames = append(observedHostnames, hostname)
	}

	return meshroute_gateway.SortByHostname(hostInfosByHostname)
}

func generateEnvoyRouteEntries(
	host plugin_gateway.GatewayHost,
	toRules []ruleByHostname,
	resolver model.LabelResourceIdentifierResolver,
) []route.Entry {
	var entries []route.Entry

	toRules = match.SortHostnamesOn(toRules, func(r ruleByHostname) string { return r.Hostname })

	// Index the routes by their path. There are typically multiple
	// routes per path with additional matching criteria.
	exactEntries := map[string][]route.Entry{}
	prefixEntries := map[string][]route.Entry{}

	for _, rules := range toRules {
		for _, rule := range rules.Rule.Rules {
			var names []string
			for _, orig := range rules.Rule.Origins {
				names = append(names, orig.GetName())
			}
			slices.Sort(names)

			entry := makeHttpRouteEntry(strings.Join(names, "_"), rule, rules.Rule.BackendRefOrigin, resolver)

			hashedMatches := api.HashMatches(rule.Matches)
			// The rule matches if any of the matches is successful (it has OR
			// semantics). That means that we have to duplicate the route table
			// entry for each repeated match so that the rule can match any of
			// the criteria.
			for _, m := range rule.Matches {
				routeEntry := entry // Shallow copy.
				routeEntry.Match = makeRouteMatch(m)
				routeEntry.Name = string(hashedMatches)

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
	}

	return plugin_gateway.HandlePrefixMatchesAndPopulatePolicies(host, exactEntries, prefixEntries, entries)
}

func makeHttpRouteEntry(
	name string,
	rule api.Rule,
	backendRefToOrigin map[common_api.MatchesHash]model.ResourceMeta,
	resolver model.LabelResourceIdentifierResolver,
) route.Entry {
	entry := route.Entry{
		Route: name,
	}

	for _, b := range pointer.Deref(rule.Default.BackendRefs) {
		var ref *model.ResolvedBackendRef
		if origin, ok := backendRefToOrigin[api.HashMatches(rule.Matches)]; ok {
			ref = model.ResolveBackendRef(origin, b, resolver)
		}
		var dest map[string]string
		if ref == nil || ref.ResourceOrNil() == nil {
			// We have a legacy backendRef
			if !b.ReferencesRealObject() {
				var ok bool
				dest, ok = tags.FromLegacyTargetRef(b.TargetRef)
				if !ok {
					// This should be caught by validation
					continue
				}
			} else {
				// We have a real backendRef but it's not valid
				dest = map[string]string{
					mesh_proto.ServiceTag: metadata.UnresolvedBackendServiceTag,
				}
			}
		}
		target := route.Destination{
			Destination:   dest,
			BackendRef:    ref,
			Weight:        uint32(pointer.DerefOr(b.Weight, 1)),
			Policies:      nil,
			RouteProtocol: core_mesh.ProtocolHTTP,
		}

		entry.Action.Forward = append(entry.Action.Forward, target)
	}

	for _, f := range pointer.Deref(rule.Default.Filters) {
		if r := f.RequestRedirect; r != nil {
			redirection := &route.Redirection{
				Status:     uint32(pointer.DerefOr(r.StatusCode, 302)),
				Scheme:     pointer.Deref(r.Scheme),
				Host:       string(pointer.Deref(r.Hostname)),
				Port:       uint32(pointer.Deref(r.Port)),
				StripQuery: true,
			}
			if p := r.Path; p != nil {
				rewrite := &route.Rewrite{}
				switch p.Type {
				case api.ReplaceFullPathType:
					rewrite.ReplaceFullPath = p.ReplaceFullPath
				case api.ReplacePrefixMatchType:
					rewrite.ReplacePrefixMatch = p.ReplacePrefixMatch
				}
				redirection.PathRewrite = rewrite
			}
			entry.Action.Redirect = redirection
		} else if m := f.RequestMirror; m != nil {
			decimal, err := common_api.NewDecimalFromIntOrString(
				pointer.DerefOr(f.RequestMirror.Percentage, intstr.FromInt(100)),
			)
			if err != nil {
				continue
			}
			tags, ok := tags.FromLegacyTargetRef(m.BackendRef.TargetRef)
			if !ok {
				continue
			}
			entry.Mirror = &route.Mirror{
				Percentage: decimal.InexactFloat64(),
				Forward: route.Destination{
					Destination: tags,
				},
			}
		} else if h := f.RequestHeaderModifier; h != nil {
			if entry.RequestHeaders == nil {
				entry.RequestHeaders = &route.Headers{}
			}

			for _, s := range h.Set {
				entry.RequestHeaders.Replace = append(
					entry.RequestHeaders.Replace, route.Pair(string(s.Name), string(s.Value)))
			}

			for _, s := range h.Add {
				entry.RequestHeaders.Append = append(
					entry.RequestHeaders.Append, route.Pair(string(s.Name), string(s.Value)))
			}

			entry.RequestHeaders.Delete = append(
				entry.RequestHeaders.Delete, h.Remove...)
		} else if h := f.ResponseHeaderModifier; h != nil {
			if entry.ResponseHeaders == nil {
				entry.ResponseHeaders = &route.Headers{}
			}

			for _, s := range h.Set {
				entry.ResponseHeaders.Replace = append(
					entry.ResponseHeaders.Replace, route.Pair(string(s.Name), string(s.Value)))
			}

			for _, s := range h.Add {
				entry.ResponseHeaders.Append = append(
					entry.ResponseHeaders.Append, route.Pair(string(s.Name), string(s.Value)))
			}

			entry.ResponseHeaders.Delete = append(
				entry.ResponseHeaders.Delete, h.Remove...)
		} else if r := f.URLRewrite; r != nil {
			rewrite := route.Rewrite{}

			if r.Path != nil {
				switch r.Path.Type {
				case api.ReplaceFullPathType:
					rewrite.ReplaceFullPath = r.Path.ReplaceFullPath
				case api.ReplacePrefixMatchType:
					rewrite.ReplacePrefixMatch = r.Path.ReplacePrefixMatch
				}
			}

			if r.Hostname != nil {
				rewrite.ReplaceHostname = pointer.To(string(*r.Hostname))
			}

			if r.HostToBackendHostname {
				rewrite.HostToBackendHostname = true
			}

			entry.Rewrite = &rewrite
		}
	}
	return entry
}

func makeRouteMatch(ruleMatch api.Match) route.Match {
	match := route.Match{}

	if p := ruleMatch.Path; p != nil {
		switch p.Type {
		case api.Exact:
			match.ExactPath = p.Value
		case api.PathPrefix:
			match.PrefixPath = p.Value
		case api.RegularExpression:
			match.RegexPath = p.Value
		}
	} else {
		// Envoy routes require a path match, so if the route
		// didn't specify, we match any path so that the additional
		// match criteria will be applied.
		match.PrefixPath = "/"
	}

	if m := ruleMatch.Method; m != nil {
		match.Method = string(*m)
	}

	for _, h := range ruleMatch.Headers {
		typ := pointer.DerefOr(h.Type, common_api.HeaderMatchExact)
		switch typ {
		case common_api.HeaderMatchExact:
			match.ExactHeader = append(
				match.ExactHeader, route.Pair(string(h.Name), string(h.Value)),
			)
		case common_api.HeaderMatchRegularExpression:
			match.RegexHeader = append(
				match.RegexHeader, route.Pair(string(h.Name), string(h.Value)),
			)
		case common_api.HeaderMatchAbsent:
			match.AbsentHeader = append(match.AbsentHeader, string(h.Name))
		case common_api.HeaderMatchPresent:
			match.PresentHeader = append(match.PresentHeader, string(h.Name))
		case common_api.HeaderMatchPrefix:
			match.PrefixHeader = append(
				match.PrefixHeader, route.Pair(string(h.Name), string(h.Value)),
			)
		}
	}

	for _, q := range ruleMatch.QueryParams {
		switch q.Type {
		case api.ExactQueryMatch:
			match.ExactQuery = append(
				match.ExactQuery, route.Pair(q.Name, q.Value),
			)
		case api.RegularExpressionQueryMatch:
			match.RegexQuery = append(
				match.ExactQuery, route.Pair(q.Name, q.Value),
			)
		}
	}

	return match
}
