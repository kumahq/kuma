package v1alpha1

func (cfg *Configuration) GetControlPlane(name string) (int, *ControlPlane) {
	for i, p := range cfg.ControlPlanes {
		if p.Name == name {
			return i, p
		}
	}
	return -1, nil
}

func (cfg *Configuration) AddControlPlane(cp *ControlPlane) {
	i, p := cfg.GetControlPlane(cp.Name)
	if p != nil {
		cfg.ControlPlanes[i] = cp
	} else {
		cfg.ControlPlanes = append(cfg.ControlPlanes, cp)
	}
}
