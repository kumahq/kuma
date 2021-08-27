package context

type Dashboard struct {
	FileName string
	Content  string
}

type MetricsTemplateArgs struct {
	Namespace         string
	Mesh              string
	KumaCpAddress     string
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
			WithoutPrometheus: false,
			WithoutGrafana:    false,
		},
	}
}
