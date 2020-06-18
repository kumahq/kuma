// Package kds provides support of Kuma Discovery Service, extension of xDS
package kds

import (
	"github.com/Kong/kuma/pkg/core/resources/apis/mesh"
	"github.com/Kong/kuma/pkg/core/resources/model"
)

const (
	prefix = "type.googleapis.com/kuma.mesh.v1alpha1."
)

var (
	SupportedTypes = []model.ResourceType{
		mesh.MeshType,
		mesh.DataplaneType, // for Ingress
		mesh.CircuitBreakerType,
		mesh.FaultInjectionType,
		mesh.HealthCheckType,
		mesh.TrafficLogType,
		mesh.TrafficPermissionType,
		mesh.TrafficRouteType,
		mesh.TrafficTraceType,
		mesh.ProxyTemplateType,
	}
)

func TypeURL(resourceType model.ResourceType) string {
	return prefix + string(resourceType)
}
