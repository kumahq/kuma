package xds

import (
    envoy_cluster "github.com/envoyproxy/go-control-plane/envoy/config/cluster/v3"
    envoy_core "github.com/envoyproxy/go-control-plane/envoy/config/core/v3"
    envoy_type "github.com/envoyproxy/go-control-plane/envoy/type/v3"
    "github.com/pkg/errors"

    "github.com/kumahq/kuma/pkg/core"
    core_mesh "github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
    api "github.com/kumahq/kuma/pkg/plugins/policies/meshhealthcheck/api/v1alpha1"
    util_proto "github.com/kumahq/kuma/pkg/util/proto"
)

type Configurer struct {
    Conf     api.Conf
    Protocol core_mesh.Protocol
}

type HCProtocol string

const (
    HCProtocolTCP  = "tcp"
    HCProtocolHTTP = "http"
    HCProtocolGRPC = "grpc"
    HCNone         = "none"
)

func (e *Configurer) Configure(cluster *envoy_cluster.Cluster) error {
    activeChecks := e.Conf

    healthPanicThreshold(cluster, activeChecks.HealthyPanicThreshold)
    failTrafficOnPanic(cluster, activeChecks.FailTrafficOnPanic)

    tcp := activeChecks.Tcp
    http := activeChecks.Http
    grpc := activeChecks.Grpc

    hcType := selectHealthCheckType(e.Protocol, tcp, http, grpc)

    if hcType != HCNone {
        defaultHealthCheck := buildHealthCheck(activeChecks)
        var healthChecker interface{}
        if hcType == HCProtocolTCP {
            healthChecker = tcpHealthCheck(tcp)
        } else if hcType == HCProtocolHTTP {
            healthChecker = httpHealthCheck(e.Protocol, http)
        } else if hcType == HCProtocolGRPC {
            healthChecker = grpcHealthCheck(grpc)
        }

        healthCheck := addHealthChecker(defaultHealthCheck, healthChecker)
        cluster.HealthChecks = append(cluster.HealthChecks, healthCheck)
    }

    return nil
}

func selectHealthCheckType(protocol core_mesh.Protocol, tcp *api.TcpHealthCheck, http *api.HttpHealthCheck, grpc *api.GrpcHealthCheck) HCProtocol {
    // match exact
    if (protocol == core_mesh.ProtocolHTTP || protocol == core_mesh.ProtocolHTTP2) && http != nil && http.Disabled == false {
        return HCProtocolHTTP
    }
    if protocol == core_mesh.ProtocolGRPC && grpc != nil && grpc.Disabled == false {
        return HCProtocolGRPC
    }
    if protocol == core_mesh.ProtocolTCP && tcp != nil && tcp.Disabled == false {
        return HCProtocolTCP
    }

    // match fallback HTTP
    if (protocol == core_mesh.ProtocolHTTP || protocol == core_mesh.ProtocolHTTP2) && http != nil && http.Disabled == true && tcp != nil && tcp.Disabled == false {
        return HCProtocolTCP
    }

    // match fallback GRPC
    if protocol == core_mesh.ProtocolGRPC && grpc != nil && grpc.Disabled == true && tcp != nil && tcp.Disabled == false {
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

func mapHttpHeaders(headers *[]api.HeaderValueOption) []*envoy_core.HeaderValueOption {
    var envoyHeaders []*envoy_core.HeaderValueOption
    if headers != nil {
        for _, header := range *headers {
            hvo := &envoy_core.HeaderValueOption{
                Header: &envoy_core.HeaderValue{
                    Key:   header.Header.Key,
                    Value: header.Header.Value,
                },
            }

            if header.Append != nil {
                hvo.Append = util_proto.Bool(*header.Append)
            }

            envoyHeaders = append(envoyHeaders, hvo)
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
                Text: *tcpConf.Send,
            },
        }
    }

    if tcpConf.Receive != nil {
        var receive []*envoy_core.HealthCheck_Payload
        for _, r := range *tcpConf.Receive {
            receive = append(receive, &envoy_core.HealthCheck_Payload{
                Payload: &envoy_core.HealthCheck_Payload_Text{
                    Text: r,
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
    httpConf *api.HttpHealthCheck,
) *envoy_core.HealthCheck_HttpHealthCheck_ {
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

func grpcHealthCheck(
    grpcConf *api.GrpcHealthCheck,
) *envoy_core.HealthCheck_GrpcHealthCheck_ {

    grpcHealthCheck := envoy_core.HealthCheck_GrpcHealthCheck{
        ServiceName: grpcConf.ServiceName,
        Authority: grpcConf.Authority,
    }

    return &envoy_core.HealthCheck_GrpcHealthCheck_{
        GrpcHealthCheck: &grpcHealthCheck,
    }
}

func healthPanicThreshold(cluster *envoy_cluster.Cluster, value *int32) {
    if value == nil {
        return
    }
    if cluster.CommonLbConfig == nil {
        cluster.CommonLbConfig = &envoy_cluster.Cluster_CommonLbConfig{}
    }
    cluster.CommonLbConfig.HealthyPanicThreshold = &envoy_type.Percent{Value: float64(*value)}
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
    hc := &envoy_core.HealthCheck{
        HealthChecker: &envoy_core.HealthCheck_TcpHealthCheck_{
            TcpHealthCheck: &envoy_core.HealthCheck_TcpHealthCheck{},
        },
        Interval:           util_proto.Duration(conf.Interval.Duration),
        Timeout:            util_proto.Duration(conf.Timeout.Duration),
        UnhealthyThreshold: util_proto.UInt32(uint32(conf.UnhealthyThreshold)),
        HealthyThreshold:   util_proto.UInt32(uint32(conf.HealthyThreshold)),
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
        hc.EventLogPath = *conf.EventLogPath
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
