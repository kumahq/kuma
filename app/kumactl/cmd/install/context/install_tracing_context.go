package context

type TracingTemplateArgs struct {
	Namespace string
}

type InstallTracingContext struct {
	TemplateArgs TracingTemplateArgs
}

func DefaultInstallTracingContext() InstallTracingContext {
	return InstallTracingContext{
		TemplateArgs: TracingTemplateArgs{
			Namespace: "kuma-tracing",
		},
	}
}
