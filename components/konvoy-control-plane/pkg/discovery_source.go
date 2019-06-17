package pkg

type Service struct {
	Name string
	Endpoints []Workload
}

type Workload struct {
	Labels []string
	Address string
	Port uint32
}

type State struct {
	Zone string
	Services []Service
}

type ServicesSource interface {
	StateChanges() <-chan State
}
