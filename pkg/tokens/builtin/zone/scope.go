package zone

import "slices"

const (
	IngressScope string = "ingress"
	EgressScope  string = "egress"
)

var FullScope = []string{
	IngressScope,
	EgressScope,
}

func InScope(scope []string, s string) bool {
	return slices.Contains(scope, s)
}
