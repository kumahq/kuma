package once

import "sync"

type Once struct {
	syncOnce sync.Once
	Value    interface{}
	Err      error
}

func (o *Once) Do(f func() (interface{}, error)) {
	o.syncOnce.Do(func() {
		o.Value, o.Err = f()
	})
}

func NewMap() *Map {
	return &Map{
		m: map[string]*Once{},
	}
}

type Map struct {
	mtx sync.Mutex
	m   map[string]*Once
}

func (c *Map) Get(key string) *Once {
	c.mtx.Lock()
	defer c.mtx.Unlock()
	o, exist := c.m[key]
	if !exist {
		o = &Once{}
		c.m[key] = o
	}
	return o
}

func (c *Map) Delete(key string) {
	c.mtx.Lock()
	delete(c.m, key)
	c.mtx.Unlock()
}
