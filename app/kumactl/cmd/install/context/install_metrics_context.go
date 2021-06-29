package context

import kuma_version "github.com/kumahq/kuma/pkg/version"

type Dashboard struct {
	FileName string
	Content  string
}

type MetricsTemplateArgs struct {
	Namespace               string
	Mesh                    string
	KumaPrometheusSdImage   string
	KumaPrometheusSdVersion string
	KumaCpAddress           string
	WithoutPrometheus       bool
	WithoutGrafana          bool
	Dashboards              []Dashboard
}

type InstallMetricsContext struct {
	TemplateArgs MetricsTemplateArgs
}

func DefaultInstallMetricsContext() InstallMetricsContext {
	return InstallMetricsContext{
		TemplateArgs: MetricsTemplateArgs{
			Namespace:               "kuma-metrics",
			KumaPrometheusSdImage:   "docker.io/kumahq/kuma-prometheus-sd",
			KumaPrometheusSdVersion: kuma_version.Build.GitTag,
			KumaCpAddress:           "grpc://kuma-control-plane.kuma-system:5676",
			WithoutPrometheus:       false,
			WithoutGrafana:          false,
		},
	}
}
