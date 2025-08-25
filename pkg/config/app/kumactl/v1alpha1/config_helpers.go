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

func (cfg *Configuration) AddContext(c *Context, force bool) bool {
	idx, old := cfg.GetContext(c.Name)
	if old != nil {
		if !force {
			return false
		}
		cfg.Contexts[idx] = c
	} else {
		cfg.Contexts = append(cfg.Contexts, c)
	}
	return true
}

func (cfg *Configuration) RemoveContext(name string) bool {
	i, old := cfg.GetContext(name)
	if old == nil {
		return false
	}
	cfg.Contexts = append(cfg.Contexts[:i], cfg.Contexts[i+1:]...)
	if cfg.CurrentContext == name {
		cfg.CurrentContext = ""
		if len(cfg.Contexts) > 0 {
			cfg.CurrentContext = cfg.Contexts[0].Name
		}
	}
	return true
}

func (cfg *Configuration) SwitchContext(name string) bool {
	_, ctx := cfg.GetContext(name)
	if ctx == nil {
		return false
	}
	cfg.CurrentContext = ctx.Name
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

func (cfg *Configuration) AddControlPlane(cp *ControlPlane, force bool) bool {
	idx, old := cfg.GetControlPlane(cp.Name)
	if old != nil {
		if !force {
			return false
		}
		cfg.ControlPlanes[idx] = cp
	} else {
		cfg.ControlPlanes = append(cfg.ControlPlanes, cp)
	}

	return true
}

func (cfg *Configuration) RemoveControlPlane(name string) bool {
	i, old := cfg.GetControlPlane(name)
	if old == nil {
		return false
	}
	for _, context := range cfg.Contexts {
		if context.ControlPlane == name {
			cfg.RemoveContext(context.Name)
		}
	}
	cfg.ControlPlanes = append(cfg.ControlPlanes[:i], cfg.ControlPlanes[i+1:]...)
	return true
}

func (cfg *ControlPlaneCoordinates_ApiServer) HasCerts() bool {
	if cfg == nil {
		return false
	}
	return cfg.ClientCertFile != "" && cfg.ClientKeyFile != "" && cfg.CaCertFile != ""
}
