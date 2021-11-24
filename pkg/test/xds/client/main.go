package main

import (
	"fmt"
	"math/rand"
	"os"
	"time"

	envoy_resource "github.com/envoyproxy/go-control-plane/pkg/resource/v3"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"go.uber.org/multierr"

	"github.com/kumahq/kuma/api/mesh/v1alpha1"
	"github.com/kumahq/kuma/pkg/core"
	"github.com/kumahq/kuma/pkg/core/resources/model/rest"
	kuma_log "github.com/kumahq/kuma/pkg/log"
	"github.com/kumahq/kuma/pkg/test/xds/client/stream"
)

func newRootCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "kuma-xds-client",
		Short: "Kuma xDS client",
		Long:  `Kuma xDS client.`,
		PersistentPreRun: func(_ *cobra.Command, _ []string) {
			core.SetLogger(core.NewLogger(kuma_log.DebugLevel))
		},
	}
	// sub-commands
	cmd.AddCommand(newRunCmd())
	return cmd
}

func newRunCmd() *cobra.Command {
	log := core.Log.WithName("kuma-xds-client").WithName("run")
	args := struct {
		xdsServerAddress string
		total            int32
		rampUpPeriod     time.Duration
	}{
		xdsServerAddress: "grpcs://localhost:5678",
		total:            100,
		rampUpPeriod:     30 * time.Second,
	}
	cmd := &cobra.Command{
		Use:   "run",
		Short: "Start xDS client(s) that simulate Envoy",
		Long:  `Start xDS client(s) that simulate Envoy.`,
		RunE: func(cmd *cobra.Command, _ []string) error {
			log.Info("going to start xDS clients (Envoy simulators)", "total", args.total)
			errCh := make(chan error, 1)
			for i := 0; i < int(args.total); i++ {
				id := fmt.Sprintf("default.dataplane-%d", i)
				nodeLog := log.WithName("envoy-simulator").WithValues("idx", i, "ID", id)
				nodeLog.Info("creating an xDS client ...")

				go func(i int) {
					dp := &rest.Resource{
						Meta: rest.ResourceMeta{Mesh: "default", Name: fmt.Sprintf("dataplane-%d", i), Type: "Dataplane"},
						Spec: &v1alpha1.Dataplane{
							Networking: &v1alpha1.Dataplane_Networking{
								Address: "127.0.0.1",
								Inbound: []*v1alpha1.Dataplane_Networking_Inbound{
									{Port: uint32(8080), Tags: map[string]string{"kuma.io/service": "service-0"}},
									{Port: uint32(8081), Tags: map[string]string{"kuma.io/service": "service-1"}},
									{Port: uint32(8082), Tags: map[string]string{"kuma.io/service": "service-2"}},
									{Port: uint32(8083), Tags: map[string]string{"kuma.io/service": "service-3"}},
									{Port: uint32(8084), Tags: map[string]string{"kuma.io/service": "service-4"}},
									{Port: uint32(8085), Tags: map[string]string{"kuma.io/service": "service-5"}},
									{Port: uint32(8086), Tags: map[string]string{"kuma.io/service": "service-6"}},
									{Port: uint32(8087), Tags: map[string]string{"kuma.io/service": "service-7"}},
									{Port: uint32(8088), Tags: map[string]string{"kuma.io/service": "service-8"}},
									{Port: uint32(8089), Tags: map[string]string{"kuma.io/service": "service-9"}},
								},
								Outbound: []*v1alpha1.Dataplane_Networking_Outbound{
									{Address: "127.0.0.1", Port: 11000, Tags: map[string]string{"kuma.io/service": "service-0"}},
									{Address: "127.0.0.1", Port: 11001, Tags: map[string]string{"kuma.io/service": "service-1"}},
									{Address: "127.0.0.1", Port: 11002, Tags: map[string]string{"kuma.io/service": "service-2"}},
									{Address: "127.0.0.1", Port: 11003, Tags: map[string]string{"kuma.io/service": "service-3"}},
									{Address: "127.0.0.1", Port: 11004, Tags: map[string]string{"kuma.io/service": "service-4"}},
									{Address: "127.0.0.1", Port: 11005, Tags: map[string]string{"kuma.io/service": "service-5"}},
									{Address: "127.0.0.1", Port: 11006, Tags: map[string]string{"kuma.io/service": "service-6"}},
									{Address: "127.0.0.1", Port: 11007, Tags: map[string]string{"kuma.io/service": "service-7"}},
									{Address: "127.0.0.1", Port: 11008, Tags: map[string]string{"kuma.io/service": "service-8"}},
									{Address: "127.0.0.1", Port: 11009, Tags: map[string]string{"kuma.io/service": "service-9"}},
								},
							},
						},
					}

					// add some jitter
					delay := time.Duration(int64(float64(args.rampUpPeriod.Nanoseconds()) * rand.Float64()))
					// wait
					<-time.After(delay)
					// proceed

					errCh <- func() (errs error) {
						client, err := stream.New(args.xdsServerAddress)
						if err != nil {
							return errors.Wrap(err, "failed to connect to xDS server")
						}
						defer func() {
							nodeLog.Info("closing a connection ...")
							if err := client.Close(); err != nil {
								errs = multierr.Append(errs, errors.Wrapf(err, "failed to close a connection"))
							}
						}()

						nodeLog.Info("opening an xDS stream ...")
						stream, err := client.StartStream()
						if err != nil {
							return errors.Wrap(err, "failed to start an xDS stream")
						}
						defer func() {
							nodeLog.Info("closing an xDS stream ...")
							if err := stream.Close(); err != nil {
								errs = multierr.Append(errs, errors.Wrapf(err, "failed to close an xDS stream"))
							}
						}()

						nodeLog.Info("requesting Listeners")
						e := stream.Request(id, envoy_resource.ListenerType, dp)
						if e != nil {
							return errors.Wrapf(e, "failed to request %q", envoy_resource.ListenerType)
						}

						nodeLog.Info("requesting Clusters")
						e = stream.Request(id, envoy_resource.ClusterType, dp)
						if e != nil {
							return errors.Wrapf(e, "failed to request %q", envoy_resource.ClusterType)
						}

						nodeLog.Info("requesting Endpoints")
						e = stream.Request(id, envoy_resource.EndpointType, dp)
						if e != nil {
							return errors.Wrapf(e, "failed to request %q", envoy_resource.EndpointType)
						}

						for {
							nodeLog.Info("waiting for a discovery response ...")
							resp, err := stream.WaitForResources()
							if err != nil {
								return errors.Wrap(err, "failed to receive a discovery response")
							}
							nodeLog.Info("received xDS resources", "type", resp.TypeUrl, "version", resp.VersionInfo, "nonce", resp.Nonce, "resources", len(resp.Resources))

							if err := stream.ACK(resp.TypeUrl); err != nil {
								return errors.Wrap(err, "failed to ACK a discovery response")
							}
							nodeLog.Info("ACKed discovery response", "type", resp.TypeUrl, "version", resp.VersionInfo, "nonce", resp.Nonce)
						}
					}()
				}(i)
			}

			err := <-errCh

			return errors.Wrap(err, "one of xDS clients (Envoy simulators) terminated with an error")
		},
	}
	// flags
	cmd.PersistentFlags().StringVar(&args.xdsServerAddress, "xds-server-address", args.xdsServerAddress, "address of xDS server")
	cmd.PersistentFlags().Int32Var(&args.total, "total", args.total, "number of dataplanes to emulate")
	cmd.PersistentFlags().DurationVar(&args.rampUpPeriod, "rampup-period", args.rampUpPeriod, "ramp up period")
	return cmd
}

func main() {
	if err := newRootCmd().Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
