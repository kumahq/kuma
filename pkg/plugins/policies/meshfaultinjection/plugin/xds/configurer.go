package xds

import (
	envoy_listener "github.com/envoyproxy/go-control-plane/envoy/config/listener/v3"
	envoy_http_fault "github.com/envoyproxy/go-control-plane/envoy/extensions/filters/http/fault/v3"
	envoy_hcm "github.com/envoyproxy/go-control-plane/envoy/extensions/filters/network/http_connection_manager/v3"

	"github.com/kumahq/kuma/pkg/core/kri"
	bldrs_matchers "github.com/kumahq/kuma/pkg/envoy/builders/xds/matchers"
	"github.com/kumahq/kuma/pkg/plugins/policies/core/rules/inbound"
	policies_xds "github.com/kumahq/kuma/pkg/plugins/policies/core/xds"
	policies_api "github.com/kumahq/kuma/pkg/plugins/policies/meshfaultinjection/api/v1alpha1"
	"github.com/kumahq/kuma/pkg/util/pointer"
	util_proto "github.com/kumahq/kuma/pkg/util/proto"
	listeners_v3 "github.com/kumahq/kuma/pkg/xds/envoy/listeners/v3"
)

type Configurer struct {
	Rules []*inbound.Rule
}

func (c *Configurer) ConfigureHttpListener(filterChain *envoy_listener.FilterChain) error {
	// Do not add new faults if old ones were applied
	for _, filter := range filterChain.Filters {
		if filter.Name == "envoy.filters.http.fault" {
			return nil
		}
	}
	if err := listeners_v3.UpdateHTTPConnectionManager(filterChain, c.addFaultFilters); err != nil {
		return err
	}
	return nil
}

func (c *Configurer) addFaultFilters(hcm *envoy_hcm.HttpConnectionManager) error {
	var fiFilters []*envoy_hcm.HttpFilter

	for _, rule := range c.Rules {
		matchesConf := rule.Conf.(*policies_api.Rule)
		matcher := bldrs_matchers.Matcher(
			bldrs_matchers.NewMatcherBuilder().Configure(
				bldrs_matchers.FieldMatcher(
					bldrs_matchers.NewFieldMatcher().Configure(
						bldrs_matchers.NotMatches(
							pointer.Deref(matchesConf.Matches),
							bldrs_matchers.NewOnMatch().Configure(bldrs_matchers.SkipFilterAction()),
						),
					),
				)),
		)

		for _, fault := range pointer.Deref(matchesConf.Default.Http) {
			faultConfig, _ := configureFault(fault)
			faultFilter, err := util_proto.MarshalAnyDeterministic(faultConfig)
			if err != nil {
				return err
			}

			extensionWithMatcher, err := bldrs_matchers.NewExtensionWithMatcher().
				Configure(matcher).
				Configure(bldrs_matchers.Filter("envoy.filters.http.fault", faultFilter)).
				Build()
			if err != nil {
				return err
			}
			typedExtension, err := util_proto.MarshalAnyDeterministic(extensionWithMatcher)
			if err != nil {
				return err
			}

			fiFilters = append(fiFilters, &envoy_hcm.HttpFilter{
				Name: kri.FromResourceMeta(rule.Origin.Resource, policies_api.MeshFaultInjectionType).String(),
				ConfigType: &envoy_hcm.HttpFilter_TypedConfig{
					TypedConfig: typedExtension,
				},
			})
		}
	}
	return policies_xds.InsertHTTPFiltersBeforeRouter(hcm, fiFilters...)
}

func configureFault(fault policies_api.FaultInjectionConf) (*envoy_http_fault.HTTPFault, error) {
	faultConfig := &envoy_http_fault.HTTPFault{}
	delay, err := convertDelay(fault.Delay)
	if err != nil {
		return nil, err
	}
	faultConfig.Delay = delay

	abort, err := convertAbort(fault.Abort)
	if err != nil {
		return nil, err
	}
	faultConfig.Abort = abort

	rrl, err := convertResponseRateLimit(fault.ResponseBandwidth)
	if err != nil {
		return nil, err
	}
	faultConfig.ResponseRateLimit = rrl

	return faultConfig, nil
}
