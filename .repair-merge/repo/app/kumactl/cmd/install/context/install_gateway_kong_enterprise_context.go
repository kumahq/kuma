package context

type InstallGatewayKongEnterpriseArgs struct {
	Namespace   string
	LicensePath string
}

type InstallGatewayKongEnterpriseContext struct {
	Args InstallGatewayKongEnterpriseArgs
}

func DefaultInstallGatewayKongEnterpriseContext() InstallGatewayKongEnterpriseContext {
	return InstallGatewayKongEnterpriseContext{
		Args: InstallGatewayKongEnterpriseArgs{
			Namespace: "kong-enterprise-gateway",
		},
	}
}
