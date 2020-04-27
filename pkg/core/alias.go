package core

import (
	"time"

	kube_uuid "k8s.io/apimachinery/pkg/util/uuid"
	kube_log "sigs.k8s.io/controller-runtime/pkg/log"
	kube_signals "sigs.k8s.io/controller-runtime/pkg/manager/signals"

	kuma_log "github.com/Kong/kuma/pkg/log"
)

var (
	Log       = kube_log.Log
	NewLogger = kuma_log.NewLogger
	SetLogger = kube_log.SetLogger
	Now       = time.Now

	SetupSignalHandler = kube_signals.SetupSignalHandler

	NewUUID = func() string {
		return string(kube_uuid.NewUUID())
	}
)
