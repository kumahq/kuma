package accesslog

import (
	envoy_accesslog "github.com/envoyproxy/go-control-plane/envoy/config/accesslog/v3"
	envoy_core "github.com/envoyproxy/go-control-plane/envoy/config/core/v3"
	access_loggers_file "github.com/envoyproxy/go-control-plane/envoy/extensions/access_loggers/file/v3"
	access_loggers_grpc "github.com/envoyproxy/go-control-plane/envoy/extensions/access_loggers/grpc/v3"
	access_loggers_otel "github.com/envoyproxy/go-control-plane/envoy/extensions/access_loggers/open_telemetry/v3"
	"github.com/pkg/errors"
	"google.golang.org/protobuf/proto"

	. "github.com/kumahq/kuma/pkg/envoy/builders/common"
	util_proto "github.com/kumahq/kuma/pkg/util/proto"
	"google.golang.org/protobuf/types/known/wrapperspb"
	matcherv3 "github.com/envoyproxy/go-control-plane/envoy/type/matcher/v3"
)

func NewBuilder() *Builder[envoy_accesslog.AccessLog] {
	return &Builder[envoy_accesslog.AccessLog]{}
}

func Config[R any](name string, builder *Builder[R]) Configurer[envoy_accesslog.AccessLog] {
	return func(accessLog *envoy_accesslog.AccessLog) error {
		r, err := builder.Build()
		if err != nil {
			return err
		}
		msg, ok := any(r).(proto.Message)
		if !ok {
			return errors.Errorf("%T is not proto.Message", r)
		}
		marshaled, err := util_proto.MarshalAnyDeterministic(msg)
		if err != nil {
			return err
		}
		accessLog.Name = name
		accessLog.ConfigType = &envoy_accesslog.AccessLog_TypedConfig{
			TypedConfig: marshaled,
		}
		return nil
	}
}

func MetadataFilter(matchIfKeyNotFound bool, matcherBuilder *Builder[matcherv3.MetadataMatcher]) Configurer[envoy_accesslog.AccessLog] {
	return func(accessLog *envoy_accesslog.AccessLog) error {
		matcher, err := matcherBuilder.Build()
		if err != nil {
			return err
		}
		accessLog.Filter = &envoy_accesslog.AccessLogFilter{
			FilterSpecifier: &envoy_accesslog.AccessLogFilter_MetadataFilter{
				MetadataFilter: &envoy_accesslog.MetadataFilter{
					MatchIfKeyNotFound: &wrapperspb.BoolValue{Value: matchIfKeyNotFound},
					Matcher:            matcher,
				},
			},
		}
		return nil
	}
}

func NewFileBuilder() *Builder[access_loggers_file.FileAccessLog] {
	return &Builder[access_loggers_file.FileAccessLog]{}
}

func Path(p string) Configurer[access_loggers_file.FileAccessLog] {
	return func(a *access_loggers_file.FileAccessLog) error {
		a.Path = p
		return nil
	}
}

func NewOtelBuilder() *Builder[access_loggers_otel.OpenTelemetryAccessLogConfig] {
	return &Builder[access_loggers_otel.OpenTelemetryAccessLogConfig]{}
}

func CommonConfig(logName string, clusterName string) Configurer[access_loggers_otel.OpenTelemetryAccessLogConfig] {
	return func(otel *access_loggers_otel.OpenTelemetryAccessLogConfig) error {
		otel.CommonConfig = &access_loggers_grpc.CommonGrpcAccessLogConfig{
			LogName:             logName,
			TransportApiVersion: envoy_core.ApiVersion_V3,
			GrpcService: &envoy_core.GrpcService{
				TargetSpecifier: &envoy_core.GrpcService_EnvoyGrpc_{
					EnvoyGrpc: &envoy_core.GrpcService_EnvoyGrpc{
						ClusterName: clusterName,
					},
				},
			},
		}
		return nil
	}
}
