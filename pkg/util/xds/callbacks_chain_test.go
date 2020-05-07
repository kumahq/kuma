package xds_test

import (
	"context"
	"fmt"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	. "github.com/Kong/kuma/pkg/util/xds"

	envoy "github.com/envoyproxy/go-control-plane/envoy/api/v2"
	envoy_xds "github.com/envoyproxy/go-control-plane/pkg/server/v2"
)

var _ = Describe("CallbacksChain", func() {

	var first, second CallbacksFuncs

	type methodCall struct {
		obj    string
		method string
		args   []interface{}
	}
	var calls []methodCall

	BeforeEach(func() {
		calls = make([]methodCall, 0)
		first = CallbacksFuncs{
			OnStreamOpenFunc: func(ctx context.Context, streamID int64, typ string) error {
				calls = append(calls, methodCall{"1st", "OnStreamOpen()", []interface{}{ctx, streamID, typ}})
				return fmt.Errorf("1st: OnStreamOpen()")
			},
			OnStreamClosedFunc: func(streamID int64) {
				calls = append(calls, methodCall{"1st", "OnStreamClosed()", []interface{}{streamID}})
			},
			OnStreamRequestFunc: func(streamID int64, req *envoy.DiscoveryRequest) error {
				calls = append(calls, methodCall{"1st", "OnStreamRequest()", []interface{}{streamID, req}})
				return fmt.Errorf("1st: OnStreamRequest()")
			},
			OnStreamResponseFunc: func(streamID int64, req *envoy.DiscoveryRequest, resp *envoy.DiscoveryResponse) {
				calls = append(calls, methodCall{"1st", "OnStreamResponse()", []interface{}{streamID, req, resp}})
			},
		}
		second = CallbacksFuncs{
			OnStreamOpenFunc: func(ctx context.Context, streamID int64, typ string) error {
				calls = append(calls, methodCall{"2nd", "OnStreamOpen()", []interface{}{ctx, streamID, typ}})
				return fmt.Errorf("2nd: OnStreamOpen()")
			},
			OnStreamClosedFunc: func(streamID int64) {
				calls = append(calls, methodCall{"2nd", "OnStreamClosed()", []interface{}{streamID}})
			},
			OnStreamRequestFunc: func(streamID int64, req *envoy.DiscoveryRequest) error {
				calls = append(calls, methodCall{"2nd", "OnStreamRequest()", []interface{}{streamID, req}})
				return fmt.Errorf("2nd: OnStreamRequest()")
			},
			OnStreamResponseFunc: func(streamID int64, req *envoy.DiscoveryRequest, resp *envoy.DiscoveryResponse) {
				calls = append(calls, methodCall{"2nd", "OnStreamResponse()", []interface{}{streamID, req, resp}})
			},
		}
	})

	Describe("OnStreamOpen", func() {
		It("should be called sequentially and aggregate errors", func() {
			// given
			ctx := context.Background()
			streamID := int64(1)
			typ := "xDS"
			// setup
			chain := CallbacksChain{first, second}

			// when
			err := chain.OnStreamOpen(ctx, streamID, typ)

			// then
			Expect(calls).To(Equal([]methodCall{
				methodCall{"1st", "OnStreamOpen()", []interface{}{ctx, streamID, typ}},
				methodCall{"2nd", "OnStreamOpen()", []interface{}{ctx, streamID, typ}},
			}))
			// and
			Expect(err).To(MatchError("1st: OnStreamOpen(); 2nd: OnStreamOpen()"))
		})
	})
	Describe("OnStreamClose", func() {
		It("should be called in reverse order", func() {
			// given
			streamID := int64(1)
			// setup
			chain := CallbacksChain{first, second}

			// when
			chain.OnStreamClosed(streamID)

			// then
			Expect(calls).To(Equal([]methodCall{
				{"2nd", "OnStreamClosed()", []interface{}{streamID}},
				{"1st", "OnStreamClosed()", []interface{}{streamID}},
			}))
		})
	})
	Describe("OnStreamRequest", func() {
		It("should be called sequentially and aggregate errors", func() {
			// given
			streamID := int64(1)
			req := &envoy.DiscoveryRequest{}

			// setup
			chain := CallbacksChain{first, second}

			// when
			err := chain.OnStreamRequest(streamID, req)

			// then
			Expect(calls).To(Equal([]methodCall{
				{"1st", "OnStreamRequest()", []interface{}{streamID, req}},
				{"2nd", "OnStreamRequest()", []interface{}{streamID, req}},
			}))
			// and
			Expect(err).To(MatchError("1st: OnStreamRequest(); 2nd: OnStreamRequest()"))
		})
	})
	Describe("OnStreamResponse", func() {
		It("should be called in reverse order", func() {
			// given
			chain := CallbacksChain{first, second}
			streamID := int64(1)
			req := &envoy.DiscoveryRequest{}
			resp := &envoy.DiscoveryResponse{}

			// when
			chain.OnStreamResponse(streamID, req, resp)

			// then
			Expect(calls).To(Equal([]methodCall{
				{"2nd", "OnStreamResponse()", []interface{}{streamID, req, resp}},
				{"1st", "OnStreamResponse()", []interface{}{streamID, req, resp}},
			}))
		})
	})
})

var _ envoy_xds.Callbacks = CallbacksFuncs{}

type CallbacksFuncs struct {
	OnStreamOpenFunc   func(context.Context, int64, string) error
	OnStreamClosedFunc func(int64)

	OnStreamRequestFunc  func(int64, *envoy.DiscoveryRequest) error
	OnStreamResponseFunc func(int64, *envoy.DiscoveryRequest, *envoy.DiscoveryResponse)

	OnFetchRequestFunc  func(context.Context, *envoy.DiscoveryRequest) error
	OnFetchResponseFunc func(*envoy.DiscoveryRequest, *envoy.DiscoveryResponse)
}

func (f CallbacksFuncs) OnStreamOpen(ctx context.Context, streamID int64, typ string) error {
	if f.OnStreamOpenFunc != nil {
		return f.OnStreamOpenFunc(ctx, streamID, typ)
	}
	return nil
}
func (f CallbacksFuncs) OnStreamClosed(streamID int64) {
	if f.OnStreamClosedFunc != nil {
		f.OnStreamClosedFunc(streamID)
	}
}
func (f CallbacksFuncs) OnStreamRequest(streamID int64, req *envoy.DiscoveryRequest) error {
	if f.OnStreamRequestFunc != nil {
		return f.OnStreamRequestFunc(streamID, req)
	}
	return nil
}
func (f CallbacksFuncs) OnStreamResponse(streamID int64, req *envoy.DiscoveryRequest, resp *envoy.DiscoveryResponse) {
	if f.OnStreamResponseFunc != nil {
		f.OnStreamResponseFunc(streamID, req, resp)
	}
}
func (f CallbacksFuncs) OnFetchRequest(ctx context.Context, req *envoy.DiscoveryRequest) error {
	if f.OnFetchRequestFunc != nil {
		return f.OnFetchRequestFunc(ctx, req)
	}
	return nil
}
func (f CallbacksFuncs) OnFetchResponse(req *envoy.DiscoveryRequest, resp *envoy.DiscoveryResponse) {
	if f.OnFetchResponseFunc != nil {
		f.OnFetchResponseFunc(req, resp)
	}
}
