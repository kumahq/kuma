package xds

import (
	envoy_accesslog "github.com/envoyproxy/go-control-plane/envoy/config/accesslog/v3"

	common_api "github.com/kumahq/kuma/v2/api/common/v1alpha1"
	bldrs_accesslog "github.com/kumahq/kuma/v2/pkg/envoy/builders/accesslog"
	. "github.com/kumahq/kuma/v2/pkg/envoy/builders/common"
	"github.com/kumahq/kuma/v2/pkg/plugins/policies/core/rules/inbound"
	api "github.com/kumahq/kuma/v2/pkg/plugins/policies/meshaccesslog/api/v1alpha1"
	"github.com/kumahq/kuma/v2/pkg/util/pointer"
	listeners_v3 "github.com/kumahq/kuma/v2/pkg/xds/envoy/listeners/v3"
)

// BuildAccessLogBuildersFromRules turns a list of sorted inbound rules (most
// specific first, per inbound.SortRules) into access log builders implementing
// first-match-wins semantics via a CEL ExpressionFilter on each entry.
//
// For each rule, every backend produces one access log builder configured with
// `myMatchExpr && !(prior0Expr) && !(prior1Expr) ...`. The first rule with no
// match (and no priors) produces an unfiltered entry — preserving today's
// catch-all behavior when no matches are used.
//
// TODO: on zone proxies (ZoneIngress/ZoneEgress), filter chains already match
// per-SNI via MatchServerNames, so encoding `connection.requested_server_name`
// in CEL duplicates the filter chain match. A future optimization can install
// access logs on the matching filter chain directly and drop the SNI term from
// CEL; for now CEL handles both SpiffeID and SNI uniformly across sidecar
// inbound, ZoneEgress, and ZoneIngress.
func BuildAccessLogBuildersFromRules(
	rules []*inbound.Rule,
	defaultFormat string,
	endpointsAcc *EndpointAccumulator,
	values listeners_v3.KumaValues,
	accessLogSocketPath string,
) []*Builder[envoy_accesslog.AccessLog] {
	var result []*Builder[envoy_accesslog.AccessLog]
	var priors []*common_api.Match
	for _, rule := range rules {
		conf, ok := rule.Conf.(api.Conf)
		if !ok {
			priors = append(priors, rule.Match)
			continue
		}
		expr := ComposeExpr(rule.Match, priors)
		for _, backend := range pointer.Deref(conf.Backends) {
			b := BaseAccessLogBuilder(backend, defaultFormat, endpointsAcc, values, accessLogSocketPath)
			if b == nil {
				continue
			}
			if expr != "" {
				b = b.Configure(bldrs_accesslog.CELFilter(expr))
			}
			result = append(result, b)
		}
		priors = append(priors, rule.Match)
	}
	return result
}
