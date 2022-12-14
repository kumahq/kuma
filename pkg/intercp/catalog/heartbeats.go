package catalog

import (
	"sync"
)

type Heartbeats struct {
	instances map[Instance]struct{}
	sync.Mutex
}

func NewHeartbeats() *Heartbeats {
	return &Heartbeats{
		instances: map[Instance]struct{}{},
	}
}

func (h *Heartbeats) ResetAndCollect() []Instance {
	h.Lock()
	currentInstances := h.instances
	h.instances = map[Instance]struct{}{}
	h.Unlock()
	var instances []Instance
	for k := range currentInstances {
		instances = append(instances, k)
	}
	return instances
}

func (h *Heartbeats) Add(instance Instance) {
	h.Lock()
	h.instances[instance] = struct{}{}
	h.Unlock()
}

func (h *Heartbeats) Remove(instance Instance) {
	h.Lock()
	delete(h.instances, instance)
	h.Unlock()
}
