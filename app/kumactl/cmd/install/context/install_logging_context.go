package context

type LoggingTemplateArgs struct {
	Namespace string
}

type InstallLoggingContext struct {
	TemplateArgs LoggingTemplateArgs
}

func DefaultInstallLoggingContext() InstallLoggingContext {
	return InstallLoggingContext{
		TemplateArgs: LoggingTemplateArgs{
			Namespace: "kuma-logging",
		},
	}
}
