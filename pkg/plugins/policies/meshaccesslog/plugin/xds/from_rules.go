package xds

import (
	envoy_accesslog "github.com/envoyproxy/go-control-plane/envoy/config/accesslog/v3"

	bldrs_accesslog "github.com/kumahq/kuma/v2/pkg/envoy/builders/accesslog"
	. "github.com/kumahq/kuma/v2/pkg/envoy/builders/common"
	"github.com/kumahq/kuma/v2/pkg/plugins/policies/core/rules/inbound"
	api "github.com/kumahq/kuma/v2/pkg/plugins/policies/meshaccesslog/api/v1alpha1"
	"github.com/kumahq/kuma/v2/pkg/util/pointer"
	listeners_v3 "github.com/kumahq/kuma/v2/pkg/xds/envoy/listeners/v3"
)

// BuildAccessLogBuildersFromRules turns inbound rules into access log builders.
// Each rule fires independently: every backend on every rule whose Match holds
// emits a log entry. Rules with overlapping matches produce logs to every
// matching backend (intentional — access logging fans out).
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
	for _, rule := range rules {
		conf, ok := rule.Conf.(api.Conf)
		if !ok {
			continue
		}
		expr := MatchToCEL(rule.Match)
		for _, backend := range ResolveBackends(pointer.Deref(conf.Backends), endpointsAcc) {
			b := BaseAccessLogBuilder(backend, defaultFormat, endpointsAcc, values, accessLogSocketPath)
			if expr != "" {
				b = b.Configure(bldrs_accesslog.CELFilter(expr))
			}
			result = append(result, b)
		}
	}
	return result
}
