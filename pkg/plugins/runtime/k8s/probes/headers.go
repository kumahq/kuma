package probes

import (
	"fmt"

	corev1 "k8s.io/api/core/v1"
)

const (
	KumaProbeHeaderPrefix = "X-Kuma-Probes-"
	HeaderNameTimeout     = KumaProbeHeaderPrefix + "Timeout"
	HeaderNameHost        = KumaProbeHeaderPrefix + "Host"
	HeaderNameScheme      = KumaProbeHeaderPrefix + "Scheme"
	HeaderNameGRPCService = KumaProbeHeaderPrefix + "GRPC-Service"
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
