package mux

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"net/url"
	"os"

	"github.com/pkg/errors"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/keepalive"
	"google.golang.org/grpc/metadata"

	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
	"github.com/kumahq/kuma/pkg/config/multizone"
	"github.com/kumahq/kuma/pkg/core"
	"github.com/kumahq/kuma/pkg/core/runtime/component"
	"github.com/kumahq/kuma/pkg/metrics"
)

var (
	muxClientLog = core.Log.WithName("kds-mux-client")
)

type client struct {
	callbacks Callbacks
	globalURL string
	clientID  string
	config    multizone.KdsClientConfig
	metrics   metrics.Metrics
	ctx       context.Context
}

func NewClient(globalURL string, clientID string, callbacks Callbacks, config multizone.KdsClientConfig, metrics metrics.Metrics, ctx context.Context) component.Component {
	return &client{
		callbacks: callbacks,
		globalURL: globalURL,
		clientID:  clientID,
		config:    config,
		metrics:   metrics,
		ctx:       ctx,
	}
}

func (c *client) Start(stop <-chan struct{}) (errs error) {
	u, err := url.Parse(c.globalURL)
	if err != nil {
		return err
	}
	dialOpts := c.metrics.GRPCClientInterceptors()
	dialOpts = append(dialOpts, grpc.WithDefaultCallOptions(
		grpc.MaxCallSendMsgSize(int(c.config.MaxMsgSize)),
		grpc.MaxCallRecvMsgSize(int(c.config.MaxMsgSize))),
		grpc.WithKeepaliveParams(keepalive.ClientParameters{
			Time:                grpcKeepAliveTime,
			Timeout:             grpcKeepAliveTime,
			PermitWithoutStream: true,
		}),
	)
	switch u.Scheme {
	case "grpc":
		dialOpts = append(dialOpts, grpc.WithTransportCredentials(insecure.NewCredentials()))
	case "grpcs":
		tlsConfig, err := tlsConfig(c.config.RootCAFile)
		if err != nil {
			return errors.Wrap(err, "could not ")
		}
		dialOpts = append(dialOpts, grpc.WithTransportCredentials(credentials.NewTLS(tlsConfig)))
	default:
		return errors.Errorf("unsupported scheme %q. Use one of %s", u.Scheme, []string{"grpc", "grpcs"})
	}
	conn, err := grpc.Dial(u.Host, dialOpts...)
	if err != nil {
		return err
	}
	defer func() {
		if err := conn.Close(); err != nil {
			errs = errors.Wrapf(err, "failed to close a connection")
		}
	}()
	muxClient := mesh_proto.NewMultiplexServiceClient(conn)

	withKDSCtx, cancel := context.WithCancel(metadata.AppendToOutgoingContext(c.ctx,
		"client-id", c.clientID,
		KDSVersionHeaderKey, KDSVersionV3,
	))
	defer cancel()

	log := muxClientLog.WithValues("client-id", c.clientID)
	log.Info("initializing Kuma Discovery Service (KDS) stream for global-zone sync of resources")
	stream, err := muxClient.StreamMessage(withKDSCtx)
	if err != nil {
		return err
	}
	session := NewSession("global", stream)
	if err := c.callbacks.OnSessionStarted(session); err != nil {
		log.Error(err, "closing KDS stream after callback error")
		return err
	}
	select {
	case <-stop:
		log.Info("KDS stream stopped", "reason", err)
		if err := stream.CloseSend(); err != nil {
			log.Error(err, "CloseSend returned an error")
		}
		cancel() // In this case we cancel the context early
		err = <-session.Error()
	case err = <-session.Error():
		log.Error(err, "KDS stream failed prematurely, will restart in background")
		if err := stream.CloseSend(); err != nil {
			log.Error(err, "CloseSend returned an error")
		}
	}
	return err
}

func (c *client) NeedLeaderElection() bool {
	return true
}

func tlsConfig(rootCaFile string) (*tls.Config, error) {
	if rootCaFile == "" {
		return &tls.Config{
			InsecureSkipVerify: true,
		}, nil
	}
	roots := x509.NewCertPool()
	caCert, err := os.ReadFile(rootCaFile)
	if err != nil {
		return nil, errors.Wrapf(err, "could not read certificate %s", rootCaFile)
	}
	ok := roots.AppendCertsFromPEM(caCert)
	if !ok {
		return nil, errors.New("failed to parse root certificate")
	}
	return &tls.Config{RootCAs: roots}, nil
}
