package cmd

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/spf13/cobra"

	"github.com/kumahq/kuma/test/server/types"
)

func newEchoHTTPCmd() *cobra.Command {
	args := struct {
		port     uint32
		instance string
		tls      bool
		crtFile  string
		keyFile  string
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
			if args.tls {
				return http.ListenAndServeTLS(fmt.Sprintf(":%d", args.port), args.crtFile, args.keyFile, nil)
			}
			return http.ListenAndServe(fmt.Sprintf(":%d", args.port), nil)
		},
	}
	cmd.PersistentFlags().Uint32Var(&args.port, "port", 10011, "port server is listening on")
	cmd.PersistentFlags().StringVar(&args.instance, "instance", "unknown", "will be included in response")
	cmd.PersistentFlags().BoolVar(&args.tls, "tls", false, "")
	cmd.PersistentFlags().StringVar(&args.crtFile, "crt", "./test/server/certs/server.crt", "")
	cmd.PersistentFlags().StringVar(&args.keyFile, "key", "./test/server/certs/server.key", "")
	return cmd
}
