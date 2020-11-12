package listeners

import (
	envoy_listener "github.com/envoyproxy/go-control-plane/envoy/api/v2/listener"
	envoy_kafka "github.com/envoyproxy/go-control-plane/envoy/config/filter/network/kafka_broker/v2alpha1"

	util_xds "github.com/kumahq/kuma/pkg/util/xds"

	"github.com/kumahq/kuma/pkg/util/proto"
)

func Kafka(statsName string) FilterChainBuilderOpt {
	return FilterChainBuilderOptFunc(func(config *FilterChainBuilderConfig) {
		config.Add(&KafkaConfigurer{
			statsName: statsName,
		})
	})
}

type KafkaConfigurer struct {
	statsName string
}

func (c *KafkaConfigurer) Configure(filterChain *envoy_listener.FilterChain) error {
	pbst, err := proto.MarshalAnyDeterministic(
		&envoy_kafka.KafkaBroker{
			StatPrefix: util_xds.SanitizeMetric(c.statsName),
		})
	if err != nil {
		return err
	}

	filterChain.Filters = append([]*envoy_listener.Filter{
		{
			Name: "envoy.filters.network.kafka_broker",
			ConfigType: &envoy_listener.Filter_TypedConfig{
				TypedConfig: pbst,
			},
		},
	}, filterChain.Filters...)
	return nil
}
