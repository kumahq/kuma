package context

type InstallGatewayKongArgs struct {
	Namespace string
}

type InstallGatewayKongContext struct {
	Args InstallGatewayKongArgs
}

func DefaultInstallGatewayKongContext() InstallGatewayKongContext {
	return InstallGatewayKongContext{
		Args: InstallGatewayKongArgs{
			Namespace: "kuma-gateway",
		},
	}
}
