package xds

import (
	envoy_listener "github.com/envoyproxy/go-control-plane/envoy/config/listener/v3"

	core_xds "github.com/kumahq/kuma/pkg/core/xds"
	api "github.com/kumahq/kuma/pkg/plugins/policies/meshfaultinjection/api/v1alpha1"
	listeners_v3 "github.com/kumahq/kuma/pkg/xds/envoy/listeners/v3"
)

func FaultInjectionConfigurationFromPolicy(fi *api.FaultInjectionConf, from core_xds.Subset) *listeners_v3.FaultInjectionConfiguration {
	if fi == nil {
		return &listeners_v3.FaultInjectionConfiguration{}
	}

	faultInjectionConf := listeners_v3.FaultInjectionConfiguration{}
	if fi.Abort != nil {
		faultInjectionConf.Abort = &listeners_v3.Abort{
			HttpStatus: uint32(*fi.Abort.HttpStatus),
			Percentage: float64(*fi.Abort.Percentage),
		}
	}
	if fi.Delay != nil {
		faultInjectionConf.Delay = &listeners_v3.Delay{
			Value : fi.Delay.Value.Duration,
			Percentage: float64(*fi.Abort.Percentage),
		}
	}
	if fi.ResponseBandwidth != nil {
		faultInjectionConf.ResponseBandwidth = &listeners_v3.ResponseBandwidth{
			Limit : *fi.ResponseBandwidth.Limit,
			Percentage: float64(*fi.Abort.Percentage),
		}
	}
	
	if from != nil {
		sourceTags := []map[string]string{}
		for _, tag := range from {
			value := map[string]string{
				tag.Key : tag.Value,
			}
			sourceTags = append(sourceTags, value)
		}
		faultInjectionConf.SourceTags = sourceTags
	}

	return &faultInjectionConf
}

type Configurer struct {
	FaultInjections []*api.FaultInjectionConf
	From core_xds.Subset
}

func (c *Configurer) ConfigureFilterChain(filterChain *envoy_listener.FilterChain) error {
	configs := []*listeners_v3.FaultInjectionConfiguration{}
	for _, fi := range c.FaultInjections {
		configs = append(configs, FaultInjectionConfigurationFromPolicy(fi, c.From))
	}
	configurer := listeners_v3.FaultInjectionConfigurer{
		FaultInjections: configs,
	}
	err := configurer.Configure(filterChain)
	if err != nil {
		return err
	}
	return nil
}
