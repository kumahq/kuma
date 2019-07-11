package v1alpha1

func (cfg *Configuration) GetCurrent() *Context {
	_, c := cfg.GetContext(cfg.CurrentContext)
	return c
}

func (cfg *Configuration) GetContext(name string) (int, *Context) {
	for i, c := range cfg.Contexts {
		if c.Name == name {
			return i, c
		}
	}
	return -1, nil
}

func (cfg *Configuration) AddContext(c *Context) {
	i, old := cfg.GetContext(c.Name)
	if old != nil {
		cfg.Contexts[i] = c
	} else {
		cfg.Contexts = append(cfg.Contexts, c)
	}
}

func (cfg *Configuration) GetControlPlane(name string) (int, *ControlPlane) {
	for i, p := range cfg.ControlPlanes {
		if p.Name == name {
			return i, p
		}
	}
	return -1, nil
}

func (cfg *Configuration) AddControlPlane(cp *ControlPlane) {
	i, old := cfg.GetControlPlane(cp.Name)
	if old != nil {
		cfg.ControlPlanes[i] = cp
	} else {
		cfg.ControlPlanes = append(cfg.ControlPlanes, cp)
	}
}
