package stream

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"net/url"

	envoy_core "github.com/envoyproxy/go-control-plane/envoy/config/core/v3"
	envoy_discovery "github.com/envoyproxy/go-control-plane/envoy/service/discovery/v3"
	"github.com/pkg/errors"
	"google.golang.org/genproto/googleapis/rpc/status"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
	"google.golang.org/protobuf/types/known/structpb"

	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
	"github.com/kumahq/kuma/pkg/core/resources/model/rest"
	util_proto "github.com/kumahq/kuma/pkg/util/proto"
)

type Client struct {
	conn   *grpc.ClientConn
	client envoy_discovery.AggregatedDiscoveryServiceClient
}

type Stream struct {
	stream         envoy_discovery.AggregatedDiscoveryService_StreamAggregatedResourcesClient
	latestACKed    map[string]*envoy_discovery.DiscoveryResponse
	latestReceived map[string]*envoy_discovery.DiscoveryResponse
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
		// #nosec G402 -- it's acceptable as this is only to be used in testing
		dialOpts = append(dialOpts, grpc.WithTransportCredentials(credentials.NewTLS(&tls.Config{
			InsecureSkipVerify: true,
		})))
	default:
		return nil, errors.Errorf("unsupported scheme %q. Use one of %s", url.Scheme, []string{"grpc", "grpcs"})
	}
	conn, err := grpc.NewClient(url.Host, dialOpts...)
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
		latestACKed:    make(map[string]*envoy_discovery.DiscoveryResponse),
		latestReceived: make(map[string]*envoy_discovery.DiscoveryResponse),
	}, nil
}

func (c *Client) Close() error {
	return c.conn.Close()
}

func (s *Stream) Request(clientId string, typ string, dp rest.Resource) error {
	dpJSON, err := json.Marshal(dp)
	if err != nil {
		return err
	}
	version := &mesh_proto.Version{
		KumaDp: &mesh_proto.KumaDpVersion{
			Version:   "0.0.1",
			GitTag:    "v0.0.1",
			GitCommit: "91ce236824a9d875601679aa80c63783fb0e8725",
			BuildDate: "2019-08-07T11:26:06Z",
		},
		Envoy: &mesh_proto.EnvoyVersion{
			Version: "1.15.0",
			Build:   "hash/1.15.0/RELEASE",
		},
	}
	md := &structpb.Struct{
		Fields: map[string]*structpb.Value{
			"dataplane.resource": {Kind: &structpb.Value_StringValue{StringValue: string(dpJSON)}},
			"version": {
				Kind: &structpb.Value_StructValue{
					StructValue: util_proto.MustToStruct(version),
				},
			},
		},
	}
	return s.stream.Send(&envoy_discovery.DiscoveryRequest{
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
	err := s.stream.Send(&envoy_discovery.DiscoveryRequest{
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
	return s.stream.Send(&envoy_discovery.DiscoveryRequest{
		VersionInfo:   latestACKed.GetVersionInfo(),
		ResponseNonce: latestReceived.Nonce,
		ResourceNames: []string{},
		TypeUrl:       typ,
		ErrorDetail: &status.Status{
			Message: fmt.Sprintf("%s", err),
		},
	})
}

func (s *Stream) WaitForResources() (*envoy_discovery.DiscoveryResponse, error) {
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
