package xds

import "context"

type NoopCallbacks struct{}

func (*NoopCallbacks) OnStreamOpen(context.Context, int64, string) error {
	return nil
}

func (*NoopCallbacks) OnStreamClosed(int64) {
}

func (*NoopCallbacks) OnStreamRequest(int64, DiscoveryRequest) error {
	return nil
}

func (*NoopCallbacks) OnStreamResponse(int64, DiscoveryRequest, DiscoveryResponse) {
}

func (*NoopCallbacks) OnDeltaStreamOpen(context.Context, int64, string) error {
	return nil
}

func (*NoopCallbacks) OnDeltaStreamClosed(int64) {
}

func (*NoopCallbacks) OnStreamDeltaRequest(int64, DeltaDiscoveryRequest) error {
	return nil
}

func (*NoopCallbacks) OnStreamDeltaResponse(int64, DeltaDiscoveryRequest, DeltaDiscoveryResponse) {
}

var _ Callbacks = &NoopCallbacks{}
