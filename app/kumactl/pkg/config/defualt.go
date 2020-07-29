package config

import config_proto "github.com/kumahq/kuma/pkg/config/app/kumactl/v1alpha1"

func DefaultConfiguration() config_proto.Configuration {
	return config_proto.Configuration{
		ControlPlanes: []*config_proto.ControlPlane{
			{
				Name: "local",
				Coordinates: &config_proto.ControlPlaneCoordinates{
					ApiServer: &config_proto.ControlPlaneCoordinates_ApiServer{
						Url: "http://localhost:5681",
					},
				},
			},
		},
		Contexts: []*config_proto.Context{
			{
				Name:         "local",
				ControlPlane: "local",
			},
		},
		CurrentContext: "local",
	}
}
