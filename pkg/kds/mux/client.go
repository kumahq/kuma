package mux

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"net/url"
	"os"

	"github.com/go-logr/logr"
	"github.com/pkg/errors"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/keepalive"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"

	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
	config "github.com/kumahq/kuma/pkg/config/app/kuma-cp"
	"github.com/kumahq/kuma/pkg/config/multizone"
	"github.com/kumahq/kuma/pkg/core"
	"github.com/kumahq/kuma/pkg/core/resources/registry"
	"github.com/kumahq/kuma/pkg/core/runtime/component"
	"github.com/kumahq/kuma/pkg/kds/service"
	"github.com/kumahq/kuma/pkg/metrics"
)

var muxClientLog = core.Log.WithName("kds-mux-client")

type client struct {
	callbacks           Callbacks
	globalToZoneCb      OnGlobalToZoneSyncStartedFunc
	zoneToGlobalCb      OnZoneToGlobalSyncStartedFunc
	globalURL           string
	clientID            string
	config              multizone.KdsClientConfig
	experimantalConfig  config.ExperimentalConfig
	metrics             metrics.Metrics
	ctx                 context.Context
	envoyAdminProcessor service.EnvoyAdminProcessor
	additionalMetadata  []multizone.MetadataKeyValue
}

func NewClient(ctx context.Context, globalURL string, clientID string, callbacks Callbacks, globalToZoneCb OnGlobalToZoneSyncStartedFunc, zoneToGlobalCb OnZoneToGlobalSyncStartedFunc, config multizone.KdsClientConfig, experimantalConfig config.ExperimentalConfig, metrics metrics.Metrics, envoyAdminProcessor service.EnvoyAdminProcessor, additionalMetadata []multizone.MetadataKeyValue, ) component.Component {
	return &client{
		ctx:                 ctx,
		callbacks:           callbacks,
		globalToZoneCb:      globalToZoneCb,
		zoneToGlobalCb:      zoneToGlobalCb,
		globalURL:           globalURL,
		clientID:            clientID,
		config:              config,
		experimantalConfig:  experimantalConfig,
		metrics:             metrics,
		envoyAdminProcessor: envoyAdminProcessor,
		additionalMetadata: additionalMetadata,
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

	withKDSCtx, cancel := context.WithCancel(metadata.AppendToOutgoingContext(c.ctx,
		"client-id", c.clientID,
		KDSVersionHeaderKey, KDSVersionV3,
	))

	for _, meta := range c.additionalMetadata {
		withKDSCtx = metadata.AppendToOutgoingContext(withKDSCtx, meta.Key, meta.Value)
	}
	defer cancel()

	log := muxClientLog.WithValues("client-id", c.clientID)
	errorCh := make(chan error)

	go c.startXDSConfigs(withKDSCtx, log, conn, stop, errorCh)
	go c.startStats(withKDSCtx, log, conn, stop, errorCh)
	go c.startClusters(withKDSCtx, log, conn, stop, errorCh)
	if c.experimantalConfig.KDSDeltaEnabled {
		go c.startGlobalToZoneSync(withKDSCtx, log, conn, stop, errorCh)
		go c.startZoneToGlobalSync(withKDSCtx, log, conn, stop, errorCh)
	} else {
		go c.startKDSMultiplex(withKDSCtx, log, conn, stop, errorCh)
	}

	select {
	case <-stop:
		cancel()
		return errs
	case err = <-errorCh:
		return err
	}
}

func (c *client) startKDSMultiplex(ctx context.Context, log logr.Logger, conn *grpc.ClientConn, stop <-chan struct{}, errorCh chan error) {
	muxClient := mesh_proto.NewMultiplexServiceClient(conn)
	log.Info("initializing Kuma Discovery Service (KDS) stream for global-zone sync of resources")
	stream, err := muxClient.StreamMessage(ctx)
	if err != nil {
		errorCh <- err
		return
	}
	bufferSize := len(registry.Global().ObjectTypes())
	// nolint:contextcheck
	session := NewSession("global", stream, uint32(bufferSize), c.config.MsgSendTimeout.Duration)
	if err := c.callbacks.OnSessionStarted(session); err != nil {
		log.Error(err, "closing KDS stream after callback error")
		errorCh <- err
		return
	}
	select {
	case <-stop:
		log.Info("KDS stream stopped", "reason", err)
		if err := stream.CloseSend(); err != nil {
			log.Error(err, "CloseSend returned an error")
		}
		err = <-session.Error()
		errorCh <- err
	case err = <-session.Error():
		log.Error(err, "KDS stream failed prematurely, will restart in background")
		if err := stream.CloseSend(); err != nil {
			log.Error(err, "CloseSend returned an error")
		}
		errorCh <- err
		return
	}
}

func (c *client) startGlobalToZoneSync(ctx context.Context, log logr.Logger, conn *grpc.ClientConn, stop <-chan struct{}, errorCh chan error) {
	kdsClient := mesh_proto.NewKDSSyncServiceClient(conn)
	log.Info("initializing Kuma Discovery Service (KDS) stream for global to zone sync of resources with delta xDS")
	stream, err := kdsClient.GlobalToZoneSync(ctx)
	if err != nil {
		errorCh <- err
		return
	}
	if err := c.globalToZoneCb.OnGlobalToZoneSyncStarted(stream); err != nil {
		errorCh <- errors.Wrap(err, "closing Global to Zone Sync stream after callback error")
		return
	}
	<-stop
	log.Info("Global to Zone Sync rpc stream stopped")
	if err := stream.CloseSend(); err != nil {
		errorCh <- errors.Wrap(err, "CloseSend returned an error")
	}
}

func (c *client) startZoneToGlobalSync(ctx context.Context, log logr.Logger, conn *grpc.ClientConn, stop <-chan struct{}, errorCh chan error) {
	kdsClient := mesh_proto.NewKDSSyncServiceClient(conn)
	log.Info("initializing Kuma Discovery Service (KDS) stream for zone to global sync of resources with delta xDS")
	stream, err := kdsClient.ZoneToGlobalSync(ctx)
	if err != nil {
		errorCh <- err
		return
	}
	if err := c.zoneToGlobalCb.OnZoneToGlobalSyncStarted(stream); err != nil {
		errorCh <- errors.Wrap(err, "closing Zone to Global Sync stream after callback error")
		return
	}
	<-stop
	log.Info("Zone to Global Sync rpc stream stopped")
	if err := stream.CloseSend(); err != nil {
		errorCh <- errors.Wrap(err, "CloseSend returned an error")
	}
}

func (c *client) startXDSConfigs(
	ctx context.Context,
	log logr.Logger,
	conn *grpc.ClientConn,
	stop <-chan struct{},
	errorCh chan error,
) {
	client := mesh_proto.NewGlobalKDSServiceClient(conn)
	log = log.WithValues("rpc", "XDS Configs")
	log.Info("initializing rpc stream for executing config dump on data plane proxies")
	stream, err := client.StreamXDSConfigs(ctx)
	if err != nil {
		errorCh <- err
		return
	}

	processingErrorsCh := make(chan error)
	go c.envoyAdminProcessor.StartProcessingXDSConfigs(stream, processingErrorsCh)
	c.handleProcessingErrors(stream, log, stop, processingErrorsCh, errorCh)
}

func (c *client) startStats(
	ctx context.Context,
	log logr.Logger,
	conn *grpc.ClientConn,
	stop <-chan struct{},
	errorCh chan error,
) {
	client := mesh_proto.NewGlobalKDSServiceClient(conn)
	log = log.WithValues("rpc", "stats")
	log.Info("initializing rpc stream for executing stats on data plane proxies")
	stream, err := client.StreamStats(ctx)
	if err != nil {
		errorCh <- err
		return
	}

	processingErrorsCh := make(chan error)
	go c.envoyAdminProcessor.StartProcessingStats(stream, processingErrorsCh)
	c.handleProcessingErrors(stream, log, stop, processingErrorsCh, errorCh)
}

func (c *client) startClusters(
	ctx context.Context,
	log logr.Logger,
	conn *grpc.ClientConn,
	stop <-chan struct{},
	errorCh chan error,
) {
	client := mesh_proto.NewGlobalKDSServiceClient(conn)
	log = log.WithValues("rpc", "clusters")
	log.Info("initializing rpc stream for executing clusters on data plane proxies")
	stream, err := client.StreamClusters(ctx)
	if err != nil {
		errorCh <- err
		return
	}

	processingErrorsCh := make(chan error)
	go c.envoyAdminProcessor.StartProcessingClusters(stream, processingErrorsCh)
	c.handleProcessingErrors(stream, log, stop, processingErrorsCh, errorCh)
}

func (c *client) handleProcessingErrors(
	stream grpc.ClientStream,
	log logr.Logger,
	stop <-chan struct{},
	processingErrorsCh chan error,
	errorCh chan error,
) {
	select {
	case <-stop:
		log.Info("Envoy Admin rpc stream stopped")
		if err := stream.CloseSend(); err != nil {
			log.Error(err, "CloseSend returned an error")
		}
	case err := <-processingErrorsCh:
		if status.Code(err) == codes.Unimplemented {
			log.Error(err, "Envoy Admin rpc stream failed, because Global CP does not implement this rpc. Upgrade Global CP.")
			// backwards compatibility. Do not rethrow error, so KDS multiplex can still operate.
			return
		}
		log.Error(err, "Envoy Admin rpc stream failed prematurely, will restart in background")
		if err := stream.CloseSend(); err != nil {
			log.Error(err, "CloseSend returned an error")
		}
		errorCh <- err
		return
	}
}

func (c *client) NeedLeaderElection() bool {
	return true
}

func tlsConfig(rootCaFile string) (*tls.Config, error) {
	if rootCaFile == "" {
		// #nosec G402 -- we allow this when not specifying a CA
		return &tls.Config{
			InsecureSkipVerify: true,
			MinVersion:         tls.VersionTLS12,
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
	return &tls.Config{RootCAs: roots, MinVersion: tls.VersionTLS12}, nil
}
