package controllers

import (
	"github.com/pkg/errors"

	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
	"github.com/kumahq/kuma/pkg/plugins/runtime/k8s/metadata"
)

func GetEnvoyConfiguration(deltaXds bool, annotations metadata.Annotations) (*mesh_proto.EnvoyConfiguration, error) {
	envoyConfig := &mesh_proto.EnvoyConfiguration{
		XdsTransportProtocolVariant: mesh_proto.EnvoyConfiguration_GRPC,
	}
	if deltaXds {
		envoyConfig.XdsTransportProtocolVariant = mesh_proto.EnvoyConfiguration_DELTA_GRPC
	}
	xdsTransportProtocolVariant, exist := annotations.GetString(metadata.KumaXdsTransportProtocolVariant)
	if exist {
		switch xdsTransportProtocolVariant {
		case mesh_proto.EnvoyConfiguration_DELTA_GRPC.String():
			envoyConfig.XdsTransportProtocolVariant = mesh_proto.EnvoyConfiguration_DELTA_GRPC
		case mesh_proto.EnvoyConfiguration_GRPC.String():
			envoyConfig.XdsTransportProtocolVariant = mesh_proto.EnvoyConfiguration_GRPC
		default:
			return nil, errors.Errorf("invalid xds transport protocol variant '%s'", xdsTransportProtocolVariant)
		}
	}
	return envoyConfig, nil
}
