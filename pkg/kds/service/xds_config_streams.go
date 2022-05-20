package service

import (
	"sync"

	"github.com/pkg/errors"

	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
)

type XDSConfigStreams interface {
	Send(zone string, req *mesh_proto.XDSConfigRequest) error
	WatchResponse(zone string, reqID string, resp chan *mesh_proto.XDSConfigResponse) error
	DeleteWatch(zone string, reqID string)

	ZoneConnected(zone string, stream mesh_proto.GlobalKDSService_StreamXDSConfigsServer)
	ZoneDisconnected(zone string)
	ResponseReceived(zone string, resp *mesh_proto.XDSConfigResponse) error
}

type xdsConfigStreams struct {
	streamForZone map[string]*xdsConfigStream
	sync.Mutex    // protects streamForZone
}

func (x *xdsConfigStreams) ResponseReceived(zone string, resp *mesh_proto.XDSConfigResponse) error {
	stream, err := x.zoneStream(zone)
	if err != nil {
		return err
	}
	stream.Lock()
	ch, ok := stream.watchForRequestId[resp.RequestId]
	stream.Unlock()
	if !ok {
		return errors.Errorf("callback for request Id %s not found", resp.RequestId)
	}
	ch <- resp
	return nil
}

func NewXdsConfigStreams() XDSConfigStreams {
	return &xdsConfigStreams{
		streamForZone: map[string]*xdsConfigStream{},
	}
}

func (x *xdsConfigStreams) ZoneConnected(zone string, stream mesh_proto.GlobalKDSService_StreamXDSConfigsServer) {
	x.Lock()
	defer x.Unlock()
	x.streamForZone[zone] = &xdsConfigStream{
		stream:            stream,
		watchForRequestId: map[string]chan *mesh_proto.XDSConfigResponse{},
	}
}

func (x *xdsConfigStreams) zoneStream(zone string) (*xdsConfigStream, error) {
	x.Lock()
	defer x.Unlock()
	stream, ok := x.streamForZone[zone]
	if !ok {
		return nil, errors.Errorf("zone %s is not connected", zone)
	}
	return stream, nil
}

func (x *xdsConfigStreams) ZoneDisconnected(zone string) {
	x.Lock()
	defer x.Unlock()
	delete(x.streamForZone, zone)
}

type xdsConfigStream struct {
	stream            mesh_proto.GlobalKDSService_StreamXDSConfigsServer
	watchForRequestId map[string]chan *mesh_proto.XDSConfigResponse
	sync.Mutex        // protects watchForRequestId
}

func (x *xdsConfigStreams) Send(zone string, req *mesh_proto.XDSConfigRequest) error {
	stream, err := x.zoneStream(zone)
	if err != nil {
		return err
	}
	return stream.stream.Send(req)
}

func (x *xdsConfigStreams) WatchResponse(zone string, reqID string, resp chan *mesh_proto.XDSConfigResponse) error {
	stream, err := x.zoneStream(zone)
	if err != nil {
		return err
	}
	stream.Lock()
	defer stream.Unlock()
	stream.watchForRequestId[reqID] = resp
	return nil
}

func (x *xdsConfigStreams) DeleteWatch(zone string, reqID string) {
	stream, err := x.zoneStream(zone)
	if err != nil {
		return // zone was already deleted
	}
	stream.Lock()
	defer stream.Unlock()
	delete(stream.watchForRequestId, reqID)
}
