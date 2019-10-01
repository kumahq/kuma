package envoy

import (
	"fmt"
	"github.com/Kong/kuma/api/mesh/v1alpha1"
	core_xds "github.com/Kong/kuma/pkg/core/xds"
	"github.com/envoyproxy/go-control-plane/envoy/api/v2/core"
	accesslog "github.com/envoyproxy/go-control-plane/envoy/config/accesslog/v2"
	filter_accesslog "github.com/envoyproxy/go-control-plane/envoy/config/filter/accesslog/v2"
	"github.com/envoyproxy/go-control-plane/pkg/util"
	"github.com/gogo/protobuf/types"
	"github.com/pkg/errors"
	"strings"
)

const AccessLogDefaultFormat = "[%START_TIME%] %DOWNSTREAM_REMOTE_ADDRESS%(%KUMA_DOWNSTREAM_CLUSTER%)->%UPSTREAM_HOST%(%UPSTREAM_CLUSTER%) took %DURATION%ms, sent %BYTES_SENT% bytes, received: %BYTES_RECEIVED% bytes\n"

func convertLoggingBackends(backends []*v1alpha1.LoggingBackend, proxy *core_xds.Proxy) ([]*filter_accesslog.AccessLog, error) {
	var result []*filter_accesslog.AccessLog
	for _, backend := range backends {
		log, err := convertLoggingBackend(backend, proxy)
		if err != nil {
			return nil, err
		}
		result = append(result, log)
	}
	return result, nil
}

func convertLoggingBackend(backend *v1alpha1.LoggingBackend, proxy *core_xds.Proxy) (*filter_accesslog.AccessLog, error) {
	format := AccessLogDefaultFormat
	if backend.Format != "" {
		format = backend.Format
	}
	// if dataplane has no service - fill this with placeholder. Otherwise take the first service
	service := proxy.Dataplane.Spec.GetIdentifyingService()
	format = strings.ReplaceAll(format, "%KUMA_DOWNSTREAM_CLUSTER%", service)

	if file, ok := backend.GetType().(*v1alpha1.LoggingBackend_File_); ok {
		return fileAccessLog(format, file)
	} else if tcp, ok := backend.GetType().(*v1alpha1.LoggingBackend_Tcp_); ok {
		return tcpAccessLog(format, tcp, proxy.Id)
	} else {
		return nil, errors.Errorf("could not convert LoggingBackend of type %T to AccessLog", backend.GetType())
	}
}

func tcpAccessLog(format string, tcp *v1alpha1.LoggingBackend_Tcp_, id core_xds.ProxyId) (*filter_accesslog.AccessLog, error) {
	fileAccessLog := &accesslog.HttpGrpcAccessLogConfig{
		CommonConfig: &accesslog.CommonGrpcAccessLogConfig{
			LogName: fmt.Sprintf("%s;%s", tcp.Tcp.Address, format),
			GrpcService: &core.GrpcService{
				TargetSpecifier: &core.GrpcService_GoogleGrpc_{
					GoogleGrpc: &core.GrpcService_GoogleGrpc{
						StatPrefix: "grpc_service",
						TargetUri:  fmt.Sprintf("unix:///tmp/kuma-access-logs-%s-%s.sock", id.Name, id.Mesh),
					},
				},
			},
		},
	}
	marshalled, err := types.MarshalAny(fileAccessLog)
	if err != nil {
		return nil, errors.Wrap(err, "could not marshall FileAccessLog")
	}
	return &filter_accesslog.AccessLog{
		Name: util.HTTPGRPCAccessLog,
		ConfigType: &filter_accesslog.AccessLog_TypedConfig{
			TypedConfig: marshalled,
		},
	}, nil
}

func fileAccessLog(format string, file *v1alpha1.LoggingBackend_File_) (*filter_accesslog.AccessLog, error) {
	fileAccessLog := &accesslog.FileAccessLog{
		AccessLogFormat: &accesslog.FileAccessLog_Format{
			Format: format,
		},
		Path: file.File.Path,
	}
	marshalled, err := types.MarshalAny(fileAccessLog)
	if err != nil {
		return nil, errors.Wrap(err, "could not marshall FileAccessLog")
	}
	return &filter_accesslog.AccessLog{
		Name: util.FileAccessLog,
		ConfigType: &filter_accesslog.AccessLog_TypedConfig{
			TypedConfig: marshalled,
		},
	}, nil
}
