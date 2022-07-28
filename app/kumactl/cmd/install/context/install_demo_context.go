package context

type InstallDemoArgs struct {
	Zone           string
	Namespace      string
	WithoutGateway bool
}

type InstallDemoContext struct {
	Args InstallDemoArgs
}

func DefaultInstallDemoContext() InstallDemoContext {
	return InstallDemoContext{
		Args: InstallDemoArgs{
			Zone:           "local",
			Namespace:      "kuma-demo",
			WithoutGateway: false,
		},
	}
}
