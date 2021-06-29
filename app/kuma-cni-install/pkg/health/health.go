package health

import (
	"net/http"
	"sync/atomic"
)

func Start() *atomic.Value {
	httpServer := http.NewServeMux()
	status := &atomic.Value{}
	status.Store(false)

	httpServer.HandleFunc("/healthz", health)
	httpServer.HandleFunc("/readyz", ready(status))

	go func() {
		_ = http.ListenAndServe(":8000", httpServer)
	}()

	return status
}

func health(w http.ResponseWriter, _ *http.Request) {
	w.WriteHeader(http.StatusOK)
}

func ready(ready *atomic.Value) http.HandlerFunc {
	return func(w http.ResponseWriter, request *http.Request) {
		if ready == nil || !ready.Load().(bool) {
			http.Error(w, "Not ready yet", http.StatusServiceUnavailable)
			return
		}
		w.WriteHeader(http.StatusOK)
	}
}
