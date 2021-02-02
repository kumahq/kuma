package hooks

type Hooks struct {
	resourceSetHooks []ResourceSetHook
}

func (h *Hooks) AddResourceSetHook(hook ResourceSetHook) {
	h.resourceSetHooks = append(h.resourceSetHooks, hook)
}

func (h *Hooks) ResourceSetHooks() []ResourceSetHook {
	return h.resourceSetHooks
}
