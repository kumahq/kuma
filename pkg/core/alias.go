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

	SetupSignalHandler = func() (context.Context, context.Context, <-chan struct{}) {
		gracefulCtx, gracefulCancel := context.WithCancel(context.Background())
		ctx, cancel := context.WithCancel(context.Background())
		c := make(chan os.Signal, 3)
		signal.Notify(c, syscall.SIGINT, syscall.SIGTERM, syscall.SIGUSR2)
		usr2Notify := make(chan struct{}, 1)
		go func() {
			logger := Log.WithName("runtime")

			var stopSignalCancel func(os.Signal)
			thirdStopSignal := func(signal os.Signal) {
				logger.Info("received third signal, force exit", "signal", signal.String())
				os.Exit(1)
			}
			secondStopSignal := func(signal os.Signal) {
				logger.Info("received second signal, stopping instance", "signal", signal.String())
				cancel()
				stopSignalCancel = thirdStopSignal
			}
			firstStopSignal := func(signal os.Signal) {
				logger.Info("received signal, stopping instance gracefully", "signal", signal.String())
				gracefulCancel()
				close(usr2Notify)
				stopSignalCancel = secondStopSignal
			}
			stopSignalCancel = firstStopSignal
			for {
				s := <-c
				switch s {
				case syscall.SIGINT, syscall.SIGTERM:
					stopSignalCancel(s)
				case syscall.SIGUSR2:
					select {
					case usr2Notify <- struct{}{}:
					default:
					}
				}
			}
		}()
		return gracefulCtx, ctx, usr2Notify
	}
)

func NewUUID() string {
	return uuid.NewString()
}
