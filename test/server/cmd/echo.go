package cmd

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	"html"
	"net"
	"net/http"
	"os"
	"strconv"
	"sync"
	"time"

	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/spf13/cobra"
	"go.opentelemetry.io/otel/exporters/prometheus"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"

	"github.com/kumahq/kuma/test/server/types"
)

const secondaryInboundPort = 9090

func newEchoHTTPCmd() *cobra.Command {
	counters := newCounters()

	args := struct {
		ip       string
		port     uint32
		instance string
		tls      bool
		tls13    bool
		crtFile  string
		keyFile  string
		probes   bool
	}{}
	cmd := &cobra.Command{
		Use:   "echo",
		Short: "Run Test Server with generic echo response",
		Long:  `Run Test Server with generic echo response.`,
		RunE: func(cmd *cobra.Command, _ []string) error {
			handleEcho := func(writer http.ResponseWriter, request *http.Request) {
				headers := request.Header
				handleDelay(headers)
				headers.Add("host", request.Host)

				if n, id, ok := parseSucceedAfterNHeaders(headers); ok {
					if counters.get(id) <= n {
						writer.WriteHeader(http.StatusServiceUnavailable)
						return
					}
				}

				resp := &types.EchoResponse{
					Instance: args.instance,
					Received: types.EchoResponseReceived{
						Method:  request.Method,
						Path:    request.URL.Path,
						Headers: headers,
					},
				}
				respBody, err := json.Marshal(resp)
				if err != nil {
					if _, err := writer.Write([]byte(`could not marshal json`)); err != nil {
						panic(err)
					}
					writer.WriteHeader(500)
				}
				writer.WriteHeader(http.StatusOK)
				if _, err := writer.Write(respBody); err != nil {
					panic(err)
				}
			}

			promExporter, err := prometheus.New(prometheus.WithoutCounterSuffixes())
			if err != nil {
				return err
			}
			sdkmetric.NewMeterProvider(sdkmetric.WithReader(promExporter))
			promHandler := promhttp.Handler()

			http.HandleFunc("/", handleEcho)
			http.HandleFunc("/metrics", func(writer http.ResponseWriter, request *http.Request) {
				promHandler.ServeHTTP(writer, request)
			})
			if args.probes {
				http.HandleFunc("/probes", func(writer http.ResponseWriter, request *http.Request) {
					switch request.URL.Query().Get("type") {
					case "readiness":
						writer.WriteHeader(http.StatusOK)
						if _, err := writer.Write([]byte("I'm ready!")); err != nil {
							panic(err)
						}
					case "liveness":
						writer.WriteHeader(http.StatusOK)
						if _, err := writer.Write([]byte("I'm alive!")); err != nil {
							panic(err)
						}
					case "":
						writer.WriteHeader(http.StatusBadRequest)
						if _, err := writer.Write([]byte("no probe's type provided")); err != nil {
							panic(err)
						}
					default:
						writer.WriteHeader(http.StatusBadRequest)
						if _, err := writer.Write([]byte(fmt.Sprintf("unknown probe type: %q", html.EscapeString(request.URL.Query().Get("type"))))); err != nil {
							panic(err)
						}
					}
					if _, err := writer.Write([]byte("\n")); err != nil {
						panic(err)
					}
				})
			}
			srv := http.Server{
				Addr:              net.JoinHostPort(args.ip, strconv.Itoa(int(args.port))),
				ReadHeaderTimeout: time.Second,
			}
			if args.tls && args.tls13 {
				srv.TLSConfig = &tls.Config{
					MinVersion: tls.VersionTLS13,
					MaxVersion: tls.VersionTLS13,
				}
			}
			if args.tls {
				return srv.ListenAndServeTLS(args.crtFile, args.keyFile)
			}
			secondInboundMux := http.NewServeMux()
			secondInboundMux.HandleFunc("/", handleEcho)
			go func() {
				_ = http.ListenAndServe(net.JoinHostPort(args.ip, strconv.Itoa(secondaryInboundPort)), secondInboundMux) // nolint: gosec
			}()
			return srv.ListenAndServe()
		},
	}
	cmd.PersistentFlags().Uint32Var(&args.port, "port", 10011, "port server is listening on")
	cmd.PersistentFlags().StringVar(&args.ip, "ip", "0.0.0.0", "ip server is listening on")
	r, err := os.Hostname()
	if r == "" || err != nil {
		r = "unknown"
	}
	cmd.PersistentFlags().StringVar(&args.instance, "instance", r, "will be included in response")
	cmd.PersistentFlags().BoolVar(&args.tls, "tls", false, "run the server with TLS enabled")
	cmd.PersistentFlags().BoolVar(&args.tls13, "tls13", false, "run the server with TLS 1.3, requires enabled tls")
	cmd.PersistentFlags().StringVar(&args.crtFile, "crt", "./test/server/certs/server.crt", "path to the server's TLS cert")
	cmd.PersistentFlags().StringVar(&args.keyFile, "key", "./test/server/certs/server.key", "path to the server's TLS key")
	cmd.PersistentFlags().BoolVar(&args.probes, "probes", false, "generate readiness and liveness endpoints")
	return cmd
}

type counters struct {
	sync.RWMutex
	counters map[string]int
}

func newCounters() *counters {
	return &counters{counters: map[string]int{}}
}

func (c *counters) get(hash string) int {
	c.Lock()
	defer c.Unlock()
	c.counters[hash]++
	return c.counters[hash]
}

func parseSucceedAfterNHeaders(headers http.Header) (int, string, bool) {
	id := headers.Get("x-succeed-after-n-id")
	if id == "" {
		return 0, "", false
	}

	nHeader := headers.Get("x-succeed-after-n")
	n, err := strconv.Atoi(nHeader)
	if err != nil || n < 2 {
		return 0, "", false
	}

	return n, id, true
}

func handleDelay(headers http.Header) {
	delayHeader := headers.Get("x-set-response-delay-ms")
	delay, err := strconv.Atoi(delayHeader)
	if err != nil {
		return
	}
	time.Sleep(time.Duration(delay) * time.Millisecond)
}
