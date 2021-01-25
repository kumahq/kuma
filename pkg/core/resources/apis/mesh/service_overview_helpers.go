package mesh

func (t *ServiceOverviewResource) GetStatus() Status {
	if t.Spec.Online == 0 {
		return Offline
	}
	if t.Spec.Online == t.Spec.Total {
		return Online
	}
	return PartiallyDegraded
}
