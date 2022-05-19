package context

type Dashboard struct {
	FileName string
	Content  string
}

type ObservabilityTemplateArgs struct {
	Namespace         string
	Mesh              string
	KumaCpAddress     string
	JaegerAddress     string
	LokiAddress       string
	PrometheusAddress string
	KumaCpApiAddress  string
	WithoutPrometheus bool
	WithoutGrafana    bool
	WithoutLoki       bool
	WithoutJaeger     bool
	Dashboards        []Dashboard
}

type InstallObservabilityContext struct {
	TemplateArgs ObservabilityTemplateArgs
}

func DefaultInstallObservabilityContext() InstallObservabilityContext {
	return InstallObservabilityContext{
		TemplateArgs: ObservabilityTemplateArgs{
			Namespace:         "kuma-observability",
			KumaCpAddress:     "http://kuma-control-plane.kuma-system:5676",
			KumaCpApiAddress:  "http://kuma-control-plane.kuma-system:5681",
			JaegerAddress:     "http://jaeger-query.kuma-observability",
			LokiAddress:       "http://loki.kuma-observability:3100",
			PrometheusAddress: "http://prometheus-server.kuma-observability",
			WithoutPrometheus: false,
			WithoutGrafana:    false,
			WithoutLoki:       false,
			WithoutJaeger:     false,
		},
	}
}
