package envoy

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/kumahq/kuma/pkg/core"
	"github.com/kumahq/kuma/pkg/core/runtime/component"
)

var _ component.GracefulComponent = &DPServerComponent{}

var dpServerLog = core.Log.WithName("kuma-dp").WithName("run").WithName("dpServer")

type DPServerComponent struct {
	mux         *http.ServeMux
	port        int
	notifyDrain chan<- struct{}
	done        chan struct{}
}

func NewDPServer(port int, notifyDrain chan<- struct{}) component.GracefulComponent {
	mux := http.NewServeMux()
	mux.HandleFunc("/drain", func(resp http.ResponseWriter, req *http.Request) {
		if req.Method != http.MethodPost {
			resp.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
		select {
		case notifyDrain <- struct{}{}:
		default:
		}
		resp.WriteHeader(http.StatusOK)
	})
	return &DPServerComponent{
		mux:         mux,
		port:        port,
		notifyDrain: notifyDrain,
		done:        make(chan struct{}),
	}
}

func (*DPServerComponent) NeedLeaderElection() bool {
	return false
}

func (dp *DPServerComponent) Start(stop <-chan struct{}) error {
	server := &http.Server{
		Handler:           dp.mux,
		Addr:              fmt.Sprintf("localhost:%d", dp.port),
		ReadHeaderTimeout: time.Second,
	}
	go func() {
		<-stop
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()
		if err := server.Shutdown(ctx); err != nil {
			dpServerLog.Error(err, "couldn't shutdown server")
		}
		close(dp.notifyDrain)
		close(dp.done)
	}()
	err := server.ListenAndServe()
	if errors.Is(err, http.ErrServerClosed) {
		return nil
	}
	return err
}

func (dp *DPServerComponent) WaitForDone() {
	<-dp.done
}
