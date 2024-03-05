package builders

import (
	"github.com/kumahq/kuma/pkg/core/xds"
	"github.com/kumahq/kuma/pkg/test/resources/builders"
)

type EndpointBuilder struct {
	res *xds.Endpoint
}

func Endpoint() *EndpointBuilder {
	return &EndpointBuilder{res: &xds.Endpoint{
		Target:          "192.168.0.1",
		UnixDomainPath:  "",
		Port:            0,
		Tags:            nil,
		Weight:          0,
		Locality:        nil,
		ExternalService: nil,
	}}
}

func (e *EndpointBuilder) Build() *xds.Endpoint {
	return e.res
}

func (e *EndpointBuilder) With(fn func(*xds.Endpoint)) *EndpointBuilder {
	fn(e.res)
	return e
}

func (e *EndpointBuilder) WithTags(tagsKV ...string) *EndpointBuilder {
	e.WithTagsMap(builders.TagsKVToMap(tagsKV))
	return e
}

func (e *EndpointBuilder) WithTagsMap(tags map[string]string) *EndpointBuilder {
	e.res.Tags = tags
	return e
}

func (e *EndpointBuilder) AddTagsMap(tags map[string]string) *EndpointBuilder {
	for k, v := range tags {
		e.res.Tags[k] = v
	}
	return e
}

func (e *EndpointBuilder) WithTarget(target string) *EndpointBuilder {
	e.res.Target = target
	return e
}

func (e *EndpointBuilder) WithPort(port uint32) *EndpointBuilder {
	e.res.Port = port
	return e
}

func (e *EndpointBuilder) WithWeight(weight uint32) *EndpointBuilder {
	e.res.Weight = weight
	return e
}

func (e *EndpointBuilder) WithZone(zone string) *EndpointBuilder {
	if e.res.Locality == nil {
		e.res.Locality = &xds.Locality{}
	}
	e.res.Locality.Zone = zone
	return e
}

func (e *EndpointBuilder) WithPriority(priority uint32) *EndpointBuilder {
	if e.res.Locality == nil {
		e.res.Locality = &xds.Locality{}
	}
	e.res.Locality.Priority = priority
	return e
}

func (e *EndpointBuilder) WithExternalService(externalService *xds.ExternalService) *EndpointBuilder {
	e.res.ExternalService = externalService
	return e
}
