package verifier

import (
	v2 "github.com/envoyproxy/go-control-plane/envoy/api/v2"

	"github.com/Kong/kuma/pkg/core/resources/store"
	test_grpc "github.com/Kong/kuma/pkg/test/grpc"
)

type TestContext interface {
	Store() store.ResourceStore
	Stream() *test_grpc.MockStream
	Stop() chan struct{}
	SaveLastResponse(typ string, response *v2.DiscoveryResponse)
	LastResponse(typeURL string) *v2.DiscoveryResponse
}

type TestContextImpl struct {
	ResourceStore store.ResourceStore
	MockStream    *test_grpc.MockStream
	StopCh        chan struct{}
	Responses     map[string]*v2.DiscoveryResponse
}

func (t *TestContextImpl) Store() store.ResourceStore {
	return t.ResourceStore
}

func (t *TestContextImpl) Stream() *test_grpc.MockStream {
	return t.MockStream
}

func (t *TestContextImpl) Stop() chan struct{} {
	return t.StopCh
}

func (t *TestContextImpl) SaveLastResponse(typ string, response *v2.DiscoveryResponse) {
	t.Responses[typ] = response
}

func (t *TestContextImpl) LastResponse(typ string) *v2.DiscoveryResponse {
	return t.Responses[typ]
}
