package stream

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"net/url"

	envoy "github.com/envoyproxy/go-control-plane/envoy/api/v2"
	envoy_core "github.com/envoyproxy/go-control-plane/envoy/api/v2/core"
	envoy_discovery "github.com/envoyproxy/go-control-plane/envoy/service/discovery/v2"
	"github.com/pkg/errors"
	"google.golang.org/genproto/googleapis/rpc/status"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
	"google.golang.org/protobuf/types/known/structpb"

	"github.com/kumahq/kuma/pkg/core/resources/model/rest"
)

type Client struct {
	conn   *grpc.ClientConn
	client envoy_discovery.AggregatedDiscoveryServiceClient
}

type Stream struct {
	stream         envoy_discovery.AggregatedDiscoveryService_StreamAggregatedResourcesClient
	latestACKed    map[string]*envoy.DiscoveryResponse
	latestReceived map[string]*envoy.DiscoveryResponse
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
	client := envoy_discovery.NewAggregatedDiscoveryServiceClient(conn)
	return &Client{
		conn:   conn,
		client: client,
	}, nil
}

func (c *Client) StartStream() (*Stream, error) {
	ctx := metadata.NewOutgoingContext(context.Background(), metadata.MD{})
	stream, err := c.client.StreamAggregatedResources(ctx)
	if err != nil {
		return nil, err
	}
	return &Stream{
		stream:         stream,
		latestACKed:    make(map[string]*envoy.DiscoveryResponse),
		latestReceived: make(map[string]*envoy.DiscoveryResponse),
	}, nil
}

func (c *Client) Close() error {
	return c.conn.Close()
}

func (s *Stream) Request(clientId string, typ string, dp *rest.Resource) error {
	dpJSON, err := json.Marshal(dp)
	if err != nil {
		return err
	}
	md := &structpb.Struct{
		Fields: map[string]*structpb.Value{
			"dataplane.resource": {Kind: &structpb.Value_StringValue{StringValue: string(dpJSON)}},
		},
	}
	return s.stream.Send(&envoy.DiscoveryRequest{
		VersionInfo:   "",
		ResponseNonce: "",
		Node: &envoy_core.Node{
			Id:       clientId,
			Metadata: md,
		},
		ResourceNames: []string{},
		TypeUrl:       typ,
	})
}

func (s *Stream) ACK(typ string) error {
	latestReceived := s.latestReceived[typ]
	if latestReceived == nil {
		return nil
	}
	err := s.stream.Send(&envoy.DiscoveryRequest{
		VersionInfo:   latestReceived.VersionInfo,
		ResponseNonce: latestReceived.Nonce,
		ResourceNames: []string{},
		TypeUrl:       typ,
	})
	if err == nil {
		s.latestACKed = s.latestReceived
	}
	return err
}

func (s *Stream) NACK(typ string, err error) error {
	latestReceived := s.latestReceived[typ]
	if latestReceived == nil {
		return nil
	}
	latestACKed := s.latestACKed[typ]
	return s.stream.Send(&envoy.DiscoveryRequest{
		VersionInfo:   latestACKed.GetVersionInfo(),
		ResponseNonce: latestReceived.Nonce,
		ResourceNames: []string{},
		TypeUrl:       typ,
		ErrorDetail: &status.Status{
			Message: fmt.Sprintf("%s", err),
		},
	})
}

func (s *Stream) WaitForResources() (*envoy.DiscoveryResponse, error) {
	resp, err := s.stream.Recv()
	if err != nil {
		return nil, err
	}
	s.latestReceived[resp.TypeUrl] = resp
	return resp, nil
}

func (s *Stream) Close() error {
	return s.stream.CloseSend()
}
