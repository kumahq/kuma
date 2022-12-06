package intercp

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"time"

	"github.com/pkg/errors"
	"github.com/sethvargo/go-retry"

	"github.com/kumahq/kuma/api/system/v1alpha1"
	"github.com/kumahq/kuma/pkg/core"
	"github.com/kumahq/kuma/pkg/core/resources/manager"
	"github.com/kumahq/kuma/pkg/core/runtime"
	"github.com/kumahq/kuma/pkg/core/runtime/component"
	"github.com/kumahq/kuma/pkg/core/user"
	"github.com/kumahq/kuma/pkg/intercp/catalog"
	"github.com/kumahq/kuma/pkg/intercp/client"
	"github.com/kumahq/kuma/pkg/intercp/server"
	intercp_tls "github.com/kumahq/kuma/pkg/intercp/tls"
)

var log = core.Log.WithName("inter-cp")

func Setup(rt runtime.Runtime) error {
	cfg := rt.Config().InterCp
	defaults := &intercp_tls.DefaultsComponent{
		ResManager: rt.ResourceManager(),
		Log:        log.WithName("defaults"),
	}

	heartbeats := catalog.NewHeartbeats()
	c := catalog.NewConfigCatalog(rt.ResourceManager())

	instance := catalog.Instance{
		Id:          rt.GetInstanceId(),
		Address:     cfg.Catalog.InstanceAddress,
		InterCpPort: cfg.Server.Port,
	}

	ctx := user.Ctx(context.Background(), user.ControlPlane)
	registerComponent := component.ComponentFunc(func(stop <-chan struct{}) error {
		ca, serverCert, clientCert, err := generateCerts(ctx, rt.ReadOnlyResourceManager(), cfg.Catalog.InstanceAddress)
		if err != nil {
			return errors.Wrap(err, "could not generate certificates to start inter-cp server")
		}

		interCpServer, err := server.New(cfg.Server, rt.Metrics(), serverCert, *ca)
		if err != nil {
			return errors.Wrap(err, "could not start inter-cp server")
		}
		v1alpha1.RegisterInterCpPingServiceServer(interCpServer.GrpcServer(), catalog.NewServer(heartbeats, rt.LeaderInfo()))

		clientTLSConfig := client.TLSConfig{
			CaCert:     *ca,
			ClientCert: clientCert,
		}
		return rt.Add(
			interCpServer,
			catalog.NewHeartbeatComponent(c, instance, cfg.Catalog.HeartbeatInterval.Duration, func(serverURL string) (catalog.Client, error) {
				conn, err := client.New(serverURL, &clientTLSConfig)
				if err != nil {
					return nil, errors.Wrap(err, "could not create inter-cp client")
				}
				return catalog.NewGRPCClient(conn), nil
			}),
		)
	})

	return rt.Add(
		defaults,
		catalog.NewWriter(c, heartbeats, instance, cfg.Catalog.WriterInterval.Duration),
		registerComponent,
	)
}

func generateCerts(ctx context.Context, resManager manager.ReadOnlyResourceManager, instanceId string) (
	caCert *x509.Certificate, serverCert tls.Certificate, clientCert tls.Certificate, err error,
) {
	backoff := retry.WithMaxRetries(300, retry.NewConstant(1*time.Second))
	var ca tls.Certificate
	// we need to retry because the CA may not be created yet
	err = retry.Do(ctx, backoff, func(ctx context.Context) error {
		loadedCa, err := intercp_tls.LoadCA(ctx, resManager)
		if err != nil {
			return retry.RetryableError(err)
		}
		ca = loadedCa
		return nil
	})
	if err != nil {
		return
	}
	if len(ca.Certificate) != 1 {
		err = errors.New("there should be exactly one certificate")
		return
	}
	caCert, err = x509.ParseCertificate(ca.Certificate[0])
	if err != nil {
		return
	}
	serverCert, err = intercp_tls.GenerateServerCert(ca, instanceId)
	if err != nil {
		return
	}
	clientCert, err = intercp_tls.GenerateClientCert(ca, instanceId)
	if err != nil {
		return
	}
	return
}
