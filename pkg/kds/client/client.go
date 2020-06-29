package client

import (
	"context"
	"crypto/tls"
	"net/url"

	"github.com/pkg/errors"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/metadata"

	mesh_proto "github.com/Kong/kuma/api/mesh/v1alpha1"
)

type KDSClient interface {
	StartStream(clientId string) (KDSStream, error)
	Close() error
}

type client struct {
	conn   *grpc.ClientConn
	client mesh_proto.KumaDiscoveryServiceClient
}

var _ KDSClient = &client{}

func New(serverURL string) (KDSClient, error) {
	u, err := url.Parse(serverURL)
	if err != nil {
		return nil, err
	}
	var dialOpts []grpc.DialOption
	switch u.Scheme {
	case "http":
		dialOpts = append(dialOpts, grpc.WithInsecure())
	case "https":
		dialOpts = append(dialOpts, grpc.WithTransportCredentials(credentials.NewTLS(&tls.Config{
			InsecureSkipVerify: true, // it's acceptable since we don't pass any secrets to the server
		})))
	default:
		return nil, errors.Errorf("unsupported scheme %q. Use one of %s", u.Scheme, []string{"grpc", "grpcs"})
	}
	conn, err := grpc.Dial(u.Host, dialOpts...)
	if err != nil {
		return nil, err
	}
	c := mesh_proto.NewKumaDiscoveryServiceClient(conn)
	return &client{
		conn:   conn,
		client: c,
	}, nil
}

func (c *client) StartStream(clientId string) (KDSStream, error) {
	ctx := metadata.NewOutgoingContext(context.Background(), metadata.MD{})
	stream, err := c.client.StreamKumaResources(ctx)
	if err != nil {
		return nil, err
	}
	return NewKDSStream(stream, clientId), nil
}

func (c *client) Close() error {
	return c.conn.Close()
}
