package core

import (
	kube_types "k8s.io/apimachinery/pkg/types"
	kube_uuid "k8s.io/apimachinery/pkg/util/uuid"
	kube_ctrl "sigs.k8s.io/controller-runtime"
	kube_log "sigs.k8s.io/controller-runtime/pkg/log"
	kube_log_zap "sigs.k8s.io/controller-runtime/pkg/log/zap"
)

type NamespacedName = kube_types.NamespacedName

var (
	Log       = kube_ctrl.Log
	NewLogger = kube_log_zap.Logger
	SetLogger = kube_log.SetLogger

	SetupSignalHandler = kube_ctrl.SetupSignalHandler

	NewUUID = func() string {
		return string(kube_uuid.NewUUID())
	}
)
