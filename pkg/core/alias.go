package core

import (
	"context"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/google/uuid"
	kube_ctrl "sigs.k8s.io/controller-runtime"

	kuma_log "github.com/kumahq/kuma/pkg/log"
)

var (
	// TODO remove dependency on kubernetes see: https://github.com/kumahq/kuma/issues/2798
	Log                   = kube_ctrl.Log
	NewLogger             = kuma_log.NewLogger
	NewLoggerTo           = kuma_log.NewLoggerTo
	NewLoggerWithRotation = kuma_log.NewLoggerWithRotation
	SetLogger             = kube_ctrl.SetLogger
	Now                   = time.Now

	SetupSignalHandler = func() (context.Context, context.Context, <-chan struct{}) {
		gracefulCtx, gracefulCancel := context.WithCancel(context.Background())
		ctx, cancel := context.WithCancel(context.Background())
		c := make(chan os.Signal, 3)
		signal.Notify(c, handledSignals...)
		usr2Notify := make(chan struct{}, 1)
		go func() {
			logger := Log.WithName("runtime")

			var numberOfStopSignals uint
			for {
				s := <-c
				switch s {
				case syscall.SIGINT, syscall.SIGTERM:
					switch numberOfStopSignals {
					case 0:
						logger.Info("received signal, stopping instance gracefully", "signal", s.String())
						gracefulCancel()
						close(usr2Notify)
						usr2Notify = nil
					case 1:
						logger.Info("received second signal, stopping instance", "signal", s.String())
						cancel()
					default:
						logger.Info("received third signal, force exit", "signal", s.String())
						os.Exit(1)
					}
					numberOfStopSignals++
				case usr2:
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
