package v1alpha1

import (
	common_api "github.com/kumahq/kuma/api/common/v1alpha1"
)

func (m *MeshTrace) GetTargetRef() *common_api.TargetRef {
	if m == nil {
		return nil
	}
	return m.TargetRef
}

func (m *MeshTrace) GetDefault() *Conf {
	if m == nil {
		return nil
	}
	return m.Default
}

func (c *Conf) GetBackends() []*Backend {
	if c == nil {
		return nil
	}
	return c.Backends
}

func (c *Conf) GetSampling() *Sampling {
	if c == nil {
		return nil
	}
	return c.Sampling
}

func (c *Conf) GetTags() []*Tag {
	if c == nil {
		return nil
	}
	return c.Tags
}

func (b *Backend) GetZipkin() *ZipkinBackend {
	if b == nil {
		return nil
	}
	return b.Zipkin
}

func (b *Backend) GetDatadog() *DatadogBackend {
	if b == nil {
		return nil
	}
	return b.Datadog
}

func (z *ZipkinBackend) GetUrl() string {
	if z == nil {
		return ""
	}
	return z.Url
}

func (z *ZipkinBackend) GetTraceId128Bit() bool {
	if z == nil {
		return false
	}
	return z.TraceId128Bit
}

func (z *ZipkinBackend) GetApiVersion() string {
	if z == nil {
		return ""
	}
	return z.ApiVersion
}

func (z *ZipkinBackend) GetSharedSpanContext() *common_api.BoolValue {
	if z == nil {
		return nil
	}
	return z.SharedSpanContext
}

func (d *DatadogBackend) GetUrl() string {
	if d == nil {
		return ""
	}
	return d.Url
}

func (d *DatadogBackend) GetSplitService() bool {
	if d == nil {
		return false
	}
	return d.SplitService
}

func (s *Sampling) GetOverall() *common_api.UInt32Value {
	if s == nil {
		return nil
	}
	return s.Overall
}

func (s *Sampling) GetClient() *common_api.UInt32Value {
	if s == nil {
		return nil
	}
	return s.Client
}

func (s *Sampling) GetRandom() *common_api.UInt32Value {
	if s == nil {
		return nil
	}
	return s.Random
}

func (t *Tag) GetName() string {
	if t == nil {
		return ""
	}
	return t.Name
}

func (t *Tag) GetLiteral() string {
	if t == nil {
		return ""
	}
	return t.Literal
}

func (t *Tag) GetHeader() *HeaderTag {
	if t == nil {
		return nil
	}
	return t.Header
}

func (h *HeaderTag) GetName() string {
	if h == nil {
		return ""
	}
	return h.Name
}

func (h *HeaderTag) GetDefault() string {
	if h == nil {
		return ""
	}
	return h.Default
}
