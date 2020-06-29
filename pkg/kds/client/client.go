package client

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"io/ioutil"
	"net/url"

	"github.com/pkg/errors"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/metadata"

	mesh_proto "github.com/Kong/kuma/api/mesh/v1alpha1"
	kds_config "github.com/Kong/kuma/pkg/config/kds"
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

func New(serverURL string, config *kds_config.KdsClientConfig) (KDSClient, error) {
	u, err := url.Parse(serverURL)
	if err != nil {
		return nil, err
	}
	var dialOpts []grpc.DialOption
	switch u.Scheme {
	case "grpc":
		dialOpts = append(dialOpts, grpc.WithInsecure())
	case "grpcs":
		tlsConfig, err := tlsConfig(config.RootCAFile)
		if err != nil {
			return nil, errors.Wrap(err, "could not ")
		}
		dialOpts = append(dialOpts, grpc.WithTransportCredentials(credentials.NewTLS(tlsConfig)))
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

func tlsConfig(rootCaFile string) (*tls.Config, error) {
	if rootCaFile == "" {
		return &tls.Config{
			InsecureSkipVerify: true,
		}, nil
	}
	roots := x509.NewCertPool()
	caCert, err := ioutil.ReadFile(rootCaFile)
	if err != nil {
		return nil, errors.Wrapf(err, "could not read certificate %s", rootCaFile)
	}
	ok := roots.AppendCertsFromPEM(caCert)
	if !ok {
		return nil, errors.New("failed to parse root certificate")
	}
	return &tls.Config{RootCAs: roots}, nil
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
