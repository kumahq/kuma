package zone

const (
	IngressScope string = "ingress"
	EgressScope  string = "egress"
)

var FullScope = []string{
	IngressScope,
	EgressScope,
}

func InScope(scope []string, s string) bool {
	for _, item := range scope {
		if item == s {
			return true
		}
	}

	return false
}
