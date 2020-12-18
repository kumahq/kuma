package xds

import "context"

type NoopCallbacks struct {
}

func (n *NoopCallbacks) OnStreamOpen(context.Context, int64, string) error {
	return nil
}

func (n *NoopCallbacks) OnStreamClosed(int64) {
	return
}

func (n *NoopCallbacks) OnStreamRequest(int64, DiscoveryRequest) error {
	return nil
}

func (n *NoopCallbacks) OnStreamResponse(int64, DiscoveryRequest, DiscoveryResponse) {
	return
}

var _ Callbacks = &NoopCallbacks{}
