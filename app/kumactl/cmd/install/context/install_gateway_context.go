package context

type InstallGatewayArgs struct {
	Type      string
	Namespace string
}

type InstallGatewayContext struct {
	Args           InstallGatewayArgs
	AvailableTypes map[string]struct{}
}

func DefaultInstallGatewayContext() InstallGatewayContext {
	return InstallGatewayContext{
		Args: InstallGatewayArgs{
			Namespace: "kuma-gateway",
		},
		AvailableTypes: map[string]struct{}{"kong": {}},
	}
}
