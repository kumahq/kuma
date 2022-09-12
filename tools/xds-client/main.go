package main

import (
	"encoding/binary"
	"fmt"
	"math/rand"
	"net"
	"os"
	"time"

	envoy_resource "github.com/envoyproxy/go-control-plane/pkg/resource/v3"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"go.uber.org/multierr"

	"github.com/kumahq/kuma/api/mesh/v1alpha1"
	"github.com/kumahq/kuma/pkg/core"
	"github.com/kumahq/kuma/pkg/core/resources/model/rest/unversioned"
	rest_v1alpha1 "github.com/kumahq/kuma/pkg/core/resources/model/rest/v1alpha1"
	kuma_log "github.com/kumahq/kuma/pkg/log"
	"github.com/kumahq/kuma/tools/xds-client/stream"
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
		dps              int
		services         int
		inbounds         int
		outbounds        int
		rampUpPeriod     time.Duration
	}{
		xdsServerAddress: "grpcs://localhost:5678",
		dps:              100,
		services:         50,
		inbounds:         1,
		outbounds:        3,
		rampUpPeriod:     30 * time.Second,
	}
	cmd := &cobra.Command{
		Use:   "run",
		Short: "Start xDS client(s) that simulate Envoy",
		Long:  `Start xDS client(s) that simulate Envoy.`,
		RunE: func(cmd *cobra.Command, _ []string) error {
			ipRand := rand.Uint32()
			log.Info("going to start xDS clients (Envoy simulators)", "dps", args.dps)
			errCh := make(chan error, 1)
			for i := 0; i < args.dps; i++ {
				id := fmt.Sprintf("default.dataplane-%d", i)
				nodeLog := log.WithName("envoy-simulator").WithValues("idx", i, "ID", id)
				nodeLog.Info("creating an xDS client ...")

				go func(i int) {
					buf := make([]byte, 4)
					binary.LittleEndian.PutUint32(buf, ipRand+uint32(i))
					ip := net.IP(buf).String()

					dpSpec := &v1alpha1.Dataplane{
						Networking: &v1alpha1.Dataplane_Networking{
							Address: ip,
						},
					}
					for j := 0; j < args.inbounds; j++ {
						service := fmt.Sprintf("service-%d", rand.Int()%args.services)
						dpSpec.Networking.Inbound = append(dpSpec.Networking.Inbound, &v1alpha1.Dataplane_Networking_Inbound{
							Port: uint32(8080 + j),
							Tags: map[string]string{
								"kuma.io/service":  service,
								"kuma.io/protocol": "http",
							},
						})
					}
					for j := 0; j < args.outbounds; j++ {
						service := fmt.Sprintf("service-%d", rand.Int()%args.services)
						dpSpec.Networking.Outbound = append(dpSpec.Networking.Outbound, &v1alpha1.Dataplane_Networking_Outbound{
							Port: uint32(10080 + j), Tags: map[string]string{"kuma.io/service": service},
						})
					}

					dp := &unversioned.Resource{
						Meta: rest_v1alpha1.ResourceMeta{Mesh: "default", Name: fmt.Sprintf("dataplane-%d", i), Type: "Dataplane"},
						Spec: dpSpec,
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
	cmd.PersistentFlags().IntVar(&args.dps, "dps", args.dps, "number of dataplanes to emulate")
	cmd.PersistentFlags().IntVar(&args.services, "services", args.services, "number of services")
	cmd.PersistentFlags().IntVar(&args.inbounds, "inbounds", args.inbounds, "number of inbounds")
	cmd.PersistentFlags().IntVar(&args.outbounds, "outbounds", args.outbounds, "number of outbounds")
	cmd.PersistentFlags().DurationVar(&args.rampUpPeriod, "rampup-period", args.rampUpPeriod, "ramp up period")
	return cmd
}

func main() {
	if err := newRootCmd().Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
