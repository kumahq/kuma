package verifier

import (
	"sync"

	envoy_sd "github.com/envoyproxy/go-control-plane/envoy/service/discovery/v3"

	"github.com/kumahq/kuma/pkg/core/resources/store"
	test_grpc "github.com/kumahq/kuma/pkg/test/grpc"
)

type TestContext interface {
	Store() store.ResourceStore
	ServerStream() *test_grpc.MockServerStream
	ClientStream() *test_grpc.MockClientStream
	Stop() chan struct{}
	SaveLastResponse(typ string, response *envoy_sd.DiscoveryResponse)
	LastResponse(typeURL string) *envoy_sd.DiscoveryResponse
	SaveLastACKedResponse(typ string, response *envoy_sd.DiscoveryResponse)
	LastACKedResponse(typ string) *envoy_sd.DiscoveryResponse
	WaitGroup() *sync.WaitGroup
}

type TestContextImpl struct {
	ResourceStore      store.ResourceStore
	MockStream         *test_grpc.MockServerStream
	MockClientStream   *test_grpc.MockClientStream
	StopCh             chan struct{}
	Wg                 *sync.WaitGroup
	Responses          map[string]*envoy_sd.DiscoveryResponse
	LastACKedResponses map[string]*envoy_sd.DiscoveryResponse
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

func (t *TestContextImpl) SaveLastResponse(typ string, response *envoy_sd.DiscoveryResponse) {
	t.Responses[typ] = response
}

func (t *TestContextImpl) LastResponse(typ string) *envoy_sd.DiscoveryResponse {
	return t.Responses[typ]
}

func (t *TestContextImpl) SaveLastACKedResponse(typ string, response *envoy_sd.DiscoveryResponse) {
	t.LastACKedResponses[typ] = response
}

func (t *TestContextImpl) LastACKedResponse(typ string) *envoy_sd.DiscoveryResponse {
	return t.LastACKedResponses[typ]
}

func (t *TestContextImpl) WaitGroup() *sync.WaitGroup {
	return t.Wg
}
