package clusters

import (
	"encoding/hex"

	envoy_cluster "github.com/envoyproxy/go-control-plane/envoy/config/cluster/v3"
	envoy_core "github.com/envoyproxy/go-control-plane/envoy/config/core/v3"
	envoy_type "github.com/envoyproxy/go-control-plane/envoy/type/v3"
	"github.com/pkg/errors"
	"google.golang.org/protobuf/types/known/wrapperspb"

	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
	"github.com/kumahq/kuma/pkg/core"
	core_mesh "github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
	util_proto "github.com/kumahq/kuma/pkg/util/proto"
)

var log = core.Log.WithName("xds").WithName("health-check-configurer")

type HealthCheckConfigurer struct {
	HealthCheck *core_mesh.HealthCheckResource
	Protocol    core_mesh.Protocol
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
		appendAction := envoy_core.HeaderValueOption_OVERWRITE_IF_EXISTS_OR_ADD
		if header.Append.Value {
			appendAction = envoy_core.HeaderValueOption_APPEND_IF_EXISTS_OR_ADD
		}
		envoyHeaders = append(envoyHeaders, &envoy_core.HeaderValueOption{
			Header: &envoy_core.HeaderValue{
				Key:   header.Header.Key,
				Value: header.Header.Value,
			},
			AppendAction: appendAction,
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
	protocol core_mesh.Protocol,
	httpConf *mesh_proto.HealthCheck_Conf_Http,
) *envoy_core.HealthCheck_HttpHealthCheck_ {
	var expectedStatuses []*envoy_type.Int64Range
	for _, status := range httpConf.ExpectedStatuses {
		expectedStatuses = append(
			expectedStatuses,
			mapUInt32ToInt64Range(status.Value),
		)
	}

	codecClientType := envoy_type.CodecClientType_HTTP1
	if protocol == core_mesh.ProtocolHTTP2 {
		codecClientType = envoy_type.CodecClientType_HTTP2
	}

	httpHealthCheck := envoy_core.HealthCheck_HttpHealthCheck{
		Path:                httpConf.Path,
		RequestHeadersToAdd: mapHttpHeaders(httpConf.RequestHeadersToAdd),
		ExpectedStatuses:    expectedStatuses,
		CodecClientType:     codecClientType,
	}

	return &envoy_core.HealthCheck_HttpHealthCheck_{
		HttpHealthCheck: &httpHealthCheck,
	}
}

func healthPanicThreshold(cluster *envoy_cluster.Cluster, value *wrapperspb.FloatValue) {
	if value == nil {
		return
	}
	if cluster.CommonLbConfig == nil {
		cluster.CommonLbConfig = &envoy_cluster.Cluster_CommonLbConfig{}
	}
	cluster.CommonLbConfig.HealthyPanicThreshold = &envoy_type.Percent{Value: float64(value.Value)}
}

func failTrafficOnPanic(cluster *envoy_cluster.Cluster, value *wrapperspb.BoolValue) {
	if value == nil {
		return
	}
	if cluster.CommonLbConfig == nil {
		cluster.CommonLbConfig = &envoy_cluster.Cluster_CommonLbConfig{}
	}
	if cluster.CommonLbConfig.GetLocalityWeightedLbConfig() != nil {
		// used load balancing type doesn't support 'fail_traffic_on_panic', right now we don't use
		// 'locality_weighted_lb_config' in Kuma, locality aware load balancing is implemented based on priority levels
		log.Error(
			errors.New("unable to set 'fail_traffic_on_panic' for 'locality_weighted_lb_config' load balancer"),
			"unable to configure 'fail_traffic_on_panic', parameter is ignored")
		return
	}
	if cluster.CommonLbConfig.LocalityConfigSpecifier == nil {
		cluster.CommonLbConfig.LocalityConfigSpecifier = &envoy_cluster.Cluster_CommonLbConfig_ZoneAwareLbConfig_{
			ZoneAwareLbConfig: &envoy_cluster.Cluster_CommonLbConfig_ZoneAwareLbConfig{},
		}
	}
	if zoneAwareLbConfig := cluster.CommonLbConfig.GetZoneAwareLbConfig(); zoneAwareLbConfig != nil {
		zoneAwareLbConfig.FailTrafficOnPanic = value.GetValue()
	}
}

func buildHealthCheck(conf *mesh_proto.HealthCheck_Conf) *envoy_core.HealthCheck {
	return &envoy_core.HealthCheck{
		HealthChecker: &envoy_core.HealthCheck_TcpHealthCheck_{
			TcpHealthCheck: &envoy_core.HealthCheck_TcpHealthCheck{},
		},
		Interval:                     conf.Interval,
		Timeout:                      conf.Timeout,
		UnhealthyThreshold:           util_proto.UInt32(conf.UnhealthyThreshold),
		HealthyThreshold:             util_proto.UInt32(conf.HealthyThreshold),
		InitialJitter:                conf.InitialJitter,
		IntervalJitter:               conf.IntervalJitter,
		IntervalJitterPercent:        conf.IntervalJitterPercent,
		EventLogPath:                 conf.EventLogPath,
		AlwaysLogHealthCheckFailures: conf.AlwaysLogHealthCheckFailures.GetValue(),
		NoTrafficInterval:            conf.NoTrafficInterval,
		ReuseConnection:              conf.ReuseConnection,
	}
}

func addHealthChecker(healthCheck *envoy_core.HealthCheck, healthChecker interface{}) *envoy_core.HealthCheck {
	switch hc := healthChecker.(type) {
	case *envoy_core.HealthCheck_HttpHealthCheck_:
		healthCheck.HealthChecker = hc
	case *envoy_core.HealthCheck_TcpHealthCheck_:
		healthCheck.HealthChecker = hc
	}

	return healthCheck
}

func (e *HealthCheckConfigurer) Configure(cluster *envoy_cluster.Cluster) error {
	if e.HealthCheck == nil || e.HealthCheck.Spec.Conf == nil {
		return nil
	}

	activeChecks := e.HealthCheck.Spec.Conf

	healthPanicThreshold(cluster, activeChecks.GetHealthyPanicThreshold())
	failTrafficOnPanic(cluster, activeChecks.GetFailTrafficOnPanic())

	tcp := activeChecks.GetTcp()
	http := activeChecks.GetHttp()

	if tcp == nil && http == nil {
		cluster.HealthChecks = append(cluster.HealthChecks, buildHealthCheck(activeChecks))

		return nil
	}

	if tcp != nil {
		defaultHealthCheck := buildHealthCheck(activeChecks)
		healthChecker := tcpHealthCheck(tcp)
		healthCheck := addHealthChecker(defaultHealthCheck, healthChecker)

		cluster.HealthChecks = append(cluster.HealthChecks, healthCheck)
	}

	if http != nil {
		defaultHealthCheck := buildHealthCheck(activeChecks)
		healthChecker := httpHealthCheck(e.Protocol, http)
		healthCheck := addHealthChecker(defaultHealthCheck, healthChecker)

		cluster.HealthChecks = append(cluster.HealthChecks, healthCheck)
	}

	return nil
}
