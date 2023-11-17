package builders

import "github.com/kumahq/kuma/pkg/core/xds"

type EndpointMapBuilder struct {
	res xds.EndpointMap
}

func EndpointMap() *EndpointMapBuilder {
	return &EndpointMapBuilder{res: xds.EndpointMap{}}
}

func (em *EndpointMapBuilder) Build() xds.EndpointMap {
	return em.res
}

func (em *EndpointMapBuilder) With(fn func(xds.EndpointMap)) *EndpointMapBuilder {
	fn(em.res)
	return em
}

func (em *EndpointMapBuilder) AddEndpoint(service xds.ServiceName, endpoint *EndpointBuilder) *EndpointMapBuilder {
	em.res[service] = append(em.res[service], *endpoint.Build())
	return em
}

func (em *EndpointMapBuilder) AddEndpoints(service xds.ServiceName, endpoints ...*EndpointBuilder) *EndpointMapBuilder {
	for _, endpoint := range endpoints {
		em.res[service] = append(em.res[service], *endpoint.Build())
	}
	return em
}
