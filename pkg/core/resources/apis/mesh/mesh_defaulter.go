package mesh

func (mesh *MeshResource) Default() {
	// default settings for Prometheus metrics
	if mesh.Spec.Metrics != nil {
		if mesh.Spec.Metrics.Prometheus != nil {
			if mesh.Spec.Metrics.Prometheus.Port == 0 {
				mesh.Spec.Metrics.Prometheus.Port = 5670
			}
			if mesh.Spec.Metrics.Prometheus.Path == "" {
				mesh.Spec.Metrics.Prometheus.Path = "/metrics"
			}
		}
	}
}
