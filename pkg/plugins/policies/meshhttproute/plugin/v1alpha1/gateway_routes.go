package v1alpha1

import (
	"context"
	"slices"
	"sort"
	"strings"

	"k8s.io/apimachinery/pkg/util/intstr"

	common_api "github.com/kumahq/kuma/api/common/v1alpha1"
	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
	"github.com/kumahq/kuma/pkg/core/permissions"
	core_mesh "github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
	"github.com/kumahq/kuma/pkg/core/resources/model"
	core_xds "github.com/kumahq/kuma/pkg/core/xds"
	"github.com/kumahq/kuma/pkg/plugins/policies/core/rules"
	api "github.com/kumahq/kuma/pkg/plugins/policies/meshhttproute/api/v1alpha1"
	plugin_gateway "github.com/kumahq/kuma/pkg/plugins/runtime/gateway"
	"github.com/kumahq/kuma/pkg/plugins/runtime/gateway/match"
	"github.com/kumahq/kuma/pkg/plugins/runtime/gateway/route"
	"github.com/kumahq/kuma/pkg/util/pointer"
	xds_context "github.com/kumahq/kuma/pkg/xds/context"
	envoy_names "github.com/kumahq/kuma/pkg/xds/envoy/names"
	"github.com/kumahq/kuma/pkg/xds/envoy/tags"
	xds_topology "github.com/kumahq/kuma/pkg/xds/topology"
)

type hostnameTags struct {
	Hostname string
	Tags     map[string]string
}
type listenersHostnames struct {
	listener  *mesh_proto.MeshGateway_Listener
	hostnames []hostnameTags
}

func CollectListenerInfos(
	ctx context.Context,
	meshCtx xds_context.MeshContext,
	gateway *core_mesh.MeshGatewayResource,
	proxy *core_xds.Proxy,
	rawRules rules.GatewayRules,
) []plugin_gateway.GatewayListenerInfo {
	networking := proxy.Dataplane.Spec.GetNetworking()
	listenersByPort := map[uint32]listenersHostnames{}
	for _, listener := range gateway.Spec.GetConf().GetListeners() {
		if listener.Protocol == mesh_proto.MeshGateway_Listener_TCP {
			continue
		}

		listenerAcc, ok := listenersByPort[listener.GetPort()]
		if !ok {
			listenerAcc = listenersHostnames{
				listener: listener,
			}
		}
		hostname := listener.GetNonEmptyHostname()
		listenerAcc.hostnames = append(listenerAcc.hostnames, hostnameTags{
			Hostname: hostname,
			Tags: mesh_proto.Merge(
				networking.GetGateway().GetTags(),
				gateway.Spec.GetTags(),
				listener.GetTags(),
			),
		})
		listenersByPort[listener.GetPort()] = listenerAcc
	}

	var infos []plugin_gateway.GatewayListenerInfo

	for port, listener := range listenersByPort {
		externalServices := meshCtx.Resources.ExternalServices()

		matchedExternalServices := permissions.MatchExternalServicesTrafficPermissions(
			proxy.Dataplane, externalServices, meshCtx.Resources.TrafficPermissions(),
		)

		outboundEndpoints := core_xds.EndpointMap{}
		for k, v := range meshCtx.EndpointMap {
			outboundEndpoints[k] = v
		}

		esEndpoints := xds_topology.BuildExternalServicesEndpointMap(
			ctx,
			meshCtx.Resource,
			matchedExternalServices,
			meshCtx.DataSourceLoader,
			proxy.Zone,
		)
		for k, v := range esEndpoints {
			outboundEndpoints[k] = v
		}

		hostInfos := SortRulesToHosts(
			meshCtx.Resources.MeshLocalResources,
			rawRules,
			networking.Address,
			listener.listener,
			listener.hostnames,
		)
		infos = append(infos, plugin_gateway.GatewayListenerInfo{
			Proxy:             proxy,
			Gateway:           gateway,
			HostInfos:         hostInfos,
			ExternalServices:  externalServices,
			OutboundEndpoints: outboundEndpoints,
			Listener: plugin_gateway.GatewayListener{
				Port:     port,
				Protocol: listener.listener.GetProtocol(),
				ResourceName: envoy_names.GetGatewayListenerName(
					gateway.Meta.GetName(),
					listener.listener.GetProtocol().String(),
					port,
				),
				CrossMesh: listener.listener.GetCrossMesh(),
				Resources: listener.listener.GetResources(),
			},
		})
	}

	return infos
}

func SortRulesToHosts(
	meshLocalResources xds_context.ResourceMap,
	rawRules rules.GatewayRules,
	address string,
	listener *mesh_proto.MeshGateway_Listener,
	hostnameTags []hostnameTags,
) []plugin_gateway.GatewayHostInfo {
	var hostInfos []plugin_gateway.GatewayHostInfo

	for _, hostnameTag := range hostnameTags {
		inboundListener := rules.NewInboundListenerHostname(
			address,
			listener.GetPort(),
			hostnameTag.Hostname,
		)
		rawRules, ok := rawRules.ToRules.ByListenerAndHostname[inboundListener]
		if !ok {
			continue
		}
		var ruleHostnames []string
		rulesByHostname := map[string][]ToRouteRule{}
		for _, rawRule := range rawRules {
			conf := rawRule.Conf.(api.PolicyDefault)
			rule := ToRouteRule{
				Subset:    rawRule.Subset,
				Rules:     conf.Rules,
				Hostnames: conf.Hostnames,
				Origin:    rawRule.Origin,
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
				rulesByHostname[hostname] = append(accRule, rule)
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
			var rules []ToRouteRule
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
			// Create an info for every hostname match
			// We may end up duplicating info more than once so we copy it here
			host := plugin_gateway.GatewayHost{
				Hostname: hostnameMatch,
				Routes:   nil,
				Policies: map[model.ResourceType][]match.RankedPolicy{},
				TLS:      listener.Tls,
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
			hostInfo.AppendEntries(GenerateEnvoyRouteEntries(host, rules))
			hostInfos = append(hostInfos, hostInfo)
		}
	}
	sort.Slice(hostInfos, func(i, j int) bool {
		return hostInfos[i].Host.Hostname > hostInfos[j].Host.Hostname
	})

	return hostInfos
}

func GenerateEnvoyRouteEntries(host plugin_gateway.GatewayHost, toRules []ToRouteRule) []route.Entry {
	var entries []route.Entry

	// Index the routes by their path. There are typically multiple
	// routes per path with additional matching criteria.
	exactEntries := map[string][]route.Entry{}
	prefixEntries := map[string][]route.Entry{}

	for _, rules := range toRules {
		for _, rule := range rules.Rules {
			var names []string
			for _, orig := range rules.Origin {
				names = append(names, orig.GetName())
			}
			slices.Sort(names)
			entry := makeHttpRouteEntry(strings.Join(names, "_"), rule)

			// The rule matches if any of the matches is successful (it has OR
			// semantics). That means that we have to duplicate the route table
			// entry for each repeated match so that the rule can match any of
			// the criteria.
			for _, m := range rule.Matches {
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
	}

	return plugin_gateway.HandlePrefixMatchesAndPopulatePolicies(host, exactEntries, prefixEntries, entries)
}

func makeHttpRouteEntry(name string, rule api.Rule) route.Entry {
	entry := route.Entry{
		Route: name,
	}

	for _, b := range *rule.Default.BackendRefs {
		dest, ok := tags.FromTargetRef(b.TargetRef)
		if !ok {
			// This should be caught by validation
			continue
		}
		target := route.Destination{
			Destination:   dest,
			Weight:        uint32(*b.Weight),
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
			tags, ok := tags.FromTargetRef(m.BackendRef)
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
					entry.RequestHeaders.Replace, route.Pair(string(s.Name), string(s.Value)))
			}

			for _, s := range h.Add {
				entry.ResponseHeaders.Append = append(
					entry.RequestHeaders.Append, route.Pair(string(s.Name), string(s.Value)))
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
