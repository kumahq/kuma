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

func (cfg *Configuration) AddContext(c *Context) bool {
	_, old := cfg.GetContext(c.Name)
	if old != nil {
		return false
	}
	cfg.Contexts = append(cfg.Contexts, c)
	return true
}

func (cfg *Configuration) GetControlPlane(name string) (int, *ControlPlane) {
	for i, p := range cfg.ControlPlanes {
		if p.Name == name {
			return i, p
		}
	}
	return -1, nil
}

func (cfg *Configuration) AddControlPlane(cp *ControlPlane) bool {
	_, old := cfg.GetControlPlane(cp.Name)
	if old != nil {
		return false
	}
	cfg.ControlPlanes = append(cfg.ControlPlanes, cp)
	return true
}
