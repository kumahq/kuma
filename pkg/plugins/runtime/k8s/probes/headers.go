package probes

import (
	"fmt"
	corev1 "k8s.io/api/core/v1"
)

const (
	KumaProbeHeaderPrefix = "x-kuma-probes-"
	HeaderNameTimeout     = KumaProbeHeaderPrefix + "timeout"
	HeaderNameHost        = KumaProbeHeaderPrefix + "host"
	HeaderNameScheme      = KumaProbeHeaderPrefix + "scheme"
	HeaderNameGRPCService = KumaProbeHeaderPrefix + "grpc-service"
)

func TimeoutHeader(timeoutSeconds int32) corev1.HTTPHeader {
	return corev1.HTTPHeader{
		Name:  HeaderNameTimeout,
		Value: fmt.Sprintf("%d", timeoutSeconds),
	}
}

func HostHeader(host string) corev1.HTTPHeader {
	return corev1.HTTPHeader{
		Name:  HeaderNameHost,
		Value: host,
	}
}

func SchemeHeader(scheme corev1.URIScheme) corev1.HTTPHeader {
	return corev1.HTTPHeader{
		Name:  HeaderNameScheme,
		Value: string(scheme),
	}
}

func GRPCServiceHeader(service string) corev1.HTTPHeader {
	return corev1.HTTPHeader{
		Name:  HeaderNameGRPCService,
		Value: service,
	}
}
