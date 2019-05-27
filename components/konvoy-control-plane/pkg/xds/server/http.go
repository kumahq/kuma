package server

import (
	"context"
	"fmt"
	"log"
	"net/http"

	xds "github.com/envoyproxy/go-control-plane/pkg/server"
)

func RunHttpGateway(ctx context.Context, srv xds.Server, port int) {
	httpServer := &http.Server{Addr: fmt.Sprintf(":%d", port), Handler: &xds.HTTPGateway{Server: srv}}
	go func() {
		if err := httpServer.ListenAndServe(); err != nil {
			log.Println(err)
		}
	}()
	log.Printf("xDS HTTP server listening on %d\n", port)

	<-ctx.Done()
	httpServer.Shutdown(ctx)
}
