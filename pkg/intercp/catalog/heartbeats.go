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

func (h *Heartbeats) Collect() []Instance {
	h.Lock()
	defer h.Unlock()
	var instances []Instance
	for k := range h.instances {
		instances = append(instances, k)
	}
	h.instances = map[Instance]struct{}{}
	return instances
}

func (h *Heartbeats) Add(instance Instance) {
	h.Lock()
	h.instances[instance] = struct{}{}
	h.Unlock()
}
