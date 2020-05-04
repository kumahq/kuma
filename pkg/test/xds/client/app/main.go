package main

import (
	"fmt"
	"io/ioutil"
	"math/rand"
	"os"
	"time"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"go.uber.org/multierr"

	"github.com/ghodss/yaml"

	envoy_resource "github.com/envoyproxy/go-control-plane/pkg/resource/v2"

	"github.com/Kong/kuma/pkg/core"
	kuma_log "github.com/Kong/kuma/pkg/log"

	xds_client "github.com/Kong/kuma/pkg/test/xds/client"
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
		configFile       string
		rampUpPeriod     time.Duration
	}{
		xdsServerAddress: "grpc://localhost:5678",
		configFile:       "",
		rampUpPeriod:     30 * time.Second,
	}
	cmd := &cobra.Command{
		Use:   "run",
		Short: "Start xDS client(s) that simulate Envoy",
		Long:  `Start xDS client(s) that simulate Envoy.`,
		RunE: func(cmd *cobra.Command, _ []string) error {
			configBytes, err := ioutil.ReadFile(args.configFile)
			if err != nil {
				return errors.Wrapf(err, "failed to read config file from %q", args.configFile)
			}

			type Node struct {
				// ID of the Envoy node.
				ID string `json:"id,omitempty"`
			}
			type Config struct {
				// List of Envoy nodes to simulate.
				Nodes []Node `json:"nodes,omitempty"`
			}

			config := Config{}
			err = yaml.Unmarshal(configBytes, &config)
			if err != nil {
				return errors.Wrap(err, "failed to unmarshal config")
			}

			log.Info("going to start xDS clients (Envoy simulators)", "total", len(config.Nodes))

			errCh := make(chan error, 1)

			for i, node := range config.Nodes {
				nodeLog := log.WithName("envoy-simulator").WithValues("idx", i, "ID", node.ID)
				nodeLog.Info("creating an xDS client ...")

				go func(i int, node Node) {
					// add some jitter
					delay := time.Duration(int64(float64(args.rampUpPeriod.Nanoseconds()) * rand.Float64()))
					// wait
					<-time.After(delay)
					// proceed

					errCh <- func() (errs error) {
						client, err := xds_client.New(args.xdsServerAddress)
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
						e := stream.Request(node.ID, envoy_resource.ListenerType)
						if e != nil {
							return errors.Wrapf(e, "failed to request %q", envoy_resource.ListenerType)
						}

						nodeLog.Info("requesting Clusters")
						e = stream.Request(node.ID, envoy_resource.ClusterType)
						if e != nil {
							return errors.Wrapf(e, "failed to request %q", envoy_resource.ClusterType)
						}

						nodeLog.Info("requesting Endpoints")
						e = stream.Request(node.ID, envoy_resource.EndpointType)
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
				}(i, node)
			}

			err = <-errCh

			return errors.Wrap(err, "one of xDS clients (Envoy simulators) terminated with an error")
		},
	}
	// flags
	cmd.PersistentFlags().StringVar(&args.xdsServerAddress, "xds-server-address", args.xdsServerAddress, "address of xDS server")
	cmd.PersistentFlags().StringVar(&args.configFile, "config-file", args.configFile, "path to a config file")
	_ = cmd.MarkFlagRequired("config-file")
	cmd.PersistentFlags().DurationVar(&args.rampUpPeriod, "rampup-period", args.rampUpPeriod, "ramp up period")
	return cmd
}

func main() {
	if err := newRootCmd().Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
