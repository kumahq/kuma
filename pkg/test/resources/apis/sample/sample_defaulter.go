package sample

func (r *TrafficRouteResource) Default() {
	if r.Spec.Path == "" {
		r.Spec.Path = "/default"
	}
}
