package cmd

import (
	"time"

	"github.com/kumahq/kuma/pkg/plugins/runtime/k8s/metadata"
	"github.com/kumahq/kuma/pkg/plugins/runtime/k8s/util"
)

const (
	injectAnnotationKey    = metadata.KumaSidecarInjectionAnnotation
	sidecarStatusKey       = metadata.KumaSidecarInjectedAnnotation
	podRetrievalMaxRetries = 30
	podRetrievalInterval   = 1 * time.Second

	KUMAINIT = util.KumaInitContainerName
)
