package xds

import (
	"encoding/hex"
	"strings"
	"time"

	envoy_cluster "github.com/envoyproxy/go-control-plane/envoy/config/cluster/v3"
	envoy_core "github.com/envoyproxy/go-control-plane/envoy/config/core/v3"
	envoy_healthchecks "github.com/envoyproxy/go-control-plane/envoy/extensions/health_check/event_sinks/file/v3"
	envoy_type "github.com/envoyproxy/go-control-plane/envoy/type/v3"
	"github.com/pkg/errors"
	"k8s.io/apimachinery/pkg/util/intstr"

	common_api "github.com/kumahq/kuma/api/common/v1alpha1"
	"github.com/kumahq/kuma/api/mesh/v1alpha1"
	"github.com/kumahq/kuma/pkg/core"
	core_mesh "github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
	api "github.com/kumahq/kuma/pkg/plugins/policies/meshhealthcheck/api/v1alpha1"
	"github.com/kumahq/kuma/pkg/util/pointer"
	util_proto "github.com/kumahq/kuma/pkg/util/proto"
	"github.com/kumahq/kuma/pkg/xds/envoy/tags"
)

type Configurer struct {
	Conf     api.Conf
	Protocol core_mesh.Protocol
	Tags     v1alpha1.MultiValueTagSet
}

type HCProtocol string

const (
	HCProtocolTCP  = "tcp"
	HCProtocolHTTP = "http"
	HCProtocolGRPC = "grpc"
	HCNone         = "none"
)

const (
	defaultInterval           = 1 * time.Minute
	defaultTimeout            = 15 * time.Second
	defaultUnhealthyThreshold = int32(5)
	defaultHealthyThreshold   = int32(1)

	defaultHTTPPath = "/"
)

func (e *Configurer) Configure(cluster *envoy_cluster.Cluster) error {
	activeChecks := e.Conf

	err := healthPanicThreshold(cluster, activeChecks.HealthyPanicThreshold)
	if err != nil {
		return err
	}
	failTrafficOnPanic(cluster, activeChecks.FailTrafficOnPanic)

	tcp := activeChecks.Tcp
	http := activeChecks.Http
	grpc := activeChecks.Grpc

	hcType := selectHealthCheckType(e.Protocol, tcp, http, grpc)

	if hcType != HCNone {
		defaultHealthCheck := buildHealthCheck(activeChecks)
		var healthChecker interface{}

		switch hcType {
		case HCProtocolTCP:
			healthChecker = tcpHealthCheck(tcp)
		case HCProtocolHTTP:
			healthChecker = httpHealthCheck(e.Protocol, http, e.Tags)
		case HCProtocolGRPC:
			healthChecker = grpcHealthCheck(grpc)
		}

		healthCheck := addHealthChecker(defaultHealthCheck, healthChecker)
		cluster.HealthChecks = append(cluster.HealthChecks, healthCheck)
	}

	return nil
}

func selectHealthCheckType(protocol core_mesh.Protocol, tcp *api.TcpHealthCheck, http *api.HttpHealthCheck, grpc *api.GrpcHealthCheck) HCProtocol {
	// match exact
	if (protocol == core_mesh.ProtocolHTTP || protocol == core_mesh.ProtocolHTTP2) && http != nil && !pointer.Deref(http.Disabled) {
		return HCProtocolHTTP
	}
	if protocol == core_mesh.ProtocolGRPC && grpc != nil && !pointer.Deref(grpc.Disabled) {
		return HCProtocolGRPC
	}
	if protocol == core_mesh.ProtocolTCP && tcp != nil && !pointer.Deref(tcp.Disabled) {
		return HCProtocolTCP
	}

	// match fallback HTTP
	if (protocol == core_mesh.ProtocolHTTP || protocol == core_mesh.ProtocolHTTP2) && http != nil && pointer.Deref(http.Disabled) && tcp != nil && !pointer.Deref(tcp.Disabled) {
		return HCProtocolTCP
	}

	// match fallback GRPC
	if protocol == core_mesh.ProtocolGRPC && grpc != nil && pointer.Deref(grpc.Disabled) && tcp != nil && !pointer.Deref(tcp.Disabled) {
		return HCProtocolTCP
	}

	return HCNone
}

func mapUInt32ToInt64Range(value uint32) *envoy_type.Int64Range {
	return &envoy_type.Int64Range{
		Start: int64(value),
		End:   int64(value) + 1,
	}
}

func mapHttpHeaders(headers *api.HeaderModifier, srcTags v1alpha1.MultiValueTagSet) []*envoy_core.HeaderValueOption {
	var envoyHeaders []*envoy_core.HeaderValueOption
	if len(srcTags) > 0 {
		envoyHeaders = append(envoyHeaders, &envoy_core.HeaderValueOption{
			Header: &envoy_core.HeaderValue{
				Key:   tags.TagsHeaderName,
				Value: tags.Serialize(srcTags),
			},
		})
	}
	for _, header := range pointer.Deref(pointer.Deref(headers).Add) {
		for _, val := range strings.Split(string(header.Value), ",") {
			envoyHeaders = append(envoyHeaders, &envoy_core.HeaderValueOption{
				Header: &envoy_core.HeaderValue{
					Key:   string(header.Name),
					Value: val,
				},
				AppendAction: envoy_core.HeaderValueOption_APPEND_IF_EXISTS_OR_ADD,
			})
		}
	}
	for _, header := range pointer.Deref(pointer.Deref(headers).Set) {
		for _, val := range strings.Split(string(header.Value), ",") {
			envoyHeaders = append(envoyHeaders, &envoy_core.HeaderValueOption{
				Header: &envoy_core.HeaderValue{
					Key:   string(header.Name),
					Value: val,
				},
				AppendAction: envoy_core.HeaderValueOption_OVERWRITE_IF_EXISTS_OR_ADD,
			})
		}
	}
	return envoyHeaders
}

func tcpHealthCheck(
	tcpConf *api.TcpHealthCheck,
) *envoy_core.HealthCheck_TcpHealthCheck_ {
	tcpHealthCheck := envoy_core.HealthCheck_TcpHealthCheck{}

	if tcpConf.Send != nil {
		tcpHealthCheck.Send = &envoy_core.HealthCheck_Payload{
			Payload: &envoy_core.HealthCheck_Payload_Text{
				Text: hex.EncodeToString([]byte(*tcpConf.Send)),
			},
		}
	}

	if tcpConf.Receive != nil {
		var receive []*envoy_core.HealthCheck_Payload
		for _, r := range *tcpConf.Receive {
			receive = append(receive, &envoy_core.HealthCheck_Payload{
				Payload: &envoy_core.HealthCheck_Payload_Text{
					Text: hex.EncodeToString([]byte(r)),
				},
			})
		}
		tcpHealthCheck.Receive = receive
	}

	return &envoy_core.HealthCheck_TcpHealthCheck_{
		TcpHealthCheck: &tcpHealthCheck,
	}
}

func httpHealthCheck(protocol core_mesh.Protocol, httpConf *api.HttpHealthCheck, srcTags v1alpha1.MultiValueTagSet) *envoy_core.HealthCheck_HttpHealthCheck_ {
	var expectedStatuses []*envoy_type.Int64Range
	if httpConf.ExpectedStatuses != nil {
		for _, status := range *httpConf.ExpectedStatuses {
			expectedStatuses = append(
				expectedStatuses,
				mapUInt32ToInt64Range(uint32(status)),
			)
		}
	}

	codecClientType := envoy_type.CodecClientType_HTTP1
	if protocol == core_mesh.ProtocolHTTP2 {
		codecClientType = envoy_type.CodecClientType_HTTP2
	}

	path := defaultHTTPPath
	if httpConf.Path != nil {
		path = *httpConf.Path
	}

	httpHealthCheck := envoy_core.HealthCheck_HttpHealthCheck{
		Path:                path,
		RequestHeadersToAdd: mapHttpHeaders(httpConf.RequestHeadersToAdd, srcTags),
		ExpectedStatuses:    expectedStatuses,
		CodecClientType:     codecClientType,
	}

	return &envoy_core.HealthCheck_HttpHealthCheck_{
		HttpHealthCheck: &httpHealthCheck,
	}
}

func grpcHealthCheck(
	grpcConf *api.GrpcHealthCheck,
) *envoy_core.HealthCheck_GrpcHealthCheck_ {
	return &envoy_core.HealthCheck_GrpcHealthCheck_{
		GrpcHealthCheck: &envoy_core.HealthCheck_GrpcHealthCheck{
			ServiceName: pointer.Deref(grpcConf.ServiceName),
			Authority:   pointer.Deref(grpcConf.Authority),
		},
	}
}

func healthPanicThreshold(cluster *envoy_cluster.Cluster, value *intstr.IntOrString) error {
	if value == nil {
		return nil
	}
	if cluster.CommonLbConfig == nil {
		cluster.CommonLbConfig = &envoy_cluster.Cluster_CommonLbConfig{}
	}
	percentage, err := envoyPercent(*value)
	if err != nil {
		return err
	}
	cluster.CommonLbConfig.HealthyPanicThreshold = percentage
	return nil
}

func failTrafficOnPanic(cluster *envoy_cluster.Cluster, value *bool) {
	if value == nil {
		return
	}
	if cluster.CommonLbConfig == nil {
		cluster.CommonLbConfig = &envoy_cluster.Cluster_CommonLbConfig{}
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
		cluster.CommonLbConfig.LocalityConfigSpecifier = &envoy_cluster.Cluster_CommonLbConfig_ZoneAwareLbConfig_{
			ZoneAwareLbConfig: &envoy_cluster.Cluster_CommonLbConfig_ZoneAwareLbConfig{},
		}
	}
	if zoneAwareLbConfig := cluster.CommonLbConfig.GetZoneAwareLbConfig(); zoneAwareLbConfig != nil {
		zoneAwareLbConfig.FailTrafficOnPanic = *value
	}
}

func buildHealthCheck(conf api.Conf) *envoy_core.HealthCheck {
	interval := defaultInterval
	if conf.Interval != nil {
		interval = conf.Interval.Duration
	}

	timeout := defaultTimeout
	if conf.Timeout != nil {
		timeout = conf.Timeout.Duration
	}

	unhealthyThreshold := defaultUnhealthyThreshold
	if conf.UnhealthyThreshold != nil {
		unhealthyThreshold = *conf.UnhealthyThreshold
	}

	healthyThreshold := defaultHealthyThreshold
	if conf.HealthyThreshold != nil {
		healthyThreshold = *conf.HealthyThreshold
	}

	hc := &envoy_core.HealthCheck{
		HealthChecker: &envoy_core.HealthCheck_TcpHealthCheck_{
			TcpHealthCheck: &envoy_core.HealthCheck_TcpHealthCheck{},
		},
		Interval:           util_proto.Duration(interval),
		Timeout:            util_proto.Duration(timeout),
		UnhealthyThreshold: util_proto.UInt32(uint32(unhealthyThreshold)),
		HealthyThreshold:   util_proto.UInt32(uint32(healthyThreshold)),
	}

	if conf.InitialJitter != nil {
		hc.InitialJitter = util_proto.Duration(conf.InitialJitter.Duration)
	}
	if conf.IntervalJitter != nil {
		hc.IntervalJitter = util_proto.Duration(conf.IntervalJitter.Duration)
	}
	if conf.IntervalJitterPercent != nil {
		hc.IntervalJitterPercent = uint32(*conf.IntervalJitterPercent)
	}
	if conf.EventLogPath != nil {
		config := envoy_healthchecks.HealthCheckEventFileSink{
			EventLogPath: *conf.EventLogPath,
		}
		hc.EventLogger = []*envoy_core.TypedExtensionConfig{{
			Name:        "envoy.health_check.event_sinks.file",
			TypedConfig: util_proto.MustMarshalAny(&config),
		}}
	}
	if conf.AlwaysLogHealthCheckFailures != nil {
		hc.AlwaysLogHealthCheckFailures = *conf.AlwaysLogHealthCheckFailures
	}
	if conf.NoTrafficInterval != nil {
		hc.NoTrafficInterval = util_proto.Duration(conf.NoTrafficInterval.Duration)
	}
	if conf.ReuseConnection != nil {
		hc.ReuseConnection = util_proto.Bool(*conf.ReuseConnection)
	}

	return hc
}

func addHealthChecker(healthCheck *envoy_core.HealthCheck, healthChecker interface{}) *envoy_core.HealthCheck {
	if httpHc, ok := healthChecker.(*envoy_core.HealthCheck_HttpHealthCheck_); ok {
		healthCheck.HealthChecker = httpHc
	} else if tcpHc, ok := healthChecker.(*envoy_core.HealthCheck_TcpHealthCheck_); ok {
		healthCheck.HealthChecker = tcpHc
	} else if grpcHc, ok := healthChecker.(*envoy_core.HealthCheck_GrpcHealthCheck_); ok {
		healthCheck.HealthChecker = grpcHc
	}

	return healthCheck
}

func envoyPercent(intOrStr intstr.IntOrString) (*envoy_type.Percent, error) {
	decimal, err := common_api.NewDecimalFromIntOrString(intOrStr)
	if err != nil {
		return nil, err
	}
	value, _ := decimal.Float64()
	return &envoy_type.Percent{
		Value: value,
	}, nil
}
