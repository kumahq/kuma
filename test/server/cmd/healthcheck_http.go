package cmd

import (
	"fmt"
	"net/http"
	"time"

	"github.com/spf13/cobra"
)

func newHealthCheckHTTP() *cobra.Command {
	args := struct {
		port          uint32
		healthMethod  string
		contentMethod string
		content       string
	}{}
	cmd := &cobra.Command{
		Use:   "http",
		Short: "Run Test Server for HTTP Health Check test",
		Long:  `Run Test Server for HTTP Health Check test.`,
		RunE: func(cmd *cobra.Command, _ []string) error {
			http.HandleFunc(fmt.Sprintf("/%s", args.healthMethod), func(writer http.ResponseWriter, request *http.Request) {
				writer.WriteHeader(http.StatusOK)
			})
			http.HandleFunc(fmt.Sprintf("/%s", args.contentMethod), func(writer http.ResponseWriter, request *http.Request) {
				if _, err := writer.Write([]byte(args.content)); err != nil {
					http.Error(writer, err.Error(), http.StatusInternalServerError)
				}
			})
			return (&http.Server{Addr: fmt.Sprintf(":%d", args.port), ReadHeaderTimeout: time.Second}).ListenAndServe()
		},
	}
	cmd.PersistentFlags().Uint32Var(&args.port, "port", 10011, "port server is listening on")
	cmd.PersistentFlags().StringVar(&args.healthMethod, "health-method", "health", "method for health checking, returns 200 OK")
	cmd.PersistentFlags().StringVar(&args.contentMethod, "content-method", "content", "method for returning value defined in '--content'")
	cmd.PersistentFlags().StringVar(&args.content, "content", "response", "value for method defined in '--content-method'")
	return cmd
}
