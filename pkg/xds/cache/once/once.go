package once

import "sync"

type once struct {
	syncOnce sync.Once
	Value    interface{}
	Err      error
}

func (o *once) Do(f func() (interface{}, error)) {
	o.syncOnce.Do(func() {
		o.Value, o.Err = f()
	})
}

func newMap() *omap {
	return &omap{
		m: map[string]*once{},
	}
}

type omap struct {
	mtx sync.Mutex
	m   map[string]*once
}

func (c *omap) Get(key string) (*once, bool) {
	c.mtx.Lock()
	defer c.mtx.Unlock()
	o, exist := c.m[key]
	if !exist {
		o = &once{}
		c.m[key] = o
		return o, true
	}
	return o, false
}

func (c *omap) Delete(key string) {
	c.mtx.Lock()
	delete(c.m, key)
	c.mtx.Unlock()
}
