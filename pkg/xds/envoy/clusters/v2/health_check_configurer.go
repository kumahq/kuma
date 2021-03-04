package clusters

import (
	"encoding/hex"

	envoy_api "github.com/envoyproxy/go-control-plane/envoy/api/v2"
	envoy_core "github.com/envoyproxy/go-control-plane/envoy/api/v2/core"
	envoy_type "github.com/envoyproxy/go-control-plane/envoy/type"
	"github.com/golang/protobuf/ptypes/wrappers"
	"github.com/pkg/errors"

	"github.com/kumahq/kuma/pkg/core"

	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
	mesh_core "github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
)

type HealthCheckConfigurer struct {
	HealthCheck *mesh_core.HealthCheckResource
}

var _ ClusterConfigurer = &HealthCheckConfigurer{}

func mapUInt32ToInt64Range(value uint32) *envoy_type.Int64Range {
	return &envoy_type.Int64Range{
		Start: int64(value),
		End:   int64(value) + 1,
	}
}

func mapHttpHeaders(
	headers []*mesh_proto.HealthCheck_Conf_Http_HeaderValueOption,
) []*envoy_core.HeaderValueOption {
	var envoyHeaders []*envoy_core.HeaderValueOption
	for _, header := range headers {
		envoyHeaders = append(envoyHeaders, &envoy_core.HeaderValueOption{
			Header: &envoy_core.HeaderValue{
				Key:   header.Header.Key,
				Value: header.Header.Value,
			},
			Append: header.Append,
		})
	}
	return envoyHeaders
}

func tcpHealthCheck(
	tcpConf *mesh_proto.HealthCheck_Conf_Tcp,
) *envoy_core.HealthCheck_TcpHealthCheck_ {
	tcpHealthCheck := envoy_core.HealthCheck_TcpHealthCheck{}

	if tcpConf.Send != nil {
		tcpHealthCheck.Send = &envoy_core.HealthCheck_Payload{
			Payload: &envoy_core.HealthCheck_Payload_Text{
				Text: hex.EncodeToString(tcpConf.Send.Value),
			},
		}
	}

	if tcpConf.Receive != nil {
		var receive []*envoy_core.HealthCheck_Payload
		for _, r := range tcpConf.Receive {
			receive = append(receive, &envoy_core.HealthCheck_Payload{
				Payload: &envoy_core.HealthCheck_Payload_Text{
					Text: hex.EncodeToString(r.Value),
				},
			})
		}
		tcpHealthCheck.Receive = receive
	}

	return &envoy_core.HealthCheck_TcpHealthCheck_{
		TcpHealthCheck: &tcpHealthCheck,
	}
}

func httpHealthCheck(
	httpConf *mesh_proto.HealthCheck_Conf_Http,
) *envoy_core.HealthCheck_HttpHealthCheck_ {
	var expectedStatuses []*envoy_type.Int64Range
	for _, status := range httpConf.ExpectedStatuses {
		expectedStatuses = append(
			expectedStatuses,
			mapUInt32ToInt64Range(status.Value),
		)
	}

	httpHealthCheck := envoy_core.HealthCheck_HttpHealthCheck{
		Path:                httpConf.Path,
		RequestHeadersToAdd: mapHttpHeaders(httpConf.RequestHeadersToAdd),
		ExpectedStatuses:    expectedStatuses,
		CodecClientType:     envoy_type.CodecClientType_HTTP2,
	}

	return &envoy_core.HealthCheck_HttpHealthCheck_{
		HttpHealthCheck: &httpHealthCheck,
	}
}

func healthPanicThreshold(cluster *envoy_api.Cluster, value *wrappers.FloatValue) {
	if value == nil {
		return
	}
	if cluster.CommonLbConfig == nil {
		cluster.CommonLbConfig = &envoy_api.Cluster_CommonLbConfig{}
	}
	cluster.CommonLbConfig.HealthyPanicThreshold = &envoy_type.Percent{Value: float64(value.Value)}
}

func failTrafficOnPanic(cluster *envoy_api.Cluster, value *wrappers.BoolValue) {
	if value == nil {
		return
	}
	if cluster.CommonLbConfig == nil {
		cluster.CommonLbConfig = &envoy_api.Cluster_CommonLbConfig{}
	}
	if cluster.CommonLbConfig.GetLocalityWeightedLbConfig() != nil {
		// used load balancing type doesn't support 'fail_traffic_on_panic', right now we don't use
		// 'locality_weighted_lb_config' in Kuma, locality aware load balancing is implemented based on priority levels
		core.Log.WithName("health-check-configurer").Error(
			errors.New("unable to set 'fail_traffic_on_panic' for 'locality_weighted_lb_config' load balancer"),
			"unable to configure 'fail_traffic_on_panic', parameter is ignored")
		return
	}
	if cluster.CommonLbConfig.LocalityConfigSpecifier == nil {
		cluster.CommonLbConfig.LocalityConfigSpecifier = &envoy_api.Cluster_CommonLbConfig_ZoneAwareLbConfig_{
			ZoneAwareLbConfig: &envoy_api.Cluster_CommonLbConfig_ZoneAwareLbConfig{},
		}
	}
	cluster.CommonLbConfig.GetZoneAwareLbConfig().FailTrafficOnPanic = value.GetValue()
}

func (e *HealthCheckConfigurer) Configure(cluster *envoy_api.Cluster) error {
	if e.HealthCheck == nil || e.HealthCheck.Spec.Conf == nil {
		return nil
	}
	activeChecks := e.HealthCheck.Spec.Conf
	healthCheck := envoy_core.HealthCheck{
		HealthChecker: &envoy_core.HealthCheck_TcpHealthCheck_{
			TcpHealthCheck: &envoy_core.HealthCheck_TcpHealthCheck{},
		},
		Interval:                     activeChecks.Interval,
		Timeout:                      activeChecks.Timeout,
		UnhealthyThreshold:           &wrappers.UInt32Value{Value: activeChecks.UnhealthyThreshold},
		HealthyThreshold:             &wrappers.UInt32Value{Value: activeChecks.HealthyThreshold},
		InitialJitter:                activeChecks.InitialJitter,
		IntervalJitter:               activeChecks.IntervalJitter,
		IntervalJitterPercent:        activeChecks.IntervalJitterPercent,
		EventLogPath:                 activeChecks.EventLogPath,
		AlwaysLogHealthCheckFailures: activeChecks.AlwaysLogHealthCheckFailures.GetValue(),
		NoTrafficInterval:            activeChecks.NoTrafficInterval,
	}

	healthPanicThreshold(cluster, activeChecks.GetHealthyPanicThreshold())
	failTrafficOnPanic(cluster, activeChecks.GetFailTrafficOnPanic())

	if tcp := activeChecks.GetTcp(); tcp != nil {
		healthCheck.HealthChecker = tcpHealthCheck(tcp)
	}

	if http := activeChecks.GetHttp(); http != nil {
		healthCheck.HealthChecker = httpHealthCheck(http)
	}

	cluster.HealthChecks = append(cluster.HealthChecks, &healthCheck)
	return nil
}
