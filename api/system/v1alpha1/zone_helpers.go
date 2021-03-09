package v1alpha1

func (x *Zone) IsEnabled() bool {
	if x.Enabled == nil {
		return true
	}
	return x.Enabled.GetValue()
}
