package server

import (
	"context"
	"fmt"
	"log"
	"net/http"
)

func RunDiagnosticsServer(ctx context.Context, port int) {
	mux := http.NewServeMux()
	mux.HandleFunc("/ready", func(resp http.ResponseWriter, _ *http.Request) {
		resp.WriteHeader(http.StatusOK)
	})
	mux.HandleFunc("/healthy", func(resp http.ResponseWriter, _ *http.Request) {
		resp.WriteHeader(http.StatusOK)
	})

	httpServer := &http.Server{Addr: fmt.Sprintf(":%d", port), Handler: mux}
	go func() {
		if err := httpServer.ListenAndServe(); err != nil {
			log.Println(err)
		}
	}()
	log.Printf("diagnostics server listening on %d\n", port)

	<-ctx.Done()
	httpServer.Shutdown(ctx)
}
