package v3

import (
	envoy_listener "github.com/envoyproxy/go-control-plane/envoy/config/listener/v3"
	envoy_hcm "github.com/envoyproxy/go-control-plane/envoy/extensions/filters/network/http_connection_manager/v3"

	core_xds "github.com/kumahq/kuma/pkg/core/xds"
	util_proto "github.com/kumahq/kuma/pkg/util/proto"
	util_xds "github.com/kumahq/kuma/pkg/util/xds"
)

type HttpConnectionManagerConfigurer struct {
	StatsName                string
	ForwardClientCertDetails bool
	NormalizePath            bool
	InternalAddresses        []core_xds.InternalAddress
}

func (c *HttpConnectionManagerConfigurer) Configure(filterChain *envoy_listener.FilterChain) error {
	config := &envoy_hcm.HttpConnectionManager{
		StatPrefix:  util_xds.SanitizeMetric(c.StatsName),
		CodecType:   envoy_hcm.HttpConnectionManager_AUTO,
		HttpFilters: []*envoy_hcm.HttpFilter{},
		// notice that route configuration is left up to other configurers
	}

	if c.ForwardClientCertDetails {
		config.ForwardClientCertDetails = envoy_hcm.HttpConnectionManager_SANITIZE_SET
		config.SetCurrentClientCertDetails = &envoy_hcm.HttpConnectionManager_SetCurrentClientCertDetails{
			Uri: true,
		}
	}

	if c.NormalizePath {
		config.NormalizePath = util_proto.Bool(true)
	}

	if len(c.InternalAddresses) == 0 {
		c.InternalAddresses = core_xds.LocalHostAddresses
	}
	config.InternalAddressConfig = &envoy_hcm.HttpConnectionManager_InternalAddressConfig{
		UnixSockets: false,
		CidrRanges:  core_xds.InternalAddressToEnvoyCIDRs(c.InternalAddresses),
	}

	pbst, err := util_proto.MarshalAnyDeterministic(config)
	if err != nil {
		return err
	}

	filterChain.Filters = append(filterChain.Filters, &envoy_listener.Filter{
		Name: "envoy.filters.network.http_connection_manager",
		ConfigType: &envoy_listener.Filter_TypedConfig{
			TypedConfig: pbst,
		},
	})
	return nil
}
