package xds

import "context"

type NoopCallbacks struct{}

func (n *NoopCallbacks) OnStreamOpen(context.Context, int64, string) error {
	return nil
}

func (n *NoopCallbacks) OnStreamClosed(int64) {
}

func (n *NoopCallbacks) OnStreamRequest(int64, DiscoveryRequest) error {
	return nil
}

func (n *NoopCallbacks) OnStreamResponse(int64, DiscoveryRequest, DiscoveryResponse) {
}

func (n *NoopCallbacks) OnDeltaStreamOpen(context.Context, int64, string) error {
	return nil
}

func (n *NoopCallbacks) OnDeltaStreamClosed(int64) {
}

func (n *NoopCallbacks) OnStreamDeltaRequest(int64, DeltaDiscoveryRequest) error {
	return nil
}

func (n *NoopCallbacks) OnStreamDeltaResponse(int64, DeltaDiscoveryRequest, DeltaDiscoveryResponse) {
}

var _ Callbacks = &NoopCallbacks{}
