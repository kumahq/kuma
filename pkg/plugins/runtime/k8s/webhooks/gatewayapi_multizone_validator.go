package webhooks

import (
	"context"
	"errors"
	"net/http"

	admission "k8s.io/api/admission/v1"
	kube_webhook "sigs.k8s.io/controller-runtime/pkg/webhook"
	kube_admission "sigs.k8s.io/controller-runtime/pkg/webhook/admission"
	gatewayapi "sigs.k8s.io/gateway-api/apis/v1alpha2"

	config_core "github.com/kumahq/kuma/pkg/config/core"
	"github.com/kumahq/kuma/pkg/plugins/runtime/k8s/controllers/gatewayapi/common"
)

func NewGatewayAPIMultizoneValidator(cpMode config_core.CpMode) *kube_admission.Webhook {
	return &kube_admission.Webhook{
		Handler: &GatewayAPIMultizoneValidator{
			CpMode: cpMode,
		},
	}
}

var GatewayAPINotSupportedErr = errors.New("GatewayAPI of Kuma is only supported in Standalone deployments")

type GatewayAPIMultizoneValidator struct {
	CpMode  config_core.CpMode
	Decoder *kube_admission.Decoder
}

var _ kube_webhook.AdmissionHandler = &GatewayAPIMultizoneValidator{}

func (g *GatewayAPIMultizoneValidator) InjectDecoder(d *kube_admission.Decoder) error {
	g.Decoder = d
	return nil
}

func (g *GatewayAPIMultizoneValidator) Handle(_ context.Context, req kube_admission.Request) kube_admission.Response {
	if req.Operation == admission.Create {
		gatewayClass := &gatewayapi.GatewayClass{}
		if err := g.Decoder.Decode(req, gatewayClass); err != nil {
			return kube_admission.Errored(http.StatusBadRequest, err)
		}
		if g.CpMode != config_core.Standalone && gatewayClass.Spec.ControllerName == common.ControllerName {
			return kube_admission.Errored(http.StatusBadRequest, GatewayAPINotSupportedErr)
		}
	}
	return kube_admission.Allowed("")
}
