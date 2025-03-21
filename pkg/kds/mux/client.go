package mux

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"net/url"
	"os"
	"time"

	"github.com/go-logr/logr"
	"github.com/pkg/errors"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/encoding/gzip"
	"google.golang.org/grpc/keepalive"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"

	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
	config "github.com/kumahq/kuma/pkg/config/app/kuma-cp"
	"github.com/kumahq/kuma/pkg/config/multizone"
	"github.com/kumahq/kuma/pkg/core"
	"github.com/kumahq/kuma/pkg/core/runtime/component"
	"github.com/kumahq/kuma/pkg/kds"
	"github.com/kumahq/kuma/pkg/kds/service"
	"github.com/kumahq/kuma/pkg/metrics"
	"github.com/kumahq/kuma/pkg/version"
)

var muxClientLog = core.Log.WithName("kds-mux-client")

type client struct {
	globalToZoneCb      OnGlobalToZoneSyncStartedFunc
	zoneToGlobalCb      OnZoneToGlobalSyncStartedFunc
	globalURL           string
	clientID            string
	config              multizone.KdsClientConfig
	experimantalConfig  config.ExperimentalConfig
	metrics             metrics.Metrics
	ctx                 context.Context
	envoyAdminProcessor service.EnvoyAdminProcessor
}

func NewClient(ctx context.Context, globalURL string, clientID string, globalToZoneCb OnGlobalToZoneSyncStartedFunc, zoneToGlobalCb OnZoneToGlobalSyncStartedFunc, config multizone.KdsClientConfig, experimantalConfig config.ExperimentalConfig, metrics metrics.Metrics, envoyAdminProcessor service.EnvoyAdminProcessor) component.Component {
	return &client{
		ctx:                 ctx,
		globalToZoneCb:      globalToZoneCb,
		zoneToGlobalCb:      zoneToGlobalCb,
		globalURL:           globalURL,
		clientID:            clientID,
		config:              config,
		experimantalConfig:  experimantalConfig,
		metrics:             metrics,
		envoyAdminProcessor: envoyAdminProcessor,
	}
}

func (c *client) Start(stop <-chan struct{}) (errs error) {
	u, err := url.Parse(c.globalURL)
	if err != nil {
		return err
	}
	dialOpts := c.metrics.GRPCClientInterceptors()
	dialOpts = append(dialOpts, grpc.WithUserAgent(version.Build.UserAgent("kds")), grpc.WithDefaultCallOptions(
		grpc.UseCompressor(gzip.Name),
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
		tlsConfig, err := tlsConfig(c.config.RootCAFile, c.config.TlsSkipVerify)
		if err != nil {
			return errors.Wrap(err, "could not ")
		}
		dialOpts = append(dialOpts, grpc.WithTransportCredentials(credentials.NewTLS(tlsConfig)))
	default:
		return errors.Errorf("unsupported scheme %q. Use one of %s", u.Scheme, []string{"grpc", "grpcs"})
	}
	conn, err := grpc.NewClient(u.Host, dialOpts...)
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
		kds.FeaturesMetadataKey, kds.FeatureZonePingHealth,
		kds.FeaturesMetadataKey, kds.FeatureHashSuffix,
	))
	defer cancel()

	log := muxClientLog.WithValues("client-id", c.clientID)
	errorCh := make(chan error)

	go c.startHealthCheck(withKDSCtx, log, conn, errorCh)

	go c.startXDSConfigs(withKDSCtx, log, conn, errorCh)
	go c.startStats(withKDSCtx, log, conn, errorCh)
	go c.startClusters(withKDSCtx, log, conn, errorCh)
	go c.startGlobalToZoneSync(withKDSCtx, log, conn, errorCh)
	go c.startZoneToGlobalSync(withKDSCtx, log, conn, errorCh)

	select {
	case <-stop:
		cancel()
		return errs
	case err = <-errorCh:
		cancel()
		return err
	}
}

func (c *client) startGlobalToZoneSync(ctx context.Context, log logr.Logger, conn *grpc.ClientConn, errorCh chan error) {
	kdsClient := mesh_proto.NewKDSSyncServiceClient(conn)
	log = log.WithValues("rpc", "global-to-zone")
	log.Info("initializing Kuma Discovery Service (KDS) stream for global to zone sync of resources with delta xDS")
	stream, err := kdsClient.GlobalToZoneSync(ctx)
	if err != nil {
		errorCh <- err
		return
	}
	processingErrorsCh := make(chan error)
	c.globalToZoneCb.OnGlobalToZoneSyncStarted(stream, processingErrorsCh)
	c.handleProcessingErrors(stream, log, processingErrorsCh, errorCh)
}

func (c *client) startZoneToGlobalSync(ctx context.Context, log logr.Logger, conn *grpc.ClientConn, errorCh chan error) {
	kdsClient := mesh_proto.NewKDSSyncServiceClient(conn)
	log = log.WithValues("rpc", "zone-to-global")
	log.Info("initializing Kuma Discovery Service (KDS) stream for zone to global sync of resources with delta xDS")
	stream, err := kdsClient.ZoneToGlobalSync(ctx)
	if err != nil {
		errorCh <- err
		return
	}
	processingErrorsCh := make(chan error)
	c.zoneToGlobalCb.OnZoneToGlobalSyncStarted(stream, processingErrorsCh)
	c.handleProcessingErrors(stream, log, processingErrorsCh, errorCh)
}

func (c *client) startXDSConfigs(
	ctx context.Context,
	log logr.Logger,
	conn *grpc.ClientConn,
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
	c.handleProcessingErrors(stream, log, processingErrorsCh, errorCh)
}

func (c *client) startStats(
	ctx context.Context,
	log logr.Logger,
	conn *grpc.ClientConn,
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
	c.handleProcessingErrors(stream, log, processingErrorsCh, errorCh)
}

func (c *client) startClusters(
	ctx context.Context,
	log logr.Logger,
	conn *grpc.ClientConn,
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
	c.handleProcessingErrors(stream, log, processingErrorsCh, errorCh)
}

func (c *client) startHealthCheck(
	ctx context.Context,
	log logr.Logger,
	conn *grpc.ClientConn,
	errorCh chan error,
) {
	client := mesh_proto.NewGlobalKDSServiceClient(conn)
	log = log.WithValues("rpc", "healthcheck")
	log.Info("starting")

	prevInterval := 5 * time.Minute
	ticker := time.NewTicker(prevInterval)
	defer ticker.Stop()
	for {
		log.Info("sending health check")
		resp, err := client.HealthCheck(ctx, &mesh_proto.ZoneHealthCheckRequest{})
		if err != nil && !errors.Is(err, context.Canceled) {
			if status.Code(err) == codes.Unimplemented {
				log.Info("health check unimplemented in server, stopping")
				return
			}
			log.Error(err, "health check failed")
			errorCh <- errors.Wrap(err, "zone health check request failed")
		} else if interval := resp.Interval.AsDuration(); interval > 0 {
			if prevInterval != interval {
				prevInterval = interval
				log.Info("Global CP requested new healthcheck interval", "interval", interval)
			}
			ticker.Reset(interval)
		}

		select {
		case <-ticker.C:
			continue
		case <-ctx.Done():
			log.Info("stopping")
			return
		}
	}
}

func (c *client) handleProcessingErrors(
	stream grpc.ClientStream,
	log logr.Logger,
	processingErrorsCh chan error,
	errorCh chan error,
) {
	err := <-processingErrorsCh
	if status.Code(err) == codes.Unimplemented {
		log.Error(err, "rpc stream failed, because global CP does not implement this rpc. Upgrade remote CP.")
		// backwards compatibility. Do not rethrow error, so KDS multiplex can still operate.
		return
	}
	if errors.Is(err, context.Canceled) {
		log.Info("rpc stream shutting down")
		// Let's not propagate this error further as we've already cancelled the context
		err = nil
	} else {
		log.Error(err, "rpc stream failed prematurely, will restart in background")
	}
	if err := stream.CloseSend(); err != nil {
		log.Error(err, "CloseSend returned an error")
	}
	if err != nil {
		errorCh <- err
	}
}

func (c *client) NeedLeaderElection() bool {
	return true
}

func tlsConfig(rootCaFile string, skipVerify bool) (*tls.Config, error) {
	tlsConfig := &tls.Config{
		InsecureSkipVerify: skipVerify, // #nosec G402 -- we let the user decide if they want to ignore verification
		MinVersion:         tls.VersionTLS12,
	}
	if rootCaFile != "" {
		roots := x509.NewCertPool()
		caCert, err := os.ReadFile(rootCaFile)
		if err != nil {
			return nil, errors.Wrapf(err, "could not read certificate %s", rootCaFile)
		}
		ok := roots.AppendCertsFromPEM(caCert)
		if !ok {
			return nil, errors.New("failed to parse root certificate")
		}
		tlsConfig.RootCAs = roots
	}
	return tlsConfig, nil
}
