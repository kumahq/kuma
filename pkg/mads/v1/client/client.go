package client

import (
	"context"
	"crypto/tls"
	"fmt"
	"net/url"

	envoy_core "github.com/envoyproxy/go-control-plane/envoy/config/core/v3"
	envoy_sd "github.com/envoyproxy/go-control-plane/envoy/service/discovery/v3"
	"github.com/pkg/errors"
	"google.golang.org/genproto/googleapis/rpc/status"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"

	observability_v1 "github.com/kumahq/kuma/api/observability/v1"
	mads_v1 "github.com/kumahq/kuma/pkg/mads/v1"
	util_proto "github.com/kumahq/kuma/pkg/util/proto"
)

type Client struct {
	conn   *grpc.ClientConn
	client observability_v1.MonitoringAssignmentDiscoveryServiceClient
}

type Stream struct {
	stream         observability_v1.MonitoringAssignmentDiscoveryService_StreamMonitoringAssignmentsClient
	latestACKed    *envoy_sd.DiscoveryResponse
	latestReceived *envoy_sd.DiscoveryResponse
}

func New(serverURL string) (*Client, error) {
	url, err := url.Parse(serverURL)
	if err != nil {
		return nil, err
	}
	var dialOpts []grpc.DialOption
	switch url.Scheme {
	case "grpc":
		dialOpts = append(dialOpts, grpc.WithTransportCredentials(insecure.NewCredentials()))
	case "grpcs":
		dialOpts = append(dialOpts, grpc.WithTransportCredentials(credentials.NewTLS(&tls.Config{MinVersion: tls.VersionTLS12})))
	default:
		return nil, errors.Errorf("unsupported scheme %q. Use one of %s", url.Scheme, []string{"grpc", "grpcs"})
	}
	conn, err := grpc.NewClient(url.Host, dialOpts...)
	if err != nil {
		return nil, err
	}
	client := observability_v1.NewMonitoringAssignmentDiscoveryServiceClient(conn)
	return &Client{
		conn:   conn,
		client: client,
	}, nil
}

func (c *Client) StartStream() (*Stream, error) {
	ctx := metadata.NewOutgoingContext(context.Background(), metadata.MD{})
	stream, err := c.client.StreamMonitoringAssignments(ctx)
	if err != nil {
		return nil, err
	}
	return &Stream{
		stream: stream,
	}, nil
}

func (c *Client) Close() error {
	return c.conn.Close()
}

func (s *Stream) RequestAssignments(clientId string) error {
	return s.stream.Send(&envoy_sd.DiscoveryRequest{
		VersionInfo:   "",
		ResponseNonce: "",
		Node: &envoy_core.Node{
			Id: clientId,
		},
		ResourceNames: []string{},
		TypeUrl:       mads_v1.MonitoringAssignmentType,
	})
}

func (s *Stream) ACK() error {
	if s.latestReceived == nil {
		return nil
	}
	err := s.stream.Send(&envoy_sd.DiscoveryRequest{
		VersionInfo:   s.latestReceived.VersionInfo,
		ResponseNonce: s.latestReceived.Nonce,
		ResourceNames: []string{},
		TypeUrl:       mads_v1.MonitoringAssignmentType,
	})
	if err == nil {
		s.latestACKed = s.latestReceived
	}
	return err
}

func (s *Stream) NACK(err error) error {
	if s.latestReceived == nil {
		return nil
	}
	return s.stream.Send(&envoy_sd.DiscoveryRequest{
		VersionInfo:   s.latestACKed.GetVersionInfo(),
		ResponseNonce: s.latestReceived.Nonce,
		ResourceNames: []string{},
		TypeUrl:       mads_v1.MonitoringAssignmentType,
		ErrorDetail: &status.Status{
			Message: fmt.Sprintf("%s", err),
		},
	})
}

func (s *Stream) WaitForAssignments() ([]*observability_v1.MonitoringAssignment, error) {
	resp, err := s.stream.Recv()
	if err != nil {
		return nil, err
	}
	s.latestReceived = resp
	assignments := make([]*observability_v1.MonitoringAssignment, len(resp.Resources))
	for i := range resp.Resources {
		assignment := &observability_v1.MonitoringAssignment{}
		if err := util_proto.UnmarshalAnyTo(resp.Resources[i], assignment); err != nil {
			return nil, errors.Wrapf(err, "failed to unmarshal MADS response: %v", resp)
		}
		assignments[i] = assignment
	}
	return assignments, nil
}

func (s *Stream) Close() error {
	return s.stream.CloseSend()
}
