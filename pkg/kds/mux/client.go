package mux

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	std_errors "errors"
	"net/url"
	"os"
	"time"

	"github.com/envoyproxy/go-control-plane/pkg/server/delta/v3"
	"github.com/go-logr/logr"
	"github.com/pkg/errors"
	"golang.org/x/sync/errgroup"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/encoding/gzip"
	"google.golang.org/grpc/keepalive"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"

	mesh_proto "github.com/kumahq/kuma/v2/api/mesh/v1alpha1"
	"github.com/kumahq/kuma/v2/pkg/config"
	config_kumacp "github.com/kumahq/kuma/v2/pkg/config/app/kuma-cp"
	"github.com/kumahq/kuma/v2/pkg/config/core/resources/store"
	"github.com/kumahq/kuma/v2/pkg/config/multizone"
	"github.com/kumahq/kuma/v2/pkg/core"
	core_model "github.com/kumahq/kuma/v2/pkg/core/resources/model"
	core_runtime "github.com/kumahq/kuma/v2/pkg/core/runtime"
	"github.com/kumahq/kuma/v2/pkg/core/runtime/component"
	"github.com/kumahq/kuma/v2/pkg/kds"
	"github.com/kumahq/kuma/v2/pkg/kds/service"
	kds_client_v2 "github.com/kumahq/kuma/v2/pkg/kds/v2/client"
	kds_server_v2 "github.com/kumahq/kuma/v2/pkg/kds/v2/server"
	kds_sync_store "github.com/kumahq/kuma/v2/pkg/kds/v2/store"
	"github.com/kumahq/kuma/v2/pkg/metrics"
	resources_k8s "github.com/kumahq/kuma/v2/pkg/plugins/resources/k8s"
	"github.com/kumahq/kuma/v2/pkg/util/pointer"
	"github.com/kumahq/kuma/v2/pkg/version"
)

var muxClientLog = core.Log.WithName("kds-mux-client")

type client struct {
	globalURL           string
	clientID            string
	config              multizone.KdsClientConfig
	experimantalConfig  config_kumacp.ExperimentalConfig
	metrics             metrics.Metrics
	ctx                 context.Context
	envoyAdminProcessor service.EnvoyAdminProcessor
	deltaServer         delta.Server
	typesSentByGlobal   []core_model.ResourceType
	rt                  core_runtime.Runtime
	resourceSyncer      kds_sync_store.ResourceSyncer
}

func NewClient(
	ctx context.Context,
	globalURL string,
	clientID string,
	config multizone.KdsClientConfig,
	experimantalConfig config_kumacp.ExperimentalConfig,
	metrics metrics.Metrics,
	envoyAdminProcessor service.EnvoyAdminProcessor,
	resourceSyncer kds_sync_store.ResourceSyncer,
	rt core_runtime.Runtime,
	deltaServer delta.Server,
) component.Component {
	return &client{
		ctx:                 ctx,
		globalURL:           globalURL,
		clientID:            clientID,
		config:              config,
		experimantalConfig:  experimantalConfig,
		metrics:             metrics,
		envoyAdminProcessor: envoyAdminProcessor,
		resourceSyncer:      resourceSyncer,
		rt:                  rt,
		deltaServer:         deltaServer,
		typesSentByGlobal:   rt.KDSContext().TypesSentByGlobal,
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

	cfgJson, err := config.ConfigForDisplay(pointer.To(c.rt.Config()))
	if err != nil {
		errorCh <- errors.Wrap(err, "could not marshall config to json")
		return
	}

	group, innerCtx := errgroup.WithContext(ctx)
	stream, err := kdsClient.GlobalToZoneSync(innerCtx)
	if err != nil {
		errorCh <- err
		return
	}
	defer func() {
		if err := stream.CloseSend(); err != nil {
			log.Error(err, "CloseSend returned an error")
		}
	}()

	syncClient := kds_client_v2.NewKDSSyncClient(
		log,
		c.typesSentByGlobal,
		kds_client_v2.NewDeltaKDSStream(stream, c.clientID, c.rt.GetInstanceId(), cfgJson),
		kds_sync_store.ZoneSyncCallback(
			stream.Context(),
			c.resourceSyncer,
			c.rt.Config().Store.Type == store.KubernetesStore,
			resources_k8s.NewSimpleKubeFactory(),
			c.rt.Config().Store.Kubernetes.SystemNamespace,
		),
		c.rt.Config().Multizone.Zone.KDS.ResponseBackoff.Duration,
	)

	if err := syncClient.Receive(innerCtx, group); err != nil && !std_errors.Is(err, context.Canceled) {
		errorCh <- errors.Wrap(err, "GlobalToZoneSyncClient finished with an error")
		return
	}

	log.V(1).Info("GlobalToZoneSyncClient finished gracefully")
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
	defer func() {
		if err := stream.CloseSend(); err != nil {
			log.Error(err, "CloseSend returned an error")
		}
	}()

	log.Info("ZoneToGlobalSync new session created")
	errorStream := NewErrorRecorderStream(kds_server_v2.NewServerStream(stream))
	err = c.deltaServer.DeltaStreamHandler(errorStream, "")
	if err == nil {
		err = errorStream.Err()
	}

	if err != nil && !errors.Is(err, context.Canceled) {
		errorCh <- errors.Wrap(err, "ZoneToGlobalSync finished with an error")
		return
	}

	log.V(1).Info("ZoneToGlobalSync finished gracefully")
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
