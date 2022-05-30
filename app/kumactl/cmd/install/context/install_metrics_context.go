package context

type MetricsTemplateArgs struct {
	Namespace         string
	Mesh              string
	KumaCpAddress     string
	JaegerAddress     string
	LokiAddress       string
	KumaCpApiAddress  string
	WithoutPrometheus bool
	WithoutGrafana    bool
	Dashboards        []Dashboard
}

type InstallMetricsContext struct {
	TemplateArgs MetricsTemplateArgs
}

func DefaultInstallMetricsContext() InstallMetricsContext {
	return InstallMetricsContext{
		TemplateArgs: MetricsTemplateArgs{
			Namespace:         "kuma-metrics",
			KumaCpAddress:     "http://kuma-control-plane.kuma-system:5676",
			KumaCpApiAddress:  "http://kuma-control-plane.kuma-system:5681",
			JaegerAddress:     "http://jaeger-query.kuma-tracing",
			LokiAddress:       "http://loki.kuma-logging:3100",
			WithoutPrometheus: false,
			WithoutGrafana:    false,
		},
	}
}
