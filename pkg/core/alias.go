package core

import (
	"context"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/google/uuid"
	kube_log "sigs.k8s.io/controller-runtime/pkg/log"

	kuma_log "github.com/kumahq/kuma/pkg/log"
)

var (
	// TODO remove dependency on kubernetes see: https://github.com/kumahq/kuma/issues/2798
	Log                   = kube_log.Log
	NewLogger             = kuma_log.NewLogger
	NewLoggerTo           = kuma_log.NewLoggerTo
	NewLoggerWithRotation = kuma_log.NewLoggerWithRotation
	SetLogger             = kube_log.SetLogger
	Now                   = time.Now

	SetupSignalHandler = func() (context.Context, context.Context) {
		gracefulCtx, gracefulCancel := context.WithCancel(context.Background())
		ctx, cancel := context.WithCancel(context.Background())
		c := make(chan os.Signal, 3)
		signal.Notify(c, syscall.SIGINT, syscall.SIGTERM)
		go func() {
			logger := Log.WithName("runtime")
			s := <-c
			logger.Info("received signal, stopping instance gracefully", "signal", s.String())
			gracefulCancel()
			s = <-c
			logger.Info("received second signal, stopping instance", "signal", s.String())
			cancel()
			s = <-c
			logger.Info("received third signal, force exit", "signal", s.String())
			os.Exit(1)
		}()
		return gracefulCtx, ctx
	}
)

func NewUUID() string {
	return uuid.NewString()
}
