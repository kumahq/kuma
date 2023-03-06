package mux

import (
	"context"
	"io"
	"sync"

	envoy_sd "github.com/envoyproxy/go-control-plane/envoy/service/discovery/v3"
	"google.golang.org/grpc/metadata"
)

type kdsServerStream struct {
	ctx        context.Context
	sendBuffer chan *envoy_sd.DeltaDiscoveryResponse
	recvBuffer chan *envoy_sd.DeltaDiscoveryRequest

	// Protects the send-buffer against writing on a closed channel, this is needed as we don't control in which goroutine `Send` will be called.
	lock   sync.Mutex
	closed bool
}

func (k *kdsServerStream) Send(response *envoy_sd.DeltaDiscoveryResponse) error {
	k.lock.Lock()
	defer k.lock.Unlock()
	if k.closed {
		return io.EOF
	}
	k.sendBuffer <- response
	return nil
}

func (k *kdsServerStream) Recv() (*envoy_sd.DeltaDiscoveryRequest, error) {
	r, more := <-k.recvBuffer
	if !more {
		return nil, io.EOF
	}
	return r, nil
}

func (k *kdsServerStream) SetHeader(metadata.MD) error {
	panic("not implemented")
}

func (k *kdsServerStream) SendHeader(metadata.MD) error {
	panic("not implemented")
}

func (k *kdsServerStream) SetTrailer(metadata.MD) {
	panic("not implemented")
}

func (k *kdsServerStream) Context() context.Context {
	return k.ctx
}

func (k *kdsServerStream) SendMsg(m interface{}) error {
	panic("not implemented")
}

func (k *kdsServerStream) RecvMsg(m interface{}) error {
	panic("not implemented")
}

func (k *kdsServerStream) close() {
	k.lock.Lock()
	defer k.lock.Unlock()

	k.closed = true
	close(k.sendBuffer)
	close(k.recvBuffer)
}