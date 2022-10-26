package v1alpha1

import (
	common_api "github.com/kumahq/kuma/api/common/v1alpha1"
)

func (d *DoNothingPolicy) GetTargetRef() *common_api.TargetRef {
	if d == nil {
		return nil
	}
	return d.TargetRef
}

func (d *DoNothingPolicy) GetTo() []*To {
	if d == nil {
		return nil
	}
	return d.To
}

func (d *DoNothingPolicy) GetFrom() []*From {
	if d == nil {
		return nil
	}
	return d.From
}

func (t *To) GetTargetRef() *common_api.TargetRef {
	if t == nil {
		return nil
	}
	return t.TargetRef
}

func (t *To) GetDefault() *Conf {
	if t == nil {
		return nil
	}
	return t.Default
}

func (f *From) GetTargetRef() *common_api.TargetRef {
	if f == nil {
		return nil
	}
	return f.TargetRef
}

func (f *From) GetDefault() *Conf {
	if f == nil {
		return nil
	}
	return f.Default
}

func (c *Conf) GetEnableDoNothing() bool {
	if c == nil {
		return false
	}
	return c.EnableDoNothing
}
