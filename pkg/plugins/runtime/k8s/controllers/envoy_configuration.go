package controllers

import (
	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
	"github.com/kumahq/kuma/pkg/plugins/runtime/k8s/metadata"
)

func GetEnvoyConfiguration(deltaXds bool, annotations metadata.Annotations) *mesh_proto.EnvoyConfiguration {
	envoyConfig := &mesh_proto.EnvoyConfiguration{
		XdsTransportProtocolVariant: mesh_proto.EnvoyConfiguration_GRPC,
	}
	if deltaXds {
		envoyConfig.XdsTransportProtocolVariant = mesh_proto.EnvoyConfiguration_DELTA_GRPC
	}
	xdsTransportProtocolVariant, exist := annotations.GetString(metadata.KumaXdsTransportProtocolVariant)
	if exist {
		switch xdsTransportProtocolVariant {
		case "DELTA_GRPC":
			envoyConfig.XdsTransportProtocolVariant = mesh_proto.EnvoyConfiguration_DELTA_GRPC
		case "GRPC":
			envoyConfig.XdsTransportProtocolVariant = mesh_proto.EnvoyConfiguration_GRPC
		}
	}
	return envoyConfig
}
