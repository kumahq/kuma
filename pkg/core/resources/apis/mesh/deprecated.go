package mesh

func (t *MeshResource) Deprecations() []string {
	if t.Spec.MeshServices == nil { //nolint:staticcheck // deprecated on purpose
		return nil
	}
	return []string{
		"meshServices was removed in 3.0 and is ignored: every mesh behaves as Exclusive. Drop the field before it is reserved again in a future release.",
	}
}
