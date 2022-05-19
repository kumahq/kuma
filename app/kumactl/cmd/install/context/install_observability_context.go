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
	Components        []string
	ComponentsMap     map[string]bool
	Dashboards        []Dashboard
}

type InstallObservabilityContext struct {
	TemplateArgs ObservabilityTemplateArgs
}

func DefaultInstallObservabilityContext() InstallObservabilityContext {
	return InstallObservabilityContext{
		TemplateArgs: ObservabilityTemplateArgs{
			Namespace:         "mesh-observability",
			KumaCpAddress:     "http://kuma-control-plane.kuma-system:5676",
			KumaCpApiAddress:  "http://kuma-control-plane.kuma-system:5681",
			JaegerAddress:     "http://jaeger-query.mesh-observability",
			LokiAddress:       "http://loki.mesh-observability:3100",
			PrometheusAddress: "http://prometheus-server.mesh-observability",
			Components:        []string{"grafana", "prometheus", "loki", "jaeger"},
		},
	}
}
