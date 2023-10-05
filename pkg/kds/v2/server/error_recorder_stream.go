package server

import (
	"io"
	"sync"

	envoy_sd "github.com/envoyproxy/go-control-plane/envoy/service/discovery/v3"
	"github.com/envoyproxy/go-control-plane/pkg/server/stream/v3"
)

// ErrorRecorderStream is a DeltaStream that records an error
// We need this because go-control-plane@v0.11.1/pkg/server/delta/v3/server.go:190 swallows an error on Recv()
type ErrorRecorderStream interface {
	stream.DeltaStream
	Err() error
}

type errorRecorderStream struct {
	stream.DeltaStream
	err error
	sync.Mutex
}

var _ stream.DeltaStream = &errorRecorderStream{}

func NewErrorRecorderStream(s stream.DeltaStream) ErrorRecorderStream {
	return &errorRecorderStream{
		DeltaStream: s,
	}
}

func (e *errorRecorderStream) Recv() (*envoy_sd.DeltaDiscoveryRequest, error) {
	res, err := e.DeltaStream.Recv()
	if err != nil && err != io.EOF { // do not consider "end of stream" an error
		e.Lock()
		e.err = err
		e.Unlock()
	}
	return res, err
}

func (e *errorRecorderStream) Err() error {
	e.Lock()
	defer e.Unlock()
	return e.err
}
