package context

import "github.com/kumahq/kuma/v2/pkg/core/resources/model"

type InstallGatewayKongArgs struct {
	Namespace string
	Mesh      string
}

type InstallGatewayKongContext struct {
	Args InstallGatewayKongArgs
}

func DefaultInstallGatewayKongContext() InstallGatewayKongContext {
	return InstallGatewayKongContext{
		Args: InstallGatewayKongArgs{
			Namespace: "kuma-gateway",
			Mesh:      model.DefaultMesh,
		},
	}
}
