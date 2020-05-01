package sample

func (r *TrafficRouteResource) Default() error {
	if r.Spec.Path == "" {
		r.Spec.Path = "/default"
	}
	return nil
}
