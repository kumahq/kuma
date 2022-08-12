package cmd

import (
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"os"
	"strconv"

	"github.com/spf13/cobra"

	"github.com/kumahq/kuma/test/server/types"
)

func newEchoHTTPCmd() *cobra.Command {
	args := struct {
		ip       string
		port     uint32
		instance string
		tls      bool
		crtFile  string
		keyFile  string
		probes   bool
	}{}
	cmd := &cobra.Command{
		Use:   "echo",
		Short: "Run Test Server with generic echo response",
		Long:  `Run Test Server with generic echo response.`,
		RunE: func(cmd *cobra.Command, _ []string) error {
			http.HandleFunc("/", func(writer http.ResponseWriter, request *http.Request) {
				headers := request.Header
				headers.Add("host", request.Host)
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
						if _, err := writer.Write([]byte(fmt.Sprintf("unknown probe type: %s", request.URL.Query().Get("type")))); err != nil {
							panic(err)
						}
					}
					if _, err := writer.Write([]byte("\n")); err != nil {
						panic(err)
					}
				})
			}
			if args.tls {
				return http.ListenAndServeTLS(net.JoinHostPort(args.ip, strconv.Itoa(int(args.port))), args.crtFile, args.keyFile, nil)
			}
			return http.ListenAndServe(net.JoinHostPort(args.ip, strconv.Itoa(int(args.port))), nil)
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
	cmd.PersistentFlags().StringVar(&args.crtFile, "crt", "./test/server/certs/server.crt", "path to the server's TLS cert")
	cmd.PersistentFlags().StringVar(&args.keyFile, "key", "./test/server/certs/server.key", "path to the server's TLS key")
	cmd.PersistentFlags().BoolVar(&args.probes, "probes", false, "generate readiness and liveness endpoints")
	return cmd
}
