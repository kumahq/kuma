package client

import (
	"context"
	"crypto/tls"
	"fmt"
	"net/url"

	"github.com/golang/protobuf/ptypes"
	"github.com/pkg/errors"
	"google.golang.org/genproto/googleapis/rpc/status"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/metadata"

	envoy "github.com/envoyproxy/go-control-plane/envoy/api/v2"
	envoy_core "github.com/envoyproxy/go-control-plane/envoy/api/v2/core"

	observability_proto "github.com/Kong/kuma/api/observability/v1alpha1"

	"github.com/Kong/kuma/pkg/mads"
)

type Client struct {
	conn   *grpc.ClientConn
	client observability_proto.MonitoringAssignmentDiscoveryServiceClient
}

type Stream struct {
	stream         observability_proto.MonitoringAssignmentDiscoveryService_StreamMonitoringAssignmentsClient
	latestACKed    *envoy.DiscoveryResponse
	latestReceived *envoy.DiscoveryResponse
}

func New(serverURL string) (*Client, error) {
	url, err := url.Parse(serverURL)
	if err != nil {
		return nil, err
	}
	var dialOpts []grpc.DialOption
	switch url.Scheme {
	case "grpc":
		dialOpts = append(dialOpts, grpc.WithInsecure())
	case "grpcs":
		dialOpts = append(dialOpts, grpc.WithTransportCredentials(credentials.NewTLS(&tls.Config{
			InsecureSkipVerify: true, // it's acceptable since we don't pass any secrets to the server
		})))
	default:
		return nil, errors.Errorf("unsupported scheme %q. Use one of %s", url.Scheme, []string{"grpc", "grpcs"})
	}
	conn, err := grpc.Dial(url.Host, dialOpts...)
	if err != nil {
		return nil, err
	}
	client := observability_proto.NewMonitoringAssignmentDiscoveryServiceClient(conn)
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
	return s.stream.Send(&envoy.DiscoveryRequest{
		VersionInfo:   "",
		ResponseNonce: "",
		Node: &envoy_core.Node{
			Id: clientId,
		},
		ResourceNames: []string{},
		TypeUrl:       mads.MonitoringAssignmentType,
	})
}

func (s *Stream) ACK() error {
	if s.latestReceived == nil {
		return nil
	}
	err := s.stream.Send(&envoy.DiscoveryRequest{
		VersionInfo:   s.latestReceived.VersionInfo,
		ResponseNonce: s.latestReceived.Nonce,
		ResourceNames: []string{},
		TypeUrl:       mads.MonitoringAssignmentType,
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
	return s.stream.Send(&envoy.DiscoveryRequest{
		VersionInfo:   s.latestACKed.GetVersionInfo(),
		ResponseNonce: s.latestReceived.Nonce,
		ResourceNames: []string{},
		TypeUrl:       mads.MonitoringAssignmentType,
		ErrorDetail: &status.Status{
			Message: fmt.Sprintf("%s", err),
		},
	})
}

func (s *Stream) WaitForAssignments() ([]*observability_proto.MonitoringAssignment, error) {
	resp, err := s.stream.Recv()
	if err != nil {
		return nil, err
	}
	s.latestReceived = resp
	assignments := make([]*observability_proto.MonitoringAssignment, len(resp.Resources))
	for i := range resp.Resources {
		assignment := &observability_proto.MonitoringAssignment{}
		if err := ptypes.UnmarshalAny(resp.Resources[i], assignment); err != nil {
			return nil, errors.Wrapf(err, "failed to unmarshal MADS response: %v", resp)
		}
		assignments[i] = assignment
	}
	return assignments, nil
}

func (s *Stream) Close() error {
	return s.stream.CloseSend()
}
