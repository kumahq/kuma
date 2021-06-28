// Package kds provides support of Kuma Discovery Service, extension of xDS
package kds

import (
	"github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
	"github.com/kumahq/kuma/pkg/core/resources/apis/system"
	"github.com/kumahq/kuma/pkg/core/resources/model"
)

const (
	googleApis   = "type.googleapis.com/"
	KumaResource = googleApis + "kuma.mesh.v1alpha1.KumaResource"
)

var (
	SupportedTypes = []model.ResourceType{
		mesh.CircuitBreakerType,
		mesh.DataplaneType,
		mesh.ZoneIngressType,
		mesh.ZoneIngressInsightType,
		mesh.DataplaneInsightType,
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
		system.SecretType,
		system.GlobalSecretType,
		system.ConfigType,
	}
)
