package zone

const (
	// TODO (bartsmykla): uncomment when Zone Token will be available for dataplanes
	// 	and ingresses
	// DataplaneScope string = "dataplane
	// IngressScope string = "ingress"
	EgressScope string = "egress"
)

var FullScope = []string{
	// TODO (bartsmykla): uncomment when Zone Token will be available for dataplanes
	// 	and ingresses
	// DataplaneScope,
	// IngressScope,
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
