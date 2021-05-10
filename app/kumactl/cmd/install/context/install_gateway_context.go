package context

type InstallGatewayArgs struct {
	Type      string
	Namespace string
}

type InstallGatewayContext struct {
	Args InstallGatewayArgs
}

func DefaultInstallGatewayContext() InstallGatewayContext {
	return InstallGatewayContext{
		Args: InstallGatewayArgs{
			Namespace: "kuma-gateway",
		},
	}
}
