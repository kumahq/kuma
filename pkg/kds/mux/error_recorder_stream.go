package mux

import (
	"io"
	"sync"

	envoy_sd "github.com/envoyproxy/go-control-plane/envoy/service/discovery/v3"
	"github.com/envoyproxy/go-control-plane/pkg/server/stream/v3"
	"github.com/pkg/errors"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/proto"
)

// ErrorRecorderStream is a DeltaStream that records an error
// We need this because go-control-plane@v0.11.1/pkg/server/delta/v3/server.go:190 swallows an error on Recv()
type ErrorRecorderStream interface {
	stream.DeltaStream
	Err() error
}

// scalars only, holding the response would pin its payload for the stream's lifetime
type largestResponse struct {
	typeURL   string
	size      int
	resources int
}

type errorRecorderStream struct {
	stream.DeltaStream
	err     error
	largest largestResponse
	sync.Mutex
}

var _ stream.DeltaStream = &errorRecorderStream{}

func NewErrorRecorderStream(s stream.DeltaStream) ErrorRecorderStream {
	return &errorRecorderStream{
		DeltaStream: s,
	}
}

func (e *errorRecorderStream) Send(resp *envoy_sd.DeltaDiscoveryResponse) error {
	size := proto.Size(resp)
	e.Lock()
	if size > e.largest.size {
		e.largest = largestResponse{typeURL: resp.GetTypeUrl(), size: size, resources: len(resp.GetResources())}
	}
	e.Unlock()
	return e.describeOversized(e.DeltaStream.Send(resp))
}

func (e *errorRecorderStream) Recv() (*envoy_sd.DeltaDiscoveryRequest, error) {
	res, err := e.DeltaStream.Recv()
	if err != nil && err != io.EOF { // do not consider "end of stream" an error
		err = e.describeOversized(err)
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

func (e *errorRecorderStream) describeOversized(err error) error {
	if status.Code(err) != codes.ResourceExhausted {
		return err
	}
	e.Lock()
	defer e.Unlock()
	if e.largest.size == 0 {
		return err
	}
	return errors.Wrapf(err, "largest response sent on this stream: type=%s size=%d resources=%d",
		e.largest.typeURL, e.largest.size, e.largest.resources)
}
