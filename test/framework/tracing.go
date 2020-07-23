package framework

type Tracing interface {
	ZipkinCollectorURL() string
	TracedServices() ([]string, error)
}
