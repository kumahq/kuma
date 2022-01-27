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
	NewLoggerWithRotation = kuma_log.NewLoggerWithRotation
	SetLogger             = kube_log.SetLogger
	Now                   = time.Now
	TempDir               = os.TempDir

	SetupSignalHandler = func() context.Context {
		ctx, cancel := context.WithCancel(context.Background())
		c := make(chan os.Signal, 2)
		signal.Notify(c, syscall.SIGINT, syscall.SIGTERM)
		go func() {
			s := <-c
			Log.Info("Received signal, stopping instance", "signal", s.String())
			cancel()
			s = <-c
			Log.Info("Received second signal, force exit", "signal", s.String())
			os.Exit(1)
		}()
		return ctx
	}
)

func NewUUID() string {
	return uuid.NewString()
}
