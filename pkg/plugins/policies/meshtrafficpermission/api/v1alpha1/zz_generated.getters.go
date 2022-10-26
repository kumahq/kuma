package v1alpha1

import (
	common_api "github.com/kumahq/kuma/api/common/v1alpha1"
)

func (m *MeshTrafficPermission) GetTargetRef() *common_api.TargetRef {
	if m == nil {
		return nil
	}
	return m.TargetRef
}

func (m *MeshTrafficPermission) GetFrom() []*From {
	if m == nil {
		return nil
	}
	return m.From
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

func (c *Conf) GetAction() Action {
	if c == nil {
		return ""
	}
	return c.Action
}
