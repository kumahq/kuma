package v1alpha1

import (
	common_api "github.com/kumahq/kuma/api/common/v1alpha1"
)

func (m *MeshAccessLog) GetTargetRef() *common_api.TargetRef {
	if m == nil {
		return nil
	}
	return m.TargetRef
}

func (m *MeshAccessLog) GetTo() []*To {
	if m == nil {
		return nil
	}
	return m.To
}

func (m *MeshAccessLog) GetFrom() []*From {
	if m == nil {
		return nil
	}
	return m.From
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

func (c *Conf) GetBackends() []*Backend {
	if c == nil {
		return nil
	}
	return c.Backends
}

func (b *Backend) GetTcp() *TCPBackend {
	if b == nil {
		return nil
	}
	return b.Tcp
}

func (b *Backend) GetFile() *FileBackend {
	if b == nil {
		return nil
	}
	return b.File
}

func (t *TCPBackend) GetFormat() *Format {
	if t == nil {
		return nil
	}
	return t.Format
}

func (t *TCPBackend) GetAddress() string {
	if t == nil {
		return ""
	}
	return t.Address
}

func (f *FileBackend) GetFormat() *Format {
	if f == nil {
		return nil
	}
	return f.Format
}

func (f *FileBackend) GetPath() string {
	if f == nil {
		return ""
	}
	return f.Path
}

func (f *Format) GetPlain() string {
	if f == nil {
		return ""
	}
	return f.Plain
}

func (f *Format) GetJson() []*JsonValue {
	if f == nil {
		return nil
	}
	return f.Json
}

func (j *JsonValue) GetKey() string {
	if j == nil {
		return ""
	}
	return j.Key
}

func (j *JsonValue) GetValue() string {
	if j == nil {
		return ""
	}
	return j.Value
}
