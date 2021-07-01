// Package kds provides support of Kuma Discovery Service, extension of xDS
package kds

import (
	"github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
	"github.com/kumahq/kuma/pkg/core/resources/apis/system"
	"github.com/kumahq/kuma/pkg/core/resources/model"
)

const (
	googleApis = "type.googleapis.com/"

	// KumaResource is the type URL of the KumaResource protobuf.
	KumaResource = googleApis + "kuma.mesh.v1alpha1.KumaResource"
)

var (
	// SupportedTypes is a list of Kuma types that may be exchanged by KDS peers.
	SupportedTypes = []model.ResourceType{
		mesh.CircuitBreakerType,
		mesh.DataplaneInsightType,
		mesh.DataplaneType,
		mesh.ExternalServiceType,
		mesh.FaultInjectionType,
		mesh.HealthCheckType,
		mesh.MeshType,
		mesh.ProxyTemplateType,
		mesh.RateLimitType,
		mesh.RetryType,
		mesh.TimeoutType,
		mesh.TrafficLogType,
		mesh.TrafficPermissionType,
		mesh.TrafficRouteType,
		mesh.TrafficTraceType,
		mesh.ZoneIngressInsightType,
		mesh.ZoneIngressType,
		system.ConfigType,
		system.GlobalSecretType,
		system.SecretType,
	}
)
