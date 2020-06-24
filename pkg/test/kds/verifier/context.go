package verifier

import (
	"sync"

	v2 "github.com/envoyproxy/go-control-plane/envoy/api/v2"

	"github.com/Kong/kuma/pkg/core/resources/store"
	test_grpc "github.com/Kong/kuma/pkg/test/grpc"
)

type TestContext interface {
	Store() store.ResourceStore
	ServerStream() *test_grpc.MockServerStream
	ClientStream() *test_grpc.MockClientStream
	Stop() chan struct{}
	SaveLastResponse(typ string, response *v2.DiscoveryResponse)
	LastResponse(typeURL string) *v2.DiscoveryResponse
	SaveLastACKedResponse(typ string, response *v2.DiscoveryResponse)
	LastACKedResponse(typ string) *v2.DiscoveryResponse
	WaitGroup() *sync.WaitGroup
}

type TestContextImpl struct {
	ResourceStore      store.ResourceStore
	MockStream         *test_grpc.MockServerStream
	MockClientStream   *test_grpc.MockClientStream
	StopCh             chan struct{}
	Wg                 *sync.WaitGroup
	Responses          map[string]*v2.DiscoveryResponse
	LastACKedResponses map[string]*v2.DiscoveryResponse
}

func (t *TestContextImpl) Store() store.ResourceStore {
	return t.ResourceStore
}

func (t *TestContextImpl) ServerStream() *test_grpc.MockServerStream {
	return t.MockStream
}

func (t *TestContextImpl) ClientStream() *test_grpc.MockClientStream {
	return t.MockClientStream
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

func (t *TestContextImpl) SaveLastACKedResponse(typ string, response *v2.DiscoveryResponse) {
	t.LastACKedResponses[typ] = response
}

func (t *TestContextImpl) LastACKedResponse(typ string) *v2.DiscoveryResponse {
	return t.LastACKedResponses[typ]
}

func (t *TestContextImpl) WaitGroup() *sync.WaitGroup {
	return t.Wg
}
