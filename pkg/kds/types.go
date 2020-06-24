// Package kds provides support of Kuma Discovery Service, extension of xDS
package kds

import (
	"github.com/Kong/kuma/pkg/core/resources/apis/mesh"
	"github.com/Kong/kuma/pkg/core/resources/apis/system"
	"github.com/Kong/kuma/pkg/core/resources/model"
)

const (
	googleApis   = "type.googleapis.com/"
	KumaResource = googleApis + "kuma.mesh.v1alpha1.KumaResource"
)

var (
	SupportedTypes = []model.ResourceType{
		mesh.MeshType,
		mesh.DataplaneType,
		mesh.CircuitBreakerType,
		mesh.FaultInjectionType,
		mesh.HealthCheckType,
		mesh.TrafficLogType,
		mesh.TrafficPermissionType,
		mesh.TrafficRouteType,
		mesh.TrafficTraceType,
		mesh.ProxyTemplateType,
		system.SecretType,
	}
)
